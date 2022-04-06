package types

import (
	"crypto/md5"
	"math/big"

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

// Token balance of a owner at given block number.
type TokenBalance struct {
	Owner common.Address `json:"owner"`
	// Zero address means ETH/BNB
	Token       common.Address `json:"token"`
	Balance     *big.Int
	BlockNumber int64 `json:"block_number"`
}
