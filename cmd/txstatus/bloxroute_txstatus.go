package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crypto-crawler/bloxroute-go/client"
	"github.com/crypto-crawler/bloxroute-go/types"
)

// doc: https://docs.bloxroute.com/streams/txstatus

// Subscribe to the `transactionStatus` stream from bloXroute cloud API.
func main() {
	certFile := flag.String("cert", "external_gateway_cert.pem", "The cert file")
	keyFile := flag.String("key", "external_gateway_key.pem", "The key file")
	gatewayUrl := flag.String("gateway", "", "The gateway url")
	header := flag.String("header", "", "The authorization header")
	flag.Parse()
	if *gatewayUrl == "" || *certFile == "" || *keyFile == "" || *header == "" {
		flag.Usage()
		return
	}

	// catch Ctrl+C
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	stopCh := make(chan struct{})

	pendingTxCh := make(chan *types.Transaction)
	{
		log.Println("Connecting to bloXroute gateway")
		bloXrouteClient, err := client.NewBloXrouteClientToGateway(*gatewayUrl, *header, stopCh)
		if err != nil {
			log.Fatal(err)
		}
		err = bloXrouteClient.SubscribeNewTxs([]string{"tx_hash", "raw_tx"}, "", pendingTxCh)
		if err != nil {
			log.Fatal(err)
		}
	}

	statusCh := make(chan *types.TxStatus)
	log.Println("Connecting to bloXroute cloud")
	transactionStatusClient, err := client.NewTransactionStatusClient(*header, stopCh, statusCh)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for txStatus := range statusCh {
			bytes, _ := json.Marshal(txStatus)
			fmt.Println(string(bytes))
		}
	}()

	for {
		select {
		case <-signals:
			log.Println("Ctrl+C detected, exiting...")
			close(stopCh)
			time.Sleep(1 * time.Second) // give some time for other goroutines to stop
			close(statusCh)
			return
		case txJson := <-pendingTxCh:
			txRaw, err := hex.DecodeString(txJson.RawTx[2:])
			if err == nil {
				err = transactionStatusClient.StartMonitorTransaction([][]byte{txRaw}, false)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
