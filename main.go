package main

import (
	"fmt"
	"math"
	"syscall"
	"time"
	"unsafe"

	//	"golang.org/x/sys/windows"

	"github.com/ccherng/go-xaudio2"
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

const (
	FormatTag        = 3 // WAVE_FORMAT_IEEE_FLOAT
	Channels         = 2
	SampleBits       = 32
	SampleRate       = 44100
	SamplesPerBuffer = 2048
)

func encode(x Keyboard) string {
	for i := 0; i < 256/8; i++ {
		// v := x[i]&1 | x[i]&1
	}
	return ""
}

func main() {
	var context *xa2.IXAudio2
	xa2.XAudio2Create(&context)

	var master *xa2.IXAudio2MasteringVoice
	context.CreateMasteringVoice(&master, 0, 0, 0, 0, nil)

	format := xa2.WAVEFORMATEX{
		WFormatTag:      FormatTag,
		NChannels:       Channels,
		NSamplesPerSec:  SampleRate,
		NAvgBytesPerSec: SampleRate * Channels * SampleBits / 8,
		NBlockAlign:     Channels * SampleBits / 8,
		WBitsPerSample:  SampleBits,
	}

	const N = 2

	var history struct {
		Value [2 * 2 * SampleRate]float32
		Index int
	}

	var blocks [N][SamplesPerBuffer * Channels]float32
	var emptyBlock [SamplesPerBuffer * Channels]float32
	var buffers [N]xa2.XAUDIO2_BUFFER
	var last = 0

	for i := range blocks {
		block := &blocks[i]
		buffer := &buffers[i]

		buffer.AudioBytes = uint32(len(block)) * 2
		buffer.PAudioData = (*uint8)(unsafe.Pointer(&block[0]))
	}

	var voice *xa2.IXAudio2SourceVoice
	var state xa2.XAUDIO2_VOICE_STATE

	context.CreateSourceVoice(&voice, &format, 0, 2.0, nil, nil, nil)
	voice.Start(0, 0)

	type Note struct {
		Volume float32
		P0, P1 float32
		S0, S1 float32
	}

	var notes []Note

	var keys Keyboard
	for {
		if KeyDown(0x1b) {
			return
		}
		keys.Update()

		for i, key := range Scale {
			if keys.Pressed(byte(key)) {
				f0 := float32(220 * math.Exp2(1+float64(i)/12))
				f1 := float32(220*math.Exp2(1+float64(i)/12) + 5)
				notes = append(notes, Note{
					Volume: 0.7,
					P0:     0, P1: 0,
					S0: f0 * tau / SampleRate, S1: f1 * tau / SampleRate,
				})
			}
		}

		if keys.Pressed('4') {
			fmt.Println("ACTIVE ", len(notes))
			for _, note := range notes {
				fmt.Println("  - ", note)
			}
		}

		voice.GetState(&state)
		if state.BuffersQueued > 1 {
			continue
		}

		block := &blocks[last]
		buffer := &buffers[last]
		last = last + 1
		if last >= N {
			last = 0
		}

		copy(block[:], emptyBlock[:])

		next := make([]Note, 0, len(notes))
		for _, note := range notes {
			for i := 0; i < SamplesPerBuffer; i += 2 {
				note.P0 += note.S0
				note.P1 += note.S1

				v0, v1 := sin(note.P0), sin(note.P1)
				block[i] += note.Volume * v0
				block[i+1] += note.Volume * v1
				note.Volume *= 0.9997
			}
			if note.Volume > 0.001 {
				next = append(next, note)
			}
		}
		notes = next

		k := (history.Index - SampleRate/3) / 2 * 2
		if k < 0 {
			k += len(history.Value)
		}

		for i := 0; i < SamplesPerBuffer; i += 2 {
			block[i] += 0.3 * history.Value[k+1]
			block[i+1] += 0.3 * history.Value[k]
			if k += 2; k >= len(history.Value) {
				k = 0
			}

			history.Value[history.Index] = block[i]
			history.Value[history.Index+1] = block[i+1]
			if history.Index += 2; history.Index >= len(history.Value) {
				history.Index = 0
			}
		}

		voice.GetState(&state)
		if state.BuffersQueued == 0 {
			fmt.Println("STARVED")
		}

		//start := qpc.Now()
		voice.SubmitSourceBuffer(buffer, nil)
		//stop := qpc.Now()
		//fmt.Println(stop.Sub(start).Nanoseconds())
	}

	time.Sleep(time.Second * 2)
}

func sin(x float32) float32 { return float32(math.Sin(float64(x))) }
func cos(x float32) float32 { return float32(math.Cos(float64(x))) }

const (
	tau = 2 * math.Pi
)
