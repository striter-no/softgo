package api

import (
	"time"
)

type FPSCounter struct {
	frames         int
	lastUpdate     time.Time
	currentFPS     float64
	updateInterval time.Duration
}

func NewFPSCounter(interval time.Duration) *FPSCounter {
	return &FPSCounter{
		lastUpdate:     time.Now(),
		updateInterval: interval,
	}
}

func (f *FPSCounter) Tick() float64 {
	f.frames++
	now := time.Now()
	elapsed := now.Sub(f.lastUpdate)

	if elapsed >= f.updateInterval {
		f.currentFPS = float64(f.frames) / elapsed.Seconds()

		f.frames = 0
		f.lastUpdate = now
	}

	return f.currentFPS
}
