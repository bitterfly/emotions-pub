package emotions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// FindClosestPower finds the closest power of two above the given number
func FindClosestPower(x int) int {
	p := 1

	for p < x {
		p *= 2
	}

	return p
}

//IsPowerOfTwo returns whether a number is a power of two
func IsPowerOfTwo(x int) bool {
	return (x & (x - 1)) == 0
}

// PrintCoefficients prints only those fourier coefficients that are greater than ɛ
func PrintCoefficients(coefficients []Complex) {
	fmt.Printf("Number of coefficients: %d\n", len(coefficients))
	for i, c := range coefficients {
		if Magnitude(c) > EPS {
			fmt.Printf("%d: %s\n", i, c)
		}
	}
}

// PlotSignal draws a graph of the given signal and saves it into a file
func PlotSignal(data []float64, file string) {
	plots, err := plot.New()
	if err != nil {
		panic(err)
	}

	ymin := data[0]
	ymax := data[0]
	s := make(plotter.XYs, len(data))
	for i := 0; i < len(data); i++ {
		s[i].X = float64(i)
		// s[i].X = math.Log(float64(i) + 1)
		s[i].Y = data[i]
		if s[i].Y > ymax {
			ymax = s[i].Y
		}
		if s[i].Y < ymin {
			ymin = s[i].Y
		}
		// s[i].Y = math.Log(data[i])
	}

	line, _ := plotter.NewLine(s)
	line.Color = color.RGBA{0, 100, 88, 255}
	line.Width = vg.Points(10)

	plots.Y.Max = ymax + 2.0
	plots.Y.Min = ymin - 2.0
	plots.Add(line)
	plots.HideAxes()

	if err := plots.Save(64*vg.Inch, 32*vg.Inch, file); err != nil {
		panic(err)
	}
}

// PlotSignal draws a graph of the given signal and saves it into a file
func PlotSignal2(c []Complex, data []float64, file string) {
	plots, err := plot.New()
	if err != nil {
		panic(err)
	}

	s := make(plotter.XYs, len(c))
	for i := 0; i < len(c); i++ {
		s[i].X = float64(i)
		// s[i].Y = math.Log(Power(c[i]))
		s[i].Y = Power(c[i])
	}

	line, _ := plotter.NewLine(s)
	line.Color = color.RGBA{0, 100, 88, 255}
	line.Width = vg.Points(20)

	plots.Add(line)

	s = make(plotter.XYs, len(data))
	for i := 0; i < len(data); i++ {
		s[i].X = float64(i)
		s[i].Y = data[i]
	}
	line2, _ := plotter.NewLine(s)
	// exp.Width = vg.Points(2)
	line2.Width = vg.Points(40)
	line2.Color = color.RGBA{255, 60, 88, 255}

	plots.Add(line2)

	if err := plots.Save(64*vg.Inch, 32*vg.Inch, file); err != nil {
		panic(err)
	}
}

func PlotClusters(data []MfccClusterisable, k int, file string) {
	plots, err := plot.New()
	if err != nil {
		panic(err)
	}

	s := make(plotter.XYs, len(data))
	clusters := make([]string, len(data))

	for i, d := range data {
		s[i].X = d.coefficients[0]
		s[i].Y = d.coefficients[1]
		clusters[i] = fmt.Sprintf("%d", d.clusterID)
	}

	scatter, _ := plotter.NewScatter(s)
	palette := palette.Heat(k, 1)

	scatter.GlyphStyleFunc = func(i int) draw.GlyphStyle {
		return draw.GlyphStyle{Color: palette.Colors()[data[i].clusterID], Radius: vg.Points(3), Shape: draw.CircleGlyph{}}
	}

	plots.Add(scatter)

	if err := plots.Save(32*vg.Inch, 16*vg.Inch, file); err != nil {
		panic(err)
	}
}

func PlotBarSignal(data []float64, file string) {
	v := make(plotter.Values, len(data))
	min := data[0]
	for i := range v {
		v[i] = data[i]
		if data[i] < min {
			min = data[i]
		}
	}

	for i := range v {
		v[i] -= min
	}

	bars, err := plotter.NewBarChart(v, 2)
	if err != nil {
		panic(err)
	}

	bars.Color = color.RGBA{0, 100, 88, 255}
	bars.Width = vg.Points(10)

	plotc, err := plot.New()
	if err != nil {
		panic(err)
	}
	plotc.X.Min = 0
	plotc.X.Max = float64(len(data))
	plotc.HideAxes()

	plotc.X.Label.Text = "Frequency"
	plotc.Y.Label.Text = "Log(Energy)"
	plotc.Y.Label.Font.Size = 124
	plotc.X.Label.Font.Size = 124

	plotc.Add(bars)
	if err := plotc.Save(32*vg.Inch, 32*vg.Inch, file); err != nil {
		panic(err)
	}
}

