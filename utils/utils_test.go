package utils

import (
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/crypto-crawler/bloxroute-go/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestSubscribePairReserves(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	bloXrouteClient, err := client.NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	pairReservesCh := make(chan *PairReserves)
	pairs := []common.Address{
		common.HexToAddress("0x58f876857a02d6762e0101bb5c46a8c1ed44dc16"),
		common.HexToAddress("0x7efaef62fddcca950418312c6c91aef321375a00"),
		common.HexToAddress("0x0ed7e52944161450477ee417de9cd3a859b14fd0"),
		common.HexToAddress("0x16b9a82891338f9ba80e2d6970fdda79d1eb0dae"),
		common.HexToAddress("0x2354ef4df11afacb85a5c7f98b624072eccddbb1"),
	}

	err = SubscribePairReserves(bloXrouteClient, pairs, pairReservesCh)
	assert.NoError(t, err)

	pairReserves := <-pairReservesCh
	assert.NotEmpty(t, pairReserves.Pair)
	assert.Greater(t, pairReserves.Reserve0.Cmp(big.NewInt(0)), 0)
	assert.Greater(t, pairReserves.Reserve1.Cmp(big.NewInt(0)), 0)

	close(stopCh)
}

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
