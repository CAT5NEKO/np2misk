package entity

import (
	"strings"
	"testing"
)

func TestTrack_FormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		track    *Track
		expected string
	}{
		{
			name:     "formats message correctly",
			track:    NewTrack("Song Title", "Artist Name", "Album Name", "https://open.spotify.com/track/123", 10000),
			expected: "\U0001F3B5 #なうぷれ : Song Title / Artist Name (Album Name)\nhttps://open.spotify.com/track/123",
		},
		{
			name:     "handles multiple artists",
			track:    NewTrack("Song", "Artist1, Artist2", "Album", "https://spotify.com/track/456", 5000),
			expected: "\U0001F3B5 #なうぷれ : Song / Artist1, Artist2 (Album)\nhttps://spotify.com/track/456",
		},
		{
			name:     "handles empty URL",
			track:    NewTrack("Song", "Artist", "Album", "", 5000),
			expected: "\U0001F3B5 #なうぷれ : Song / Artist (Album)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.track.FormatMessage()
			if got != tt.expected {
				t.Errorf("FormatMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTrack_FormatMessage_TruncatesLongText(t *testing.T) {
	longTitle := strings.Repeat("a", MaxNoteLength)
	track := NewTrack(longTitle, "Artist", "Album", "url", 5000)
	msg := track.FormatMessage()

	if len(msg) > MaxNoteLength {
		t.Errorf("FormatMessage() length = %d, want <= %d", len(msg), MaxNoteLength)
	}
	if !strings.HasSuffix(msg, "...") {
		t.Error("FormatMessage() should end with '...' when truncated")
	}
}

func TestTrack_IsSameAs(t *testing.T) {
	track := NewTrack("Test Song", "Test Artist", "Test Album", "https://spotify.com", 5000)

	tests := []struct {
		name     string
		title    string
		expected bool
	}{
		{
			name:     "returns true for same title",
			title:    "Test Song",
			expected: true,
		},
		{
			name:     "returns false for different title",
			title:    "Different Song",
			expected: false,
		},
		{
			name:     "returns false for empty title",
			title:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := track.IsSameAs(tt.title)
			if got != tt.expected {
				t.Errorf("IsSameAs() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTrack_IsPlayedEnough(t *testing.T) {
	tests := []struct {
		name          string
		progress      int64
		minProgressMs int64
		expected      bool
	}{
		{
			name:          "returns true when progress exceeds minimum",
			progress:      10000,
			minProgressMs: 5000,
			expected:      true,
		},
		{
			name:          "returns true when progress equals minimum",
			progress:      5000,
			minProgressMs: 5000,
			expected:      true,
		},
		{
			name:          "returns false when progress is below minimum",
			progress:      3000,
			minProgressMs: 5000,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := NewTrack("Test", "Artist", "Album", "url", tt.progress)
			got := track.IsPlayedEnough(tt.minProgressMs)
			if got != tt.expected {
				t.Errorf("IsPlayedEnough() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTrack_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected bool
	}{
		{name: "valid title", title: "Song", expected: true},
		{name: "empty title", title: "", expected: false},
		{name: "whitespace only", title: "   ", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := NewTrack(tt.title, "Artist", "Album", "url", 5000)
			if got := track.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTrack_IsMusic(t *testing.T) {
	tests := []struct {
		name      string
		trackType TrackType
		expected  bool
	}{
		{name: "music track", trackType: TrackTypeMusic, expected: true},
		{name: "episode", trackType: TrackTypeEpisode, expected: false},
		{name: "ad", trackType: TrackTypeAd, expected: false},
		{name: "unknown", trackType: TrackTypeUnknown, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := NewTrackWithType("Song", "Artist", "Album", "url", 5000, tt.trackType)
			if got := track.IsMusic(); got != tt.expected {
				t.Errorf("IsMusic() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTrack_IsPostable(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		trackType TrackType
		expected  bool
	}{
		{name: "valid music", title: "Song", trackType: TrackTypeMusic, expected: true},
		{name: "empty title music", title: "", trackType: TrackTypeMusic, expected: false},
		{name: "valid episode", title: "Episode", trackType: TrackTypeEpisode, expected: false},
		{name: "empty episode", title: "", trackType: TrackTypeEpisode, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track := NewTrackWithType(tt.title, "Artist", "Album", "url", 5000, tt.trackType)
			if got := track.IsPostable(); got != tt.expected {
				t.Errorf("IsPostable() = %v, want %v", got, tt.expected)
			}
		})
	}
}
