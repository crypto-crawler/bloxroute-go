package client

import (
	"math/big"
	"os"
	"testing"

	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestSubscribePairReservesDeprecated(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	bloXrouteClient, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	clientExt := NewBloXrouteClientExtended(bloXrouteClient, stopCh)

	pairReservesCh := make(chan *types.PairReserves)
	pairs := []common.Address{
		common.HexToAddress("0x58f876857a02d6762e0101bb5c46a8c1ed44dc16"),
		common.HexToAddress("0x7efaef62fddcca950418312c6c91aef321375a00"),
		common.HexToAddress("0x0ed7e52944161450477ee417de9cd3a859b14fd0"),
		common.HexToAddress("0x16b9a82891338f9ba80e2d6970fdda79d1eb0dae"),
		common.HexToAddress("0x2354ef4df11afacb85a5c7f98b624072eccddbb1"),
	}

	err = clientExt.SubscribePairReservesDeprecated(pairs, pairReservesCh)
	assert.NoError(t, err)

	pairReserves := <-pairReservesCh
	assert.NotEmpty(t, pairReserves.Pair)
	assert.Greater(t, pairReserves.Reserve0.Cmp(big.NewInt(0)), 0)
	assert.Greater(t, pairReserves.Reserve1.Cmp(big.NewInt(0)), 0)

	close(stopCh)
}

func TestSubscribePairReserves(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	bloXrouteClient, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	clientExt := NewBloXrouteClientExtended(bloXrouteClient, stopCh)

	pairReservesCh := make(chan *types.PairReserves)
	pairs := []common.Address{
		common.HexToAddress("0x58f876857a02d6762e0101bb5c46a8c1ed44dc16"),
		common.HexToAddress("0x7efaef62fddcca950418312c6c91aef321375a00"),
		common.HexToAddress("0x0ed7e52944161450477ee417de9cd3a859b14fd0"),
		common.HexToAddress("0x16b9a82891338f9ba80e2d6970fdda79d1eb0dae"),
		common.HexToAddress("0x2354ef4df11afacb85a5c7f98b624072eccddbb1"),
	}

	err = clientExt.SubscribePairReserves(pairs, pairReservesCh)
	assert.NoError(t, err)

	pairReserves := <-pairReservesCh
	assert.NotEmpty(t, pairReserves.Pair)
	assert.Greater(t, pairReserves.Reserve0.Cmp(big.NewInt(0)), 0)
	assert.Greater(t, pairReserves.Reserve1.Cmp(big.NewInt(0)), 0)

	close(stopCh)
}

func TestSubscribePairReservesForBenchmark(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	bloXrouteClient, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)

	clientExt := NewBloXrouteClientExtended(bloXrouteClient, stopCh)

	pairReservesCh := make(chan *types.PairReserves)
	pairs := []common.Address{
		common.HexToAddress("0x58f876857a02d6762e0101bb5c46a8c1ed44dc16"),
		common.HexToAddress("0x7efaef62fddcca950418312c6c91aef321375a00"),
		common.HexToAddress("0x0ed7e52944161450477ee417de9cd3a859b14fd0"),
		common.HexToAddress("0x16b9a82891338f9ba80e2d6970fdda79d1eb0dae"),
		common.HexToAddress("0x2354ef4df11afacb85a5c7f98b624072eccddbb1"),
	}

	err = clientExt.SubscribePairReservesForBenchmark(pairs, pairReservesCh)
	assert.NoError(t, err)

	pairReserves := <-pairReservesCh
	assert.NotEmpty(t, pairReserves.Pair)
	assert.Greater(t, pairReserves.Reserve0.Cmp(big.NewInt(0)), 0)
	assert.Greater(t, pairReserves.Reserve1.Cmp(big.NewInt(0)), 0)
	assert.Greater(t, pairReserves.BlockTimestampLast, uint32(0))
	close(stopCh)
}

func TestSubscribeBalance(t *testing.T) {
	certFile := os.Getenv("BLOXROUTE_CERT_FILE")
	keyFile := os.Getenv("BLOXROUTE_KEY_FILE")
	if certFile == "" || keyFile == "" {
		assert.FailNow(t, "Please provide the bloXroute cert and key files path in the environment variable variable")
	}

	stopCh := make(chan struct{})
	bloXrouteClient, err := NewBloXrouteClientToCloud("BSC-Mainnet", certFile, keyFile, stopCh)
	assert.NoError(t, err)
	clientExt := NewBloXrouteClientExtended(bloXrouteClient, stopCh)

	owners := []common.Address{common.HexToAddress("0x95eA23508ecc3521081E72352C13707F6b179Fc1")}
	tokens := []common.Address{
		ZeroAddress, // BNB
		common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"), // WBNB
		common.HexToAddress("0x55d398326f99059fF775485246999027B3197955"), // USDT
		common.HexToAddress("0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"), // BUSD
	}

	balanceCh, err := clientExt.SubscribeBalance(owners, tokens)
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		balance := <-balanceCh
		if balance.Token == ZeroAddress {
			assert.Equal(t, balance.Balance.String(), "2618353928878013888")
		}
	}

	close(stopCh)
}

func TestDecodeFixedArray(t *testing.T) {
	hexStr := "0000000000000000000000000000000000000000000063b1c073d196bec87b70000000000000000000000000000000000000000000a35dcd58131666ba651e10"
	arr, err := decodeFixedArray(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, "63b1c073d196bec87b70", arr[0].Text(16))
	assert.Equal(t, "a35dcd58131666ba651e10", arr[1].Text(16))
}

func TestDecodeDynamicArray(t *testing.T) {
	hexStr := "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000063b1c073d196bec87b70000000000000000000000000000000000000000000a35dcd58131666ba651e10"
	arr, err := decodeDynamicArray(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, "63b1c073d196bec87b70", arr[0].Text(16))
	assert.Equal(t, "a35dcd58131666ba651e10", arr[1].Text(16))
}
