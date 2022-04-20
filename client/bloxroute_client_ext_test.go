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
	// The hex string below is generated by https://adibas03.github.io/online-ethereum-abi-encoder-decoder/#/encode
	hexStr := "0000000000000000000000000000000000000000000063b1c073d196bec87b70000000000000000000000000000000000000000000a35dcd58131666ba651e10"
	arr, err := decodeFixedArray(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(arr))
	assert.Equal(t, "63b1c073d196bec87b70", arr[0].Text(16))
	assert.Equal(t, "a35dcd58131666ba651e10", arr[1].Text(16))
}

func TestDecodeDynamicArray(t *testing.T) {
	// The hex string below is generated by https://adibas03.github.io/online-ethereum-abi-encoder-decoder/#/encode
	hexStr := "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000063b1c073d196bec87b70000000000000000000000000000000000000000000a35dcd58131666ba651e10"
	arr, err := decodeDynamicArray(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(arr))
	assert.Equal(t, "63b1c073d196bec87b70", arr[0].Text(16))
	assert.Equal(t, "a35dcd58131666ba651e10", arr[1].Text(16))
}

func TestDecodeTwoDimensionArray(t *testing.T) {
	// The hex string below is generated by https://adibas03.github.io/online-ethereum-abi-encoder-decoder/#/encode
	hexStr := "000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000063c4db5d580600ff82c9000000000000000000000000000000000000000000a3d27b16209517e6e55b1500000000000000000000000000000000000000000000000000000000625f77620000000000000000000000000000000000000000006cde862d673a796c5fd7b40000000000000000000000000000000000000000006ce4df7f27f0b8c15a533500000000000000000000000000000000000000000000000000000000625f77620000000000000000000000000000000000000000000ca9d3ff04a8ca6fb957eb00000000000000000000000000000000000000000000425133d69920be1ce48b00000000000000000000000000000000000000000000000000000000625f775f0000000000000000000000000000000000000000008dcdac85e90e0b91a345ce000000000000000000000000000000000000000000005658cfa8227b7d82fe1500000000000000000000000000000000000000000000000000000000625f77620000000000000000000000000000000000000000002ead0135d9dbaabd86154c0000000000000000000000000000000000000000002e9c84518776c01d761aca00000000000000000000000000000000000000000000000000000000625f7762"
	arr, err := decodeTwoDimensionArray(hexStr, 3)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(arr))
	assert.Equal(t, 3, len(arr[0]))

	assert.Equal(t, "63c4db5d580600ff82c9", arr[0][0].Text(16))
	assert.Equal(t, "a3d27b16209517e6e55b15", arr[0][1].Text(16))
	assert.Equal(t, "625f7762", arr[0][2].Text(16))

	assert.Equal(t, "6cde862d673a796c5fd7b4", arr[1][0].Text(16))
	assert.Equal(t, "6ce4df7f27f0b8c15a5335", arr[1][1].Text(16))
	assert.Equal(t, "625f7762", arr[1][2].Text(16))

	assert.Equal(t, "ca9d3ff04a8ca6fb957eb", arr[2][0].Text(16))
	assert.Equal(t, "425133d69920be1ce48b", arr[2][1].Text(16))
	assert.Equal(t, "625f775f", arr[2][2].Text(16))

	assert.Equal(t, "8dcdac85e90e0b91a345ce", arr[3][0].Text(16))
	assert.Equal(t, "5658cfa8227b7d82fe15", arr[3][1].Text(16))
	assert.Equal(t, "625f7762", arr[3][2].Text(16))

	assert.Equal(t, "2ead0135d9dbaabd86154c", arr[4][0].Text(16))
	assert.Equal(t, "2e9c84518776c01d761aca", arr[4][1].Text(16))
	assert.Equal(t, "625f7762", arr[4][2].Text(16))
}
