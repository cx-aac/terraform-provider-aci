package provider

import (
	"math"
	"math/rand"
	"time"
)

const Retries = 4
const Factor = 3
const MinDelay = 4 * time.Second

func backoff(attempts int) bool {
	if attempts >= Retries {
		return false
	}
	min := float64(MinDelay)
	backoff := min * math.Pow(Factor, float64(attempts))
	backoff = (rand.Float64()/2+0.5)*(backoff-min) + min
	time.Sleep(time.Duration(backoff))
	return true
}
