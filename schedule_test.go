package main

import (
	"errors"
	"testing"
	"time"
)

func TestInOrder_ReleasesInSendOrder(t *testing.T) {
	var q inOrder
	at := func(i int) result { return result{rtt: time.Duration(i)} }

	if got := q.add(1, at(1)); len(got) != 0 {
		t.Fatalf("seq 1 before seq 0 released %d results; want 0", len(got))
	}
	if got := q.add(2, at(2)); len(got) != 0 {
		t.Fatalf("seq 2 before seq 0 released %d results; want 0", len(got))
	}
	got := q.add(0, at(0))
	if len(got) != 3 || got[0].rtt != 0 || got[1].rtt != 1 || got[2].rtt != 2 {
		t.Fatalf("seq 0 released %v; want seqs 0,1,2 in order", got)
	}
	if got := q.add(3, at(3)); len(got) != 1 || got[0].rtt != 3 {
		t.Fatalf("seq 3 released %v; want seq 3 alone", got)
	}
}

func TestNextSend(t *testing.T) {
	t0 := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	gap := time.Second
	if got := nextSend(t0, t0.Add(10*time.Millisecond), gap); !got.Equal(t0.Add(gap)) {
		t.Errorf("on time: next = %v; want %v (stays on the grid)", got, t0.Add(gap))
	}
	late := t0.Add(5 * time.Second)
	if got := nextSend(t0, late, gap); !got.Equal(late) {
		t.Errorf("after a stall: next = %v; want now %v (no burst)", got, late)
	}
}

// A failure that takes 5 intervals must not stretch the time between sends:
// one result per interval, like ping, during an outage as well.
func TestRunFixedRate_OutageKeepsTheTimeScale(t *testing.T) {
	const (
		n        = 10
		interval = 20 * time.Millisecond
		timeout  = 100 * time.Millisecond
	)
	var got []result
	start := time.Now()
	runFixedRate(n,
		func() time.Duration { return interval },
		func() func() result {
			sentAt := time.Now()
			return func() result {
				time.Sleep(timeout)
				return result{sentAt: sentAt, err: errors.New("timeout")}
			}
		},
		func(r result) { got = append(got, r) },
	)
	elapsed := time.Since(start)

	if len(got) != n {
		t.Fatalf("handled %d results; want %d", len(got), n)
	}
	// Sequential sending would take n*(timeout+interval) = 1.2s.
	if limit := (n-1)*interval + timeout + 300*time.Millisecond; elapsed > limit {
		t.Errorf("10 failing requests took %v; want under %v (fixed rate)", elapsed, limit)
	}
	span := got[n-1].sentAt.Sub(got[0].sentAt)
	if want := (n - 1) * interval; span < want-5*time.Millisecond || span > want+100*time.Millisecond {
		t.Errorf("first to last send = %v; want about %v (one interval per request)", span, want)
	}
	for i := 1; i < n; i++ {
		if got[i].sentAt.Before(got[i-1].sentAt) {
			t.Fatalf("result %d sent before result %d: not in send order", i, i-1)
		}
	}
}

// A slow request holds back faster later ones, so the bar stays in send order.
func TestRunFixedRate_SlowRequestKeepsItsSlot(t *testing.T) {
	seq := 0
	var order []int
	runFixedRate(3,
		func() time.Duration { return 5 * time.Millisecond },
		func() func() result {
			i := seq
			seq++
			return func() result {
				if i == 0 {
					time.Sleep(60 * time.Millisecond)
				}
				return result{rtt: time.Duration(i)}
			}
		},
		func(r result) { order = append(order, int(r.rtt)) },
	)
	if len(order) != 3 || order[0] != 0 || order[1] != 1 || order[2] != 2 {
		t.Errorf("handled order %v; want [0 1 2]", order)
	}
}

func TestRecordResult_PeriodStartsAtSendTime(t *testing.T) {
	sent := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	s := &stats{min: time.Hour}
	recordResult(s, result{sentAt: sent, err: errors.New("timeout")})
	if s.currentPeriod == nil || !s.currentPeriod.start.Equal(sent) {
		t.Fatalf("DOWN period start = %v; want the send time %v", s.currentPeriod, sent)
	}
	if s.failures != 1 || len(s.blocks) != 1 {
		t.Errorf("failures=%d blocks=%d; want 1 and 1", s.failures, len(s.blocks))
	}
}