// PlotCoefficients draws a bar plot of the fourier coefficients and saves it into a file
func PlotCoefficients(coefficients []Complex, file string) {
	v := make(plotter.Values, len(coefficients))
	max := 0.0
	j := 0
	for i := range v {
		v[i] = Magnitude(coefficients[i])
		if v[i]-max > EPS {
			max = v[i]
			j = i
		}
	}

	fmt.Printf("%d: %f\n", j, max)

	plotc, err := plot.New()
	if err != nil {
		panic(err)
	}

	plotc.X.Min = 0
	plotc.X.Max = float64(len(coefficients))

	plotc.X.Label.Text = "Frequency"
	plotc.Y.Label.Text = "Energy"

	bars, err := plotter.NewBarChart(v, 2)

	plotc.Add(bars)

	if err := plotc.Save(16*vg.Inch, 16*vg.Inch, file); err != nil {
		panic(err)
	}

}

func PrintFrameSlice(frames [][]float64) {
	for i, frame := range frames {
		fmt.Printf("%d\n", i)
		PlotSignal(frame, fmt.Sprintf("signal/signal%d.png", i))
		c, _ := FftReal(frame)
		PlotCoefficients(c, fmt.Sprintf("spectrum/spectrum%d.png", i))

	}
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sliceCopy(first []float64, from, to, length int) []float64 {
	second := make([]float64, length, length)
	copy(second, first[from:Min(to, len(first))])
	hanningWindow(second[0 : Min(to, len(first))-from])
	return second
}

func PlotEeg(filename string, output string) {
	ts := getEegTrainingSet(filename)
	plotEeg(ts, output)
}

func plotEeg(data []EegClusterable, file string) {
	fmt.Printf("data: %d %d %d\n", len(data), len(data[0].Data), len(data[0].Data[0]))

	plots, err := plot.New()
	if err != nil {
		panic(err)
	}

	maximums := make([]float64, len(data[0].Data), len(data[0].Data))
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i].Data); j++ {
			if data[i].Data[j][0] > maximums[0] {
				maximums[0] = data[i].Data[j][0]
			}

			if data[i].Data[j][1] > maximums[1] {
				maximums[1] = data[i].Data[j][1]
			}

			if data[i].Data[j][2] > maximums[2] {
				maximums[2] = data[i].Data[j][2]
			}

			if data[i].Data[j][3] > maximums[3] {
				maximums[3] = data[i].Data[j][3]
			}
		}
	}

	fmt.Printf("len: %d\n", len(data))

	for i := 0; i < len(data); i++ {
		s := make(plotter.XYs, len(data[i].Data), len(data[i].Data))
		for j := 0; j < len(data[i].Data); j++ {
			s[j].X = float64(j)
			s[j].Y = float64(i)
		}

		scatter, _ := plotter.NewScatter(s)
		bla := data[i]
		scatter.GlyphStyleFunc = func(k int) draw.GlyphStyle {
			return draw.GlyphStyle{
				Color:  getColour(bla.Data[k], maximums),
				Radius: vg.Points(50),
				Shape:  draw.BoxGlyph{},
			}
		}
		plots.Add(scatter)
	}

	err = plots.Save(vg.Length(2.0*50*float64(len(data[0].Data))), vg.Length(2*50*float64(len(data))), file)
	if err != nil {
		panic(err)
	}
}

func PlotEmotion(filename string, output string) {
	data := ReadXML(filename, 19)

	var cl []EegClusterable

	for i, d := range data {
		var features [][]float64
		frames := cutElectrodeIntoFrames(d, 200, 150, false)
		fouriers := fourierElectrode(frames)
		for _, f := range fouriers {
			v := make([]float64, 4, 4)
			for _, ff := range f {
				magnitude := Magnitude(ff)
				w := getRange(magnitude)
				if w == -1 {
					break
				}
				v[w] = magnitude
			}

			features = append(features, v)
		}

		for j := 0; j < len(features); j++ {
			if j > len(cl)-1 {
				cl = append(cl, EegClusterable{
					Data: make([][]float64, len(data), len(data)),
				})
			}

			cl[j].Data[i] = features[j]
		}
	}

	plotEeg(cl, output)
}

