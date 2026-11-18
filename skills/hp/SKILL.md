---
name: hp
description: >-
  Use when checking or watching a host's HTTP(S) latency or reachability with hp (hittyping, prettyping for HTTP): run a bounded probe, pick protocol, thresholds or display flags, or read its output (the glyph bar, the stats line, the exit summary and its UP/DOWN timeline). `hp --skill` prints this guide from the binary.
---

# hp: HTTP(S) latency monitor

hp sends one `HEAD` request at a time to a host and draws each round-trip as a colored glyph, like prettyping does for
ping. It measures the full request (DNS, TCP, TLS, response headers) on a fresh connection every time (keep-alive is
off), so numbers are higher than ICMP ping and include the TLS handshake. Any HTTP response is a success, a `503` or
an unfollowed `301` included: hp measures reachability and latency, not health.

## Run it bounded

hp runs until Ctrl+C unless `-c` is given. An agent always passes `-c`:

```bash
hp -c 10 -q cloudflare.com           # 10 requests, then the exit summary
hp -c 30 -i 500ms -t 2s -q 8.8.8.8   # faster probe, shorter timeout
```

- The argument is a host, an IP or `host:port`; IPv6 is bracketed for you. No argument means `1.1.1.1`. A leading
  `https://` or `http://` is stripped and the scheme comes from the flags, so `hp http://x` still uses HTTPS: pass `-1`
  for plain HTTP. A path (`example.com/health`) is not supported: hp reads it as part of the host name and exits 1.
- Exit code: `0` when the run ends (also at 100% loss), `1` for a bad flag or a host that does not resolve.
- Output always carries ANSI colors and cursor movement, even through a pipe. Read the **exit summary** (the
  plain-text block after `--- <host> hp statistics ---`); the live bar is redrawn in place and is noise in a capture.
  `-Q` hides that summary too, so never use `-Q` when you need the result.

## Read the result

```text
--- cloudflare.com hp statistics ---
10 requests, 9 ok, 1 failed, 10% loss
round-trip min/avg/max = 21/34/88 ms
timeline:
  10:02:18  UP       6s (6 ok)
  10:02:24  DOWN     1s (1 lost)
  10:02:25  UP    ← active
```

- `round-trip` is printed only when at least one request succeeded.
- `timeline` is printed only when a request failed. Each row is a period of consecutive successes (UP) or failures
  (DOWN) with its start time, length and request count; the last row is the period still running. Over 20 periods it
  keeps the first 5 and last 5. The date is added when the run spans midnight.
- A failure is any error: timeout, refused connection, TLS error, or `-2` without HTTP/2 negotiated.

The live stats line reads `failed/total (loss%) lost; min/avg/max ms; last: N ms`.

## The bar

| Glyph | Block mode (default) | Braille mode (`-b`, two requests per cell) |
|---|---|---|
| green | `▁▂▃` below `--green` (150 ms) | `⡀⡄` |
| yellow | `▄▅` below `--yellow` (400 ms) | `⡆` |
| red | `▆▇█` at or above `--yellow` | `⡇` |
| failure | red `!` | purple `⣿` (`--fail-style dots`, default) or red `!` (`--fail-style bang`) |

- Braille cells pair two requests (left dot column = older). The cell's color is the slower one's. A request still waiting for its pair when the run ends is drawn as a half cell.
- `--fail-style dots` (0.8.7+): any cell holding a failure becomes a purple `⣿`.
  `--fail-style bang` (behavior up to 0.8.6): red `!` only when both requests failed; one failure leaves its column
  blank and the cell shows the other request. Block mode ignores `--fail-style`.
- One glyph is one `--interval` of wall time, up or down (a braille cell is two). A request starts every interval
  whatever earlier ones are doing, as ping does, so loss % is the share of time the host did not answer. Glyphs are
  drawn in send order: during an outage the bar runs up to `--timeout` behind, while up to timeout ÷ interval
  requests (5 with the defaults) wait in flight.

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-c`, `--count` | 0 | requests to send, 0 = until Ctrl+C |
| `-i`, `--interval` | 1s | time from one request's start to the next |
| `-j`, `--jitter` | 0 | add a random 0..N to each interval |
| `-t`, `--timeout` | 5s | per-request timeout |
| `-1`, `--http` | | plain HTTP/1.1 (`http://`) |
| `-2`, `--http2` | | require HTTP/2; a response on another protocol counts as a failure |
| `-3`, `--http3` | | HTTP/3 over QUIC (UDP 443) |
| `-d`, `--downgrade` | | after 3 failures before the first success, step down 3 → 2 → HTTPS until one answers |
| `-D`, `--downgrade-insecure` | | as `-d`, and may fall to plain HTTP |
| `-k`, `--insecure` | | skip TLS certificate checks |
| `-m`, `--min` | 0 | ms mapped to the lowest block (env `HP_MIN`) |
| `-g`, `--green` | 150 | green below this many ms (env `HP_GREEN`) |
| `-y`, `--yellow` | 400 | yellow below this, red at or above (env `HP_YELLOW`) |
| `-b`, `--braille` | | braille bar, two requests per cell |
| `--fail-style` | dots | braille failure glyph, `dots` or `bang` (env `HP_FAIL_STYLE`) |
| `--legend` | | print the legend line |
| `--noheader` | | hide the header line |
| `-q`, `--quiet` | | hide header and legend |
| `-Q`, `--silent` | | also hide the exit summary |
| `-v`, `--version` | | print the version |
| `--skill` | | print this guide |

`-1`, `-2` and `-3` exclude each other. A flag beats its env var; a threshold flag of 0 is ignored.

## Proxies

`HTTPS_PROXY`, `HTTP_PROXY` and `ALL_PROXY` are honored, SOCKS5 included
(`HTTPS_PROXY=socks5://host:1080 hp -c 5 example.com`). With a proxy set, hp skips its local DNS check and the header
shows no resolved IP. HTTP/3 does not go through the proxy.
