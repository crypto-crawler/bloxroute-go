package utils

import (
	"log"
	"os"
	"testing"

	"github.com/crypto-crawler/bloxroute-go/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeWalletBalance(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	bloXrouteClient, err := client.NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	balancesCh := make(chan *WalletBalance)
	addresses := []common.Address{
		common.HexToAddress("0x95eA23508ecc3521081E72352C13707F6b179Fc1"),
	}

	err = SubscribeWalletBalance(bloXrouteClient, addresses, balancesCh)
	assert.NoError(t, err)

	balance := <-balancesCh
	log.Println(balance)
	if balance.Bnb != nil {
		assert.Equal(t, balance.Bnb.String(), "2618353928878013888")
	}

	balance = <-balancesCh
	log.Println(balance)
	if balance.Bnb != nil {
		assert.Equal(t, balance.Bnb.String(), "2618353928878013888")
	}

	balance = <-balancesCh
	log.Println(balance)
	if balance.Bnb != nil {
		assert.Equal(t, balance.Bnb.String(), "2618353928878013888")
	}
	close(stopCh)
}