func getColour(x []float64, maximums []float64) color.RGBA {
	r := uint8(math.Min(math.Floor(x[0]*float64(255)/maximums[0]), 255))
	g := uint8(math.Min(math.Floor(x[1]*float64(255)/maximums[1]), 255))
	b := uint8(math.Min(math.Floor(x[2]*float64(255)/maximums[2]), 255))
	a := uint8(math.Min(math.Floor(x[3]*float64(155)/maximums[3]+100), 255))

	// if r == 255 || g == 255 || b == 255 || a == 255 {
	// fmt.Printf("x: %v, r: %d g: %d b: %d a: %d\n", x, r, g, b, a)
	// }

	return color.RGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}
}

func getName(s string) string {
	if rune(s[1]) == '-' {
		return s[2:]
	}

	switch rune(s[1]) {
	case 'h':
		return "happiness"
	case 's':
		return "sadness"
	case 'a':
		return "anger"
	case 'n':
		return "neutral"
	case 'p':
		return "positive"
	case 'm':
		return "negative"
	default:
		panic(s)
	}
}

//IsZero checks if float is small enough
func IsZero(x []float64) bool {
	for _, xx := range x {
		if xx > 0.00001 || xx > -0.00001 {
			return false
		}
	}
	return true
}

//ParseArguments receives command arguments and separates them on spaces
func ParseArguments(args []string) map[string][]string {
	var arguments []string
	var emotion string
	emotions := make(map[string][]string)

	for i := 0; i < len(args); i++ {
		if rune(args[i][0]) == '-' {
			if len(arguments) != 0 {
				emotions[emotion] = arguments
				arguments = []string{}
			}
			emotion = getName(args[i])
		} else {
			arguments = append(arguments, args[i])
		}
	}

	emotions[emotion] = arguments

	return emotions
}

//ParseArgumentsFromFile takes a txt file in the format
//<emotion>\t<wav-file>(\t<eeg-gile>)
//and returns dictionaries with the file for each emotion (the keys are the emotions)
func ParseArgumentsFromFile(inputFilename string, multiple bool) (map[string][]string, map[string][]string, []string, error) {
	firstFiles := make(map[string][]string)
	secondFiles := make(map[string][]string)
	emotionTags := make([]string, 0, 100)

	file, err := os.Open(inputFilename)
	if err != nil {
		return nil, nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "\t")
		emotionTags = append(emotionTags, line[0])
		firstFiles[line[0]] = append(firstFiles[line[0]], line[1])
		if multiple {
			secondFiles[line[0]] = append(secondFiles[line[0]], line[2])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}

	return firstFiles, secondFiles, emotionTags, nil
}

func GetAverage(bucketSize int, frameLen int, arrayLen int) int {
	if bucketSize == 1 {
		return arrayLen
	}

	if bucketSize == 0 || bucketSize <= frameLen {
		return 1
	}

	return bucketSize / frameLen
}

func SortKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func SortKeysS(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func GetAlphaEGMs(dirname string) ([]AlphaEGM, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	egms := make([]AlphaEGM, len(files), len(files))
	for i, f := range files {
		bytes, _ := ioutil.ReadFile(filepath.Join(dirname, f.Name()))
		err := json.Unmarshal(bytes, &egms[i])
		if err != nil {
			return nil, err
		}
	}
	return egms, nil
}

func GetEGMs(dirname string) ([]EmotionGausianMixure, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	egms := make([]EmotionGausianMixure, len(files), len(files))
	for i, f := range files {
		bytes, _ := ioutil.ReadFile(filepath.Join(dirname, f.Name()))
		err := json.Unmarshal(bytes, &egms[i])
		if err != nil {
			return nil, err
		}
	}

	return egms, nil
}

func GetMagnitute(c []Complex, n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < len(data); i++ {
		data[i] = Magnitude(c[i])
	}
	return data
}

func GetPower(c []Complex, n int) []float64 {
	data := make([]float64, n, n)
	for i := 0; i < len(data); i++ {
		data[i] = math.Log(Power(c[i]))
	}
	return data
}

func ReadSpeechFeaturesOne(filename string) [][]float64 {
	wf, _ := Read(filename, 0.01, 0.97)
	return MFCCs(wf, 13, 23)
}

func ReadSpeechFeatures(filenames []string) [][]float64 {
	mfccs := make([][]float64, 0, len(filenames)*100)
	for _, f := range filenames {
		wf, _ := Read(f, 0.01, 0.97)

		mfcc := MFCCs(wf, 13, 23)
		mfccs = append(mfccs, mfcc...)
	}

	return mfccs
}

func ReadSpeechFeaturesAppend(filenames []string, features *[]([][]float64)) [][]float64 {
	mfccs := make([][]float64, 0, len(filenames)*100)
	for _, f := range filenames {
		wf, _ := Read(f, 0.01, 0.97)

		mfcc := MFCCs(wf, 13, 23)
		mfccs = append(mfccs, mfcc...)
		*features = append(*features, mfcc)
	}

	return mfccs
}
