# Multi-Target Mode Design Options

This document summarizes layout options for supporting multiple targets in hp.

## Usage

```bash
hp google.com cloudflare.com 1.1.1.1
```

---

## Layout Options

### Option 1: Stacked Rows (Simplest)

Each target gets its own row with hostname prefix, bars grow horizontally.

```
HP multi-target mode
Legend: ▁▂▃<150ms ▄▅<400ms ▆▇█>=400ms !fail

google.com     ▁▁▂▁▄▃▁▂▁▁▂▃▁▁▂▁▁▁▂▁▆▁▁▂▁▃▁▁
cloudflare.com ▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
1.1.1.1        ▁▁▁▁▂▁▁▁▁▁▁▁▂▁▁▁▁▁▄▁▁▁▁▁▁▁▁▁

               min    avg    max   loss
google.com      45ms   67ms  512ms    0%
cloudflare.com  12ms   18ms   42ms    0%
1.1.1.1          8ms   14ms   31ms    0%
```

**Pros:** Simple implementation, clear separation
**Cons:** Hostname alignment issues with variable-length names

---

### Option 2: Compact Interleaved (Single Line)

All targets share one line with letter prefixes.

```
HP google.com | cloudflare.com | 1.1.1.1
Legend: G=google C=cloudflare 1=1.1.1.1

G▁C▁1▁G▂C▁1▁G▁C▁1▁G▃C▂1▁G▁C▁1▁
```

**Pros:** Compact, shows timing relationship between targets
**Cons:** Hard to read, limited to ~3 targets, confusing

---

### Option 3: Live Dashboard (Curses-Style)

Box-drawing characters for a TUI feel.

```
┌─ google.com ──────────────────────────┐
│ ▁▁▂▁▃▁▂▁▁▂▃▁▁▂▁▁  67ms avg | 0% loss │
├─ cloudflare.com ──────────────────────┤
│ ▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁  18ms avg | 0% loss │
├─ 1.1.1.1 ─────────────────────────────┤
│ ▁▁▁▁▂▁▁▁▁▁▁▁▂▁▁▁  14ms avg | 0% loss │
└───────────────────────────────────────┘
```

**Pros:** Pretty, professional look
**Cons:** Complex terminal control, box characters may not render everywhere

---

## Alignment Solutions (for Option 1)

### A. Labels Above Bars ✓ SELECTED

```
google.com
▁▁▂▁▄▃▁▂▁▁▂▃▁▁▂▁▁▁▂▁▆▁▁▂▁▃▁▁

cloudflare.com
▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁

1.1.1.1
▁▁▁▁▂▁▁▁▁▁▁▁▁▂▁▁▁▁▁▄▁▁▁▁▁▁▁▁
```

**Pros:** Clean, no alignment issues, clear separation
**Cons:** Uses more vertical space

---

### B. Short Numeric/Letter Prefixes

```
Legend: [1]=google.com [2]=cloudflare.com [3]=1.1.1.1

1: ▁▁▂▁▄▃▁▂▁▁▂▃▁▁▂▁▁▁▂▁▆▁▁▂▁▃▁▁
2: ▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
3: ▁▁▁▁▂▁▁▁▁▁▁▁▁▂▁▁▁▁▁▄▁▁▁▁▁▁▁▁
```

**Pros:** Compact, perfectly aligned
**Cons:** Requires mental mapping of numbers to hosts

---

### C. Right-Aligned Labels with Stats

```
▁▁▂▁▄▃▁▂▁▁▂▃▁▁▂▁▁▁▂▁▆▁▁▂▁▃▁▁  google.com     67ms
▁▁▁▁▁▁▂▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁  cloudflare.com 18ms
▁▁▁▁▂▁▁▁▁▁▁▁▁▂▁▁▁▁▁▄▁▁▁▁▁▁▁▁  1.1.1.1        14ms
```

**Pros:** Bars start at same position, live stats visible
**Cons:** Labels at end feel backwards, wrapping is awkward

---

## Decision

**Selected: Option 1 + Alignment Solution A (Labels Above Bars)**

This provides the cleanest output with no alignment issues while remaining simple to implement.
