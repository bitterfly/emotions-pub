package emotions

import (
	"fmt"
	"math"
	"os"
)

const FRAME_IN_MS = 25
const STEP_IN_MS = 10

func rectangular(x int, n int) float64 {
	return 1
}

func hanning(x int, n int) float64 {
	return 0.5 - 0.5*math.Cos(float64(x*2)*math.Pi/float64(n))
}

// so when x == n-1 -> 2pi
func hamming(x int, n int) float64 {
	return 0.54 - 0.46*math.Cos(float64(x*2)*math.Pi/float64(n))
}

func window(content []float64, windowFunction func(int, int) float64) {
	for x, y := range content {
		content[x] = y * windowFunction(x, len(content))
	}
}

func hammingWindow(content []float64) {
	for x, y := range content {
		content[x] = y * hamming(x, len(content))
	}
}

func hanningWindow(content []float64) {
	for x, y := range content {
		content[x] = y * hanning(x, len(content))
	}
}

func rectangularWindow(content []float64) {
	for x, y := range content {
		content[x] = y
	}
}

// CutSliceIntoFrames receives float array and cuts it into frames
// It takes approximately 25mS of real data and then pads it with zeroes to the first power of two
// It takes frames with len 25ms with step of 10ms and applies a window function to the frame (hamming)
func CutSliceIntoFrames(data []float64, sampleRate uint32, frameInMs int, stepInMs int, verbose bool) [][]float64 {
	realSamplesPerFrame := int((float64(frameInMs) / 1000.0) * float64(sampleRate))

	samplesPerFrame := FindClosestPower(int(realSamplesPerFrame))
	step := int((float64(stepInMs) / 1000.0) * float64(sampleRate))

	if verbose {
		fmt.Fprintf(os.Stderr, "Samples: %d\n", len(data))
		fmt.Fprintf(os.Stderr, "Real samples per frame for %dms: %d\n", frameInMs, realSamplesPerFrame)
		fmt.Fprintf(os.Stderr, "Samples per frame: %d\nStep: %d\n", samplesPerFrame, step)

		fmt.Fprintf(os.Stderr, "Which is %.3fms long\n", 1000.0*float64(samplesPerFrame)/float64(sampleRate))
	}

	numFrames := (len(data) - realSamplesPerFrame) / step

	if verbose {
		fmt.Fprintf(os.Stderr, "Frames in file: %d\n====================\n", numFrames)
	}
	frames := make([][]float64, numFrames, numFrames)

	for frame := 0; frame < numFrames; frame++ {
		i := frame * step

		frames[frame] = sliceCopyWithWindow(data, i, i+realSamplesPerFrame, samplesPerFrame)
	}
	return frames
}

func PutWindow(data []float64, win string) []float64 {
	second := make([]float64, len(data), len(data))
	copy(second, data)
	if win == "rec" {
		rectangularWindow(second)
	} else if win == "han" {
		hanningWindow(second)
	} else {
		hammingWindow(second)
	}
	return second
}

func sliceCopyWithWindow(first []float64, from, to, length int) []float64 {
	second := make([]float64, length, length)
	copy(second, first[from:Min(to, len(first))])
	hammingWindow(second[0 : Min(to, len(first))-from])
	return second
}

//CutWavFileIntoFrames takes a wavfiles and cuts it into frames
func CutWavFileIntoFrames(wf WavFile) [][]float64 {
	return CutSliceIntoFrames(wf.data, wf.sampleRate, FRAME_IN_MS, STEP_IN_MS, false)
}
