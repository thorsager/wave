package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const RiffChunkID uint32 = 0x52494646    // 'RIFF'
const WaveForamt uint32 = 0x57415645     // 'WAVE'
const FmtSubChunkID uint32 = 0x666d7420  // 'fmt '
const DataSubChunkID uint32 = 0x64617461 // 'data'
const ListSubChunkID uint32 = 0x4c495354 // 'LIST'
const InfoSubChunkID uint32 = 0x494e464f // 'INFO'

const PCMAudioFormat uint16 = 0x0001
const PCMSubChunkFMTSize uint32 = 0x0010

type Encoder struct {
	writeSeeker io.WriteSeeker

	audioFormat   uint16
	numChannels   uint16
	sampleRate    uint32
	byteRate      uint32
	blockAlign    uint16
	bitsPerSample uint16

	numberOfSamples uint32
	headerWritten   bool
	info            InfoList
}

func NewEncoder(ws io.WriteSeeker, audioFormat uint16, channels uint16, sampleRate uint32, bitsPerSample uint16) *Encoder {
	encoder := &Encoder{
		audioFormat:   audioFormat,
		numChannels:   channels,
		sampleRate:    sampleRate,
		bitsPerSample: bitsPerSample,
		writeSeeker:   ws,
		byteRate:      uint32(sampleRate*uint32(channels)*uint32(bitsPerSample)) / 8,
		blockAlign:    channels * bitsPerSample / 8,
	}
	return encoder
}

func (e *Encoder) AddInfo(info ...Info) {
	e.info = append(e.info, info...)
}

func (e *Encoder) Write(data []byte) (int, error) {
	if !e.headerWritten {
		err := e.writeHeader(0)
		if err != nil {
			return 0, err
		}
	}

	dataBitLength := len(data) * 8
	if dataBitLength%int(e.bitsPerSample*e.numChannels) != 0 {
		return 0, fmt.Errorf("data is not a multiple of the bits per sample times number of channels")
	}
	numberOfSamples := dataBitLength / int(e.bitsPerSample)

	n, err := e.writeSeeker.Write(data)
	if err != nil {
		return 0, err
	}
	e.numberOfSamples += uint32(numberOfSamples)
	return n, err
}

func (e *Encoder) WriteSample(s any) error {
	if !e.headerWritten {
		err := e.writeHeader(0)
		if err != nil {
			return err
		}
	}
	sampleLength := binary.Size(s) * 8
	if sampleLength != int(e.bitsPerSample) {
		return fmt.Errorf("invalid sample length: %d", sampleLength)
	}
	err := binary.Write(e.writeSeeker, binary.LittleEndian, s)
	if err != nil {
		return err
	}
	e.numberOfSamples++
	return nil
}

func (e *Encoder) Close() error {
	if e.numberOfSamples%uint32(e.numChannels) != 0 {
		return fmt.Errorf("invalid sample count %d for %d channels", e.numberOfSamples, e.numChannels)
	}
	ms, err := e.writeMeta()
	if err != nil {
		return err
	}
	err = e.writeHeader(ms)
	if err != nil {
		return err
	}
	switch e.writeSeeker.(type) {
	case *os.File:
		return e.writeSeeker.(*os.File).Sync()
	}
	return nil
}

func (e *Encoder) writeHeader(metaSize uint32) error {
	subChunk2Size := e.numberOfSamples * uint32(e.bitsPerSample) / 8
	chunkSize := 32 + subChunk2Size + metaSize

	if e.headerWritten {
		// seek and update chunkSize
		_, err := e.writeSeeker.Seek(4, io.SeekStart)
		if err != nil {
			return err
		}
		err = binary.Write(e.writeSeeker, binary.LittleEndian, chunkSize)
		if err != nil {
			return err
		}

		// seek and update subChunkSize
		_, err = e.writeSeeker.Seek(40, io.SeekStart)
		if err != nil {
			return err
		}
		err = binary.Write(e.writeSeeker, binary.LittleEndian, subChunk2Size)
		if err != nil {
			return err
		}

		// seek endOfFile
		_, err = e.writeSeeker.Seek(0, io.SeekEnd)
		return err
	}

	buffer := make([]byte, 44)
	binary.BigEndian.PutUint32(buffer[0:], RiffChunkID)
	binary.LittleEndian.PutUint32(buffer[4:], chunkSize)
	binary.BigEndian.PutUint32(buffer[8:], WaveForamt)

	binary.BigEndian.PutUint32(buffer[12:], FmtSubChunkID)
	binary.LittleEndian.PutUint32(buffer[16:], PCMSubChunkFMTSize)
	binary.LittleEndian.PutUint16(buffer[20:], PCMAudioFormat)
	binary.LittleEndian.PutUint16(buffer[22:], e.numChannels)
	binary.LittleEndian.PutUint32(buffer[24:], e.sampleRate)
	binary.LittleEndian.PutUint32(buffer[28:], e.byteRate)
	binary.LittleEndian.PutUint16(buffer[32:], e.blockAlign)
	binary.LittleEndian.PutUint16(buffer[34:], e.bitsPerSample)

	binary.BigEndian.PutUint32(buffer[36:], DataSubChunkID)
	binary.LittleEndian.PutUint32(buffer[40:], subChunk2Size)

	_, err := e.writeSeeker.Write(buffer)
	if err != nil {
		return err
	}
	e.headerWritten = true

	return nil
}
func (e *Encoder) writeMeta() (uint32, error) {
	if !e.info.Contains(MarkerISFT) {
		e.info = append(e.info, Info{MarkerISFT, "github.com/thorsager/wave"})
	}
	chunkData, err := encode(e.info)
	if err != nil {
		return 0, err
	}
	listBuff := bytes.NewBuffer(nil)
	err = binary.Write(listBuff, binary.BigEndian, ListSubChunkID)
	if err != nil {
		return 0, err
	}
	err = binary.Write(listBuff, binary.LittleEndian, uint32(len(chunkData)))
	if err != nil {
		return 0, err
	}
	err = binary.Write(listBuff, binary.BigEndian, chunkData)
	if err != nil {
		return 0, err
	}
	n, err := e.writeSeeker.Write(listBuff.Bytes())
	return uint32(n), err
}
