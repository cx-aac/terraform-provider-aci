package provider

import (
	"math"
	"math/rand"
	"time"
)

const Factor = 3
const MinDelay = 4 * time.Second
const MaxDelay = 60 * time.Second

func backoff(attempts int, maxRetries int) bool {
	if attempts > maxRetries {
		return false
	}
	min := float64(MinDelay)
	backoff := min * math.Pow(Factor, float64(attempts))
	if backoff > float64(MaxDelay) {
		backoff = float64(MaxDelay)
	}
	backoff = (rand.Float64()/2+0.5)*(backoff-min) + min
	time.Sleep(time.Duration(backoff))
	return true
}
