package echo

import (
	"errors"
	"time"

	"github.com/loov/synth/realaudio"
)

type Node struct {
	Dampening float32
	delay     time.Duration
	index     int
	buffer    []float32
}

func New(dampening float32, delay time.Duration) *Node {
	return &Node{
		delay:     delay,
		Dampening: dampening,
	}
}

func (node *Node) Init(format realaudio.Format) error {
	if format.ChannelCount != 2 {
		return errors.New("cannot handle non-stereo")
	}

	size := int(node.delay.Seconds() * float64(format.SampleRate))
	node.buffer = make([]float32, size*2)
	return nil
}

func (node *Node) Render(buffer []float32) error {
	index := node.index
	hist := node.buffer
	for i := 0; i < len(buffer); i += 2 {
		buffer[i] += 0.3 * hist[index+1]
		buffer[i+1] += 0.3 * hist[index]

		hist[index] = buffer[i]
		hist[index+1] = buffer[i+1]

		if index += 2; index >= len(hist) {
			index = 0
		}
	}
	node.index = index
	return nil
}
