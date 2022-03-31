package client

import (
	"os"
	"testing"

	"github.com/crypto-crawler/bloxroute-go/types"
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
	err = client.SubscribeNewTxs(nil, "", txCh)
	assert.NoError(t, err)

	tx := <-txCh
	assert.NotEmpty(t, tx.TxHash)

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
	err = client.SubscribeNewBlocks(nil, blockCh)
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
	err = client.SubscribeTxReceipts(nil, receiptsCh)
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
	client, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	respCh := make(chan *types.EthOnBlockResponse)
	callParams := make([]map[string]string, 0)
	callParams = append(callParams, map[string]string{"name": "blockNumber", "method": "eth_blockNumber"})
	err = client.SubscribeEthOnBlock(nil, callParams, respCh)
	assert.NoError(t, err)

	x := <-respCh
	assert.Equal(t, "blockNumber", x.Name)
	assert.NotEmpty(t, x.Response)

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
