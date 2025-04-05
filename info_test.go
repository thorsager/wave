package wave

import (
	"reflect"
	"testing"
)

func Test_encode(t *testing.T) {
	type args struct {
		il InfoList
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Title", args{InfoList{{MarkerINAM, "fooo"}}},
			[]byte{
				0x49, 0x4E, 0x46, 0x4F,
				0x49, 0x4E, 0x41, 0x4D, 0x05, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x6F, 0x00,
			},
			false,
		},
		{"Title+Date", args{InfoList{{MarkerINAM, "fooo"}, {MarkerICRD, "1977-09-26 20:00:00"}}},
			[]byte{
				0x49, 0x4E, 0x46, 0x4F,
				0x49, 0x4E, 0x41, 0x4D, 0x05, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x6F, 0x00,
				0x49, 0x43, 0x52, 0x44, 0x14, 0x00, 0x00, 0x00, 0x31, 0x39, 0x37, 0x37, 0x2D, 0x30, 0x39, 0x2D,
				0x32, 0x36, 0x20, 0x32, 0x30, 0x3A, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x00,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encode(tt.args.il)
			if (err != nil) != tt.wantErr {
				t.Errorf("encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decode(t *testing.T) {
	type args struct {
		in0 []byte
	}
	tests := []struct {
		name    string
		args    args
		want    InfoList
		wantErr bool
	}{
		{"INAM", args{
			[]byte{
				0x49, 0x4E, 0x46, 0x4F, 0x49, 0x4E, 0x41, 0x4D, 0x05, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x6F, 0x00,
			}},
			InfoList{{MarkerINAM, "fooo"}},
			false,
		},
		{"shortVale", args{
			[]byte{
				0x49, 0x4E, 0x46, 0x4F, 0x49, 0x4E, 0x41, 0x4D, 0x05, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x6F,
			}},
			nil,
			true,
		},
		{"short", args{
			[]byte{
				0x49, 0x4E, 0x46, 0x4F, 0x49, 0x4E, 0x41, 0x4D, 0x00, 0x00, 0x00, 0x00,
			}},
			InfoList{{MarkerINAM, ""}},
			false,
		},
		{"INAM,INAM", args{
			[]byte{
				0x49, 0x4E, 0x46, 0x4F,
				0x49, 0x4E, 0x41, 0x4D, 0x05, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x6F, 0x00,
				0x49, 0x4E, 0x41, 0x4D, 0x04, 0x00, 0x00, 0x00, 0x6F, 0x6F, 0x6F, 0x00,
				0x49, 0x43, 0x52, 0x44, 0x14, 0x00, 0x00, 0x00, 0x31, 0x39, 0x37, 0x37, 0x2D, 0x30, 0x39, 0x2D,
				0x32, 0x36, 0x20, 0x32, 0x30, 0x3A, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x00,
			}},
			InfoList{
				{MarkerINAM, "fooo"},
				{MarkerINAM, "ooo"},
				{MarkerICRD, "1977-09-26 20:00:00"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decode(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_asMarker(t *testing.T) {
	type args struct {
		i uint32
	}
	tests := []struct {
		name string
		args args
		want Marker
	}{
		{"INAM", args{0x494E414D}, MarkerINAM},
		{"ICRD", args{0x49435244}, MarkerICRD},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asMarker(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("asMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarker_Uint32(t *testing.T) {
	tests := []struct {
		name string
		m    Marker
		want uint32
	}{
		{"INAM", MarkerINAM, 0x494E414D},
		{"ICRD", MarkerICRD, 0x49435244},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Uint32(); got != tt.want {
				t.Errorf("Marker.Uint32() = %X, want %X", got, tt.want)
			}
		})
	}
}

func TestMarker_String(t *testing.T) {
	tests := []struct {
		name string
		m    Marker
		want string
	}{
		{"INAM", MarkerINAM, "INAM"},
		{"ICRD", MarkerICRD, "ICRD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("Marker.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMarker(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Marker
		wantErr bool
	}{
		{"INAM", args{"INAM"}, MarkerINAM, false},
		{"ICRD", args{"ICRD"}, MarkerICRD, false},
		{"Invalid long", args{"INVALID"}, Marker{}, true},
		{"Invalid short", args{"I"}, Marker{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarker(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMarker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}
