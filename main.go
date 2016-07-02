package main

import (
	"errors"
	"log"
	"math"
	"syscall"
	"time"
	"unsafe"

	"github.com/loov/synth/module/echo"
	"github.com/loov/synth/realaudio"
)

var (
	user32               = syscall.MustLoadDLL("user32.dll")
	procGetKeyboardState = user32.MustFindProc("GetKeyboardState")
	procGetKeyState      = user32.MustFindProc("GetKeyState")
)

const Scale = "Q2W3ER5T6Y7UI9O0P"

type Keyboard struct {
	Prev [256]byte
	Next [256]byte
}

func (keys *Keyboard) Update() bool {
	keys.Prev = keys.Next
	ok, _, _ := syscall.Syscall(procGetKeyboardState.Addr(), 2, uintptr(unsafe.Pointer(&keys.Next[0])), 0, 0)
	return ok != 0
}

func (keys *Keyboard) Down(key byte) bool {
	return keys.Next[key]&128 != 0
}
func (keys *Keyboard) Pressed(key byte) bool {
	return keys.Next[key]&128 != 0 && keys.Prev[key]&128 == 0
}
func (keys *Keyboard) Released(key byte) bool {
	return keys.Prev[key]&128 != 0 && keys.Next[key]&128 == 0
}

func KeyDown(key int) bool {
	v, _, _ := syscall.Syscall(procGetKeyState.Addr(), 2, uintptr(key), 0, 0)
	return v&0x8000 != 0
}

type Note struct {
	A      float32 // amplitude
	P0, P1 float32 // phase
	S0, S1 float32 // step
}

type Synth struct {
	Dampening float32

	playing    []Note
	sampleTime float32
}

func NewSynth(dampening float32) *Synth {
	return &Synth{Dampening: dampening}
}

func (synth *Synth) Init(format realaudio.Format) error {
	if format.ChannelCount != 2 {
		return errors.New("cannot handle non-stereo")
	}
	synth.sampleTime = 2 * math.Pi / float32(format.SampleRate)
	return nil
}

func (synth *Synth) PlayNote(freq float32) {
	synth.playing = append(synth.playing, Note{
		A:  0.5,
		P0: 0, P1: 0,
		S0: freq * synth.sampleTime,
		S1: (freq + 5) * synth.sampleTime,
	})
}

func (synth *Synth) Render(buffer []float32) error {
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
		if note.P0 > 2*math.Pi {
			note.P0 -= 2 * math.Pi
		}
		if note.P1 > 2*math.Pi {
			note.P1 -= 2 * math.Pi
		}
		if note.A > 0.001 {
			next = append(next, note)
		}
	}
	synth.playing = next
	return nil
}

func main() {
	synth := NewSynth(0.9997)

	var keys Keyboard
	update := realaudio.RenderFunc(func(_ []float32) error {
		if KeyDown(0x1b) {
			return realaudio.ErrDone
		}
		keys.Update()
		for i, key := range Scale {
			if keys.Pressed(byte(key)) {
				freq := float32(220 * math.Exp2(1+float64(i)/12))
				synth.PlayNote(freq)
			}
		}
		return nil
	})

	err := realaudio.Play(realaudio.Sources{
		update,
		synth,
		echo.New(0.5, 200*time.Millisecond),
		echo.New(0.1, 50*time.Millisecond),
	})
	if err != nil && err != realaudio.ErrDone {
		log.Fatal(err)
	}
}

func sin(x float32) float32 { return float32(math.Sin(float64(x))) }
func cos(x float32) float32 { return float32(math.Cos(float64(x))) }

const (
	tau = 2 * math.Pi
)
