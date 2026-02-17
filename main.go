package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	flag "github.com/spf13/pflag"
)

// displayMu serializes terminal output so the suspend handler can
// cleanly pause rendering before the process is actually stopped.
var displayMu sync.Mutex

const version = "0.8.2"

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
	hideCur    = "\033[?25l"
	showCur    = "\033[?25h"
	steadyCur  = "\033[2 q" // DECSCUSR: steady block cursor
	defaultCur = "\033[0 q" // DECSCUSR: reset to terminal default
)

// Unicode block characters for visualization
var blocks = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// Braille dot patterns for left column (dots 1,2,3,7) - 5 height levels (0-4)
var brailleLeft = []rune{0x00, 0x40, 0x44, 0x46, 0x47}

// Braille dot patterns for right column (dots 4,5,6,8) - 5 height levels (0-4)
var brailleRight = []rune{0x00, 0x80, 0xA0, 0xB0, 0xB8}

// Protocol levels for downgrade feature
const (
	protoHTTP1 = 0 // Plain HTTP/1.1 (insecure)
	protoHTTPS = 1 // HTTPS (auto-negotiate)
	protoHTTP2 = 2 // HTTP/2 (forced)
	protoHTTP3 = 3 // HTTP/3 (QUIC)
)

var protoNames = map[int]string{
	protoHTTP1: "HTTP/1.1",
	protoHTTPS: "HTTPS",
	protoHTTP2: "HTTP/2",
	protoHTTP3: "HTTP/3",
}

// Configurable thresholds (ms)
var (
	minLatency      int64 = 0 // baseline for smallest block
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
	count        int
	failures     int
	total        time.Duration
	min          time.Duration
	max          time.Duration
	last         time.Duration
	blocks       []string      // individual blocks for proper width handling
	col          int           // current column position on bar line
	lastPrinted  int           // last block index printed
	braille      bool          // braille mode enabled
	pendingRTT   time.Duration // pending RTT for braille pairing (-1 = failure, 0 = none)
	hasPending   bool          // whether there's a pending RTT
}

