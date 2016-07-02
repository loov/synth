package realaudio

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ccherng/go-xaudio2"
)

const (
	defFormatTag        = 3 // WAVE_FORMAT_IEEE_FLOAT
	defChannels         = 2
	defSampleBits       = 32
	defSampleRate       = 44100
	defSamplesPerBuffer = 1024

	defBufferCount     = 2
	defBufferCountMask = 1
)

func Play(source Source) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var context *xa2.IXAudio2
	var master *xa2.IXAudio2MasteringVoice
	var voice *xa2.IXAudio2SourceVoice

	var buffers [defBufferCount]xa2.XAUDIO2_BUFFER
	var blocks [defBufferCount][defSamplesPerBuffer * defChannels]float32
	var empty [defSamplesPerBuffer * defChannels]float32

	format := xa2.WAVEFORMATEX{
		WFormatTag:      defFormatTag,
		NChannels:       defChannels,
		NSamplesPerSec:  defSampleRate,
		NAvgBytesPerSec: defSampleRate * defChannels * defSampleBits / 8,
		NBlockAlign:     defChannels * defSampleBits / 8,
		WBitsPerSample:  defSampleBits,
	}

	if ret := xa2.XAudio2Create(&context); ret != 0 {
		return fmt.Errorf("unable to create XAudio2 context: %v", ret)
	}
	defer context.Release()

	if ret := context.CreateMasteringVoice(&master, 0, 0, 0, 0, nil); ret != 0 {
		return fmt.Errorf("unable to create XAudio2 mastering voice: %v", ret)
	}
	if ret := context.CreateSourceVoice(&voice, &format, 0, 2.0, nil, nil, nil); ret != 0 {
		return fmt.Errorf("unable to create XAudio2 source voice: %v", ret)
	}

	for i := range buffers {
		buffer := &buffers[i]
		block := &blocks[i]
		buffer.AudioBytes = uint32(len(block) * 4)
		buffer.PAudioData = (*uint8)(unsafe.Pointer(&block[0]))
	}
	if err := source.Init(Format{SampleRate: defSampleRate, ChannelCount: defChannels}); err != nil {
		return err
	}
	if ret := voice.Start(0, 0); ret != 0 {
		return fmt.Errorf("unable to start XAudio2 voice: %v", ret)
	}

	var nextBlock int
	var state xa2.XAUDIO2_VOICE_STATE
	for {
		voice.GetState(&state)
		if state.BuffersQueued > 1 {
			continue
		}

		block := &blocks[nextBlock]
		buffer := &buffers[nextBlock]
		nextBlock = (nextBlock + 1) & defBufferCountMask

		copy(block[:], empty[:])
		if err := source.Render(block[:]); err != nil {
			return err
		}

		if ret := voice.SubmitSourceBuffer(buffer, nil); ret != 0 {
			return fmt.Errorf("submitting XAudio2 source buffer failed: %v", ret)
		}
	}
}
