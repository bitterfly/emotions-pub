package emotions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"sort"
)

type Tagged struct {
	Tag  string
	Data [][]float64
}

func UnmarshallKNNEeg(filename string) ([]Tagged, error) {
	var tagged []Tagged
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &tagged)
	if err != nil {
		return nil, err
	}

	return tagged, nil
}

func UnmarshallGMMEeg(filename string) ([]EmotionGausianMixure, error) {
	return nil, nil
}

func GetμAndσTagged(tagged []Tagged) ([]float64, []float64) {
	featureVectorSize := len(tagged[0].Data[0])
	// fmt.Fprintf(os.Stderr, "FeatureVectorSize: %d\n", featureVectorSize)

	variances := make([]float64, featureVectorSize, featureVectorSize)

	expectation := make([]float64, featureVectorSize, featureVectorSize)
	expectationSquared := make([]float64, featureVectorSize, featureVectorSize)

	for t := range tagged {
		for i := range tagged[t].Data {
			for j := range tagged[t].Data[i] {
				expectation[j] += tagged[t].Data[i][j]
				expectationSquared[j] += expectation[j] * expectation[j]
			}
		}
	}

	numVectors := 0
	for t := range tagged {
		numVectors += len(tagged[t].Data)
	}

	// fmt.Fprintf(os.Stderr, "numVectors: %d\n", numVectors)
	for j := 0; j < featureVectorSize; j++ {
		expectation[j] /= float64(numVectors)
		expectationSquared[j] /= float64(numVectors)

		variances[j] = expectationSquared[j] - expectation[j]*expectation[j]
	}

	// fmt.Fprintf(os.Stderr, "Expectations: %d\n", len(expectation))
	// fmt.Fprintf(os.Stderr, "Variances: %d\n", len(variances))
	return expectation, variances
}

func findClosest(v []float64, trainSet []Tagged, trainVar []float64) string {
	minDist := math.Inf(42)
	minDistTag := ""

	for t := range trainSet {
		for tv := range trainSet[t].Data {
			curDist := mahalanobisDistance(v, trainSet[t].Data[tv], trainVar)
			if curDist < minDist {
				minDist = curDist
				minDistTag = trainSet[t].Tag
			}
		}
	}

	return minDistTag
}

func testKNN(emotion string, emotions []string, vectors [][]float64, trainSet []Tagged, trainVar []float64) (int, int, int) {
	fmt.Printf("%s\t", emotion)

	counters := make(map[string]int)
	for _, e := range emotions {
		counters[e] = 0
	}

	for v := range vectors {
		counters[findClosest(vectors[v], trainSet, trainVar)]++
	}

	sum := 0
	for _, e := range emotions {
		fmt.Printf("%d\t", counters[e])
		sum += counters[e]
	}
	fmt.Printf("\n")

	return correct(emotion, counters), counters[emotion], sum
}

func ClassifyKNN(featureType string, trainSetFilename string, bucketSize int, frameLen int, frameStep int, emotionFiles map[string][]string) error {
	trainSet, err := UnmarshallKNNEeg(trainSetFilename)
	if err != nil {
		return err
	}
	_, trainVar := GetμAndσTagged(trainSet)

	fileKeys := make([]string, 0, len(emotionFiles))
	for k := range emotionFiles {
		fileKeys = append(fileKeys, k)
	}
	sort.Strings(fileKeys)

	correctFiles := make(map[string]int, len(fileKeys))
	correctVectors := make(map[string]int, len(fileKeys))
	sumVectors := make(map[string]int, len(fileKeys))

	for _, emotion := range fileKeys {
		for _, f := range emotionFiles[emotion] {
			fmt.Printf("%s\t", emotion)
			vec := GetFourierForFile(f, 19, frameLen, frameStep)
			average := GetAverage(bucketSize, frameStep, len(vec))
			averaged := AverageSlice(vec, average)

			boolCorrect, correctVector, sumVector := testKNN(emotion, fileKeys, averaged, trainSet, trainVar)
			correctFiles[emotion] += boolCorrect
			correctVectors[emotion] += correctVector
			sumVectors[emotion] += sumVector
		}
	}
	sort.Strings(fileKeys)
	fmt.Printf("\tCorrectFiles\tCorrectVectors\n")
	for _, emotion := range fileKeys {
		fmt.Printf("%s\t%f\t%f\n", emotion, float64(correctFiles[emotion])/float64(len(emotionFiles[emotion])), float64(correctVectors[emotion])/float64(sumVectors[emotion]))
	}

	return nil
}