func main() {
	// Silence quic-go UDP buffer warnings that corrupt terminal display
	log.SetOutput(io.Discard)

	interval := flag.DurationP("interval", "i", time.Second, "interval between requests")
	jitter := flag.DurationP("jitter", "j", 0, "max random jitter to add to interval (e.g., 200ms, 3s)")
	timeout := flag.DurationP("timeout", "t", 5*time.Second, "request timeout")
	count := flag.IntP("count", "c", 0, "number of requests (0 = unlimited)")
	showLegend := flag.Bool("legend", false, "show the legend line")
	noHeader := flag.Bool("noheader", false, "hide the header line")
	useBraille := flag.BoolP("braille", "b", false, "use braille visualization (2x density)")
	quiet := flag.BoolP("quiet", "q", false, "hide header and legend")
	silent := flag.BoolP("silent", "Q", false, "hide header, legend, and final stats")
	minFlag := flag.Int64P("min", "m", 0, "min latency baseline in ms (env: HP_MIN)")
	greenFlag := flag.Int64P("green", "g", 0, "green threshold in ms (env: HP_GREEN)")
	yellowFlag := flag.Int64P("yellow", "y", 0, "yellow threshold in ms (env: HP_YELLOW)")
	insecure := flag.BoolP("insecure", "k", false, "skip TLS certificate verification")
	useHTTP1 := flag.BoolP("http", "1", false, "use plain HTTP/1.1")
	useHTTP2 := flag.BoolP("http2", "2", false, "force HTTP/2 (fail if not negotiated)")
	useHTTP3 := flag.BoolP("http3", "3", false, "use HTTP/3 (QUIC)")
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

	s := &stats{min: time.Hour, braille: *useBraille}

	// Disable terminal input processing to prevent keypresses from corrupting
	// the display (echo, VDISCARD, VREPRINT, etc.).
	restoreInput := disableInputProcessing()
	fmt.Print(steadyCur)
	cleanup := func() {
		fmt.Print(defaultCur)
		restoreInput()
	}
	setup := func() {
		restoreInput = disableInputProcessing()
		fmt.Print(steadyCur)
	}

	// Handle Ctrl-Z (suspend) and fg (resume)
	redraw := func() {
		fmt.Println() // reserve stats line
		fmt.Print(up) // move back to bar line
		redrawDisplay(s)
	}
	handleSuspendResume(cleanup, setup, redraw)

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		if !*silent {
			printFinal(displayURL, s)
		}
		cleanup()
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
			fmt.Printf("%sHittyPing (v%s) %s%s%s [%s%s%s] (%s)%s\n", gray, version, reset+bold, displayURL, reset+gray, reset, resolvedIP, gray, protoNames[currentProto], reset)
		} else {
			fmt.Printf("%sHittyPing (v%s) %s%s %s(%s)%s\n", gray, version, reset+bold, displayURL, reset+gray, protoNames[currentProto], reset)
		}
	}
	if !*noHeader && !*quiet && !*silent {
		printHeader()
	}
	if *showLegend && !*quiet && !*silent {
		if *useBraille {
			fmt.Printf("%sLegend: %s⡀⡄%s<%dms %s⡆%s<%dms %s⡇%s>=%dms %s%s!%sfail %s(2x density)%s\n",
				gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, red, bold, reset, gray, reset)
		} else {
			fmt.Printf("%sLegend: %s▁▂▃%s<%dms %s▄▅%s<%dms %s▆▇█%s>=%dms %s%s!%sfail%s\n",
				gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, red, bold, reset, reset)
		}
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
			if s.braille {
				if s.hasPending {
					// Pair with pending: pending=left, failure=right
					s.blocks = append(s.blocks, getBrailleChar(s.pendingRTT, -1))
					s.hasPending = false
				} else {
					// Store failure as pending
					s.pendingRTT = -1
					s.hasPending = true
				}
			} else {
				s.blocks = append(s.blocks, red+bold+"!"+reset)
			}

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
					if !*noHeader && !*quiet && !*silent {
						fmt.Printf("\n%s↓ Downgrading to %s (3 initial failures)%s\n", yellow, protoNames[currentProto], reset)
						printHeader()
						fmt.Println() // Reserve stats line
						fmt.Print(up) // Move back to bar line
					}

					// Skip to next iteration - don't double-print or wait
					continue
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
			if s.braille {
				if s.hasPending {
					// Pair with pending: pending=left, current=right
					s.blocks = append(s.blocks, getBrailleChar(s.pendingRTT, rtt))
					s.hasPending = false
				} else {
					// Store as pending
					s.pendingRTT = rtt
					s.hasPending = true
				}
			} else {
				s.blocks = append(s.blocks, getBlock(rtt))
			}
		}
		printDisplay(s)
		requestNum++
		if *count > 0 && requestNum >= *count {
			if !*silent {
				printFinal(displayURL, s)
			}
			cleanup()
			os.Exit(0)
		}
		sleepDuration := *interval
		if *jitter > 0 {
			sleepDuration += time.Duration(rand.Int63n(int64(*jitter)))
		}
		time.Sleep(sleepDuration)
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
	_ = resp.Body.Close()

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

// getBrailleHeight returns a height level (1-4) for braille rendering
// Minimum is 1 to ensure at least one dot is always visible
func getBrailleHeight(rtt time.Duration) int {
	if rtt < 0 {
		return -1 // failure
	}
	ms := rtt.Milliseconds()

	if ms < greenThreshold {
		// Green zone: heights 1-2
		progress := ms - minLatency
		span := greenThreshold - minLatency
		if span > 0 && progress > span/2 {
			return 2
		}
		return 1
	} else if ms < yellowThreshold {
		// Yellow zone: height 3
		return 3
	} else {
		// Red zone: height 4
		return 4
	}
}

