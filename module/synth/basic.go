package synth

import (
	"errors"

	"github.com/loov/synth/realaudio"
)

type Note struct {
	A      float32 // amplitude
	P0, P1 float32 // phase
	S0, S1 float32 // step
}

type Basic struct {
	Dampening float32

	playing    []Note
	sampleTime float32
}

func NewBasic(dampening float32) *Basic {
	return &Basic{Dampening: dampening}
}

func (synth *Basic) Init(format realaudio.Format) error {
	if format.ChannelCount != 2 {
		return errors.New("cannot handle non-stereo")
	}
	synth.sampleTime = tau / float32(format.SampleRate)
	return nil
}

func (synth *Basic) PlayNote(freq float32) {
	synth.playing = append(synth.playing, Note{
		A:  0.5,
		P0: 0, P1: 0,
		S0: freq * synth.sampleTime,
		S1: (freq + 5) * synth.sampleTime,
	})
}

func (synth *Basic) Render(buffer []float32) error {
	next := synth.playing[:0]
	damp := synth.Dampening
	for _, note := range synth.playing {
		for i := 0; i < len(buffer); i += 2 {
			note.P0 += note.S0
			note.P1 += note.S1

			v0, v1 := sin(note.P0), sin(note.P1)
			buffer[i] += note.A * v0
			buffer[i+1] += note.A * v1
			note.A *= damp
		}
		if note.P0 > tau {
			note.P0 -= tau
		}
		if note.P1 > tau {
			note.P1 -= tau
		}
		if note.A > 0.001 {
			next = append(next, note)
		}
	}
	synth.playing = next
	return nil
}
