package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
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

// Configurable thresholds (ms)
var (
	greenThreshold  int64 = 150
	yellowThreshold int64 = 400
)

// getEnvInt returns the env var value as int64, or the default if not set/invalid
func getEnvInt(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return def
}

// getTermWidth returns terminal width, defaulting to 80
func getTermWidth() int {
	type winsize struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	var ws winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	if ws.Col == 0 {
		return 80
	}
	return int(ws.Col)
}

type stats struct {
	count    int
	failures int
	total    time.Duration
	min      time.Duration
	max      time.Duration
	last     time.Duration
	blocks   []string // individual blocks for proper width handling
}

func main() {
	interval := flag.Duration("i", time.Second, "interval between requests")
	timeout := flag.Duration("t", 5*time.Second, "request timeout")
	noLegend := flag.Bool("nolegend", false, "hide the legend line")
	greenFlag := flag.Int64("green", 0, "green threshold in ms (env: HITTYPING_GREEN)")
	yellowFlag := flag.Int64("yellow", 0, "yellow threshold in ms (env: HITTYPING_YELLOW)")
	flag.Parse()

	// Apply thresholds: env vars first, flags override
	greenThreshold = getEnvInt("HITTYPING_GREEN", greenThreshold)
	yellowThreshold = getEnvInt("HITTYPING_YELLOW", yellowThreshold)
	if *greenFlag > 0 {
		greenThreshold = *greenFlag
	}
	if *yellowFlag > 0 {
		yellowThreshold = *yellowFlag
	}

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
		fmt.Printf("%sLegend: %s▁▂▃%s<%dms %s▄▅%s<%dms %s▆▇█%s>=%dms %s×%sfail%s\n",
			gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, gray, reset, reset)
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
			s.blocks = append(s.blocks, gray+"×"+reset)
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
			s.blocks = append(s.blocks, getBlock(rtt))
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

	var idx int
	var color string

	if ms < greenThreshold {
		// Green zone: blocks 0-2 (▁▂▃)
		color = green
		idx = int(ms * 3 / greenThreshold)
		if idx > 2 {
			idx = 2
		}
	} else if ms < yellowThreshold {
		// Yellow zone: blocks 3-4 (▄▅)
		color = yellow
		progress := ms - greenThreshold
		span := yellowThreshold - greenThreshold
		idx = 3 + int(progress*2/span)
		if idx > 4 {
			idx = 4
		}
	} else {
		// Red zone: blocks 5-7 (▆▇█)
		color = red
		// Scale red from yellowThreshold to 2x yellowThreshold
		progress := ms - yellowThreshold
		span := yellowThreshold // red zone spans another yellowThreshold worth
		idx = 5 + int(progress*3/span)
		if idx > 7 {
			idx = 7
		}
	}

	if idx < 0 {
		idx = 0
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

	// Get terminal width and calculate visible blocks
	width := getTermWidth()
	visibleCount := len(s.blocks)
	startIdx := 0
	if visibleCount > width-1 {
		startIdx = visibleCount - (width - 1)
		visibleCount = width - 1
	}

	// Build visible bar
	var bar strings.Builder
	for i := startIdx; i < startIdx+visibleCount; i++ {
		bar.WriteString(s.blocks[i])
	}

	// Bar line with cursor at end
	fmt.Printf("%s%s%s", col0, clearLn, bar.String())
	// Move down, print stats, move back up to end of bar
	fmt.Printf("%s%s%s%d/%s%d%s %s(%2d%%) lost;%s %d/%s%d%s/%d%sms; last:%s %s%d%s%sms%s%s",
		down, col0, clearLn,
		s.failures, bold, total, reset,
		gray, lossPct, reset,
		minMs, bold, avg.Milliseconds(), reset, s.max.Milliseconds(), gray, reset,
		bold, s.last.Milliseconds(), reset, gray, reset,
		up)
	// Move cursor to end of bar (column = visible blocks + 1)
	fmt.Printf("\033[%dG", visibleCount+1)
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
