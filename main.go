package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	flag "github.com/spf13/pflag"
)

const version = "0.5.0"

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
	saveCur = "\033[s"
	restCur = "\033[u"
)

// Unicode block characters for visualization
var blocks = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// Configurable thresholds (ms)
var (
	minLatency      int64 = 0   // baseline for smallest block
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
	col      int      // current column position on bar line
}

func main() {
	interval := flag.DurationP("interval", "i", time.Second, "interval between requests")
	timeout := flag.DurationP("timeout", "t", 5*time.Second, "request timeout")
	noLegend := flag.BoolP("nolegend", "n", false, "hide the legend line")
	minFlag := flag.Int64P("min", "m", 0, "min latency baseline in ms (env: HP_MIN)")
	greenFlag := flag.Int64P("green", "g", 0, "green threshold in ms (env: HP_GREEN)")
	yellowFlag := flag.Int64P("yellow", "y", 0, "yellow threshold in ms (env: HP_YELLOW)")
	insecure := flag.BoolP("insecure", "k", false, "skip TLS certificate verification")
	useHTTP := flag.Bool("http", false, "use plain HTTP instead of HTTPS")
	useHTTP3 := flag.Bool("http3", false, "use HTTP/3 (QUIC) - requires build with -tags http3")
	showVersion := flag.BoolP("version", "v", false, "show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("hp (hittyping) version %s\n", version)
		os.Exit(0)
	}

	// Apply thresholds: env vars first, flags override
	minLatency = getEnvInt("HP_MIN", minLatency)
	greenThreshold = getEnvInt("HP_GREEN", greenThreshold)
	yellowThreshold = getEnvInt("HP_YELLOW", yellowThreshold)
	if *minFlag > 0 {
		minLatency = *minFlag
	}
	if *greenFlag > 0 {
		greenThreshold = *greenFlag
	}
	if *yellowFlag > 0 {
		yellowThreshold = *yellowFlag
	}

	url := "https://1.1.1.1"
	if flag.NArg() > 0 {
		url = flag.Arg(0)
		if *useHTTP {
			// Force HTTP
			url = strings.TrimPrefix(url, "https://")
			url = strings.TrimPrefix(url, "http://")
			url = "http://" + url
		} else if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
	} else if *useHTTP {
		url = "http://1.1.1.1"
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

	// Check HTTP/3 availability
	if *useHTTP3 {
		if !http3Available {
			fmt.Fprintln(os.Stderr, "HTTP/3 not compiled in. Rebuild with: go build -tags http3")
			os.Exit(1)
		}
		if *useHTTP {
			fmt.Fprintln(os.Stderr, "Cannot use --http with --http3")
			os.Exit(1)
		}
	}

	// Determine protocol for display
	protocol := "HTTPS"
	if *useHTTP {
		protocol = "HTTP"
	} else if *useHTTP3 {
		protocol = "HTTP/3"
	}

	// Print header
	fmt.Printf("%sHP %s %s(%s)%s\n", gray, displayURL, gray, protocol, reset)
	if !*noLegend {
		fmt.Printf("%sLegend: %s▁▂▃%s<%dms %s▄▅%s<%dms %s▆▇█%s>=%dms %s%s!%sfail%s\n",
			gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, red, bold, reset, reset)
	}
	fmt.Println() // Reserve stats line
	fmt.Print(up) // Move back to bar line

	// Create HTTP client based on protocol choice
	var client *http.Client
	if *useHTTP3 {
		client = newHTTP3Client(*timeout, *insecure)
	} else {
		client = &http.Client{
			Timeout: *timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: *insecure,
				},
				DisableKeepAlives: true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	for {
		rtt, err := measureRTT(client, url)
		if err != nil {
			s.failures++
			s.blocks = append(s.blocks, red+bold+"!"+reset)
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
		// Green zone: blocks 0-2 (▁▂▃), scaled from minLatency
		color = green
		if ms <= minLatency {
			idx = 0
		} else {
			progress := ms - minLatency
			span := greenThreshold - minLatency
			if span > 0 {
				idx = int(progress * 3 / span)
			}
			if idx > 2 {
				idx = 2
			}
		}
	} else if ms < yellowThreshold {
		// Yellow zone: blocks 3-4 (▄▅)
		color = yellow
		progress := ms - greenThreshold
		span := yellowThreshold - greenThreshold
		if span > 0 {
			idx = 3 + int(progress*2/span)
		} else {
			idx = 3
		}
		if idx > 4 {
			idx = 4
		}
	} else {
		// Red zone: blocks 5-7 (▆▇█)
		color = red
		// Scale red from yellowThreshold to 2x yellowThreshold
		progress := ms - yellowThreshold
		span := yellowThreshold // red zone spans another yellowThreshold worth
		if span > 0 {
			idx = 5 + int(progress*3/span)
		} else {
			idx = 5
		}
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

	width := getTermWidth()

	// Print just the latest block (incremental)
	if len(s.blocks) > 0 {
		fmt.Print(s.blocks[len(s.blocks)-1])
		s.col++
	}

	// Check if we need to wrap to next line for the NEXT block
	if s.col >= width-1 {
		// Move to stats line, print newline to scroll, move back up, clear line
		fmt.Print(down + "\n" + up + col0 + clearLn)
		s.col = 0
	}

	// Save cursor, print stats below, restore cursor
	fmt.Printf("%s%s%s%s%d/%s%d%s %s(%2d%%) lost;%s %d/%s%d%s/%d%sms; last:%s %s%d%s%sms%s%s",
		saveCur, down, col0, clearLn,
		s.failures, bold, total, reset,
		gray, lossPct, reset,
		minMs, bold, avg.Milliseconds(), reset, s.max.Milliseconds(), gray, reset,
		bold, s.last.Milliseconds(), reset, gray, reset,
		restCur)
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

	fmt.Printf("\n\n%s--- %s hp statistics ---%s\n", gray, url, reset)
	fmt.Printf("%d requests, %d ok, %d failed, %d%% loss\n", total, s.count, s.failures, lossPct)
	if s.count > 0 {
		fmt.Printf("round-trip min/avg/max = %d/%d/%d ms\n", minMs, avg.Milliseconds(), s.max.Milliseconds())
	}
}
