package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/spiceai/sping/pkg/ping"
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: aping <url>")
		return
	}

	url, err := url.Parse(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	if url.Scheme == "" {
		url.Scheme = "https"
	}

	pingClient := ping.NewPingClient(url)

	fmt.Printf("SPING %s\n", aurora.BrightCyan(url.String()))

	timer := time.NewTimer(0)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT)

	go func() {
		for {
			select {
			case <-signalChannel:
				return
			case <-timer.C:
				pingClient.Ping()
				timer.Reset(time.Second)
			}
		}
	}()

	<-signalChannel

	pingClient.PrintStats()

	fmt.Println()
}
