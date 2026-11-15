package main

import (
	"crypto/tls"
	_ "embed"
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

const version = "0.8.8"

//go:embed skills/hp/SKILL.md
var skillText string

const (
	// ANSI colors
	green      = "\033[32m"
	yellow     = "\033[33m"
	red        = "\033[31m"
	purple     = "\033[35m"
	gray       = "\033[90m"
	bold       = "\033[1m"
	reset      = "\033[0m"
	clearLn    = "\033[K"
	up         = "\033[A"
	down       = "\033[B"
	col0       = "\033[0G"
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

// failStyle selects how a failed request is drawn in braille mode.
// Block mode always draws a red "!".
type failStyle int

const (
	failDots failStyle = iota // purple full cell ⣿ for any pair holding a failure (0.8.7+)
	failBang                  // red "!" for two failures, blank side for one (up to 0.8.6)
)

var failStyleNames = map[string]failStyle{"dots": failDots, "bang": failBang}

var brailleFail = failDots

func parseFailStyle(name string) (failStyle, error) {
	if fs, ok := failStyleNames[strings.ToLower(name)]; ok {
		return fs, nil
	}
	return 0, fmt.Errorf("invalid fail style %q (want dots or bang)", name)
}

// getEnvInt returns the env var value as int64, or the default if not set/invalid
func getEnvInt(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return def
}

type period struct {
	up    bool
	start time.Time
	count int
}

type stats struct {
	count         int
	failures      int
	total         time.Duration
	min           time.Duration
	max           time.Duration
	last          time.Duration
	blocks        []string      // individual blocks for proper width handling
	col           int           // current column position on bar line
	lastPrinted   int           // last block index printed
	braille       bool          // braille mode enabled
	pendingRTT    time.Duration // pending RTT for braille pairing (-1 = failure, 0 = none)
	hasPending    bool          // whether there's a pending RTT
	periods       []period      // completed UP/DOWN periods
	currentPeriod *period       // active period (nil until first request)
}

// recordPeriod counts a request into the UP/DOWN timeline. at is when the
// request was sent, so a period starts when its first request went out.
func recordPeriod(s *stats, up bool, at time.Time) {
	if s.currentPeriod == nil {
		s.currentPeriod = &period{up: up, start: at, count: 1}
		return
	}
	if s.currentPeriod.up == up {
		s.currentPeriod.count++
		return
	}
	// State flipped — close current period and start new one
	s.periods = append(s.periods, *s.currentPeriod)
	s.currentPeriod = &period{up: up, start: at, count: 1}
}

// recordResult adds a finished request to the stats, the timeline and the bar.
func recordResult(s *stats, r result) {
	rtt := r.rtt
	if r.err != nil {
		s.failures++
		rtt = -1
	} else {
		s.count++
		s.total += rtt
		s.last = rtt
		s.min = min(s.min, rtt)
		s.max = max(s.max, rtt)
	}
	recordPeriod(s, r.err == nil, r.sentAt)

	switch {
	case s.braille && s.hasPending:
		s.blocks = append(s.blocks, getBrailleChar(s.pendingRTT, rtt))
		s.hasPending = false
	case s.braille:
		s.pendingRTT, s.hasPending = rtt, true
	case r.err != nil:
		s.blocks = append(s.blocks, red+bold+"!"+reset)
	default:
		s.blocks = append(s.blocks, getBlock(rtt))
	}
}

// lowerWorkingProto probes each protocol below current, down to minProto,
// and returns the first one that answers.
func lowerWorkingProto(host string, current, minProto int, timeout time.Duration, insecure bool) (int, bool) {
	for p := current - 1; p >= minProto; p-- {
		if _, err := measureRTT(createClient(p, timeout, insecure), getURLForProto(host, p), p); err == nil {
			return p, true
		}
	}
	return current, false
}

func closePeriods(s *stats) {
	if s.currentPeriod != nil {
		s.periods = append(s.periods, *s.currentPeriod)
		s.currentPeriod = nil
	}
}

type options struct {
	interval, jitter, timeout        *time.Duration
	count                            *int
	showLegend, noHeader, useBraille *bool
	quiet, silent, insecure          *bool
	useHTTP1, useHTTP2, useHTTP3     *bool
	downgrade, downgradeInsecure     *bool
	showVersion, showSkill           *bool
	minFlag, greenFlag, yellowFlag   *int64
	failStyle                        *string
}

func defineFlags(fs *flag.FlagSet) *options {
	return &options{
		interval:          fs.DurationP("interval", "i", time.Second, "interval between requests"),
		jitter:            fs.DurationP("jitter", "j", 0, "max random jitter to add to interval (e.g., 200ms, 3s)"),
		timeout:           fs.DurationP("timeout", "t", 5*time.Second, "request timeout"),
		count:             fs.IntP("count", "c", 0, "number of requests (0 = unlimited)"),
		showLegend:        fs.Bool("legend", false, "show the legend line"),
		noHeader:          fs.Bool("noheader", false, "hide the header line"),
		useBraille:        fs.BoolP("braille", "b", false, "use braille visualization (2x density)"),
		failStyle:         fs.String("fail-style", "", "braille failure marker: dots (purple ⣿) or bang (red !, up to 0.8.6) (env: HP_FAIL_STYLE)"),
		quiet:             fs.BoolP("quiet", "q", false, "hide header and legend"),
		silent:            fs.BoolP("silent", "Q", false, "hide header, legend, and final stats"),
		minFlag:           fs.Int64P("min", "m", 0, "min latency baseline in ms (env: HP_MIN)"),
		greenFlag:         fs.Int64P("green", "g", 0, "green threshold in ms (env: HP_GREEN)"),
		yellowFlag:        fs.Int64P("yellow", "y", 0, "yellow threshold in ms (env: HP_YELLOW)"),
		insecure:          fs.BoolP("insecure", "k", false, "skip TLS certificate verification"),
		useHTTP1:          fs.BoolP("http", "1", false, "use plain HTTP/1.1"),
		useHTTP2:          fs.BoolP("http2", "2", false, "force HTTP/2 (fail if not negotiated)"),
		useHTTP3:          fs.BoolP("http3", "3", false, "use HTTP/3 (QUIC)"),
		downgrade:         fs.BoolP("downgrade", "d", false, "auto-downgrade protocol on failures (secure only)"),
		downgradeInsecure: fs.BoolP("downgrade-insecure", "D", false, "auto-downgrade including plain HTTP"),
		showVersion:       fs.BoolP("version", "v", false, "show version and exit"),
		showSkill:         fs.Bool("skill", false, "print the agent skill (how to use hp) and exit"),
	}
}

func main() {
	// Silence quic-go UDP buffer warnings that corrupt terminal display
	log.SetOutput(io.Discard)

	o := defineFlags(flag.CommandLine)
	flag.Parse()

	if *o.showVersion {
		fmt.Printf("hp (hittyping) version %s\n", version)
		os.Exit(0)
	}
	if *o.showSkill {
		fmt.Print(skillText)
		os.Exit(0)
	}

	// Apply thresholds: env vars first, flags override
	minLatency = getEnvInt("HP_MIN", minLatency)
	greenThreshold = getEnvInt("HP_GREEN", greenThreshold)
	yellowThreshold = getEnvInt("HP_YELLOW", yellowThreshold)
	if *o.minFlag > 0 {
		minLatency = *o.minFlag
	}
	if *o.greenFlag > 0 {
		greenThreshold = *o.greenFlag
	}
	if *o.yellowFlag > 0 {
		yellowThreshold = *o.yellowFlag
	}

	styleName := os.Getenv("HP_FAIL_STYLE")
	if *o.failStyle != "" {
		styleName = *o.failStyle
	}
	if styleName != "" {
		fs, err := parseFailStyle(styleName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		brailleFail = fs
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

	// Resolve hostname to IP for display (and validate it exists).
	// Skipped when a proxy is configured: the proxy resolves the host
	// (e.g. socks5h), so the local resolver may legitimately fail.
	resolvedIP := ""
	hostForLookup := hostname(displayURL)
	if ip := net.ParseIP(hostForLookup); ip == nil && !proxyConfigured() {
		ips, err := net.LookupHost(hostForLookup)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot resolve %s: %v\n", hostForLookup, err)
			os.Exit(1)
		}
		if len(ips) > 0 {
			resolvedIP = ips[0]
		}
	}

	s := &stats{min: time.Hour, braille: *o.useBraille}

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
		closePeriods(s)
		if !*o.silent {
			printFinal(displayURL, s)
		}
		cleanup()
		os.Exit(0)
	}()

	// Validate protocol flags (mutually exclusive)
	protoCount := 0
	if *o.useHTTP1 {
		protoCount++
	}
	if *o.useHTTP2 {
		protoCount++
	}
	if *o.useHTTP3 {
		protoCount++
	}
	if protoCount > 1 {
		fmt.Fprintln(os.Stderr, "Cannot combine -1/--http, -2/--http2, and -3/--http3")
		os.Exit(1)
	}

	// Determine initial protocol level
	currentProto := protoHTTPS
	if *o.useHTTP1 {
		currentProto = protoHTTP1
	} else if *o.useHTTP2 {
		currentProto = protoHTTP2
	} else if *o.useHTTP3 {
		currentProto = protoHTTP3
	}

	// Determine minimum protocol level for downgrade
	minProto := protoHTTPS // secure only by default
	if *o.downgradeInsecure {
		minProto = protoHTTP1
	}
	canDowngrade := *o.downgrade || *o.downgradeInsecure

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
	if !*o.noHeader && !*o.quiet && !*o.silent {
		printHeader()
	}
	if *o.showLegend && !*o.quiet && !*o.silent {
		if *o.useBraille {
			failColor, failGlyph := purple, "⣿"
			if brailleFail == failBang {
				failColor, failGlyph = red, "!"
			}
			fmt.Printf("%sLegend: %s⡀⡄%s<%dms %s⡆%s<%dms %s⡇%s>=%dms %s%s%s%sfail %s(2x density)%s\n",
				gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, failColor, bold, failGlyph, reset, gray, reset)
		} else {
			fmt.Printf("%sLegend: %s▁▂▃%s<%dms %s▄▅%s<%dms %s▆▇█%s>=%dms %s%s!%sfail%s\n",
				gray, green, reset, greenThreshold, yellow, reset, yellowThreshold, red, reset, yellowThreshold, red, bold, reset, reset)
		}
	}
	fmt.Println() // Reserve stats line
	fmt.Print(up) // Move back to bar line

	// Create HTTP client
	url := getURLForProto(host, currentProto)
	client := createClient(currentProto, *o.timeout, *o.insecure)

	gap := func() time.Duration {
		d := *o.interval
		if *o.jitter > 0 {
			d += time.Duration(rand.Int63n(int64(*o.jitter)))
		}
		return d
	}
	// Captured when the request starts: a downgrade must not switch the
	// protocol under requests already in flight.
	prepare := func() func() result {
		c, u, p, sentAt := client, url, currentProto, time.Now()
		return func() result {
			rtt, err := measureRTT(c, u, p)
			return result{sentAt: sentAt, proto: p, rtt: rtt, err: err}
		}
	}

	consecutiveFailures := 0
	handle := func(r result) {
		recordResult(s, r)
		printDisplay(s)
		switch {
		case r.err == nil:
			consecutiveFailures = 0
		case r.proto == currentProto:
			consecutiveFailures++
		}

		// Downgrade only at startup, before the first success
		if !canDowngrade || consecutiveFailures < 3 || currentProto <= minProto || s.count > 0 {
			return
		}
		p, ok := lowerWorkingProto(host, currentProto, minProto, *o.timeout, *o.insecure)
		if !ok {
			return
		}
		currentProto = p
		url = getURLForProto(host, p)
		client = createClient(p, *o.timeout, *o.insecure)
		consecutiveFailures = 0
		if !*o.noHeader && !*o.quiet && !*o.silent {
			fmt.Printf("\n%s↓ Downgrading to %s (3 initial failures)%s\n", yellow, protoNames[currentProto], reset)
			printHeader()
			fmt.Println() // Reserve stats line
			fmt.Print(up) // Move back to bar line
		}
	}

	runFixedRate(*o.count, gap, prepare, handle)
	closePeriods(s)
	if !*o.silent {
		printFinal(displayURL, s)
	}
	cleanup()
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

// hostname strips the port and IPv6 brackets from a target, leaving the
// name to resolve: "example.com:8443" → "example.com", "[::1]:80" → "::1".
func hostname(target string) string {
	if h, _, err := net.SplitHostPort(target); err == nil {
		return h
	}
	return strings.TrimPrefix(strings.TrimSuffix(target, "]"), "[")
}

// proxyConfigured reports whether any HTTP/HTTPS proxy env var is set.
// Mirrors the variables consulted by http.ProxyFromEnvironment.
func proxyConfigured() bool {
	for _, k := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		if os.Getenv(k) != "" {
			return true
		}
	}
	return false
}

func createClient(protoLevel int, timeout time.Duration, insecure bool) *http.Client {
	if protoLevel == protoHTTP3 {
		return newHTTP3Client(timeout, insecure)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
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

	if leftHeight < 0 || rightHeight < 0 {
		if brailleFail == failDots {
			return purple + bold + "⣿" + reset
		}
		if leftHeight < 0 && rightHeight < 0 {
			return red + bold + "!" + reset
		}
		leftHeight = max(leftHeight, 0)
		rightHeight = max(rightHeight, 0)
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
			for j < len(s) && (s[j] < 'A' || s[j] > 'Z') && (s[j] < 'a' || s[j] > 'z') {
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

func fmtDuration(d time.Duration) string {
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, sec)
	}
	return fmt.Sprintf("%ds", sec)
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

	if s.failures > 0 && len(s.periods) > 0 {
		fmt.Printf("%stimeline:%s\n", gray, reset)
		periods := s.periods
		truncated := 0
		if len(periods) > 20 {
			truncated = len(periods) - 10
			head := periods[:5]
			tail := periods[len(periods)-5:]
			periods = append(head, tail...)
		}
		now := time.Now()
		// include date if session spans multiple calendar days
		tsFmt := "15:04:05"
		if len(s.periods) > 0 {
			first := s.periods[0].start
			last := now
			if first.Format("2006-01-02") != last.Format("2006-01-02") {
				tsFmt = "Jan 02 15:04:05"
			}
		}
		for i, p := range periods {
			ts := p.start.Format(tsFmt)
			var label, color, detail string
			if p.up {
				label = "UP  "
				color = green
				detail = fmt.Sprintf("(%d ok)", p.count)
			} else {
				label = "DOWN"
				color = red
				detail = fmt.Sprintf("(%d lost)", p.count)
			}
			isLast := i == len(periods)-1 && truncated == 0 ||
				i == len(periods)-1 && truncated > 0

			if isLast {
				fmt.Printf("  %s  %s%s%s  %s← active%s\n", ts, color, label, reset, gray, reset)
			} else {
				var end time.Time
				if i+1 < len(periods) {
					end = periods[i+1].start
				} else {
					end = now
				}
				dur := fmtDuration(end.Sub(p.start))
				fmt.Printf("  %s  %s%s%s  %6s %s\n", ts, color, label, reset, dur, detail)
			}

			if truncated > 0 && i == 4 {
				fmt.Printf("  %s... %d more ...%s\n", gray, truncated, reset)
			}
		}
	}
}
