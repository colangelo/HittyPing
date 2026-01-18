package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	// ANSI colors
	green   = "\033[32m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	gray    = "\033[90m"
	bold    = "\033[1m"
	reset   = "\033[0m"
	clearLn = "\033[K"
	up      = "\033[A"
	down    = "\033[B"
	col0    = "\033[0G"
)

// Unicode block characters for visualization
var blocks = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

type stats struct {
	count    int
	failures int
	total    time.Duration
	min      time.Duration
	max      time.Duration
	last     time.Duration
	bar      strings.Builder
}

func main() {
	interval := flag.Duration("i", time.Second, "interval between requests")
	timeout := flag.Duration("t", 5*time.Second, "request timeout")
	noLegend := flag.Bool("nolegend", false, "hide the legend line")
	flag.Parse()

	url := "https://1.1.1.1"
	if flag.NArg() > 0 {
		url = flag.Arg(0)
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
	}

	// Display URL without scheme
	displayURL := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")

	s := &stats{min: time.Hour}

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		printFinal(displayURL, s)
		os.Exit(0)
	}()

	// Print header
	fmt.Printf("%sHITTYPING %s%s\n", gray, displayURL, reset)
	if !*noLegend {
		fmt.Printf("%sLegend: %s▁▂▃%s<150ms %s▄▅%s<400ms %s▆▇█%s>400ms %s×%sfail%s\n",
			gray, green, reset, yellow, reset, red, reset, gray, reset, reset)
	}
	fmt.Println() // Reserve stats line
	fmt.Print(up) // Move back to bar line

	client := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			DisableKeepAlives: true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		rtt, err := measureRTT(client, url)
		if err != nil {
			s.failures++
			s.bar.WriteString(gray + "×" + reset)
		} else {
			s.count++
			s.total += rtt
			s.last = rtt
			if rtt < s.min {
				s.min = rtt
			}
			if rtt > s.max {
				s.max = rtt
			}
			s.bar.WriteString(getBlock(rtt))
		}
		printDisplay(s)
		time.Sleep(*interval)
	}
}

func measureRTT(client *http.Client, url string) (time.Duration, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	return elapsed, nil
}

func getBlock(rtt time.Duration) string {
	ms := rtt.Milliseconds()

	// Map RTT to block index (20-800ms range)
	const minRTT, maxRTT = 20, 800
	var idx int
	if ms < minRTT {
		idx = 0
	} else if ms > maxRTT {
		idx = 7
	} else {
		idx = int((ms - minRTT) * 7 / (maxRTT - minRTT))
	}

	// Color based on latency
	var color string
	if ms < 150 {
		color = green
	} else if ms < 400 {
		color = yellow
	} else {
		color = red
	}

	return color + blocks[idx] + reset
}

func printDisplay(s *stats) {
	total := s.count + s.failures
	var lossPct int
	if total > 0 {
		lossPct = s.failures * 100 / total
	}

	var avg time.Duration
	if s.count > 0 {
		avg = s.total / time.Duration(s.count)
	}

	minMs := s.min.Milliseconds()
	if s.min == time.Hour {
		minMs = 0
	}

	// Bar line
	fmt.Printf("%s%s%s", col0, clearLn, s.bar.String())
	// Move down, print stats, move back up
	fmt.Printf("%s%s%s%d/%s%d%s %s(%2d%%) lost;%s %d/%s%d%s/%d%sms; last:%s %s%d%s%sms%s%s",
		down, col0, clearLn,
		s.failures, bold, total, reset,
		gray, lossPct, reset,
		minMs, bold, avg.Milliseconds(), reset, s.max.Milliseconds(), gray, reset,
		bold, s.last.Milliseconds(), reset, gray, reset,
		up)
}

func printFinal(url string, s *stats) {
	total := s.count + s.failures
	var lossPct int
	if total > 0 {
		lossPct = s.failures * 100 / total
	}

	var avg time.Duration
	if s.count > 0 {
		avg = s.total / time.Duration(s.count)
	}

	minMs := s.min.Milliseconds()
	if s.min == time.Hour {
		minMs = 0
	}

	fmt.Printf("\n\n%s--- %s hittyping statistics ---%s\n", gray, url, reset)
	fmt.Printf("%d requests, %d ok, %d failed, %d%% loss\n", total, s.count, s.failures, lossPct)
	if s.count > 0 {
		fmt.Printf("round-trip min/avg/max = %d/%d/%d ms\n", minMs, avg.Milliseconds(), s.max.Milliseconds())
	}
}
