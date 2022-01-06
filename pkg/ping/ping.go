package ping

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/montanaflynn/stats"
)

type PingClient struct {
	url    *url.URL
	client *http.Client

	durations []float64
}

func NewPingClient(uri *url.URL) *PingClient {
	return &PingClient{
		url:    uri,
		client: &http.Client{},
	}
}

func (p *PingClient) Ping() error {
	start := time.Now()
	resp, err := p.client.Get(p.url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var status aurora.Value
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = aurora.BrightGreen(resp.Status)
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		status = aurora.BrightYellow(resp.Status)
	} else {
		status = aurora.BrightRed(resp.Status)
	}

	duration := time.Since(start)
	p.durations = append(p.durations, float64(duration))

	fmt.Printf("%s (%d bytes) from %s: time=%s\n", status, resp.ContentLength, p.url.Host, duration)

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
