package sound
import (
	_ "embed"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"io"
	"bytes"
	"time"
)

type MemoryStreamer struct {
	Samples []float64
	Position int
}

//go:embed crash.mp3
var crashSoundData []byte

var MemoryAudio [44100]byte

func PlaySoundROM(soundName string) {
	if soundName == "crash" {
		streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(crashSoundData)))
		if err != nil {
			return
		}
		defer streamer.Close()

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {})))	
	}
}

