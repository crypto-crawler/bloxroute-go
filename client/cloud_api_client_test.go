package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSendTransaction(t *testing.T) {
	var cloudApiClient *CloudApiClient = nil
	//	var bloXrouteClient *BloXrouteClient = nil
	var err error = nil
	stopCh := make(chan struct{})

	authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
	if authorizationHeader == "" {
		assert.FailNow(t, "Please provide the authorization header in the AUTHORIZATION_HEADER environment variables")
	}
	cloudApiClient, err = NewCloudApiClient(authorizationHeader, stopCh)
	assert.NoError(t, err)
	assert.NotNil(t, cloudApiClient)

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
		Hash:     "0xdf2c69ac03477a82118c7758b853c9bc7bc29667804b27b5434d6d9da86aa0ff",
	}
	raw, err := txContents.ToRaw()

	txHash, err := cloudApiClient.SendTransaction(raw, false, "BSC-Mainnet")
	fmt.Printf("txHash: %s\n", txHash)
	assert.NoError(t, err)
	assert.Equal(t, txHash.Hex(), "0x529fe2e13417033ae0ed4e7efab3fd37a09be85b3d60ddc9bdd95665b2f37ef1")
	close(stopCh)
}
