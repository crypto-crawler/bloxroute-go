package client

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	bloxroute_types "github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TransactionByHash(ethClient *ethclient.Client, txHash common.Hash) (*types.Transaction, error) {
	ctx := context.Background()
	headerCh := make(chan *types.Header)

	sub, err := ethClient.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-time.After(time.Second * 10):
			sub.Unsubscribe()
			return nil, errors.New("timeout")
		case <-headerCh:
			tx, _, err := ethClient.TransactionByHash(ctx, txHash)
			if tx != nil || err == nil {
				return tx, nil
			}
		}
	}
}

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

	var ethClient *ethclient.Client = nil
	{
		fullnodeUrl := os.Getenv("FULLNODE_URL")
		if fullnodeUrl == "" {
			assert.FailNow(t, "Please provide the fullnode URL by setting the FULLNODE_URL environment variable")
		}
		ctx := context.Background()
		ethClient, err = ethclient.DialContext(ctx, fullnodeUrl)
	}
	assert.NoError(t, err)
	assert.NotNil(t, ethClient)

	txHashCh := make(chan *bloxroute_types.Transaction)
	err = bloXrouteClient.SubscribeNewTxs([]string{"tx_hash"}, "", txHashCh)
	assert.NoError(t, err)

	txHash := <-txHashCh
	assert.NotEmpty(t, txHash.TxHash)

	tx, err := TransactionByHash(ethClient, common.HexToHash(txHash.TxHash))
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	txRaw, err := tx.MarshalBinary()
	assert.NoError(t, err)
	err = transactionStatusClient.StartMonitorTransaction([][]byte{txRaw}, false)

	txStatus := <-txStatusCh
	assert.NotNil(t, txStatus)
	assert.Equal(t, common.HexToHash(txStatus.TxHash), common.HexToHash(txHash.TxHash))

	err = transactionStatusClient.StopMonitorTransaction([]common.Hash{common.HexToHash(txStatus.TxHash)})
	assert.NoError(t, err)

	close(stopCh)
}
