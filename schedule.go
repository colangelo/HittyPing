package main

import "time"

// result is one finished request.
type result struct {
	sentAt time.Time
	proto  int
	rtt    time.Duration
	err    error
}

// inOrder releases results in send order: a result waits until every
// earlier request has finished.
type inOrder struct {
	next int
	held map[int]result
}

func (q *inOrder) add(seq int, r result) []result {
	if q.held == nil {
		q.held = map[int]result{}
	}
	q.held[seq] = r
	var out []result
	for {
		r, ok := q.held[q.next]
		if !ok {
			return out
		}
		delete(q.held, q.next)
		out = append(out, r)
		q.next++
	}
}

// nextSend returns the start time of the request after the one scheduled
// at prev. After a stall (a downgrade probe, Ctrl-Z) it returns now, so
// sending resumes with one request instead of a burst of missed ones.
func nextSend(prev, now time.Time, gap time.Duration) time.Time {
	next := prev.Add(gap)
	if next.Before(now) {
		return now
	}
	return next
}

// runFixedRate starts a request every gap(), whatever earlier requests are
// doing, so a slow or timed-out request never stretches the bar's time
// scale. prepare runs on this goroutine when a request starts and returns
// the blocking measurement; handle gets the results in send order. It
// returns once count results are handled; count 0 runs forever.
func runFixedRate(count int, gap func() time.Duration, prepare func() func() result, handle func(result)) {
	type done struct {
		seq int
		r   result
	}
	results := make(chan done)
	var order inOrder
	timer := time.NewTimer(0)
	next := time.Now()
	sent, handled := 0, 0
	for count == 0 || handled < count {
		select {
		case <-timer.C:
			run, seq := prepare(), sent
			sent++
			go func() { results <- done{seq, run()} }()
			if count == 0 || sent < count {
				next = nextSend(next, time.Now(), gap())
				timer.Reset(time.Until(next))
			}
		case d := <-results:
			for _, r := range order.add(d.seq, d.r) {
				handle(r)
				handled++
			}
		}
	}
}
