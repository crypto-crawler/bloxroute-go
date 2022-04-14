package client

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
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
	sendTxChannel          chan common.Hash
	pongCh                 chan pongMsg
	newTxsCh               chan<- *types.Transaction        // output channel for `newTxs`
	pendingTxsCh           chan<- *types.Transaction        // output channel for `pendingTxs`
	newBlocksCh            chan<- *types.Block              // output channel for `newBlocks`
	bdnBlocksCh            chan<- *types.Block              // output channel for `bdnBlocks`
	txReceiptsCh           chan<- *types.TxReceipt          // output channel for `txReceipts`
	ethOnBlockCh           chan<- *types.EthOnBlockResponse // output channel for `ethOnBlock`
	rawChannel             chan<- string                    // output channel for SubscribeRaw()
}

// stream name
const (
	newTxs     = "newTxs"
	pendingTxs = "pendingTxs"
	newBlocks  = "newBlocks"
	bdnBlocks  = "bdnBlocks"
	txReceipts = "txReceipts"
	ethOnBlock = "ethOnBlock"
	rawString  = "rawString" // for SubscribeRaw()
)

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
		sendTxChannel:          make(chan common.Hash),
		pongCh:                 make(chan pongMsg),
	}

	go client.run()

	// test the connection
	if _, err = client.Ping(); err != nil {
		return nil, err
	}

	//  Send a ping every 5 seconds to keep the WebSocket connection alive
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
				if _, err := client.Ping(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

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
		sendTxChannel:          make(chan common.Hash),
		pongCh:                 make(chan pongMsg),
	}

	go client.run()

	// test the connection
	if _, err = client.Ping(); err != nil {
		return nil, err
	}

	return client, nil
}

// Subscribe to the `newTxs` stream.
// See https://docs.bloxroute.com/streams/newtxs-and-pendingtxs
func (c *BloXrouteClient) SubscribeNewTxs(include []string, filters string, outCh chan<- *types.Transaction) (string, error) {
	return c.subscribeTransactions(newTxs, include, filters, outCh)
}

// Subscribe to the `pendingTxs` stream.
// See https://docs.bloxroute.com/streams/newtxs-and-pendingtxs
func (c *BloXrouteClient) SubscribePendingTxs(include []string, filters string, outCh chan<- *types.Transaction) (string, error) {
	return c.subscribeTransactions(pendingTxs, include, filters, outCh)
}

// Subscribe to the `newBlocks` stream.
// See https://docs.bloxroute.com/streams/newblock-stream
func (c *BloXrouteClient) SubscribeNewBlocks(include []string, outCh chan<- *types.Block) (string, error) {
	return c.subscribeBlocks(newBlocks, include, outCh)
}

// Subscribe to the `txReceipts` stream.
// See https://docs.bloxroute.com/streams/txreceipts
func (c *BloXrouteClient) SubscribeTxReceipts(include []string, outCh chan<- *types.TxReceipt) (string, error) {
	c.txReceiptsCh = outCh
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
	params = append(params, txReceipts)
	if len(include) > 0 {
		m := make(map[string][]string)
		m["include"] = include
		params = append(params, m)
	}

	return c.subscribe(params)
}

// Subscribe to the `bdnBlocks` stream.
// See https://docs.bloxroute.com/streams/bdnblocks
func (c *BloXrouteClient) SubscribeBdnBlocks(include []string, outCh chan<- *types.Block) (string, error) {
	return c.subscribeBlocks(bdnBlocks, include, outCh)
}

// Subscribe to the `ethOnBlock` stream.
// See https://docs.bloxroute.com/streams/onblock-event-stream
func (c *BloXrouteClient) SubscribeEthOnBlock(include []string, callParams []map[string]string, outCh chan<- *types.EthOnBlockResponse) (string, error) {
	c.ethOnBlockCh = outCh

	if len(include) == 0 {
		// empty means all
		include = []string{"name", "response", "block_height", "tag"}
	}

	params := make([]interface{}, 0)
	params = append(params, ethOnBlock)
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

	return c.subscribe(params)
}

// SubscribeRaw() send a raw subscription request to the server.
// This is a low-level API and should be used only if you know what you are doing.
// The first element of params must be the name of the stream.
func (c *BloXrouteClient) SubscribeRaw(subRequest string, outCh chan<- string) (string, error) {
	c.rawChannel = outCh
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
		_, ok = params[0].(string)
		if !ok {
			return "", fmt.Errorf("The first element of params must be the name of the stream.")
		}
	}

	subscriptionID, err := c.sendCommand(subRequest)
	if err != nil {
		return "", err
	}

	c.idToStreamNameMap[subscriptionID] = rawString
	return subscriptionID, nil
}

// Send a transaction via BDN.
func (client *BloXrouteClient) SendTransaction(transaction []byte, nonceMonitoring bool, blockchainNetwork string) (common.Hash, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	subRequest := fmt.Sprintf(`{"jsonrpc":"2.0","method":"blxr_tx","params":{"transaction":"%s","nonce_monitoring":"%v","blockchain_network":"%v"}}`, hex.EncodeToString(transaction), nonceMonitoring, blockchainNetwork)
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
	if err != nil {
		return common.Hash{}, err
	}

	// wait for response
	select {
	case <-time.After(3 * time.Second):
		return common.Hash{}, errors.New("SendTransaction timeout")
	case txHash := <-client.sendTxChannel:
		return txHash, nil
	}
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

	return nil
}

