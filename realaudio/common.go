package realaudio

import "errors"

type Format struct {
	SampleRate   int
	ChannelCount int
}

var (
	ErrDone = errors.New("Done")
)
