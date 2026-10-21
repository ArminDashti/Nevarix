package netmon

import "time"

// CounterBps returns bits per second from SNMP counter deltas (32- or 64-bit wrap).
func CounterBps(prev, cur uint64, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	sec := elapsed.Seconds()
	if sec <= 0 {
		return 0
	}
	// uint64 subtraction wraps on counter rollover.
	delta := cur - prev
	return float64(delta) * 8 / sec
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
