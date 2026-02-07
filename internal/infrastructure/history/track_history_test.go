package history

import (
	"testing"
	"time"
)

func TestTrackHistory_IsNewTrack(t *testing.T) {
	h := NewTrackHistory(TrackHistoryConfig{CooldownPeriod: 0})

	if !h.IsNewTrack("Song A") {
		t.Error("expected new track to return true")
	}

	h.RecordTrack("Song A")

	if h.IsNewTrack("Song A") {
		t.Error("expected same track to return false")
	}

	if !h.IsNewTrack("Song B") {
		t.Error("expected different track to return true")
	}
}

func TestTrackHistory_CooldownPeriod(t *testing.T) {
	h := NewTrackHistory(TrackHistoryConfig{CooldownPeriod: 50 * time.Millisecond})

	h.RecordTrack("Song A")

	if h.IsNewTrack("Song B") {
		t.Error("expected false during cooldown")
	}

	time.Sleep(60 * time.Millisecond)

	if !h.IsNewTrack("Song B") {
		t.Error("expected true after cooldown")
	}
}

func TestTrackHistory_Clear(t *testing.T) {
	h := NewTrackHistory(TrackHistoryConfig{CooldownPeriod: 0})

	h.RecordTrack("Song A")
	h.Clear()

	if !h.IsNewTrack("Song A") {
		t.Error("expected true after clear")
	}
}

func TestTrackHistory_DefaultCooldown(t *testing.T) {
	h := NewTrackHistory(TrackHistoryConfig{})

	h.RecordTrack("Song A")

	if h.IsNewTrack("Song B") {
		t.Error("expected default cooldown to prevent immediate new track")
	}
}