// Subscribe to `newTxs` or `pendingTxs` stream.
func (c *BloXrouteClient) subscribeTransactions(streamName string, include []string, filters string, outCh chan<- *types.Transaction) (string, error) {
	switch streamName {
	case newTxs:
		c.newTxsCh = outCh
	case pendingTxs:
		c.pendingTxsCh = outCh
	default:
		return "", fmt.Errorf("unknown stream name: %s", streamName)
	}

	if len(include) == 0 {
		// empty means all
		include = []string{"tx_hash", "tx_contents"}
	}

	params := make([]interface{}, 0)
	{
		params = append(params, streamName)
		m := make(map[string]interface{})
		if len(include) > 0 {
			m["include"] = include
		}
		if len(filters) > 0 {
			m["filters"] = filters
		}
		params = append(params, m)
	}

	return c.subscribe(params)
}

// Subscribe to `newBlocks` or `bdnBlocks` stream.
func (c *BloXrouteClient) subscribeBlocks(streamName string, include []string, outCh chan<- *types.Block) (string, error) {
	switch streamName {
	case newBlocks:
		c.newBlocksCh = outCh
	case bdnBlocks:
		c.bdnBlocksCh = outCh
	default:
		return "", fmt.Errorf("unknown stream name: %s", streamName)
	}

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

	return c.subscribe(params)
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
	// This function is always called sequentially.
	c.mu.Lock()
	defer c.mu.Unlock()
	// log.Println(subRequest)

	err := c.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
	if err != nil {
		return "", err
	}

	// wait for subscription confirmation
	select {
	case <-time.After(6 * time.Second):
		return "", fmt.Errorf("timeout %s", subRequest)
	case resp := <-c.subscriptionResponseCh:
		subscriptionID := resp.Result
		c.idToCommandMap[subscriptionID] = subRequest
		return resp.Result, nil
	}
}

// Ping() sends a ping message to websocket server.
// See https://docs.bloxroute.com/apis/ping
func (c *BloXrouteClient) Ping() (time.Time, error) {
	c.mu.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(`{"method": "ping"}`))
	c.mu.Unlock()
	if err != nil {
		return time.Now(), err
	}

	// wait for pong
	select {
	case <-time.After(3 * time.Second):
		return time.Now(), errors.New("ping timeout")
	case resp := <-c.pongCh:
		return time.Parse("2006-01-02 15:04:05.99", resp.Result.Pong)
	}
}

// Close is used to terminate our websocket client
func (c *BloXrouteClient) close() error {
	err := c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(3*time.Second))
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
				// Is it pongMsg?
				pongMsg := pongMsg{}
				json.Unmarshal(nextNotification, &pongMsg)
				if err == nil {
					if pongMsg.Result.Pong != "" {
						c.pongCh <- pongMsg
						break
					}
				}
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

			{
				// Is it a sendTxResponse?
				sendTxResp := sendTxResponse{}
				err = json.Unmarshal(nextNotification, &sendTxResp)
				if err == nil {
					if sendTxResp.Result != nil && sendTxResp.Result.TxHash != "" {
						c.sendTxChannel <- common.HexToHash(sendTxResp.Result.TxHash)
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

			err = nil
			switch streamName {
			case newTxs:
				err = processMsg(nextNotification, c.newTxsCh)
			case pendingTxs:
				err = processMsg(nextNotification, c.pendingTxsCh)
			case newBlocks:
				err = processMsg(nextNotification, c.newBlocksCh)
			case bdnBlocks:
				err = processMsg(nextNotification, c.bdnBlocksCh)
			case txReceipts:
				err = processMsg(nextNotification, c.txReceiptsCh)
			case ethOnBlock:
				err = processMsg(nextNotification, c.ethOnBlockCh)
			case rawString:
				c.rawChannel <- string(nextNotification)
			default:
				log.Panicf("Unknown stream name: %s", streamName)
			}

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// the response for sending transaction
type sendTxResponse struct {
	Id      int64  `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  *struct {
		TxHash string `json:"txHash"`
	} `json:"result,omitempty"` // subscription ID is here
}

func extractTaskDisabledEvent(response string) (map[string]string, error) {
	callParams := make(map[string]string)

	re := regexp.MustCompile("commandMethod:(\\w+)")
	match := re.FindStringSubmatch(response)
	if match == nil {
		return nil, errors.New(response)
	}
	method := match[1]

	re = regexp.MustCompile("callName:(\\w+)")
	match = re.FindStringSubmatch(response)
	if match == nil {
		return nil, errors.New(response)
	}
	name := match[1]

	re = regexp.MustCompile("callPayload:([\":{},\\w]+)")
	match = re.FindStringSubmatch(response)
	if match == nil {
		return nil, errors.New(response)
	}
	data := match[1]
	json.Unmarshal([]byte(data), &callParams)

	callParams["name"] = name
	callParams["method"] = method

	return callParams, nil
}
