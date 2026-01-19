package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

const version = "0.7.4"

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

// Protocol levels for downgrade feature
const (
	protoHTTP1  = 0 // Plain HTTP/1.1 (insecure)
	protoHTTPS  = 1 // HTTPS (auto-negotiate)
	protoHTTP2  = 2 // HTTP/2 (forced)
	protoHTTP3  = 3 // HTTP/3 (QUIC)
)

var protoNames = map[int]string{
	protoHTTP1: "HTTP/1.1",
	protoHTTPS: "HTTPS",
	protoHTTP2: "HTTP/2",
	protoHTTP3: "HTTP/3",
}

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
	count := flag.IntP("count", "c", 0, "number of requests (0 = unlimited)")
	noLegend := flag.BoolP("nolegend", "q", false, "hide the legend line")
	minFlag := flag.Int64P("min", "m", 0, "min latency baseline in ms (env: HP_MIN)")
	greenFlag := flag.Int64P("green", "g", 0, "green threshold in ms (env: HP_GREEN)")
	yellowFlag := flag.Int64P("yellow", "y", 0, "yellow threshold in ms (env: HP_YELLOW)")
	insecure := flag.BoolP("insecure", "k", false, "skip TLS certificate verification")
	useHTTP1 := flag.BoolP("http", "1", false, "use plain HTTP/1.1")
	useHTTP2 := flag.BoolP("http2", "2", false, "force HTTP/2 (fail if not negotiated)")
	useHTTP3 := flag.BoolP("http3", "3", false, "use HTTP/3 (QUIC) - requires build with -tags http3")
	downgrade := flag.BoolP("downgrade", "d", false, "auto-downgrade protocol on failures (secure only)")
	downgradeInsecure := flag.BoolP("downgrade-insecure", "D", false, "auto-downgrade including plain HTTP")
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

	// Extract host (without scheme) for building URLs dynamically
	host := "1.1.1.1"
	if flag.NArg() > 0 {
		host = flag.Arg(0)
		// Strip any existing scheme
		host = strings.TrimPrefix(host, "https://")
		host = strings.TrimPrefix(host, "http://")

		// Check if it's an IPv6 address that needs brackets
		if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
			// It's an IPv6 address, wrap in brackets
			host = "[" + host + "]"
		}
	}

	displayURL := host

	// Resolve hostname to IP for display (and validate it exists)
	resolvedIP := ""
	// Strip brackets from IPv6 for parsing and display
	hostForLookup := strings.TrimPrefix(strings.TrimSuffix(displayURL, "]"), "[")
	// Check if it's already an IP address
	if ip := net.ParseIP(hostForLookup); ip == nil {
		// It's a hostname, resolve it
		ips, err := net.LookupHost(hostForLookup)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot resolve %s: %v\n", hostForLookup, err)
			os.Exit(1)
		}
		if len(ips) > 0 {
			resolvedIP = ips[0]
		}
	}

	s := &stats{min: time.Hour}

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		printFinal(displayURL, s)
		os.Exit(0)
	}()

	// Validate protocol flags (mutually exclusive)
	protoCount := 0
	if *useHTTP1 {
		protoCount++
	}
	if *useHTTP2 {
		protoCount++
	}
	if *useHTTP3 {
		protoCount++
	}
	if protoCount > 1 {
		fmt.Fprintln(os.Stderr, "Cannot combine -1/--http, -2/--http2, and -3/--http3")
		os.Exit(1)
	}

	// Check HTTP/3 availability
	if *useHTTP3 && !http3Available {
		fmt.Fprintln(os.Stderr, "HTTP/3 not compiled in. Rebuild with: go build -tags http3")
		os.Exit(1)
	}

	// Determine initial protocol level
	currentProto := protoHTTPS
	if *useHTTP1 {
		currentProto = protoHTTP1
	} else if *useHTTP2 {
		currentProto = protoHTTP2
	} else if *useHTTP3 {
		currentProto = protoHTTP3
	}

	// Determine minimum protocol level for downgrade
	minProto := protoHTTPS // secure only by default
	if *downgradeInsecure {
		minProto = protoHTTP1
	}
	canDowngrade := *downgrade || *downgradeInsecure

	// Print header
	printHeader := func() {
		// Move to beginning of line and clear
		fmt.Print(col0 + clearLn)
		if resolvedIP != "" {
			fmt.Printf("%sHittyPing (v%s) %s [%s] (%s)%s\n", gray, version, displayURL, resolvedIP, protoNames[currentProto], reset)
		} else {
			fmt.Printf("%sHittyPing (v%s) %s (%s)%s\n", gray, version, displayURL, protoNames[currentProto], reset)
		}
	}
	printHeader()
	if !*noLegend {
		fmt.Printf("%sLegend: %s▁▂▃%s<%dms %s▄▅%s<%dms %s▆▇█%s>=%dms %s%s!%sfail%s\n",
			gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, red, bold, reset, reset)
	}
	fmt.Println() // Reserve stats line
	fmt.Print(up) // Move back to bar line

	// Create HTTP client
	url := getURLForProto(host, currentProto)
	client := createClient(currentProto, *timeout, *insecure)

	consecutiveFailures := 0
	requestNum := 0
	for {
		rtt, err := measureRTT(client, url, currentProto)
		if err != nil {
			s.failures++
			consecutiveFailures++
			s.blocks = append(s.blocks, red+bold+"!"+reset)

			// Check for downgrade (only at startup, before first successful ping)
			if canDowngrade && consecutiveFailures >= 3 && currentProto > minProto && s.count == 0 {
				// Find a working lower protocol by testing each one
				candidateProto := currentProto
				foundWorking := false
				for candidateProto > minProto {
					// Try next lower protocol
					switch candidateProto {
					case protoHTTP3:
						candidateProto = protoHTTP2
					case protoHTTP2:
						candidateProto = protoHTTPS
					case protoHTTPS:
						candidateProto = protoHTTP1
					}

					// Test this protocol silently
					testURL := getURLForProto(host, candidateProto)
					testClient := createClient(candidateProto, *timeout, *insecure)
					_, testErr := measureRTT(testClient, testURL, candidateProto)
					if testErr == nil {
						// Found working protocol
						currentProto = candidateProto
						url = testURL
						client = testClient
						foundWorking = true
						break
					}
					// Otherwise continue to even lower protocol
				}

				if foundWorking {
					consecutiveFailures = 0

					// Print downgrade message and update header
					printDisplay(s)
					fmt.Printf("\n%s↓ Downgrading to %s after 3 failures%s\n", yellow, protoNames[currentProto], reset)
					printHeader()
					fmt.Println() // Reserve stats line
					fmt.Print(up) // Move back to bar line
				}
			}
		} else {
			s.count++
			consecutiveFailures = 0 // Reset on success
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
		requestNum++
		if *count > 0 && requestNum >= *count {
			printFinal(displayURL, s)
			os.Exit(0)
		}
		time.Sleep(*interval)
	}
}

func measureRTT(client *http.Client, url string, protoLevel int) (time.Duration, error) {
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

	// Check HTTP/2 requirement
	if protoLevel == protoHTTP2 && resp.Proto != "HTTP/2.0" {
		return 0, fmt.Errorf("HTTP/2 not negotiated (got %s)", resp.Proto)
	}

	return elapsed, nil
}

func createClient(protoLevel int, timeout time.Duration, insecure bool) *http.Client {
	if protoLevel == protoHTTP3 {
		return newHTTP3Client(timeout, insecure)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
		DisableKeepAlives: true,
	}
	if protoLevel == protoHTTP2 {
		transport.ForceAttemptHTTP2 = true
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func getURLForProto(host string, protoLevel int) string {
	if protoLevel == protoHTTP1 {
		return "http://" + host
	}
	return "https://" + host
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
