package entity

import (
	"fmt"
	"strings"
)

const (
	MaxNoteLength = 3000
)

type TrackType string

const (
	TrackTypeMusic   TrackType = "track"
	TrackTypeEpisode TrackType = "episode"
	TrackTypeAd      TrackType = "ad"
	TrackTypeUnknown TrackType = "unknown"
)

type Track struct {
	Title    string
	Artist   string
	Album    string
	URL      string
	Progress int64
	Type     TrackType
}

func NewTrack(title, artist, album, url string, progress int64) *Track {
	return &Track{
		Title:    title,
		Artist:   artist,
		Album:    album,
		URL:      url,
		Progress: progress,
		Type:     TrackTypeMusic,
	}
}

func NewTrackWithType(title, artist, album, url string, progress int64, trackType TrackType) *Track {
	return &Track{
		Title:    title,
		Artist:   artist,
		Album:    album,
		URL:      url,
		Progress: progress,
		Type:     trackType,
	}
}

func (t *Track) FormatMessage() string {
	msg := fmt.Sprintf("\U0001F3B5 #なうぷれ : %s / %s (%s)", t.Title, t.Artist, t.Album)
	if t.URL != "" {
		msg += "\n" + t.URL
	}
	if len(msg) > MaxNoteLength {
		msg = msg[:MaxNoteLength-3] + "..."
	}
	return msg
}

func (t *Track) IsSameAs(title string) bool {
	return t.Title == title
}

func (t *Track) IsPlayedEnough(minProgressMs int64) bool {
	return t.Progress >= minProgressMs
}

func (t *Track) IsValid() bool {
	return strings.TrimSpace(t.Title) != ""
}

func (t *Track) IsMusic() bool {
	return t.Type == TrackTypeMusic
}

func (t *Track) IsPostable() bool {
	return t.IsValid() && t.IsMusic()
}
