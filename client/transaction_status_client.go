package client

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

// A client to subscribe to the `transactionStatus` stream.
// See https://docs.bloxroute.com/streams/txstatus
type TransactionStatusClient struct {
	conn           *websocket.Conn
	subscriptionID string
	stopCh         <-chan struct{}
	outCh          chan<- *types.TxStatus
	// To ensure there is only one concurrent conn.WriteMessage(), see
	// https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency
	mu                *sync.Mutex
	commandResponseCh chan commandResponse
	pongCh            chan pongMsg
}

// Constructor.
//
// You can pass an optional url to connect to a specific IP address,
// it defaults to wss://api.blxrbdn.com/ws if empty.
//
// Availabel IP addresses are here https://docs.bloxroute.com/introduction/cloud-api-ips
func NewTransactionStatusClient(authorizationHeader string, stopCh <-chan struct{}, outCh chan<- *types.TxStatus, url string) (*TransactionStatusClient, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = tlsConfig
	if url == "" {
		url = "wss://api.blxrbdn.com/ws"
	}
	conn, _, err := dialer.Dial(url, http.Header{"Authorization": []string{authorizationHeader}})
	if err != nil {
		return nil, err
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"jsonrpc":"2.0","method":"subscribe","params":["transactionStatus",{"include":["tx_hash","status"]}]}`))
	if err != nil {
		return nil, err
	}

	_, nextNotification, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	subResp := types.SubscriptionResponse{}
	err = json.Unmarshal(nextNotification, &subResp)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	subscriptionID := subResp.Result
	if subscriptionID == "" {
		return nil, errors.New("subscriptionID is empty")
	}

	client := &TransactionStatusClient{
		conn:              conn,
		subscriptionID:    subscriptionID,
		stopCh:            stopCh,
		outCh:             outCh,
		mu:                &sync.Mutex{},
		commandResponseCh: make(chan commandResponse),
		pongCh:            make(chan pongMsg),
	}

	go client.run()

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

	// test the connection
	if _, err = client.Ping(); err != nil {
		return nil, err
	}

	return client, nil
}

type commandResponse struct {
	Id      int64  `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  *struct {
		Success bool `json:"success"`
	} `json:"result,omitempty"` // subscription ID is here
}

// Monitor given transactions.
// transactions is a list of raw transactions.
func (client *TransactionStatusClient) StartMonitorTransaction(transactions [][]byte, monitorSpeedup bool) error {
	arr := make([]string, len(transactions))
	for i, tx := range transactions {
		arr[i] = hex.EncodeToString(tx)
	}
	bytes, _ := json.Marshal(arr)
	arrJson := string(bytes)
	subRequest := fmt.Sprintf(`{"jsonrpc":"2.0","method":"start_monitor_transaction","params":{"transactions":%s,"monitor_speedup":"%v"}}`, arrJson, monitorSpeedup)

	client.mu.Lock()
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
	client.mu.Unlock()
	if err != nil {
		return err
	}

	select {
	case <-time.After(3 * time.Second):
		return errors.New("Timeout waiting for start_monitor_transaction response")
	case resp := <-client.commandResponseCh:
		if resp.Result.Success {
			return nil
		} else {
			bytes, _ := json.Marshal(resp.Result)
			return errors.New(string(bytes))
		}
	}
}

// Stop monitoring a transaction.
func (client *TransactionStatusClient) StopMonitorTransaction(txHash common.Hash) error {
	subRequest := fmt.Sprintf(`{"jsonrpc":"2.0","method":"stop_monitor_transaction","params":{"transaction_hash":"%s"}}`, txHash.Hex()[2:])
	client.mu.Lock()
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
	client.mu.Unlock()
	if err != nil {
		return err
	}

	select {
	case <-time.After(3 * time.Second):
		return errors.New("Timeout waiting for stop_monitor_transaction response")
	case resp := <-client.commandResponseCh:
		if resp.Result.Success {
			return nil
		} else {
			bytes, _ := json.Marshal(resp.Result)
			return errors.New(string(bytes))
		}
	}
}

// Ping() sends a ping message to websocket server.
// See https://docs.bloxroute.com/apis/ping
func (c *TransactionStatusClient) Ping() (time.Time, error) {
	c.mu.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(`{"method": "ping"}`))
	c.mu.Unlock()
	if err != nil {
		return time.Now(), err
	}

	// wait for pong
	select {
	case <-time.After(3 * time.Second):
		return time.Now(), errors.New("timeout")
	case resp := <-c.pongCh:
		return time.Parse("2006-01-02 15:04:05.99", resp.Result.Pong)
	}
}

// Close is used to terminate our websocket client
func (client *TransactionStatusClient) close() error {
	err := client.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(3*time.Second))
	client.conn.Close()
	return err
}

// Run the event looop.
func (client *TransactionStatusClient) run() error {
	for {
		select {
		case <-client.stopCh:
			client.close()
			return nil
		default:
			_, nextNotification, err := client.conn.ReadMessage()
			if err != nil {
				log.Println(err)
				if ce, ok := err.(*websocket.CloseError); ok {
					switch ce.Code {
					case websocket.CloseNormalClosure,
						websocket.CloseGoingAway,
						websocket.CloseNoStatusReceived,
						websocket.CloseAbnormalClosure:
						return nil
					default:
						return err
					}
				} else {
					return err
				}
			}

			{
				// Is it pongMsg?
				pongMsg := pongMsg{}
				json.Unmarshal(nextNotification, &pongMsg)
				if err == nil {
					if pongMsg.Result.Pong != "" {
						client.pongCh <- pongMsg
						break
					}
				}
			}

			{
				m := make(map[string]interface{})
				err = json.Unmarshal(nextNotification, &m)
				if err != nil {
					log.Fatal(string(nextNotification))
				}
				if _, ok := m["error"]; ok {
					log.Fatal(string(nextNotification))
				}
				if _, ok := m["result"]; ok {
					result := m["result"].(map[string]interface{})
					if _, ok := result["success"]; ok {
						// a command response
						commandResp := commandResponse{}
						err = json.Unmarshal(nextNotification, &commandResp)
						if err == nil {
							client.commandResponseCh <- commandResp
						}
					}
					break
				}
			}

			msg := types.WebsocketMsg[*types.TxStatus]{}
			err = json.Unmarshal(nextNotification, &msg)
			if err != nil {
				return err
			}
			if len(msg.Error) > 0 {
				return errors.New(string(nextNotification))
			}

			if msg.Params.Result != nil && msg.Params.Result.TxHash != "" {
				client.outCh <- msg.Params.Result
			}
		}
	}
}
