package emotions

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
)

//WavFile implements this http://tiny.systems/software/soundProgrammer/WavFormatDocs.pdf
type WavFile struct {
	sampleRate uint32

	header []byte
	data   []float64
}

func (wf WavFile) GetSampleRate() int {
	return int(wf.sampleRate)
}

func (wf WavFile) GetLenInSeconds() float64 {
	return float64(len(wf.data)) / float64(wf.sampleRate)
}

//GetData returns the array with only the data part of the wav
func (wf WavFile) GetData() []float64 {
	return wf.data
}

// Read reads a wav file from a given filename into WavFile format
func Read(filename string, ditherCoefficient float64, preemphasisCoefficient float64) (WavFile, error) {
	reader, err := readFile(filename)
	if err != nil {
		return WavFile{}, err
	}

	wf, err := readContent(reader, preemphasisCoefficient)

	for i := 0; i < len(wf.data); i++ {
		wf.data[i] += rand.NormFloat64() * ditherCoefficient
	}

	if err != nil {
		return WavFile{}, err
	}

	return wf, nil
}

func readFile(filename string) (io.Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func readContent(reader io.Reader, preemphasisCoefficient float64) (WavFile, error) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return WavFile{}, err
	}

	wavFile := WavFile{}
	wavFile.header = content[0:12]

	beginData := 0
	lenData := 0

	//first chunk begin
	i := 12
	for i < len(content) {
		chunkId := content[i : i+4]
		chunkLen := int(getUint32LittleEndian4Byte(content[i+4:i+8])) + 8

		if string(chunkId) == "fmt " {
			wavFile.header = append(wavFile.header, content[i:i+int(chunkLen)]...)
		}

		if string(chunkId) == "data" {
			beginData = i + 8
			lenData = int(getUint32LittleEndian4Byte(content[i+4 : i+8]))
			wavFile.header = append(wavFile.header, content[i:i+8]...)
		}

		i += chunkLen
	}

	setLittleEndian4ByteToInt(uint32(36+lenData), wavFile.header, 4)

	sampleRate := getUint32LittleEndian4Byte(wavFile.header[24:28])
	wavFile.sampleRate = sampleRate

	bitsPerSample := getUint16LittleEndian2Byte(wavFile.header[34:36])

	var endianFunc func([]byte) float64
	var max float64
	if bitsPerSample == uint16(16) {
		endianFunc = func(b []byte) float64 { return float64(int16(getUint16LittleEndian2Byte(b))) }
		max = float64(math.MaxInt16 + 1)
	} else if bitsPerSample == uint16(32) {
		endianFunc = func(b []byte) float64 { return float64(int32(getUint32LittleEndian4Byte(b))) }
		max = float64(math.MaxInt32 + 1)
	} else {
		return WavFile{}, fmt.Errorf("Unknown bitsPerSample: %d", bitsPerSample)
	}
	index := 0
	data := make([]float64, lenData/2, lenData/2)
	previous := 0.0
	for b := beginData; b < beginData+lenData; b += 2 {
		current := endianFunc(content[b:b+2]) / max
		data[index] = current - preemphasisCoefficient*previous
		previous = current
		index++
	}

	wavFile.data = data

	return wavFile, nil
}

func getUint32LittleEndian4Byte(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func getUint32BigEndian4Byte(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func getUint16LittleEndian2Byte(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func getUint16BigEndian2Byte(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}

func setLittleEndian2ByteFromFloat64(f float64, b []byte) {
	i := uint16(int16(f))
	b[0] = byte(i)
	b[1] = byte(i >> 8)
}

func setLittleEndian4ByteToInt(i uint32, input []byte, index int) {
	input[index+0] = byte(i)
	input[index+1] = byte(i >> 8)
	input[index+2] = byte(i >> 16)
	input[index+3] = byte(i >> 24)
}
