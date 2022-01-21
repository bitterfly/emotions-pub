package emotions

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type EegClusterable struct {
	Class string      `json:"class"`
	Data  [][]float64 `json:"data"` //19x4
}

var waveRanges = [][2]float64{
	[2]float64{4.0, 8.0},   // δ
	[2]float64{8.0, 12.0},  // α
	[2]float64{12.0, 30.0}, // β
	[2]float64{30.0, 50.0}, // γ
}

func getRange(n float64) int {
	if n > waveRanges[len(waveRanges)-1][1] {
		return -1
	}

	if n < waveRanges[0][0] {
		return -1
	}

	for i, w := range waveRanges {
		if n > w[0] && n < w[1] {
			return i
		}
	}
	return -1
}

func getVector(line []string) []float64 {
	floatValues := make([]float64, 0, len(line))

	for _, s := range line {
		if strings.TrimSpace(s) == "" {
			continue
		}

		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(err)
		}

		floatValues = append(floatValues, v)
	}
	return floatValues
}

// ReadXML takes an xml file with eeg readings and returns a vector for each electrode in time
// where the first coordinate is the data from the first electrode and so on
func ReadXML(filename string, elNum int) [][]float64 {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("%s: %s", filename, err))
	}

	electrodes := make([][]float64, elNum, elNum)
	scanner := csv.NewReader(file)
	scanner.Comma = ' '

	for {
		line, err := scanner.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(fmt.Sprintf("%s %s", filename, err))
			}
		}

		for i, value := range getVector(line) {
			electrodes[i] = append(electrodes[i], value+rand.NormFloat64()*0.0001)
		}
	}

	return electrodes
}

func cutElectrodeIntoFrames(electrode []float64, frameLen int, frameStep int, verbose bool) [][]float64 {
	return CutSliceIntoFrames(electrode, 500, frameLen, frameStep, verbose)
}

func fourierElectrode(frames [][]float64) [][]Complex {
	fouriers := make([][]Complex, len(frames), len(frames))
	for i := 0; i < len(frames); i++ {
		fouriers[i], _ = FftReal(frames[i])
	}

	return fouriers
}

// getSignificantFreq takes fourier coefficients for each frame for an electrode
// and returns an array frameNum x 4 in which the alpha, beta, gamma, theta accumulated powers are stored
func getSignificantFreq(coefficients [][]Complex) [][]float64 {
	sFreq := make([][]float64, len(coefficients), len(coefficients))

	for i := 0; i < len(coefficients); i++ {
		sFreq[i] = make([]float64, 4, 4)
		for j := 0; j < len(coefficients[i]); j++ {
			power := Power(coefficients[i][j])
			w := getRange(IndToFreq(j, 500, len(coefficients[0])))
			if w == -1 {
				continue
			}
			sFreq[i][w] += power
		}
	}

	for _, s := range sFreq {
		for j := 0; j < len(s); j++ {
			s[j] = math.Log(s[j])
		}
	}

	return sFreq
}

func getWavesMean(coefficients [][]Complex) []float64 {
	means := make([]float64, len(waveRanges), len(waveRanges))
	for i := 0; i < len(coefficients); i++ {
		for j := 0; j < len(coefficients[0]); j++ {

			power := Power(coefficients[i][j])
			w := getRange(IndToFreq(j, 500, len(coefficients[0])))
			if w == -1 {
				continue
			}
			means[w] += power
		}
	}

	divide(&means, float64(len(coefficients)))
	return means
}

func getElectrodeWavesDistribution(electrodeData []float64, frameLen int, frameStep int) []float64 {
	frames := cutElectrodeIntoFrames(electrodeData, frameLen, frameStep, false)
	fouriers := fourierElectrode(frames)
	return getWavesMean(fouriers)
}

// GetFeatureVector returns the mean of Θ, α, β and γ waves for each of the given elNum electrodes
// returns a vector 19x4
func GetFeatureVector(filename string, elNum int, frameLen int, frameStep int) [][]float64 {
	data := ReadXML(filename, elNum)
	features := make([][]float64, len(data), len(data))
	for i, d := range data {
		features[i] = getElectrodeWavesDistribution(d, frameLen, frameStep)
	}

	return features
}

func getFeaturesFromFiles(filenames []string, frameLen int, frameStep int) []EegClusterable {
	trainingSet := make([]EegClusterable, len(filenames), len(filenames))

	for i, file := range filenames {
		filename := filepath.Base(file)
		name := filename[0 : len(filename)-len(filepath.Ext(filename))]
		newFeatures := GetFeatureVector(file, 19, frameLen, frameStep)
		trainingSet[i] = EegClusterable{
			Class: name,
			Data:  newFeatures,
		}
	}

	return trainingSet
}

func SaveEegTrainingSet(filenames []string, outputFilename string, frameLen int, frameStep int) {
	features := getFeaturesFromFiles(filenames, frameLen, frameStep)

	bytes, err := json.Marshal(features)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(outputFilename, bytes, 0644)
}

func getEegTrainingSet(filename string) []EegClusterable {
	var clusterables []EegClusterable
	bytes, _ := ioutil.ReadFile(filename)
	err := json.Unmarshal(bytes, &clusterables)
	if err != nil {
		panic(err)
	}

	return clusterables
}

