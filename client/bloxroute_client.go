package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/gorilla/websocket"
)

// A typesafe websocket client for bloXroute in Go.
type BloXrouteClient struct {
	conn   *websocket.Conn
	stopCh <-chan struct{}
	// To ensure there is only one concurrent conn.WriteMessage(), see
	// https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency
	mu *sync.Mutex
	// Subscription ID to stream name
	idToStreamNameMap map[string]string
	// Subscription ID to command
	idToCommandMap         map[string]string
	subscriptionResponseCh chan types.SubscriptionResponse
	// key is the subscription ID, value is user provided  output channel
	newTxsChannels     map[string]chan<- *types.Transaction        // output channels for `newTxs`
	pendingTxsChannels map[string]chan<- *types.Transaction        // output channels for `pendingTxs`
	newBlocksChannels  map[string]chan<- *types.Block              // output channels for `newBlocks`
	bdnBlocksChannels  map[string]chan<- *types.Block              // output channels for `bdnBlocks`
	txReceiptsChannels map[string]chan<- *types.TxReceipt          // output channels for `txReceipts`
	ethOnBlockChannels map[string]chan<- *types.EthOnBlockResponse // output channels for `ethOnBlock`
	rawChannels        map[string]chan<- string                    // output channels for SubscribeRaw()
}

// Create a bloXroute websocket client connected to bloXroute cloud.
//
// `network`, available values: Mainnet, BSC-Mainnet
func NewBloXrouteClientToCloud(network string, certFile string, keyFile string, stopCh <-chan struct{}) (*BloXrouteClient, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = tlsConfig
	// see https://docs.bloxroute.com/introduction/cloud-api-ips
	url := ""
	if network == "Mainnet" {
		url = "wss://virginia.eth.blxrbdn.com/ws"
	} else if network == "BSC-Mainnet" {
		url = "wss://virginia.bsc.blxrbdn.com/ws"
	} else {
		return nil, fmt.Errorf("invalid network: %s", network)
	}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	client := &BloXrouteClient{
		conn:                   conn,
		stopCh:                 stopCh,
		mu:                     &sync.Mutex{},
		idToStreamNameMap:      make(map[string]string),
		idToCommandMap:         make(map[string]string),
		subscriptionResponseCh: make(chan types.SubscriptionResponse),
		newTxsChannels:         make(map[string]chan<- *types.Transaction),
		pendingTxsChannels:     make(map[string]chan<- *types.Transaction),
		newBlocksChannels:      make(map[string]chan<- *types.Block),
		bdnBlocksChannels:      make(map[string]chan<- *types.Block),
		txReceiptsChannels:     make(map[string]chan<- *types.TxReceipt),
		ethOnBlockChannels:     make(map[string]chan<- *types.EthOnBlockResponse),
		rawChannels:            make(map[string]chan<- string),
	}

	client.ping()
	go client.run()

	return client, nil
}

// Create a bloXroute websocket client connected to a local gateway.
func NewBloXrouteClientToGateway(url string, authorizationHeader string, stopCh <-chan struct{}) (*BloXrouteClient, error) {
	dialer := websocket.DefaultDialer
	if url == "" {
		url = "ws://127.0.0.1:28333"
	}
	conn, _, err := dialer.Dial(url, http.Header{"Authorization": []string{authorizationHeader}})
	if err != nil {
		return nil, err
	}

	client := &BloXrouteClient{
		conn:                   conn,
		stopCh:                 stopCh,
		mu:                     &sync.Mutex{},
		idToStreamNameMap:      make(map[string]string),
		idToCommandMap:         make(map[string]string),
		subscriptionResponseCh: make(chan types.SubscriptionResponse),
		newTxsChannels:         make(map[string]chan<- *types.Transaction),
		pendingTxsChannels:     make(map[string]chan<- *types.Transaction),
		newBlocksChannels:      make(map[string]chan<- *types.Block),
		bdnBlocksChannels:      make(map[string]chan<- *types.Block),
		txReceiptsChannels:     make(map[string]chan<- *types.TxReceipt),
		ethOnBlockChannels:     make(map[string]chan<- *types.EthOnBlockResponse),
		rawChannels:            make(map[string]chan<- string),
	}

	client.ping()
	go client.run()

	return client, nil
}

// Subscribe to the `newTxs` stream.
// See https://docs.bloxroute.com/streams/newtxs-and-pendingtxs
func (c *BloXrouteClient) SubscribeNewTxs(include []string, filters string, outCh chan<- *types.Transaction) error {
	return c.subscribeTransactions("newTxs", include, filters, outCh)
}

