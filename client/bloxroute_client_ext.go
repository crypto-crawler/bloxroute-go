package client

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
)

var ZeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")

// All high-level APIs are implemented in this client.
type BloXrouteClientExtended struct {
	client                 *BloXrouteClient
	stopCh                 <-chan struct{}
	ethOnBlockChForBalance chan *types.EthOnBlockResponse
	balancCh               chan *types.TokenBalance
}

func NewBloXrouteClientExtended(client *BloXrouteClient, stopCh <-chan struct{}) *BloXrouteClientExtended {
	clientExt := &BloXrouteClientExtended{
		client:                 client,
		stopCh:                 stopCh,
		ethOnBlockChForBalance: make(chan *types.EthOnBlockResponse),
		balancCh:               make(chan *types.TokenBalance),
	}

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case event := <-clientExt.ethOnBlockChForBalance:
				balance := types.TokenBalance{}
				arr := strings.Split(event.Name, "_")
				user := common.HexToAddress(arr[0])
				token := common.HexToAddress(arr[1])

				balance.Owner = user
				balance.Token = token
				balance.Balance, _ = big.NewInt(0).SetString(event.Response, 0)

				blockNumber, ok := big.NewInt(0).SetString(event.BlockHeight, 0)
				if !ok {
					panic(event)
				}
				balance.BlockNumber = blockNumber.Int64()
				clientExt.balancCh <- &balance
			}
		}
	}()

	return clientExt
}