func ClassifyGMM(featureType string, trainSetFilename string, bucketSize int, frameLen int, frameStep int, emotionFiles map[string][]string) error {
	trainSet, err := GetEGMs(trainSetFilename)
	if err != nil {
		return err
	}

	fileKeys := make([]string, 0, len(emotionFiles))
	for k := range emotionFiles {
		fileKeys = append(fileKeys, k)
	}

	correctFiles := make(map[string]int, len(fileKeys))
	correctVectors := make(map[string]int, len(fileKeys))
	sumVectors := make(map[string]int, len(fileKeys))

	sort.Strings(fileKeys)
	for _, emotion := range fileKeys {
		for _, f := range emotionFiles[emotion] {
			vec := GetFourierForFile(f, 19, frameLen, frameStep)
			average := GetAverage(bucketSize, frameStep, len(vec))

			var averaged [][]float64
			if featureType == "de" {
				averaged = GetDE(AverageSlice(vec, average))
			} else {
				averaged = AverageSlice(vec, average)
			}

			boolCorrect, vectors, sumVector := TestGMM(emotion, fileKeys, averaged, trainSet, true)
			correctFiles[emotion] += boolCorrect
			correctVectors[emotion] += vectors[emotion]
			sumVectors[emotion] += sumVector
		}
	}
	fmt.Printf("\tCorrectFiles\tCorrectVectors\n")
	for _, emotion := range fileKeys {
		fmt.Printf("%s\t%f\t%f\n", emotion, float64(correctFiles[emotion])/float64(len(emotionFiles[emotion])), float64(correctVectors[emotion])/float64(sumVectors[emotion]))
	}

	return nil
}

func ClassifyGMMConcat(trainSetFilename string, speechFiles map[string][]string, eegFiles map[string][]string) error {
	trainSet, err := GetEGMs(trainSetFilename)
	if err != nil {
		return err
	}

	fileKeys := make([]string, 0, len(speechFiles))
	for k := range speechFiles {
		fileKeys = append(fileKeys, k)
	}

	correctFiles := make(map[string]int, len(fileKeys))
	correctVectors := make(map[string]int, len(fileKeys))
	sumVectors := make(map[string]int, len(fileKeys))

	sort.Strings(fileKeys)
	for _, emotion := range fileKeys {
		for i := 0; i < len(speechFiles[emotion]); i++ {
			eegFeatures := GetFourierForFile(eegFiles[emotion][i], 19, 200, 150)
			allSpeech := ReadSpeechFeaturesOne(speechFiles[emotion][i])
			averaged := AverageSlice(allSpeech, len(allSpeech)/len(eegFeatures))
			speechFeatures := averaged[0 : len(averaged)-(len(averaged)-len(eegFeatures))]

			boolCorrect, vectors, sumVector := TestGMM(emotion, fileKeys, Concat(speechFeatures, eegFeatures), trainSet, true)
			correctFiles[emotion] += boolCorrect
			correctVectors[emotion] += vectors[emotion]
			sumVectors[emotion] += sumVector
		}
	}
	fmt.Printf("\tCorrectFiles\tCorrectVectors\n")
	for _, emotion := range fileKeys {
		fmt.Printf("%s\t%f\t%f\n", emotion, float64(correctFiles[emotion])/float64(len(speechFiles[emotion])), float64(correctVectors[emotion])/float64(sumVectors[emotion]))
	}

	return nil
}

func GetEegFeaturesForFile(bucketSize int, file string) [][]float64 {
	frameLen := 200
	frameStep := 150

	data := GetFourierForFile(file, 19, frameLen, frameStep)
	average := GetAverage(bucketSize, frameLen, len(data))
	return AverageSlice(data, average)
}

func GetSpeechFeatureForFile(filename string) [][]float64 {
	wf, _ := Read(filename, 0.01, 0.97)
	return MFCCs(wf, 13, 23)
}

