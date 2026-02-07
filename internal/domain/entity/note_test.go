package entity

import "testing"

func TestNewNote(t *testing.T) {
	note := NewNote("test text", VisibilityHome)

	if note.Text != "test text" {
		t.Errorf("Text = %v, want %v", note.Text, "test text")
	}
	if note.Visibility != VisibilityHome {
		t.Errorf("Visibility = %v, want %v", note.Visibility, VisibilityHome)
	}
}

func TestNewNoteFromTrack(t *testing.T) {
	track := NewTrack("Song", "Artist", "Album", "https://spotify.com/track/123", 5000)
	note := NewNoteFromTrack(track, VisibilityHome)

	expectedText := track.FormatMessage()
	if note.Text != expectedText {
		t.Errorf("Text = %v, want %v", note.Text, expectedText)
	}
	if note.Visibility != VisibilityHome {
		t.Errorf("Visibility = %v, want %v", note.Visibility, VisibilityHome)
	}
}

func TestNoteVisibility(t *testing.T) {
	tests := []struct {
		name     string
		vis      NoteVisibility
		expected string
	}{
		{name: "public visibility", vis: VisibilityPublic, expected: "public"},
		{name: "home visibility", vis: VisibilityHome, expected: "home"},
		{name: "followers visibility", vis: VisibilityFollowers, expected: "followers"},
		{name: "specified visibility", vis: VisibilitySpecified, expected: "specified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.vis) != tt.expected {
				t.Errorf("Visibility = %v, want %v", tt.vis, tt.expected)
			}
		})
	}
}
