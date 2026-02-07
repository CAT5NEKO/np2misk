package repository

type TrackHistory interface {
	IsNewTrack(title string) bool
	RecordTrack(title string)
	Clear()
}