// GetFourierForFile takes a filename and numbers of electrodes and returns the fourier transform of each electrode
func GetFourierForFile(filename string, elNum int, frameLen int, frameStep int) [][]float64 {
	data := ReadXML(filename, elNum)
	return getFourier(data, frameLen, frameStep)
}

func putSign(sign string, content []string) []string {
	newContent := make([]string, len(content), len(content))

	for i := range content {
		newContent[i] = fmt.Sprintf("%s %s", sign, content[i])
	}

	return newContent
}

func readEEGfiles(filenames []string, frameLen int, frameStep int) []string {
	content := make([]string, 0, 1000)
	for _, filename := range filenames {
		cbf := GetFourierForFile(filename, 19, frameLen, frameStep)

		for _, c := range cbf {
			if !IsZero(c) {
				current := ""
				for i, cc := range c {
					current += fmt.Sprintf("%d:%f ", i+1, cc)
				}
				content = append(content, fmt.Sprintf("%s\n", current))

			}
		}
	}
	return content
}

func writeToFile(filename string, content []string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	for _, l := range content {
		fmt.Fprintf(f, l)
	}
	return nil
}

// TrainEeg takes positive, negative and neutral eeg files and trains the svm models for each against the other
func TrainEeg(eegPositiveFiles []string, eegNegativeFiles []string, eegNeutralFiles []string, outputDir string, frameLen int, frameStep int) error {
	positive := readEEGfiles(eegPositiveFiles, frameLen, frameStep)
	negative := readEEGfiles(eegNegativeFiles, frameLen, frameStep)
	neutral := readEEGfiles(eegNeutralFiles, frameLen, frameStep)

	// +1 neutral -1 positive + negative
	neutralNP := filepath.Join(outputDir, "neutral_np.txt")
	neutralNPmodel := filepath.Join(outputDir, "neutral_np.model")
	err := writeToFile(neutralNP, combineSlices(
		putSign("-1", combineSlices(negative, positive)),
		putSign("+1", neutral),
	))
	if err != nil {
		return err
	}

	// +1 positive -1 (neutral + negative)
	positiveNN := filepath.Join(outputDir, "positive_nn.txt")
	positiveNNmodel := filepath.Join(outputDir, "positive_nn.model")
	err = writeToFile(positiveNN, combineSlices(
		putSign("-1", combineSlices(negative, neutral)),
		putSign("+1", positive),
	))
	if err != nil {
		return err
	}

	// +1 negative -1 (neutral + positive)
	negativeNP := filepath.Join(outputDir, "negative_np.txt")
	negativeNPmodel := filepath.Join(outputDir, "negative_np.model")

	err = writeToFile(negativeNP, combineSlices(
		putSign("-1", combineSlices(neutral, positive)),
		putSign("+1", negative),
	))
	if err != nil {
		return err
	}

	err = exec.Command("svm_learn", negativeNP, negativeNPmodel).Run()
	if err != nil {
		return fmt.Errorf("could not train svm for file %s: %s", negativeNP, err.Error())
	}
	err = exec.Command("svm_learn", positiveNN, positiveNNmodel).Run()
	if err != nil {
		return fmt.Errorf("could not train svm for file %s: %s", positiveNN, err.Error())
	}
	err = exec.Command("svm_learn", neutralNP, neutralNPmodel).Run()
	if err != nil {
		return fmt.Errorf("could not train svm for file %s: %s", neutralNP, err.Error())
	}

	return nil
}

// getFouriers takes the inverted data (19xlen(eeg))
// then cuts the data for each electrode into frames
// For each frames we compute Fourier coefficients, then we accumulate these coefficients within the wave ranges
// then we flip the result again, so we have the feature vectors which are numFrames x (numEl * 4)
func getFourier(data [][]float64, frameLen int, frameStep int) [][]float64 {
	// fmt.Fprintf(os.Stderr, fmt.Sprintf("Data: %d x %d\n", len(data), len(data[0])))
	// elFouriers is elNum x numFrames x 4(numWaves)
	elFouriers := make([]([][]float64), len(data), len(data))

	for i, d := range data {
		frames := cutElectrodeIntoFrames(d, frameLen, frameStep, false)

		fouriers := fourierElectrode(frames)
		elFouriers[i] = getSignificantFreq(fouriers)
	}

	// fmt.Printf("El fouriers: %d x %d x %d\n", len(elFouriers), len(elFouriers[0]), len(elFouriers[0][0]))

	// fourierByFrames stores for every frame the waves for each electrode
	// dim: numFrames x (numEl * 4)
	fourierByFrame := make([][]float64, len(elFouriers[0]), len(elFouriers[0]))
	for i := range elFouriers[0] {
		fourierByFrame[i] = make([]float64, 0, len(data)*len(elFouriers[0][0]))
	}

	for en := 0; en < len(elFouriers); en++ {
		for f := 0; f < len(elFouriers[en]); f++ {
			fourierByFrame[f] = append(fourierByFrame[f], elFouriers[en][f]...)
		}
	}

	// fmt.Printf("FouriersByFrame: %d x %d\n", len(fourierByFrame), len(fourierByFrame[0]))
	return fourierByFrame
}