// Monitor the reserves of of given trading pairs on PancakeSwap/Uniswap.
//
// Please put as many addresses as possible to the pairs parameter and call this function in batch.
func (clientExt *BloXrouteClientExtended) SubscribePairReservesDeprecated(pairs []common.Address, outCh chan<- *types.PairReserves) error {
	outChTmp := make(chan *types.EthOnBlockResponse)
	callParams := make([]map[string]string, 0)
	for _, pair := range pairs {
		name := fmt.Sprintf("pair_%s", pair.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": pair.Hex(), "data": "0x0902f1ac"})
	}

	visited := make(map[uint64]bool)
	go func() {
		for {
			select {
			case <-clientExt.stopCh:
				return
			case resp := <-outChTmp:
				pair := common.HexToAddress(resp.Name[len("pair_"):])
				blockNumber, ok := big.NewInt(0).SetString(resp.BlockHeight, 0)
				if !ok {
					panic(resp)
				}
				pairReserve, err := decodeReturnedDataOfGetReservesDeprecated(pair, resp.Response, blockNumber.Int64())
				if err == nil {
					hash := pairReserve.Hash()
					if !visited[hash] {
						outCh <- pairReserve
						visited[hash] = true
					}
				}
			}
		}
	}()

	_, err := clientExt.client.SubscribeEthOnBlock(nil, callParams, outChTmp)
	return err
}

func (clientExt *BloXrouteClientExtended) SubscribePairReserves(pairs []common.Address, outCh chan<- *types.PairReserves) error {
	outChTmp := make(chan *types.EthOnBlockResponse)
	callParam := map[string]string{"method": "eth_call", "to": "0xAb3A7264ca5B849288fe6a42aBBD4d559552835F"}
	h := md5.New()
	var sb strings.Builder
	// see https://adibas03.github.io/online-ethereum-abi-encoder-decoder/#/encode
	sb.WriteString("0x407a4b080000000000000000000000000000000000000000000000000000000000000020")
	sb.WriteString(fmt.Sprintf("%064x", len(pairs)))
	for _, pair := range pairs {
		io.WriteString(h, pair.Hex())
		// padding
		sb.WriteString("000000000000000000000000")
		sb.WriteString(hex.EncodeToString(pair.Bytes()))
	}
	name := fmt.Sprintf("pairs_0x%x", h.Sum(nil))
	callParam["name"] = name
	callParam["data"] = sb.String()

	go func() {
		for {
			select {
			case <-clientExt.stopCh:
				return
			case resp := <-outChTmp:
				blockNumber, ok := big.NewInt(0).SetString(resp.BlockHeight, 0)
				if !ok {
					panic(resp)
				}
				arr, err := decodeReturnedDataOfGetReserves(pairs, resp.Response, blockNumber.Int64())
				if err == nil {
					for _, pairReserve := range arr {
						outCh <- pairReserve
					}
				}
			}
		}
	}()

	_, err := clientExt.client.SubscribeEthOnBlock(nil, []map[string]string{callParam}, outChTmp)
	return err
}

func (clientExt *BloXrouteClientExtended) SubscribePairReservesForBenchmark(pairs []common.Address, outCh chan<- *types.PairReserves) error {
	outChTmp := make(chan *types.EthOnBlockResponse)
	callParam := map[string]string{"method": "eth_call", "to": "0x45974B68d81Be55E71F7ACD5c1378a9d52CF02Be"}
	h := md5.New()
	var sb strings.Builder
	// see https://adibas03.github.io/online-ethereum-abi-encoder-decoder/#/encode
	sb.WriteString("0xef7b22d90000000000000000000000000000000000000000000000000000000000000020")
	sb.WriteString(fmt.Sprintf("%064x", len(pairs)))
	for _, pair := range pairs {
		io.WriteString(h, pair.Hex())
		// padding
		sb.WriteString("000000000000000000000000")
		sb.WriteString(hex.EncodeToString(pair.Bytes()))
	}
	name := fmt.Sprintf("pairs_0x%x", h.Sum(nil))
	callParam["name"] = name
	callParam["data"] = sb.String()

	go func() {
		for {
			select {
			case <-clientExt.stopCh:
				return
			case resp := <-outChTmp:
				blockNumber, ok := big.NewInt(0).SetString(resp.BlockHeight, 0)
				if !ok {
					panic(resp)
				}
				arr, err := decodeReturnedDataOfGetReservesForBenchmark(pairs, resp.Response, blockNumber.Int64())
				if err == nil {
					for _, pairReserve := range arr {
						outCh <- pairReserve
					}
				}
			}
		}
	}()

	_, err := clientExt.client.SubscribeEthOnBlock(nil, []map[string]string{callParam}, outChTmp)
	return err
}

// Monitor tokens' balance of specified users.
//
// Multiple calls of this function will return the same channel.
func (clientExt *BloXrouteClientExtended) SubscribeBalance(users []common.Address, tokens []common.Address) (<-chan *types.TokenBalance, error) {
	callParams := make([]map[string]string, 0)
	for _, user := range users {
		for _, token := range tokens {
			name := fmt.Sprintf("%s_%s", user.Hex(), token.Hex())
			callParam := make(map[string]string)
			if token == ZeroAddress {
				callParam = map[string]string{"name": name, "method": "eth_getBalance", "address": user.Hex(), "data": "0x0"}
			} else {
				callParam = map[string]string{"name": name, "method": "eth_call", "to": token.Hex(), "data": buildBalanceInputData(user)}
			}
			callParams = append(callParams, callParam)
		}
	}

	_, err := clientExt.client.SubscribeEthOnBlock(nil, callParams, clientExt.ethOnBlockChForBalance)
	if err != nil {
		return nil, err
	}
	return clientExt.balancCh, nil
}

// Decode the data of ethOnBlock of GetReserves()
func decodeReturnedDataOfGetReservesDeprecated(pair common.Address, hexStr string, blockNumber int64) (*types.PairReserves, error) {
	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return nil, err
	}

	pairReserve := &types.PairReserves{
		Pair:               pair,
		Reserve0:           big.NewInt(0).SetBytes(bytes[0:32]),
		Reserve1:           big.NewInt(0).SetBytes(bytes[32:64]),
		BlockTimestampLast: uint32(big.NewInt(0).SetBytes(bytes[64:]).Int64()),
		BlockNumber:        blockNumber,
	}
	return pairReserve, nil
}