// Subscribe to the `pendingTxs` stream.
// See https://docs.bloxroute.com/streams/newtxs-and-pendingtxs
func (c *BloXrouteClient) SubscribePendingTxs(include []string, filters string, outCh chan<- *types.Transaction) error {
	return c.subscribeTransactions("pendingTxs", include, filters, outCh)
}

// Subscribe to the `newBlocks` stream.
// See https://docs.bloxroute.com/streams/newblock-stream
func (c *BloXrouteClient) SubscribeNewBlocks(include []string, outCh chan<- *types.Block) error {
	return c.subscribeBlocks("newBlocks", include, outCh)
}

// Subscribe to the `txReceipts` stream.
// See https://docs.bloxroute.com/streams/txreceipts
func (c *BloXrouteClient) SubscribeTxReceipts(include []string, outCh chan<- *types.TxReceipt) error {
	if len(include) == 0 {
		// empty means all
		include = []string{
			"block_hash",
			"block_number",
			"contract_address",
			"cumulative_gas_used",
			"from", "gas_used", "logs",
			"logs_bloom",
			"status",
			"to",
			"transaction_hash",
			"transaction_index",
		}
	}

	params := make([]interface{}, 0)
	params = append(params, "txReceipts")
	if len(include) > 0 {
		m := make(map[string][]string)
		m["include"] = include
		params = append(params, m)
	}

	subscriptionID, err := c.subscribe(params)
	if err != nil {
		return err
	}
	c.txReceiptsChannels[subscriptionID] = outCh

	return nil
}

// Subscribe to the `bdnBlocks` stream.
// See https://docs.bloxroute.com/streams/bdnblocks
func (c *BloXrouteClient) SubscribeBdnBlocks(include []string, outCh chan<- *types.Block) error {
	return c.subscribeBlocks("bdnBlocks", include, outCh)
}

// Subscribe to the `ethOnBlock` stream.
// See https://docs.bloxroute.com/streams/onblock-event-stream
func (c *BloXrouteClient) SubscribeEthOnBlock(include []string, callParams []map[string]string, outCh chan<- *types.EthOnBlockResponse) error {
	if len(include) == 0 {
		// empty means all
		include = []string{"name", "response", "block_height", "tag"}
	}

	params := make([]interface{}, 0)
	params = append(params, "ethOnBlock")
	{
		m := make(map[string]interface{})
		if len(include) > 0 {
			m["include"] = include
		}
		if len(callParams) > 0 {
			m["call-params"] = callParams
		}
		params = append(params, m)
	}

	subscriptionID, err := c.subscribe(params)
	if err != nil {
		return err
	}

	c.ethOnBlockChannels[subscriptionID] = outCh

	return nil
}

// SubscribeRaw() send a raw subscription request to the server.
// This is a low-level API and should be used only if you know what you are doing.
// The first element of params must be the name of the stream.
func (c *BloXrouteClient) SubscribeRaw(subRequest string, outCh chan<- string) (string, error) {
	streamName := ""
	{
		jsonObj := make(map[string]interface{})
		err := json.Unmarshal([]byte(subRequest), &jsonObj)
		if err != nil {
			return "", err
		}
		params, ok := jsonObj["params"].([]interface{})
		if !ok {
			return "", fmt.Errorf("params must be an array")
		}
		streamName, ok = params[0].(string)
		if !ok {
			return "", fmt.Errorf("The first element of params must be the name of the stream.")
		}
	}

	subscriptionID, err := c.sendCommand(subRequest)
	if err != nil {
		return "", err
	}

	c.idToStreamNameMap[subscriptionID] = streamName
	c.rawChannels[subscriptionID] = outCh
	return subscriptionID, nil
}

func (c *BloXrouteClient) Unsubscribe(subscriptionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	streamName := c.idToStreamNameMap[subscriptionID]
	if streamName == "" {
		return fmt.Errorf("subscription ID %s does NOT exist", subscriptionID)
	}

	command := fmt.Sprintf(`{"method": "unsubscribe", "params": ["%s"]}`, subscriptionID)
	log.Println(command)
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(command))
	if err != nil {
		return err
	}

	delete(c.idToStreamNameMap, subscriptionID)
	delete(c.idToCommandMap, subscriptionID)

	if c.rawChannels[subscriptionID] != nil {
		delete(c.rawChannels, subscriptionID)
		return nil
	}

	switch streamName {
	case "newTxs":
		delete(c.newTxsChannels, subscriptionID)
	case "pendingTxs":
		delete(c.pendingTxsChannels, subscriptionID)
	case "newBlocks":
		delete(c.newBlocksChannels, subscriptionID)
	case "bdnBlocks":
		delete(c.bdnBlocksChannels, subscriptionID)
	case "txReceipts":
		delete(c.txReceiptsChannels, subscriptionID)
	case "ethOnBlock":
		delete(c.ethOnBlockChannels, subscriptionID)
	default:
		log.Panicf("Unknown stream name: %s", streamName)
	}

	return nil
}

