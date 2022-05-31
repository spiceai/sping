package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/spiceai/sping/pkg/ping"
	"github.com/valyala/fasthttp"
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

	if flag.NArg() < 1 {
		fmt.Println("Usage: aping [-interval <duration>] [-timeout <duration>] [-show-content] <url>")
		fmt.Println()
		fmt.Println("Example: aping data.spiceai.io/health -interval 5s -timeout 5s")
		return
	}

	uri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(uri)

	if err := uri.Parse(nil, []byte(flag.Arg(0))); err != nil {
		fmt.Println(err)
		return
	}

	if len(uri.Scheme()) == 0 {
		uri.SetScheme("https")
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetURI(uri)
	req.Header.SetMethod(method)
	req.Header.Set("Accept-Encoding", "gzip")

	pingClient := ping.NewPingClient(req, timeout, showContent)

	fmt.Printf("SPING %s %s\n", method, aurora.BrightCyan(uri.String()))

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
