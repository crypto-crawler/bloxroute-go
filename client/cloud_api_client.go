package client

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

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
	mu *sync.Mutex
}

func NewCloudApiClient(authorizationHeader string, stopCh <-chan struct{}) (*CloudApiClient, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = tlsConfig
	conn, _, err := dialer.Dial("wss://api.blxrbdn.com/ws", http.Header{"Authorization": []string{authorizationHeader}})
	if err != nil {
		return nil, err
	}

	client := &CloudApiClient{
		conn:   conn,
		stopCh: stopCh,
		mu:     &sync.Mutex{},
	}

	err = client.Ping()
	if err != nil {
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
	_, resp, err := client.conn.ReadMessage()
	if err != nil {
		return common.Hash{}, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(resp, &m)
	if err != nil {
		return common.Hash{}, err
	}
	if _, ok := m["error"]; ok {
		return common.Hash{}, errors.New(fmt.Sprintf("%v", m["error"]))
	}
	if _, ok := m["result"]; ok {
		sendTxResp := sendTxResponse{}
		err = json.Unmarshal(resp, &sendTxResp)
		if err != nil {
			return common.Hash{}, err
		}
		return common.HexToHash(sendTxResp.Result.TxHash), nil
	}

	return common.Hash{}, nil
}

// See https://docs.bloxroute.com/apis/ping.
func (client *CloudApiClient) Ping() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(`{"jsonrpc":"2.0","method":"ping"}`))
	if err != nil {
		return err
	}

	_, nextNotification, err := client.conn.ReadMessage()
	if err != nil {
		return err
	}
	pongMsg := pongMsg{}
	err = json.Unmarshal(nextNotification, &pongMsg)
	if err != nil {
		return err
	}
	if pongMsg.Result.Pong == "" {
		return errors.New(string(nextNotification))
	}
	return nil
}

type pongMsg struct {
	Id      int64  `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		Pong string `json:"pong"`
	} `json:"result"`
}