// Subscribe to `newTxs` or `pendingTxs` stream.
func (c *BloXrouteClient) subscribeTransactions(streamName string, include []string, filters string, outCh chan<- *types.Transaction) error {
	if len(include) == 0 {
		// empty means all
		include = []string{"tx_hash", "tx_contents"}
	}

	params := make([]interface{}, 0)
	{
		params = append(params, streamName)
		if len(include) > 0 {
			m := make(map[string][]string)
			m["include"] = include
			params = append(params, m)
		}
		if len(filters) > 0 {
			m := make(map[string]string)
			m["filters"] = filters
			params = append(params, m)
		}
	}

	subscriptionID, err := c.subscribe(params)
	if err != nil {
		return err
	}
	if streamName == "newTxs" {
		c.newTxsChannels[subscriptionID] = outCh
	} else if streamName == "pendingTxs" {
		c.pendingTxsChannels[subscriptionID] = outCh
	} else {
		log.Panicf("invalid stream name: %s", streamName)
	}

	return nil
}

// Subscribe to `newBlocks` or `bdnBlocks` stream.
func (c *BloXrouteClient) subscribeBlocks(streamName string, include []string, outCh chan<- *types.Block) error {
	if len(include) == 0 {
		// empty means all
		include = []string{"hash", "header", "transactions", "uncles"}
	}

	params := make([]interface{}, 0)
	params = append(params, streamName)
	if len(include) > 0 {
		m := make(map[string][]string)
		m["include"] = include
		params = append(params, m)
	}

	subscriptionID, err := c.subscribe(params)
	if err != nil {
		return err
	}
	if streamName == "newBlocks" {
		c.newBlocksChannels[subscriptionID] = outCh
	} else if streamName == "bdnBlocks" {
		c.bdnBlocksChannels[subscriptionID] = outCh
	} else {
		log.Panicf("invalid stream name: %s", streamName)
	}

	return nil
}

// subscribe() sends a subscription request and returns the subscription ID.
func (c *BloXrouteClient) subscribe(params []interface{}) (string, error) {
	// The first element is the stream name
	streamName := params[0].(string)
	bytes, _ := json.Marshal(params)
	subRequest := fmt.Sprintf(`{"method": "subscribe", "params": %s}`, string(bytes))
	subscriptionID, err := c.sendCommand(subRequest)
	if err != nil {
		return "", err
	}

	c.idToStreamNameMap[subscriptionID] = streamName
	return subscriptionID, nil
}

func (c *BloXrouteClient) sendCommand(subRequest string) (string, error) {
	log.Println(subRequest)
	// This function is always called sequentially.
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
	if err != nil {
		return "", err
	}
	subscriptionID, err := c.waitForSubscriptionID()
	if err != nil {
		return "", err
	}

	c.idToCommandMap[subscriptionID] = subRequest
	return subscriptionID, nil
}

func (c *BloXrouteClient) waitForSubscriptionID() (string, error) {
	select {
	case <-time.After(3 * time.Second):
		return "", errors.New("timeout")
	case resp := <-c.subscriptionResponseCh:
		return resp.Result, nil
	}
}

