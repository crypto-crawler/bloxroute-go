package types

// The first response after a subscription is sent.
type SubscriptionResponse struct {
	Id      int64  `json:"id"`
	Result  string `json:"result"` // subscription ID is here
	JsonRPC string `json:"jsonrpc"`
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

type Transaction struct {
	LocalRegion bool        `json:"localRegion,omitempty"`
	TxHash      string      `json:"txHash"`
	TxContents  *TxContents `json:"txContents,omitempty"`
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
