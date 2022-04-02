package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
)

// All cloud APIs available on wss://api.blxrbdn.com/ws are implement here.
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
func (client *CloudApiClient) SendTransaction(transactions [][]byte, nonceMonitoring bool, blockchainNetwork string) (common.Hash, error) {
	return common.Hash{}, errors.New("not implemented")
}

// See https://docs.bloxroute.com/apis/ping.
func (client *CloudApiClient) Ping() error {
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
