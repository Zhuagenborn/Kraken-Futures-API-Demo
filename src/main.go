package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"kf/authkey"
	"kf/book"
	"kf/client"
	"kf/client/api"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	orderChan := make(chan book.Order)
	auth := api.Auth{APIKey: authkey.APIKey, APISecret: authkey.APISecret}
	kraken := client.New(auth, "Kraken-Futures", orderChan)
	kraken.Start("PI_XBTUSD")
	defer kraken.Close()

	for {
		select {
		case <-sigs:
			syscall.Exit(0)
		case order := <-orderChan:
			log.Printf("%+v\n", order)
		}
	}
}
