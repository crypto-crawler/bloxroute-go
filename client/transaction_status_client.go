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
}

func NewTransactionStatusClient(authorizationHeader string, stopCh <-chan struct{}, outCh chan<- *types.TxStatus) (*TransactionStatusClient, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = tlsConfig
	conn, _, err := dialer.Dial("wss://api.blxrbdn.com/ws", http.Header{"Authorization": []string{authorizationHeader}})
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
	}

	go client.ping()
	go client.run()

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
	client.mu.Lock()
	defer client.mu.Unlock()

	arr := make([]string, len(transactions))
	for i, tx := range transactions {
		arr[i] = hex.EncodeToString(tx)
	}
	bytes, _ := json.Marshal(arr)
	arrJson := string(bytes)
	subRequest := fmt.Sprintf(`{"jsonrpc":"2.0","method":"start_monitor_transaction","params":{"transactions":%s,"monitor_speedup":"%v"}}`, arrJson, monitorSpeedup)

	err := client.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
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
	log.Println(subRequest)
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(subRequest))
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

// Sends a ping message every 5 seconds to keep the WebSocket connection alive.
// https://docs.bloxroute.com/apis/ping
func (client *TransactionStatusClient) ping() {
	ticker := time.NewTicker(5 * time.Second) // ping every 5 seconds
	for {
		select {
		case <-client.stopCh:
			ticker.Stop()
			return
		case <-ticker.C:
			client.mu.Lock()
			defer client.mu.Unlock()
			err := client.conn.WriteMessage(websocket.TextMessage, []byte(`{"method": "ping"}`))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Close is used to terminate our websocket client
func (client *TransactionStatusClient) close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	err := client.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
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
				// Is it a command response?
				commandResp := commandResponse{}
				err = json.Unmarshal(nextNotification, &commandResp)
				if err == nil {
					if commandResp.Result != nil {
						client.commandResponseCh <- commandResp
						break
					}
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
