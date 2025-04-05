package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	eov uint8 = 0x00 // End of value
)

type Marker [4]byte

func (m Marker) Uint32() uint32 {
	return binary.BigEndian.Uint32(m[:])
}
func (m Marker) String() string {
	return fmt.Sprintf("%s", m[:])
}

func asMarker(i uint32) Marker {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, i)
	var marker Marker
	copy(marker[:], buf)
	return marker
}

func ParseMarker(s string) (Marker, error) {
	var m Marker
	if len(s) != 4 {
		return m, fmt.Errorf("invalid length '%s'", s)
	}
	copy(m[:], []byte(s))
	return m, nil
}

var (
	MarkerIART Marker = [4]byte{'I', 'A', 'R', 'T'}
	MarkerISFT Marker = [4]byte{'I', 'S', 'F', 'T'}
	MarkerICRD Marker = [4]byte{'I', 'C', 'R', 'D'}
	MarkerICOP Marker = [4]byte{'I', 'C', 'O', 'P'}
	MarkerIARL Marker = [4]byte{'I', 'A', 'R', 'L'}
	MarkerINAM Marker = [4]byte{'I', 'N', 'A', 'M'}
	MarkerIENG Marker = [4]byte{'I', 'E', 'N', 'G'}
	MarkerIGNR Marker = [4]byte{'I', 'G', 'N', 'R'}
	MarkerIPRD Marker = [4]byte{'I', 'P', 'R', 'D'}
	MarkerISRC Marker = [4]byte{'I', 'S', 'R', 'C'}
	MarkerISBJ Marker = [4]byte{'I', 'S', 'B', 'J'}
	MarkerICMT Marker = [4]byte{'I', 'C', 'M', 'T'}
	MarkerITRK Marker = [4]byte{'I', 'T', 'R', 'K'}
	MarkerITCH Marker = [4]byte{'I', 'T', 'C', 'H'}
	MarkerIKEY Marker = [4]byte{'I', 'K', 'E', 'Y'}
	MarkerIMED Marker = [4]byte{'I', 'M', 'E', 'D'}
)

type InfoList []Info

type Info struct {
	Marker Marker
	Value  string
}

func encode(il InfoList) ([]byte, error) {
	if len(il) == 0 {
		return nil, nil
	}
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.BigEndian, InfoSubChunkID)
	if err != nil {
		return nil, err
	}

	for _, i := range il {
		err = binary.Write(buf, binary.BigEndian, i.Marker)
		if err != nil {
			return nil, err
		}
		err = binary.Write(buf, binary.LittleEndian, uint32(len(i.Value)+1))
		if err != nil {
			return nil, err
		}
		err = binary.Write(buf, binary.BigEndian, append([]byte(i.Value), eov))
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func decode(data []byte) (InfoList, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("insufficient data")
	}
	left := len(data)
	if binary.BigEndian.Uint32(data[0:]) != InfoSubChunkID {
		return nil, fmt.Errorf("invalid sub-chunk-id")
	}
	buf := bytes.NewBuffer(data[4:])
	left -= 4

	var il InfoList
	for left >= 8 {
		ifo := Info{}
		err := binary.Read(buf, binary.BigEndian, ifo.Marker[:])
		if err != nil {
			return nil, err
		}
		left -= 4

		var l uint32
		err = binary.Read(buf, binary.LittleEndian, &l)
		if err != nil {
			return nil, err
		}
		left -= 4
		if l > uint32(left) {
			return nil, fmt.Errorf("invalid value length")
		}

		if l > 0 {
			vb := make([]byte, l)
			err = binary.Read(buf, binary.BigEndian, vb)
			if err != nil {
				return nil, err
			}
			if vb[len(vb)-1] != eov {
				return nil, fmt.Errorf("invalid value termination")
			}
			ifo.Value = string(vb[:len(vb)-1]) // drop eov
			left -= int(l)
		}
		il = append(il, ifo)
	}
	return il, nil
}

func (il InfoList) Contains(m Marker) bool {
	for _, e := range il {
		if e.Marker == m {
			return true
		}
	}
	return false
}
