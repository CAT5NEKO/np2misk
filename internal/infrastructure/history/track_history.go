package history

import (
	"sync"
	"time"

	"np2misk/internal/domain/repository"
)

type TrackHistoryConfig struct {
	CooldownPeriod time.Duration
}

type trackHistory struct {
	mu             sync.Mutex
	lastTitle      string
	lastRecordTime time.Time
	cooldownPeriod time.Duration
}

func NewTrackHistory(cfg TrackHistoryConfig) repository.TrackHistory {
	cooldown := cfg.CooldownPeriod
	if cooldown == 0 {
		cooldown = 30 * time.Second
	}
	return &trackHistory{
		cooldownPeriod: cooldown,
	}
}

func (h *trackHistory) IsNewTrack(title string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if title == h.lastTitle {
		return false
	}

	if !h.lastRecordTime.IsZero() && time.Since(h.lastRecordTime) < h.cooldownPeriod {
		return false
	}

	return true
}

func (h *trackHistory) RecordTrack(title string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastTitle = title
	h.lastRecordTime = time.Now()
}

func (h *trackHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastTitle = ""
}