// getColorForRTT returns the ANSI color for a given RTT
func getColorForRTT(rtt time.Duration) string {
	if rtt < 0 {
		return red + bold
	}
	ms := rtt.Milliseconds()
	if ms < greenThreshold {
		return green
	} else if ms < yellowThreshold {
		return yellow
	}
	return red
}

// getBrailleChar returns a braille character combining two RTT values
// leftRTT is the first (older) reading, rightRTT is the second (newer)
// Returns the character with appropriate color
func getBrailleChar(leftRTT, rightRTT time.Duration) string {
	leftHeight := getBrailleHeight(leftRTT)
	rightHeight := getBrailleHeight(rightRTT)

	// Handle failures
	if leftHeight < 0 && rightHeight < 0 {
		return red + bold + "!" + reset
	}
	if leftHeight < 0 {
		leftHeight = 0
	}
	if rightHeight < 0 {
		rightHeight = 0
	}

	// Build braille character
	char := rune(0x2800) + brailleLeft[leftHeight] + brailleRight[rightHeight]

	// Color based on worse (higher) latency
	color := getColorForRTT(leftRTT)
	if rightRTT > leftRTT {
		color = getColorForRTT(rightRTT)
	}

	return color + string(char) + reset
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

// truncateToWidth truncates s so its visible width does not exceed w columns.
// ANSI escape sequences are preserved but do not count toward width.
// If truncated, a reset sequence is appended to avoid color bleed.
func truncateToWidth(s string, w int) string {
	if w <= 0 {
		return reset
	}
	var b strings.Builder
	vis := 0
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) {
				j++
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}
		if vis >= w {
			b.WriteString(reset)
			return b.String()
		}
		b.WriteByte(s[i])
		vis++
		i++
	}
	return b.String()
}

func printDisplay(s *stats) {
	displayMu.Lock()
	defer displayMu.Unlock()

	width := getTermWidth()

	// Print new blocks since last print (incremental)
	for s.lastPrinted < len(s.blocks) {
		fmt.Print(s.blocks[s.lastPrinted])
		s.col++
		s.lastPrinted++

		// Check if we need to wrap to next line for the NEXT block
		if s.col >= width-1 {
			// Move to stats line, print newline to scroll, move back up, clear line
			fmt.Print(down + "\n" + up + col0 + clearLn)
			s.col = 0
		}
	}

	printStats(s, width)
}

// redrawDisplay reprints the visible bar tail and stats from scratch.
// Caller must hold displayMu.
func redrawDisplay(s *stats) {
	width := getTermWidth()

	// Only redraw blocks that were on the current line before suspend.
	// s.col tracks how far along the current line we were.
	count := s.col
	if count > width-1 {
		count = width - 1
	}
	start := len(s.blocks) - count
	if start < 0 {
		start = 0
	}
	s.col = 0
	for i := start; i < len(s.blocks); i++ {
		fmt.Print(s.blocks[i])
		s.col++
	}
	s.lastPrinted = len(s.blocks)

	printStats(s, width)
}

// printStats prints the stats line below the bar and returns the cursor
// to its position on the bar line. Uses relative cursor movement instead
// of save/restore to avoid position corruption from terminal scrolling.
func printStats(s *stats, width int) {
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

	statsText := fmt.Sprintf("%d/%s%d%s %s(%2d%%) lost;%s %d/%s%d%s/%d%sms; last:%s %s%d%s%sms%s",
		s.failures, bold, total, reset,
		gray, lossPct, reset,
		minMs, bold, avg.Milliseconds(), reset, s.max.Milliseconds(), gray, reset,
		bold, s.last.Milliseconds(), reset, gray, reset)

	statsText = truncateToWidth(statsText, width)

	// \n moves to stats line (scrolls if at bottom), print stats, then
	// use relative up + column positioning to return to bar line.
	fmt.Printf("\n%s%s%s%s\033[%dG", col0, clearLn, statsText, up, s.col+1)
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
