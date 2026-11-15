package main

import (
	"net"
	"regexp"
	"strings"
	"testing"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[^A-Za-z]*[A-Za-z]?`)

func FuzzTruncateToWidth(f *testing.F) {
	f.Add("0/5 ( 0%) lost; 12/30/88ms; last: 20ms", 20)
	f.Add("\x1b[1m42\x1b[0m \x1b[90mms\x1b[0m", 3)
	f.Add("\x1b[12", 1)
	f.Add("", 0)
	f.Fuzz(func(t *testing.T, s string, w int) {
		got := truncateToWidth(s, w)
		if w <= 0 {
			if got != reset {
				t.Fatalf("truncateToWidth(%q, %d) = %q; want only a reset", s, w, got)
			}
			return
		}
		body, _ := strings.CutSuffix(got, reset)
		if !strings.HasPrefix(s, body) && !strings.HasPrefix(s, got) {
			t.Fatalf("truncateToWidth(%q, %d) = %q; not a prefix of the input", s, w, got)
		}
		if visible := len(ansiEscape.ReplaceAllString(body, "")); visible > w {
			t.Fatalf("truncateToWidth(%q, %d) = %q; %d visible bytes", s, w, got, visible)
		}
		if len(ansiEscape.ReplaceAllString(s, "")) <= w && got != s {
			t.Fatalf("truncateToWidth(%q, %d) = %q; input fits and must come back unchanged", s, w, got)
		}
	})
}

func FuzzHostname(f *testing.F) {
	f.Add("example.com", "8443")
	f.Add("127.0.0.1", "80")
	f.Add("::1", "443")
	f.Add("", "1")
	f.Fuzz(func(t *testing.T, host, port string) {
		if strings.ContainsAny(host, "[]") || strings.ContainsAny(port, ":[]") {
			t.Skip()
		}
		if got := hostname(net.JoinHostPort(host, port)); got != host {
			t.Fatalf("hostname(%q) = %q; want %q", net.JoinHostPort(host, port), got, host)
		}
		if !strings.Contains(host, ":") {
			if got := hostname(host); got != host {
				t.Fatalf("hostname(%q) = %q; want it unchanged", host, got)
			}
		}
	})
}
