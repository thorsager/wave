package wave

import (
	"io"
	"reflect"
	"testing"
)

func TestNewEncoder(t *testing.T) {
	type args struct {
		ws            io.WriteSeeker
		audioFormat   uint16
		channels      uint16
		sampleRate    uint32
		bitsPerSample uint16
	}
	tests := []struct {
		name string
		args args
		want *Encoder
	}{
		{"one", args{nil, 1, 2, 8_000, 16}, &Encoder{
			writeSeeker: nil, audioFormat: 1, numChannels: 2, sampleRate: 8_000, bitsPerSample: 16,
			byteRate: 32_000, blockAlign: 4,
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEncoder(tt.args.ws, tt.args.audioFormat, tt.args.channels, tt.args.sampleRate, tt.args.bitsPerSample); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEncoder() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
