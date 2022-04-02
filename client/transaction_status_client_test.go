package client

import (
	"encoding/hex"
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

	authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
	if authorizationHeader == "" {
		assert.FailNow(t, "Please provide the authorization header in the AUTHORIZATION_HEADER environment variables")
	}
	transactionStatusClient, err = NewTransactionStatusClient(authorizationHeader, stopCh, txStatusCh)
	assert.NoError(t, err)
	assert.NotNil(t, transactionStatusClient)

	{
		gatewayUrl := os.Getenv("GATEWAY_URL")
		if gatewayUrl == "" {
			assert.FailNow(t, "Please provide the gateway URL in the GATEWAY_URL environment variables")
		}
		// raw_tx is not available on cloud API
		bloXrouteClient, err = NewBloXrouteClientToGateway(gatewayUrl, authorizationHeader, stopCh)
		assert.NoError(t, err)
		assert.NotNil(t, bloXrouteClient)
	}

	txCh := make(chan *bloxroute_types.Transaction)
	err = bloXrouteClient.SubscribePendingTxs([]string{"tx_hash", "tx_contents", "raw_tx"}, "", txCh)
	assert.NoError(t, err)

	txJson := <-txCh
	assert.NotEmpty(t, txJson.TxHash)

	txRawComputed, err := txJson.TxContents.ToRaw()
	txRaw, err := hex.DecodeString(txJson.RawTx[2:])
	assert.NoError(t, err)
	assert.NotNil(t, txRaw)
	assert.Equal(t, txRawComputed, txRaw)

	err = transactionStatusClient.StartMonitorTransaction([][]byte{txRaw}, false)

	txStatus := <-txStatusCh
	assert.NotNil(t, txStatus)
	assert.Equal(t, common.HexToHash(txStatus.TxHash), common.HexToHash(txJson.TxHash))

	err = transactionStatusClient.StopMonitorTransaction(common.HexToHash(txStatus.TxHash))
	assert.NoError(t, err)

	close(stopCh)
}
