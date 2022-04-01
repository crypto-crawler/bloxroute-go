// High-level APIs.
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/crypto-crawler/bloxroute-go/client"
	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
)

type PairReserves struct {
	Pair               common.Address `json:"pair"`
	Reserve0           *big.Int       `json:"reserve0"`
	Reserve1           *big.Int       `json:"reserve1"`
	BlockTimestampLast uint32         `json:"block_timestamp_last"`
	BlockNumber        int64          `json:"block_number"`
}

type WalletBalance struct {
	Address            common.Address `json:"address"`
	Bnb                *big.Int       `json:"bnb"`
	Usdt               *big.Int       `json:"usdt"`
	Busd               *big.Int       `json:"busd"`
	BlockTimestampLast int64          `json:"block_timestamp_last"`
	BlockNumber        int64          `json:"block_number"`
}

var BNB = common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
var USDT = common.HexToAddress("0x55d398326f99059fF775485246999027B3197955")
var BUSD = common.HexToAddress("0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56")

func (p *PairReserves) Hash() uint64 {
	h := md5.New()
	h.Write(p.Pair.Bytes())
	h.Write(p.Reserve0.Bytes())
	h.Write(p.Reserve1.Bytes())
	{
		n := big.NewInt(0)
		n.SetInt64(p.BlockNumber)
		h.Write(n.Bytes())
	}
	bs := h.Sum(nil)

	return big.NewInt(0).SetBytes(bs).Uint64()
}

// Monitor the reserves of of given trading pairs on PancakeSwap/Uniswap, etc.
func SubscribePairReserves(bloXrouteClient *client.BloXrouteClient, pairs []common.Address, outCh chan<- *PairReserves) error {
	outChTmp := make(chan *types.EthOnBlockResponse)
	callParams := make([]map[string]string, 0)
	for _, pair := range pairs {
		name := fmt.Sprintf("pair_%s", pair.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": pair.Hex(), "data": "0x0902f1ac"})
	}

	visited := make(map[uint64]bool)
	go func() {
		for resp := range outChTmp {
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
	}()
	return bloXrouteClient.SubscribeEthOnBlock(nil, callParams, outChTmp)
}

// Monitor the balances of given addresses on Binance Chain. The balances are returned in BNB, USDT, BUSD.
func SubscribeWalletBalance(bloXrouteClient *client.BloXrouteClient, addresses []common.Address, outCh chan<- *WalletBalance) error {
	outChTmp := make(chan *types.EthOnBlockResponse)
	callParams := make([]map[string]string, 0)
	for _, address := range addresses {
		name := fmt.Sprintf("bnb_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_getBalance", "address": address.Hex(), "data": "0x0"})
		name = fmt.Sprintf("usdt_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": USDT.Hex(), "data": buildBalanceInputData(address)})
		name = fmt.Sprintf("busd_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": BUSD.Hex(), "data": buildBalanceInputData(address)})
	}

	//holder := make(map[common.Address]*WalletBalance)
	go func() {
		for resp := range outChTmp {
			var address common.Address
			balance := &WalletBalance{}
			if strings.HasPrefix(resp.Name, "bnb_") {
				address = common.HexToAddress(resp.Name[len("bnb_"):])
				balance.Bnb, _ = big.NewInt(0).SetString(resp.Response, 0)
			} else if strings.HasPrefix(resp.Name, "usdt_") {
				address = common.HexToAddress(resp.Name[len("usdt_"):])
				balance.Usdt, _ = big.NewInt(0).SetString(resp.Response, 0)
			} else if strings.HasPrefix(resp.Name, "busd_") {
				address = common.HexToAddress(resp.Name[len("busd_"):])
				balance.Busd, _ = big.NewInt(0).SetString(resp.Response, 0)
			}
			balance.Address = address
			blockNumber, ok := big.NewInt(0).SetString(resp.BlockHeight, 0)
			if !ok {
				panic(resp)
			}
			balance.BlockNumber = blockNumber.Int64()
			balance.BlockTimestampLast = time.Now().Unix()
			//holder[address] = balance
			outCh <- balance

		}
	}()
	return bloXrouteClient.SubscribeEthOnBlock(nil, callParams, outChTmp)
}

func buildBalanceInputData(address common.Address) string {
	funcSelector := []byte{0x70, 0xa0, 0x82, 0x31}
	addressBytes := common.LeftPadBytes(address.Bytes(), 32)
	var dataOut []byte
	dataOut = append(dataOut, funcSelector...)
	dataOut = append(dataOut, addressBytes...)
	return fmt.Sprintf("0x%x", dataOut)
}

// Decode the data of ethOnBlock of GetReserves()
func decodeReturnedDataOfGetReserves(pair common.Address, hexStr string, blockNumber int64) (*PairReserves, error) {
	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return nil, err
	}

	pairReserve := &PairReserves{
		Pair:               pair,
		Reserve0:           big.NewInt(0).SetBytes(bytes[0:32]),
		Reserve1:           big.NewInt(0).SetBytes(bytes[32:64]),
		BlockTimestampLast: uint32(big.NewInt(0).SetBytes(bytes[64:]).Int64()),
		BlockNumber:        blockNumber,
	}
	return pairReserve, nil
}
