package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	geth_types "github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/assert"
)

func TestToRaw(t *testing.T) {
	txContents := TxContents{
		Type:     "0x0",
		Nonce:    "0x6",
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
	assert.NoError(t, err)
	assert.NotNil(t, raw)

	tx := new(geth_types.Transaction)
	err = tx.UnmarshalBinary(raw)
	assert.NoError(t, err)
	assert.Equal(t, tx.Hash(), common.HexToHash(txContents.Hash))
}
