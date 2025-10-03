package audio

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"time"
)
var MemoryAudio [44100]byte

type PCMStreamer struct {	
	cursor int
}

func (s *PCMStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	for i := range samples {
		if s.cursor + 2 > len(MemoryAudio) {
			return i, false
		}

		v := float64(int(MemoryAudio[s.cursor])-128) / 128.0
		s.cursor++
		samples[i][0] = v
		samples[i][1] = v
	}
	return len(samples), true
}

func (s *PCMStreamer) Err() error { return nil }

func Play() {
	format := beep.Format{SampleRate: 44100, NumChannels: 2, Precision: 1}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second / 10))
	streamer := &PCMStreamer{}
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		
	})))
}
