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

var (
	interval    time.Duration
	timeout     time.Duration
	method      string
	showContent bool
)

func main() {
	flag.Parse()

	method = strings.ToUpper(method)

	fmt.Printf("%s\n", interval)

	if flag.NArg() < 1 {
		fmt.Println("Usage: aping [-interval <duration>] [-timeout <duration>] [-show-content] <url>")
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

	pingClient := ping.NewPingClient(request, timeout, showContent)

	fmt.Printf("SPING %s %s\n", method, aurora.BrightCyan(request.URL.String()))

	timer := time.NewTimer(0)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT)

	go func() {
		for {
			select {
			case <-signalChannel:
				timer.Stop()
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

func init() {
	flag.DurationVar(&interval, "interval", time.Second, "interval between pings")
	flag.DurationVar(&timeout, "timeout", time.Second*5, "timeout for each ping")
	flag.BoolVar(&showContent, "show-content", false, "show response content")
	flag.StringVar(&method, "method", "GET", "method to use for each ping")
}