func ClassifyGMMBoth(bucketSize int, frameLen int, frameStep int, speechTrainDir string, speechFiles map[string][]string, eegTrainDir string, eegFiles map[string][]string) error {
	speechAlphaTrainSet, err := GetAlphaEGMs(speechTrainDir)
	if err != nil {
		return err
	}

	speechTrainSet := make([]EmotionGausianMixure, len(speechAlphaTrainSet), len(speechAlphaTrainSet))
	for i := 0; i < len(speechAlphaTrainSet); i++ {
		speechTrainSet[i] = speechAlphaTrainSet[i].EGM
	}

	eegAlphaTrainSet, err := GetAlphaEGMs(eegTrainDir)
	if err != nil {
		return err
	}

	eegTrainSet := make([]EmotionGausianMixure, len(eegAlphaTrainSet), len(eegAlphaTrainSet))
	for i := 0; i < len(eegAlphaTrainSet); i++ {
		eegTrainSet[i] = eegAlphaTrainSet[i].EGM
	}

	fileKeys := make([]string, 0, len(speechFiles))
	for k := range speechFiles {
		fileKeys = append(fileKeys, k)
	}

	sort.Strings(fileKeys)
	speechAccuracy := make(map[string]int)
	EEGAccuracy := make(map[string]int)
	bothAccuracy := make(map[string]int)

	for _, emotion := range fileKeys {
		speechAccuracy[emotion] = 0
		EEGAccuracy[emotion] = 0
		bothAccuracy[emotion] = 0
	}

	for _, emotion := range fileKeys {
		for i := 0; i < len(speechFiles[emotion]); i++ {
			sC, eC, bC, bA := TestGMMBoth(emotion, fileKeys, speechAlphaTrainSet, speechTrainSet, speechFiles[emotion][i], eegAlphaTrainSet, eegTrainSet, eegFiles[emotion][i], bucketSize)

			speechAccuracy[emotion] += sC
			EEGAccuracy[emotion] += eC
			bothAccuracy[emotion] += bC
			fmt.Printf("%s\t", emotion)
			for _, e := range fileKeys {
				if e == bA {
					fmt.Printf("1\t")
				} else {
					fmt.Printf("0\t")
				}
			}
			fmt.Printf("%d\t%d\t%d\n", sC, eC, bC)
		}
	}
	fmt.Printf("Accuracy\n")
	for _, emotion := range fileKeys {
		fmt.Printf("%s\t%f\t%f\t%f\n", emotion, float64(speechAccuracy[emotion])/float64(len(speechFiles[emotion])), float64(EEGAccuracy[emotion])/float64(len(speechFiles[emotion])), float64(bothAccuracy[emotion])/float64(len(speechFiles[emotion])))
	}

	return nil
}

func ClassifyGMMBothConcat(speechTrainDir string, speechFiles map[string][]string, eegTrainDir string, eegFiles map[string][]string) error {
	speechAlphaTrainSet, err := GetAlphaEGMs(speechTrainDir)
	if err != nil {
		return err
	}

	eegAlphaTrainSet, err := GetAlphaEGMs(eegTrainDir)
	if err != nil {
		return err
	}

	fileKeys := make([]string, 0, len(speechFiles))
	for k := range speechFiles {
		fileKeys = append(fileKeys, k)
	}

	sort.Strings(fileKeys)
	correctFiles := make(map[string]int, len(fileKeys))
	correctVectors := make(map[string]int, len(fileKeys))
	sumVectors := make(map[string]int, len(fileKeys))

	for _, emotion := range fileKeys {
		correctFiles[emotion] = 0
		correctVectors[emotion] = 0
		sumVectors[emotion] = 0
	}

	for _, emotion := range fileKeys {
		for i := 0; i < len(speechFiles[emotion]); i++ {
			eegFeatures := GetEegFeaturesForFile(0, eegFiles[emotion][i])
			allFeatures := GetSpeechFeatureForFile(speechFiles[emotion][i])
			averaged := AverageSlice(allFeatures, len(allFeatures)/len(eegFeatures))
			speechFeatures := averaged[0 : len(averaged)-(len(averaged)-len(eegFeatures))]

			boolCorrect, vectors, sumVector := TestGMMBothConcat(emotion, fileKeys, speechAlphaTrainSet, speechFeatures, eegAlphaTrainSet, eegFeatures)

			correctFiles[emotion] += boolCorrect
			correctVectors[emotion] += vectors[emotion]
			sumVectors[emotion] += sumVector
		}
	}
	fmt.Printf("\tCorrectFiles\tCorrectVectors\n")
	for _, emotion := range fileKeys {
		fmt.Printf("%s\t%f\t%f\n", emotion, float64(correctFiles[emotion])/float64(len(speechFiles[emotion])), float64(correctVectors[emotion])/float64(sumVectors[emotion]))
	}

	return nil
}