// Sends a ping message every 5 seconds to keep the WebSocket connection alive.
// https://docs.bloxroute.com/apis/ping
func (c *BloXrouteClient) ping() {
	ticker := time.NewTicker(5 * time.Second) // ping every 5 seconds
	go func() {
		for {
			select {
			case <-c.stopCh:
				return
			case <-ticker.C:
				c.mu.Lock()
				defer c.mu.Unlock()
				err := c.conn.WriteMessage(websocket.TextMessage, []byte(`{"method": "ping"}`))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
}

// Close is used to terminate our websocket client
func (c *BloXrouteClient) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	c.conn.Close()
	return err
}

func (c *BloXrouteClient) handleTaskDisabledEvent(event *types.WebsocketMsg[types.EthOnBlockResponse]) error {
	subscriptionID := event.Params.Subscription
	subRequest := c.idToCommandMap[subscriptionID]
	if subRequest == "" {
		log.Panicf("Bug: subscription ID %s not found in idToCommandMap", subscriptionID)
	}

	// resend
	newSubscriptionID, err := c.sendCommand(subRequest)
	if err != nil {
		return err
	}

	c.idToStreamNameMap[newSubscriptionID] = c.idToStreamNameMap[subscriptionID]
	delete(c.idToStreamNameMap, subscriptionID)

	c.ethOnBlockChannels[newSubscriptionID] = c.ethOnBlockChannels[subscriptionID]
	delete(c.ethOnBlockChannels, subscriptionID)

	return nil
}

func processMsg[T any](data []byte, outCh chan<- T) error {
	msg := types.WebsocketMsg[T]{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	if len(msg.Error) > 0 {
		return errors.New(string(data))
	}

	outCh <- msg.Params.Result
	return nil
}

// Run the event looop.
func (c *BloXrouteClient) run() error {
	for {
		select {
		case <-c.stopCh:
			c.close()
			return nil
		default:
			_, nextNotification, err := c.conn.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}

			{
				// Is it a subscription response?
				subscriptionResp := types.SubscriptionResponse{}
				err = json.Unmarshal(nextNotification, &subscriptionResp)
				if err == nil {
					if subscriptionResp.Result != "" {
						c.subscriptionResponseCh <- subscriptionResp
						break
					}
				}
			}

			subscriptionID := ""
			streamName := ""
			{
				m := make(map[string]interface{})
				err := json.Unmarshal(nextNotification, &m)
				if err != nil {
					log.Fatal(string(nextNotification))
				}
				if _, ok := m["error"]; ok {
					log.Fatal(string(nextNotification))
				}
				if _, ok := m["params"]; !ok {
					log.Println(string(nextNotification))
					break
				}
				params := m["params"].(map[string]interface{})
				if _, ok := params["result"]; !ok {
					log.Println(string(nextNotification))
					break
				}
				if _, ok := params["subscription"]; !ok {
					log.Println(string(nextNotification))
					break
				}
				subscriptionID = params["subscription"].(string)
				streamName = c.idToStreamNameMap[subscriptionID]

				result := params["result"].(map[string]interface{})
				if name, ok := result["name"].(string); ok {
					if name == "TaskCompletedEvent" {
						break
					} else if name == "TaskDisabledEvent" {
						disableEvent := types.WebsocketMsg[types.EthOnBlockResponse]{}
						err := json.Unmarshal(nextNotification, &disableEvent)
						if err != nil {
							log.Panicf("Bug: failed to unmashal TaskDisabledEvent %s", string(nextNotification))
						}

						c.handleTaskDisabledEvent(&disableEvent)
						break
					}
				}
			}

			if subscriptionID == "" {
				log.Panicf("Bug: no subscription ID in %s", string(nextNotification))
			}
			if streamName == "" {
				log.Panicf("Bug: no stream name for subscription ID %s", subscriptionID)
			}

			if c.rawChannels[subscriptionID] != nil {
				c.rawChannels[subscriptionID] <- string(nextNotification)
				break
			}

			err = nil
			switch streamName {
			case "newTxs":
				outCh := c.newTxsChannels[subscriptionID]
				if outCh == nil {
					log.Printf("Bug: no output channel for subscription ID %s", subscriptionID)
				}
				err = processMsg(nextNotification, outCh)
			case "pendingTxs":
				outCh := c.pendingTxsChannels[subscriptionID]
				if outCh == nil {
					log.Printf("Bug: no output channel for subscription ID %s", subscriptionID)
				}
				err = processMsg(nextNotification, outCh)
			case "newBlocks":
				outCh := c.newBlocksChannels[subscriptionID]
				if outCh == nil {
					log.Printf("Bug: no output channel for subscription ID %s", subscriptionID)
				}
				err = processMsg(nextNotification, outCh)
			case "bdnBlocks":
				outCh := c.bdnBlocksChannels[subscriptionID]
				if outCh == nil {
					log.Printf("Bug: no output channel for subscription ID %s", subscriptionID)
				}
				err = processMsg(nextNotification, outCh)
			case "txReceipts":
				outCh := c.txReceiptsChannels[subscriptionID]
				if outCh == nil {
					log.Printf("Bug: no output channel for subscription ID %s", subscriptionID)
				}
				err = processMsg(nextNotification, outCh)
			case "ethOnBlock":
				outCh := c.ethOnBlockChannels[subscriptionID]
				if outCh == nil {
					log.Printf("Bug: no output channel for subscription ID %s", subscriptionID)
				}
				err = processMsg(nextNotification, outCh)
			default:
				log.Panicf("Unknown stream name: %s", streamName)
			}

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
