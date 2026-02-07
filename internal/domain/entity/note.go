package entity

type NoteVisibility string

const (
	VisibilityPublic    NoteVisibility = "public"
	VisibilityHome      NoteVisibility = "home"
	VisibilityFollowers NoteVisibility = "followers"
	VisibilitySpecified NoteVisibility = "specified"
)

type Note struct {
	Text       string
	Visibility NoteVisibility
}

func NewNote(text string, visibility NoteVisibility) *Note {
	return &Note{
		Text:       text,
		Visibility: visibility,
	}
}

func NewNoteFromTrack(track *Track, visibility NoteVisibility) *Note {
	return &Note{
		Text:       track.FormatMessage(),
		Visibility: visibility,
	}
}
