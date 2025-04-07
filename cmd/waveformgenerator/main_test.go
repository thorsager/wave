package main

import "testing"

func Test_mix(t *testing.T) {
	type args struct {
		samples []int16
	}
	tests := []struct {
		name string
		args args
		want int16
	}{
		{"simple", args{[]int16{10, 20}}, 15},
		{"max", args{[]int16{32767, 32767}}, 32767},
		{"min", args{[]int16{-32768, -32768}}, -32768},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mix(tt.args.samples); got != tt.want {
				t.Errorf("mix() = %v, want %v", got, tt.want)
			}
		})
	}
}
