package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	var method string
	flag.StringVar(&method, "method", "GET", "method to use for each ping")
	method = strings.ToUpper(method)

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: aping <url> [-interval <duration>] [-timeout <duration>]")
		fmt.Println()
		fmt.Println("Example: aping data.spiceai.io/health -interval 5s -timeout 5s")
		return
	}

	request, err := http.NewRequest(method, flag.Arg(0), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if request.URL.Scheme == "" {
		request, err = http.NewRequest(method, "https://"+flag.Arg(0), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	pingClient := ping.NewPingClient(request, timeout)

	fmt.Printf("SPING %s %s\n", method, aurora.BrightCyan(request.URL.String()))

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
