package main

import (
	"log"
	"math"
	"syscall"
	"time"
	"unsafe"

	"github.com/loov/synth/module/echo"
	"github.com/loov/synth/module/synth"
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

func main() {
	synth := synth.NewBasic(0.9997)

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
