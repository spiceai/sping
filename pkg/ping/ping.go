package ping

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/montanaflynn/stats"
	"github.com/valyala/fasthttp"
)

type PingClient struct {
	request     *fasthttp.Request
	client      *fasthttp.Client
	showContent bool

	durations []float64
}

func NewPingClient(request *fasthttp.Request, timeout time.Duration, showContent bool) *PingClient {
	return &PingClient{
		request:     request,
		client:      &fasthttp.Client{},
		showContent: showContent,
	}
}

func (p *PingClient) Ping() error {
	start := time.Now()

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := p.client.Do(p.request, resp); err != nil {
		return err
	}

	statusCode := resp.StatusCode()
	statusText := http.StatusText(statusCode)

	var status aurora.Value
	if statusCode >= 200 && statusCode < 300 {
		status = aurora.BrightGreen(statusText)
	} else if statusCode >= 400 && statusCode < 500 {
		status = aurora.BrightYellow(statusText)
	} else {
		status = aurora.BrightRed(statusText)
	}

	duration := time.Since(start)
	p.durations = append(p.durations, float64(duration))

	var content string
	contentLength := resp.Header.ContentLength()
	if p.showContent {
		contentEncoding := resp.Header.Peek("Content-Encoding")
		var body []byte
		var err error
		if bytes.EqualFold(contentEncoding, []byte("gzip")) {
			body, err = resp.BodyGunzip()
			if err != nil {
				return fmt.Errorf("failed to gunzip response body: %w", err)
			}
		} else {
			body = resp.Body()
		}
		content = " " + strings.TrimSpace(strings.SplitN(string(body), "\n", 2)[0])
		contentLength = len(body)
	}

	fmt.Printf("%s (%d bytes) from %s: time=%s%s\n", status, contentLength, aurora.BrightBlue(string(p.request.Host())), duration.Round(time.Microsecond), content)

	return nil
}

func (p *PingClient) PrintStats() error {
	sort.SliceStable(p.durations, func(i, j int) bool {
		return p.durations[i] < p.durations[j]
	})

	mean, err := stats.Mean(p.durations)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("count=%d\n", len(p.durations))
	fmt.Printf("min=%s\n", time.Duration(p.durations[0]))
	fmt.Printf("max=%s\n", time.Duration(p.durations[len(p.durations)-1]))
	fmt.Printf("avg=%s\n", time.Duration(mean))

	return nil
}
