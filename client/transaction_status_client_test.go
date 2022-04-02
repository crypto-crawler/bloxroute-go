package client

import (
	"os"
	"testing"

	bloxroute_types "github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestStartMonitorTransaction(t *testing.T) {
	var transactionStatusClient *TransactionStatusClient = nil
	var bloXrouteClient *BloXrouteClient = nil
	var err error = nil
	stopCh := make(chan struct{})
	txStatusCh := make(chan *bloxroute_types.TxStatus)
	{
		authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
		if authorizationHeader == "" {
			assert.FailNow(t, "Please provide the authorization header in the AUTHORIZATION_HEADER environment variables")
		}
		transactionStatusClient, err = NewTransactionStatusClient(authorizationHeader, stopCh, txStatusCh)
		assert.NoError(t, err)
		assert.NotNil(t, transactionStatusClient)
	}
	{
		certFile := os.Getenv("BLOXROUTE_CERT_FILE")
		keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
		if certFile == "" || keyFile == "" {
			assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
		}
		bloXrouteClient, err = NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
		assert.NoError(t, err)
		assert.NotNil(t, bloXrouteClient)
	}

	txCh := make(chan *bloxroute_types.Transaction)
	err = bloXrouteClient.SubscribePendingTxs([]string{"tx_hash", "tx_contents"}, "", txCh)
	assert.NoError(t, err)

	txJson := <-txCh
	assert.NotEmpty(t, txJson.TxHash)

	txRaw, err := txJson.TxContents.ToRaw()
	assert.NoError(t, err)
	assert.NotNil(t, txRaw)

	err = transactionStatusClient.StartMonitorTransaction([][]byte{txRaw}, false)

	txStatus := <-txStatusCh
	assert.NotNil(t, txStatus)
	assert.Equal(t, common.HexToHash(txStatus.TxHash), common.HexToHash(txJson.TxHash))

	err = transactionStatusClient.StopMonitorTransaction([]common.Hash{common.HexToHash(txStatus.TxHash)})
	assert.NoError(t, err)

	close(stopCh)
}
