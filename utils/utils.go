package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/big"

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
