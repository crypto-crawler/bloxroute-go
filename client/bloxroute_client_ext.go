package client

import (
	"encoding/hex"
	"fmt"
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
func (clientExt *BloXrouteClientExtended) SubscribePairReserves(pairs []common.Address, outCh chan<- *types.PairReserves) error {
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
				pairReserve, err := decodeReturnedDataOfGetReserves(pair, resp.Response, blockNumber.Int64())
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

	return clientExt.client.SubscribeEthOnBlock(nil, callParams, outChTmp)
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

	err := clientExt.client.SubscribeEthOnBlock(nil, callParams, clientExt.ethOnBlockChForBalance)
	if err != nil {
		return nil, err
	}
	return clientExt.balancCh, nil
}

// Decode the data of ethOnBlock of GetReserves()
func decodeReturnedDataOfGetReserves(pair common.Address, hexStr string, blockNumber int64) (*types.PairReserves, error) {
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

func buildBalanceInputData(owner common.Address) string {
	// balanceOf(address), see // see https://www.4byte.directory/signatures/?bytes4_signature=0x70a08231
	funcSelector := []byte{0x70, 0xa0, 0x82, 0x31}
	addressBytes := common.LeftPadBytes(owner.Bytes(), 32)
	var dataOut []byte
	dataOut = append(dataOut, funcSelector...)
	dataOut = append(dataOut, addressBytes...)
	return fmt.Sprintf("0x%x", dataOut)
}
