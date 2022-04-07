package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// The first response after a subscription is sent.
type SubscriptionResponse struct {
	Id      int64  `json:"id"`
	Result  string `json:"result"` // subscription ID is here
	JsonRPC string `json:"jsonrpc"`
}


// the response for sending transaction
type SendTxResponse struct {
	Id      int64  `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  *struct {
		TxHash string `json:"txHash"`
	} `json:"result,omitempty"` // subscription ID is here
}

// bloXroute websocket message
type WebsocketMsg[T any] struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result       T      `json:"result"`
		Subscription string `json:"subscription"`
	} `json:"params"`
	Error map[string]interface{} `json:"error,omitempty"`
}

// bloXroute tx
type TxContents struct {
	From                 string `json:"from"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	Hash                 string `json:"hash"`
	Input                string `json:"input"`
	Nonce                string `json:"nonce"`
	R                    string `json:"r"`
	S                    string `json:"s"`
	To                   string `json:"to"`
	Type                 string `json:"type"`
	V                    string `json:"v"`
	Value                string `json:"value"`
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
}

// ToRaw converts a TxContents to its RLP encoded byte array.
func (txContents *TxContents) ToRaw() ([]byte, error) {
	if txContents.Type != "0x0" {
		// TODO: for now this library only supports legacy transactions
		return nil, fmt.Errorf("tx type %s not supported", txContents.Type)
	}

	nonce, err := strconv.ParseUint(txContents.Nonce[2:], 16, 64)
	if err != nil {
		return nil, err
	}
	to := common.HexToAddress(txContents.To)
	amount, ok := big.NewInt(0).SetString(txContents.Value, 0)
	if !ok {
		return nil, errors.New("invalid amount")
	}
	gasLimit, err := strconv.ParseUint(txContents.Gas[2:], 16, 64)
	if err != nil {
		return nil, err
	}
	gasPrice, ok := big.NewInt(0).SetString(txContents.GasPrice, 0)
	if !ok {
		return nil, errors.New("invalid gas price")
	}
	data, err := hex.DecodeString(txContents.Input[2:])
	if err != nil {
		return nil, err
	}

	r, ok := big.NewInt(0).SetString(txContents.R, 0)
	if !ok {
		return nil, errors.New("invalid R")
	}
	s, ok := big.NewInt(0).SetString(txContents.S, 0)
	if !ok {
		return nil, errors.New("invalid S")
	}
	v, ok := big.NewInt(0).SetString(txContents.V, 0)
	if !ok {
		return nil, errors.New("invalid V")
	}

	// Legacy Transactions are RLP(Nonce, GasPrice, Gas, To, Value, Input, V, R, S)
	values := []interface{}{nonce, gasPrice, gasLimit, to, amount, data, v, r, s}
	buf := new(bytes.Buffer)
	err = rlp.Encode(buf, values)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type Transaction struct {
	LocalRegion bool        `json:"localRegion,omitempty"`
	TxHash      string      `json:"txHash"`
	TxContents  *TxContents `json:"txContents,omitempty"`
	RawTx       string      `json:"rawTx,omitempty"`
}

// bloXroute tx message
type TransactionMsg = WebsocketMsg[Transaction]

// bloXroute block
type Block struct {
	Hash   string `json:"hash,omitempty"`
	Header *struct {
		ParentHash       string `json:"parentHash"`
		Sha3Uncles       string `json:"sha3Uncles"`
		Miner            string `json:"miner"`
		StateRoot        string `json:"stateRoot"`
		TransactionsRoot string `json:"transactionsRoot"`
		ReceiptsRoot     string `json:"receiptsRoot"`
		LogsBloom        string `json:"logsBloom"`
		Difficulty       string `json:"difficulty"`
		Number           string `json:"number"`
		GasLimit         string `json:"gasLimit"`
		GasUsed          string `json:"gasUsed"`
		Timestamp        string `json:"timestamp"`
		ExtraData        string `json:"extraData"`
		MixHash          string `json:"mixHash"`
		Nonce            string `json:"nonce"`
	} `json:"header,omitempty"`
	Transactions []TxContents  `json:"transactions,omitempty"`
	Uncles       []interface{} `json:"uncles,omitempty"`
}

type TxReceipt struct {
	BlockHash         string `json:"block_hash"`
	BlockNumber       string `json:"block_number"`
	CumulativeGasUsed string `json:"cumulative_gas_used"`
	ContractAddress   string `json:"contract_address,omitempty"`
	From              string `json:"from"`
	GasUsed           string `json:"gas_used"`
	Logs              []struct {
		Address          string   `json:"address"`
		BlockNumber      string   `json:"blockNumber"`
		BlockHash        string   `json:"blockHash"`
		Data             string   `json:"data"`
		LogIndex         string   `json:"logIndex"`
		Removed          bool     `json:"removed"`
		Topics           []string `json:"topics,omitempty"`
		TransactionHash  string   `json:"transactionHash"`
		TransactionIndex string   `json:"transactionIndex"`
	} `json:"logs,omitempty"`
	LogsBloom        string `json:"logs_bloom"`
	Status           string `json:"status"`
	To               string `json:"to,omitempty"`
	TransactionHash  string `json:"transaction_hash"`
	TransactionIndex string `json:"transaction_index"`
	Type_            string `json:"type,omitempty"`
}

// Transaction status
type TxStatus struct {
	TxHash string `json:"txHash"`
	Status string `json:"status"`
}

type EthOnBlockResponse struct {
	Name        string `json:"name"`
	Response    string `json:"response"`
	BlockHeight string `json:"block_height,omitempty"`
	Tag         string `json:"tag,omitempty"`
}
