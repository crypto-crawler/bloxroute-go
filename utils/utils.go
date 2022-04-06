// High-level APIs.
package utils

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/crypto-crawler/bloxroute-go/client"
	"github.com/crypto-crawler/bloxroute-go/types"
	"github.com/ethereum/go-ethereum/common"
)

type WalletBalance struct {
	Address            common.Address `json:"address"`
	Bnb                *big.Int       `json:"bnb"`
	Wbnb               *big.Int       `json:"wbnb"`
	Usdt               *big.Int       `json:"usdt"`
	Busd               *big.Int       `json:"busd"`
	BlockTimestampLast int64          `json:"block_timestamp_last"`
	BlockNumber        int64          `json:"block_number"`
}

var (
	WBNB = common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
	USDT = common.HexToAddress("0x55d398326f99059fF775485246999027B3197955")
	BUSD = common.HexToAddress("0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56")
)

// Monitor the balances of given addresses on Binance Chain. The balances are returned in BNB, WBNB, USDT, BUSD.
func SubscribeWalletBalance(bloXrouteClient *client.BloXrouteClient, addresses []common.Address, outCh chan<- *WalletBalance) error {
	outChTmp := make(chan *types.EthOnBlockResponse)
	callParams := make([]map[string]string, 0)
	for _, address := range addresses {
		name := fmt.Sprintf("bnb_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_getBalance", "address": address.Hex(), "data": "0x0"})
		name = fmt.Sprintf("wbnb_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": WBNB.Hex(), "data": buildBalanceInputData(address)})
		name = fmt.Sprintf("usdt_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": USDT.Hex(), "data": buildBalanceInputData(address)})
		name = fmt.Sprintf("busd_%s", address.Hex())
		callParams = append(callParams, map[string]string{"name": name, "method": "eth_call", "to": BUSD.Hex(), "data": buildBalanceInputData(address)})
	}

	// holder := make(map[common.Address]*WalletBalance)
	go func() {
		for resp := range outChTmp {
			var address common.Address
			balance := &WalletBalance{}
			if strings.HasPrefix(resp.Name, "bnb_") {
				address = common.HexToAddress(resp.Name[len("bnb_"):])
				balance.Bnb, _ = big.NewInt(0).SetString(resp.Response, 0)
			} else if strings.HasPrefix(resp.Name, "wbnb_") {
				address = common.HexToAddress(resp.Name[len("wbnb_"):])
				balance.Wbnb, _ = big.NewInt(0).SetString(resp.Response, 0)
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
			// holder[address] = balance
			outCh <- balance

		}
	}()
	return bloXrouteClient.SubscribeEthOnBlock(nil, callParams, outChTmp)
}

func buildBalanceInputData(address common.Address) string {
	// balanceOf(address), see // see https://www.4byte.directory/signatures/?bytes4_signature=0x70a08231
	funcSelector := []byte{0x70, 0xa0, 0x82, 0x31}
	addressBytes := common.LeftPadBytes(address.Bytes(), 32)
	var dataOut []byte
	dataOut = append(dataOut, funcSelector...)
	dataOut = append(dataOut, addressBytes...)
	return fmt.Sprintf("0x%x", dataOut)
}
