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

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

type sendTxResponse struct {
	Id      int64  `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  *struct {
		TxHash string `json:"txHash"`
	} `json:"result,omitempty"` // subscription ID is here
}

// All cloud APIs available on wss://api.blxrbdn.com/ws are implemented here.
type CloudApiClient struct {
	conn   *websocket.Conn
	stopCh <-chan struct{}
	// To ensure there is only one concurrent conn.WriteMessage(), see
	// https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency
	mu     *sync.Mutex
	respCh chan sendTxResponse
	pongCh chan pongMsg
}

// Constructor.
//
// You can pass an optional url to connect to a specific IP address,
// it defaults to wss://api.blxrbdn.com/ws if empty.
//
// Availabel IP addresses are here https://docs.bloxroute.com/introduction/cloud-api-ips
func NewCloudApiClient(authorizationHeader string, stopCh <-chan struct{}, url string) (*CloudApiClient, error) {
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

	client := &CloudApiClient{
		conn:   conn,
		stopCh: stopCh,
		mu:     &sync.Mutex{},
		respCh: make(chan sendTxResponse),
		pongCh: make(chan pongMsg),
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

// Send a transaction via BDN.
func (client *CloudApiClient) SendTransaction(transaction []byte, nonceMonitoring bool, blockchainNetwork string) (common.Hash, error) {
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
		return common.Hash{}, errors.New("timeout")
	case resp := <-client.respCh:
		return common.HexToHash(resp.Result.TxHash), nil
	}
}

// Ping() sends a ping message to websocket server.
// See https://docs.bloxroute.com/apis/ping
func (c *CloudApiClient) Ping() (time.Time, error) {
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
func (c *CloudApiClient) close() error {
	err := c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(3*time.Second))
	c.conn.Close()
	return err
}

// Run the event looop.
func (c *CloudApiClient) run() error {
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
				// Is it a sendTxResponse?
				resp := sendTxResponse{}
				err = json.Unmarshal(nextNotification, &resp)
				if err == nil {
					if resp.Result != nil && resp.Result.TxHash != "" {
						c.respCh <- resp
						break
					}
				}
			}
		}
	}
}

type pongMsg struct {
	Id      int64  `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		Pong string `json:"pong"`
	} `json:"result"`
}
