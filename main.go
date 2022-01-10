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
	var interval time.Duration
	flag.DurationVar(&interval, "interval", time.Second, "interval between pings")

	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", time.Second*5, "timeout for each ping")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: aping <url> [-interval <duration>] [-timeout <duration>]")
		fmt.Println()
		fmt.Println("Example: aping data.spiceai.io/health -interval 5s -timeout 5s")
		return
	}

	url, err := url.Parse(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	if url.Scheme == "" {
		url.Scheme = "https"
	}

	pingClient := ping.NewPingClient(url, timeout)

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
				timer.Reset(interval)
			}
		}
	}()

	<-signalChannel

	pingClient.PrintStats()

	fmt.Println()
}
