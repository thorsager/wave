package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	wave "github.com/thorsager/wave"
)

var (
	flagOutput     string
	flagWaveforms  []string
	flagDuration   time.Duration
	flagSampleRate uint32
)

type waveform struct {
	channel   int
	frequency float64
	gain      float64
}
type waveformList [][]waveform

func (m waveformList) String() string {
	sb := strings.Builder{}
	for k, v := range m {
		freqs := "*none*"
		if v != nil {
			var fs []string
			for _, w := range v {
				fs = append(fs, fmt.Sprintf("%f(%f)", w.frequency, w.gain))
			}
			freqs = strings.Join(fs, ",")
		}
		sb.WriteString(fmt.Sprintf("%d: %s,", k, freqs))
	}
	return sb.String()
}

func main() {
	pflag.StringVarP(&flagOutput, "output", "o", "output.wav", "output filename")
	pflag.StringSliceVarP(&flagWaveforms, "waveform", "w", []string{"0:440:1"}, "waveform to generate format: '<channel>:<frequency>[:gain]'")
	pflag.DurationVarP(&flagDuration, "duration", "d", 5*time.Second, "duration of the waveform")
	pflag.Uint32VarP(&flagSampleRate, "sample-rate", "r", 8_000, "samples per second")
	pflag.Parse()

	numFrames := flagSampleRate * uint32(flagDuration.Seconds())

	wfm, err := parseWaveforms(flagWaveforms)
	if err != nil {
		log.Fatalf("error parsig waveforms: %s", err)
	}

	log.Printf("generating a %s waveform of %s (frames: %d)", flagDuration, wfm, numFrames)

	f, err := os.Create(flagOutput)
	if err != nil {
		log.Fatalf("error creating %s: %s", flagOutput, err)
	}
	defer f.Close()

	wavOut := wave.NewEncoder(f, wave.PCMAudioFormat, uint16(len(wfm)), flagSampleRate, 16)
	wavOut.AddInfo(wave.Info{
		Marker: wave.MarkerICRD,
		Value:  time.Now().UTC().Format(time.RFC1123),
	})
	defer func() {
		err = wavOut.Close()
		if err != nil {
			log.Fatalf("error closing writer: %s", err)
		}
	}()

	var channelNumbers []int

	for k := range wfm {
		channelNumbers = append(channelNumbers, k)
	}
	slices.Sort(channelNumbers)

	for i := range numFrames {
		for _, c := range channelNumbers { // range on channels
			wfl := wfm[c]
			var unmixed []int16
			if wfl == nil {
				unmixed = []int16{0} // all zero in empty channel
			} else {
				unmixed = make([]int16, len(wfl))
				for j, wf := range wfl { // range on waves pr. channel
					fv := math.Sin(float64(i)/float64(flagSampleRate)*wf.frequency*2*math.Pi) * wf.gain
					unmixed[j] = int16(fv * 32767)
				}
			}
			mixed := mix(unmixed)
			err = wavOut.WriteSample(mixed)
			if err != nil {
				log.Fatalf("error writing frame %d: %s", i, err)
			}
		}
	}
}

func parseWaveforms(wfs []string) (waveformList, error) {
	wfm := make(map[int][]waveform)

	mc := 0
	for _, s := range wfs {
		elms := strings.Split(s, ":")
		if len(elms) > 3 || len(elms) < 2 {
			return nil, fmt.Errorf("invalid waveform description: %s", s)
		}
		channel, err := strconv.Atoi(elms[0])
		if err != nil {
			return nil, fmt.Errorf("invalid channel '%s' in waveform description: %s", elms[0], s)
		}
		if channel < 0 {
			return nil, fmt.Errorf("invalid channel '%s' in waveform description (range 0-): %s", elms[0], s)
		}

		freq, err := strconv.ParseFloat(elms[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid frequency '%s' in waveform description: %s", elms[1], s)
		}
		if freq < 0 {
			return nil, fmt.Errorf("invalid frequency '%s' in waveform description (range 0-): %s", elms[1], s)
		}
		gain := float64(1)
		if len(elms) == 3 {
			gain, err = strconv.ParseFloat(elms[2], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid gain '%s' in waveform description: %s", elms[2], s)
			}
		}
		if gain < 0 || gain > 1 {
			return nil, fmt.Errorf("invalid gain '%s' in waveform description (range 0.0 - 1.0) : %s", elms[2], s)
		}

		wfm[channel] = append(wfm[channel], waveform{
			channel:   channel,
			frequency: freq,
			gain:      gain,
		})

		if channel > mc {
			mc = channel
		}
	}

	// pad out channels with nil waveforms
	wfl := make([][]waveform, mc+1)
	for i := 0; i <= mc; i++ {
		if w, ok := wfm[i]; !ok {
			wfl[i] = nil
		} else {
			wfl[i] = w
		}
	}

	return wfl, nil
}

func mix(samples []int16) int16 {
	if len(samples) == 1 {
		return samples[0]
	}
	var mixed int64
	for _, s := range samples {
		mixed += int64(s)
	}
	mixed /= int64(len(samples))

	// clip
	if mixed > 32767 {
		mixed = 32767
	} else if mixed < -32768 {
		mixed = -32768
	}
	return int16(mixed)
}
