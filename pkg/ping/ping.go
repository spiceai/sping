package ping

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/montanaflynn/stats"
)

type PingClient struct {
	request     *http.Request
	client      *http.Client
	showContent bool

	durations []float64
}

func NewPingClient(request *http.Request, timeout time.Duration, showContent bool) *PingClient {
	return &PingClient{
		request: request,
		client: &http.Client{
			Timeout: timeout,
		},
		showContent: showContent,
	}
}

func (p *PingClient) Ping() error {
	start := time.Now()
	resp, err := p.client.Do(p.request)
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

	var content string
	contentLength := resp.ContentLength
	if p.showContent {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		content = " " + strings.TrimSpace(strings.SplitN(string(body), "\n", 2)[0])
		contentLength = int64(len(body))
	}

	fmt.Printf("%s (%d bytes) from %s: time=%s%s\n", status, contentLength, aurora.BrightBlue(p.request.Host), duration.Round(time.Microsecond), content)

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
