package client

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/crypto-crawler/bloxroute-go/types"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestNewTxs(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	txCh := make(chan *types.Transaction)
	_, err = client.SubscribeNewTxs(nil, "", txCh)
	assert.NoError(t, err)

	tx := <-txCh
	assert.NotEmpty(t, tx.TxHash)

	close(stopCh)
}

func TestNewTxsWithFilter(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert     and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)

	authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
	if authorizationHeader == "" {
		assert.FailNow(t, "Please provide  the authorization header in the AUTHORIZATION_HEADER  environment variables")
	}
	// client, err := NewBloXrouteClientToGateway("ws://localhost:28334", authorizationHeader, stopCh)

	assert.NoError(t, err)

	txCh := make(chan *types.Transaction)
	// monitor transactions sent to PancakeSwap router
	filter := fmt.Sprintf("{to} == '%s'", "0x10ed43c718714eb63d5aa57b78b54704e256024e")
	_, err = client.SubscribeNewTxs(nil, filter, txCh)
	assert.NoError(t, err)

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case tx := <-txCh:
				assert.NotEmpty(t, tx.TxHash)
				assert.Equal(t, tx.TxContents.To, "0x10ed43c718714eb63d5aa57b78b54704e256024e")
			}
		}
	}()

	tx := <-txCh
	assert.NotEmpty(t, tx.TxHash)
	assert.Equal(t, tx.TxContents.To, "0x10ed43c718714eb63d5aa57b78b54704e256024e")
	time.Sleep(time.Second * 5)

	close(stopCh)
}

func TestNewBlocks(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	blockCh := make(chan *types.Block)
	_, err = client.SubscribeNewBlocks(nil, blockCh)
	assert.NoError(t, err)

	block := <-blockCh
	assert.NotEmpty(t, block.Hash)
	assert.NotEmpty(t, block.Transactions)

	close(stopCh)
}

func TestTxReceipts(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	receiptsCh := make(chan *types.TxReceipt)
	_, err = client.SubscribeTxReceipts(nil, receiptsCh)
	assert.NoError(t, err)

	receipt := <-receiptsCh
	assert.NotEmpty(t, receipt.TransactionHash)
	assert.NotEmpty(t, receipt.Status)

	close(stopCh)
}

func TestBlockNumberFromEthOnBlock(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	// client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)

	// stopCh := make(chan struct{})
	authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
	if authorizationHeader == "" {
		assert.FailNow(t, "Please provide the authorization header in the AUTHORIZATION_HEADER environment variables")
	}
	client, err := NewBloXrouteClientToGateway("ws://localhost:28334", authorizationHeader, stopCh)

	assert.NoError(t, err)
	respCh := make(chan *types.EthOnBlockResponse)
	callParams := make([]map[string]string, 0)
	callParams = append(callParams, map[string]string{"name": "blockNumber", "method": "eth_blockNumber"})
	_, err = client.SubscribeEthOnBlock(nil, callParams, respCh)
	assert.NoError(t, err)

	x := <-respCh
	assert.Equal(t, "blockNumber", x.Name)
	assert.NotEmpty(t, x.Response)

	close(stopCh)
}

func TestSubscribeRaw(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	outCh := make(chan string)
	subscriptionID, err := client.SubscribeRaw(`{"method":"subscribe","params":["newTxs",{"include":["tx_hash","tx_contents"]}]}`, outCh)
	assert.NoError(t, err)
	assert.NotEmpty(t, subscriptionID)

	msg := <-outCh
	assert.NotEmpty(t, msg)

	close(stopCh)
}

func TestUnsubscribeShouldFail(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	err = client.Unsubscribe("an-id-that-does-not-exist")
	assert.Error(t, err)

	close(stopCh)
}

func TestUnsubscribeShouldSucceed(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	outCh := make(chan string)
	subscriptionID, err := client.SubscribeRaw(`{"method": "subscribe", "params": ["bdnBlocks",{"include":["hash"]}]}`, outCh)
	assert.NoError(t, err)
	assert.NotEmpty(t, subscriptionID)

	err = client.Unsubscribe(subscriptionID)
	assert.NoError(t, err)

	close(stopCh)
}

func TestExtractTaskDisabledEvent(t *testing.T) {
	response := "{commandMethod:eth_call blockOffset:0 callName:pair_0xB65697ec1A73eC1bF82677e62Cb86d9369Ba6c34 callPayload:{\"data\":\"0x0902f1ac\",\"to\":\"0xB65697ec1A73eC1bF82677e62Cb86d9369Ba6c34\"} active:false}"
	callParams, err := extractTaskDisabledEvent(response)
	assert.NoError(t, err)
	assert.Equal(t, "eth_call", callParams["method"])
	assert.Equal(t, "pair_0xB65697ec1A73eC1bF82677e62Cb86d9369Ba6c34", callParams["name"])
	assert.Equal(t, "0xB65697ec1A73eC1bF82677e62Cb86d9369Ba6c34", callParams["to"])
	assert.Equal(t, "0x0902f1ac", callParams["data"])
}

func TestClientSendTransaction(t *testing.T) {
	stopCh := make(chan struct{})
	authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
	if authorizationHeader == "" {
		assert.FailNow(t, "Please provide the authorization header in the AUTHORIZATION_HEADER environment variables")
	}
	// wss://api.blxrbdn.com/ws
	url := "ws://localhost:28334"
	client, err := NewBloXrouteClientToGateway(url, authorizationHeader, stopCh)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	txContents := types.TxContents{
		Type:     "0x0",
		Nonce:    "0x7",
		Gas:      "0x11a3e",
		Value:    "0x0",
		Input:    "0xa9059cbb0000000000000000000000008894e0a0c962cb723c1976a4421c95949be2d4e30000000000000000000000000000000000000000000000246f4da6499993c000",
		V:        "0x93",
		R:        "0xa175accc671df00bfadb5627ab0a92ffe07ce9cf0115ce8720d926926cea0910",
		S:        "0x40a38a22c424bf8c04cb8556ae3aa89102429f44c23c6af83283564a75ebbdc",
		To:       "0xe9e7cea3dedca5984780bafc599bd69add087d56",
		From:     "0x17db3ed2d06fee92de7bef42c4f4edcbda7494f4",
		GasPrice: "0x12a05f200",
		Hash:     "0x529fe2e13417033ae0ed4e7efab3fd37a09be85b3d60ddc9bdd95665b2f37ef1",
	}
	raw, err := txContents.ToRaw()
	tx := new(geth_types.Transaction)
	err = tx.UnmarshalBinary(raw)
	txHash, err := client.SendTransaction(raw, false, "BSC-Mainnet")
	assert.NoError(t, err)
	assert.Equal(t, txHash.Hex(), tx.Hash().Hex())
	close(stopCh)
}