func decodeReturnedDataOfGetReserves(pairs []common.Address, hexStr string, blockNumber int64) ([]*types.PairReserves, error) {
	arr, err := decodeTwoDimensionArray(hexStr, 2)
	if err != nil {
		return nil, err
	}
	if len(arr) != len(pairs) {
		panic("Bug: not possible")
	}

	result := make([]*types.PairReserves, len(pairs))
	for i := 0; i < len(pairs); i++ {
		result[i] = &types.PairReserves{
			Pair:        pairs[i],
			Reserve0:    arr[i][0],
			Reserve1:    arr[i][1],
			BlockNumber: blockNumber,
		}
	}
	return result, nil
}

func decodeReturnedDataOfGetReservesForBenchmark(pairs []common.Address, hexStr string, blockNumber int64) ([]*types.PairReserves, error) {
	arr, err := decodeTwoDimensionArray(hexStr, 3)
	if err != nil {
		return nil, err
	}
	if len(arr) != len(pairs) {
		panic("Bug: not possible")
	}

	result := make([]*types.PairReserves, len(pairs))
	for i := 0; i < len(pairs); i++ {
		result[i] = &types.PairReserves{
			Pair:               pairs[i],
			Reserve0:           arr[i][0],
			Reserve1:           arr[i][1],
			BlockTimestampLast: uint32(arr[i][2].Int64()),
			BlockNumber:        blockNumber,
		}
	}
	return result, nil
}

func buildBalanceInputData(owner common.Address) string {
	// balanceOf(address), see // see https://www.4byte.directory/signatures/?bytes4_signature=0x70a08231
	funcSelector := []byte{0x70, 0xa0, 0x82, 0x31}
	addressBytes := common.LeftPadBytes(owner.Bytes(), 32)
	var dataOut []byte
	dataOut = append(dataOut, funcSelector...)
	dataOut = append(dataOut, addressBytes...)
	return fmt.Sprintf("0x%x", dataOut)
}

func decodeFixedArray(hexStr string) ([]*big.Int, error) {
	if len(hexStr)%64 != 0 {
		return nil, fmt.Errorf("Length must be multiple of 64")
	}
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	n := len(bytes) / 32 // number of elements
	result := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		result[i] = big.NewInt(0).SetBytes(bytes[i*32 : (i+1)*32])
	}
	return result, nil
}

// Length followed by elements
func decodeDynamicArray(hexStr string) ([]*big.Int, error) {
	if len(hexStr)%64 != 0 {
		return nil, fmt.Errorf("Length must be multiple of 64")
	}
	n, ok := big.NewInt(0).SetString(hexStr[:64], 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex string %s", hexStr[:64])
	}
	length := int(n.Int64())

	if len(hexStr) != 64+length*64 {
		return nil, fmt.Errorf("invalid hex string %s", hexStr)
	}
	hexStr = hexStr[64:]

	return decodeFixedArray(hexStr)
}

// The first dimension is a fixed length array
// ndim, the fist dimension
func decodeTwoDimensionArray(hexStr string, ndim int) ([][]*big.Int, error) {
	if hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	if hexStr[:64] != "0000000000000000000000000000000000000000000000000000000000000020" {
		return nil, fmt.Errorf("invalid hex string %s", hexStr)
	}
	if len(hexStr)%64 != 0 {
		return nil, fmt.Errorf("Length must be multiple of 64")
	}
	n, ok := big.NewInt(0).SetString(hexStr[64:128], 16)
	if !ok {
		panic("Bug: not possible")
	}
	length := int(n.Int64())

	if len(hexStr) != 128+length*64*ndim {
		return nil, fmt.Errorf("invalid hex string %s", hexStr)
	}
	hexStr = hexStr[128:]

	result := make([][]*big.Int, length)
	for i := 0; i < length; i++ {
		rowStr := hexStr[i*64*ndim : (i+1)*64*ndim]
		arr, err := decodeFixedArray(rowStr)
		if err != nil {
			return nil, err
		}
		result[i] = arr
	}

	return result, nil
}
