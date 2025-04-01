package wave

import (
	"bytes"
	"encoding/binary"
)

const (
	eov uint8 = 0x00 // End of value
)

type Marker [4]byte

var (
	MarkerIART    Marker = [4]byte{'I', 'A', 'R', 'T'}
	MarkerISFT    Marker = [4]byte{'I', 'S', 'F', 'T'}
	MarkerICRD    Marker = [4]byte{'I', 'C', 'R', 'D'}
	MarkerICOP    Marker = [4]byte{'I', 'C', 'O', 'P'}
	MarkerIARL    Marker = [4]byte{'I', 'A', 'R', 'L'}
	MarkerINAM    Marker = [4]byte{'I', 'N', 'A', 'M'}
	MarkerIENG    Marker = [4]byte{'I', 'E', 'N', 'G'}
	MarkerIGNR    Marker = [4]byte{'I', 'G', 'N', 'R'}
	MarkerIPRD    Marker = [4]byte{'I', 'P', 'R', 'D'}
	MarkerISRC    Marker = [4]byte{'I', 'S', 'R', 'C'}
	MarkerISBJ    Marker = [4]byte{'I', 'S', 'B', 'J'}
	MarkerICMT    Marker = [4]byte{'I', 'C', 'M', 'T'}
	MarkerITRK    Marker = [4]byte{'I', 'T', 'R', 'K'}
	MarkerITRKBug Marker = [4]byte{'i', 't', 'r', 'k'}
	MarkerITCH    Marker = [4]byte{'I', 'T', 'C', 'H'}
	MarkerIKEY    Marker = [4]byte{'I', 'K', 'E', 'Y'}
	MarkerIMED    Marker = [4]byte{'I', 'M', 'E', 'D'}
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

func (il InfoList) Contains(m Marker) bool {
	for _, e := range il {
		if e.Marker == m {
			return true
		}
	}
	return false
}
