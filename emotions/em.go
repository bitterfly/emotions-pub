package emotions

import (
	"fmt"
	"math"
	"os"
)

type EmotionFeature struct {
	Emotion string
	Feature []float64
}

// Gaussian represent a single gaussian
type Gaussian struct {
	Phi          float64
	Expectations []float64
	Variances    []float64
}

// GaussianMixture represent a mixture of gaussians
type GaussianMixture = []Gaussian

func zeroMixture(g GaussianMixture, K int) {
	for k := 0; k < K; k++ {
		zero(&g[k].Expectations)
		zero(&g[k].Variances)
		g[k].Phi = 0.0
	}
}

type EmotionGausianMixure struct {
	Emotion string
	GM      GaussianMixture
}

type AlphaEGM struct {
	Alpha float64
	EGM   EmotionGausianMixure
}

// GMM returns the k gaussian mixures for the given data
func GMM(mfccsFloats [][]float64, k int) GaussianMixture {
	X, expectations, variances, numInCluster := KMeans(mfccsFloats, k)

	// f, _ := os.Create("/tmp/danni")
	// defer f.Close()
	// fmt.Fprintf(f, "[%d][%d]\n", len(mfccsFloats), len(mfccsFloats[0]))
	// for i, m := range mfccsFloats {
	// 	fmt.Fprintf(f, "[%d] %v\n", i, m)
	// }

	fmt.Fprintf(os.Stderr, "\n==============EM================\n")

	gmixture := make(GaussianMixture, k, k)

	for i := 0; i < k; i++ {
		gmixture[i] = Gaussian{
			Phi:          float64(numInCluster[i]) / float64(len(X)),
			Expectations: expectations[i],
			Variances:    variances[i],
		}
	}

	return em(X, k, gmixture)
}

func em(X []MfccClusterisable, k int, gMixture GaussianMixture) GaussianMixture {
	prevLikelihood := 0.0
	likelihood := 0.0
	step := 0

	f, _ := os.Create("/tmp/foo")
	defer f.Close()

	fmt.Fprintf(f, "Expectation0\n")
	for j := 0; j < k; j++ {
		fmt.Fprintf(f, "%d %v\n", j, gMixture[j].Expectations)
	}

	fmt.Fprintf(f, "Variance0\n")
	for j := 0; j < k; j++ {
		fmt.Fprintf(f, "%d %v\n", j, gMixture[j].Variances)
	}

	for step < 200 {
		fmt.Fprintf(f, "================= %d =================\n", step)
		w := make([][]float64, len(X), len(X))
		var sum float64
		maximums := make([]float64, len(X), len(X))

		for i := 0; i < len(maximums); i++ {
			maximums[i] = math.Inf(-1)
		}

		for i := 0; i < len(X); i++ {
			w[i] = make([]float64, k, k)

			for j := 0; j < k; j++ {
				w[i][j] = math.Log(gMixture[j].Phi) + N(X[i].coefficients, gMixture[j].Expectations, gMixture[j].Variances)

				if maximums[i] < w[i][j] {
					maximums[i] = w[i][j]
				}
			}

			sum = 0
			for j := 0; j < k; j++ {
				if w[i][j] < maximums[i]-10 {
					w[i][j] = 0
				} else {
					w[i][j] = math.Exp(w[i][j] - maximums[i])
					sum += w[i][j]
				}
			}

			divide(&w[i], sum)
		}

		fmt.Fprintf(f, "W\n")
		for i := range X {
			fmt.Fprintf(f, "%d %v\n", i, w[i])
		}

		N := make([]float64, k, k)
		for i := 0; i < len(X); i++ {
			for j := 0; j < k; j++ {
				N[j] += w[i][j]
			}
		}

		fmt.Fprintf(f, "Expectation1\n")
		for j := 0; j < k; j++ {
			fmt.Fprintf(f, "%d %v\n", j, gMixture[j].Expectations)
		}

		fmt.Fprintf(f, "Variance1\n")
		for j := 0; j < k; j++ {
			fmt.Fprintf(f, "%d %v\n", j, gMixture[j].Variances)
		}

		zeroMixture(gMixture, k)

		// Expectations
		for i := 0; i < len(X); i++ {
			for j := 0; j < k; j++ {
				add(&gMixture[j].Expectations, multiplied(X[i].coefficients, w[i][j]))
			}
		}

		for j := 0; j < k; j++ {
			divide(&(gMixture[j].Expectations), N[j])
		}

		// Variances
		for i := 0; i < len(X); i++ {
			for j := 0; j < k; j++ {
				diagonal := minused(X[i].coefficients, gMixture[j].Expectations)
				square(&diagonal)

				add(&gMixture[j].Variances, multiplied(diagonal, w[i][j]))
			}
		}

		// Phi and 1/Nk
		for j := 0; j < k; j++ {
			divide(&(gMixture[j].Variances), N[j])
			gMixture[j].Phi = N[j] / float64(len(X))
		}

		fmt.Fprintf(f, "Expectation2\n")
		for j := 0; j < k; j++ {
			fmt.Fprintf(f, "%d %v\n", j, gMixture[j].Expectations)
		}

		fmt.Fprintf(f, "Variance2\n")
		for j := 0; j < k; j++ {
			fmt.Fprintf(f, "%d %v\n", j, gMixture[j].Variances)
		}

		likelihood = logLikelihood(X, k, gMixture)

		if math.IsNaN(likelihood) {
			panic(fmt.Sprintf("Likelihood is NAN, step: %d", step))
		}

		if epsDistance(likelihood, prevLikelihood, 0.00001) {
			break
		}

		prevLikelihood = likelihood
		step++
	}

	fmt.Fprintf(os.Stderr, "EM: Break on step: %d with likelihood: %f\n===================================================\n", step, likelihood)
	return gMixture
}

func epsDistance(a, b, e float64) bool {
	return (a-b < e && a-b > -e)
}

func getDeterminant(variance []float64) float64 {
	det := 1.0
	for i := 0; i < len(variance); i++ {
		det *= variance[i]
	}
	return det
}

func getLogDeterminant(variance []float64) float64 {
	det := 0.0
	for i := 0; i < len(variance); i++ {
		det += math.Log(variance[i])
	}
	return det
}

func FindBestGaussian(X []float64, k int, egmms []EmotionGausianMixure) (string, bool) {
	max := math.Inf(-42)
	argmax := ""
	failed := 0

	for _, g := range egmms {
		currEmotion, err := EvaluateVector(X, k, g.GM)
		if err != nil {
			failed++
			continue
		}
		if currEmotion > max {
			max = currEmotion
			argmax = g.Emotion
		}
	}
	return argmax, failed == len(egmms)
}

func SumGmmOverCorpus(Data []EmotionFeature, egmms []EmotionGausianMixure) (float64, bool) {
	failed := false
	sum := 0.0
	for i := 0; i < len(Data); i++ {
		cur, curFailed := GmmsProbabilityGivenClass(Data[i].Feature, Data[i].Emotion, egmms)
		if curFailed {
			failed = true
		}
		sum += cur
	}
	return sum, failed
}

func GmmsProbabilityGivenClass(X []float64, givenClass string, egmms []EmotionGausianMixure) (float64, bool) {
	res := 0.0
	sum := 0.0
	failed := false

	for _, g := range egmms {
		cur, err := EvaluateVector(X, len(egmms[0].GM), g.GM)
		if err != nil {
			failed = true
			continue
		}
		sum += cur
		if givenClass == g.Emotion {
			res = cur
		}
	}

	return res / sum, failed
}

func FindBestGaussianAlpha(X []float64, k int, egmms []AlphaEGM) (string, bool) {
	max := math.Inf(-42)
	argmax := ""
	failed := 0

	for _, g := range egmms {
		currEmotion, err := EvaluateVector(X, k, g.EGM.GM)
		if err != nil {
			failed++
			continue
		}
		if currEmotion > max {
			max = currEmotion
			argmax = g.EGM.Emotion
		}
	}
	return argmax, failed == len(egmms)
}

func FindBestGaussianAlphaFloat(X []float64, e string, k int, egmms []AlphaEGM) float64 {
	for _, g := range egmms {
		if g.EGM.Emotion == e {
			currEmotion, _ := EvaluateVector(X, k, g.EGM.GM)
			return currEmotion
		}
	}
	return -1
}

func sumGaussianAlpha(X []float64, k int, egmms []AlphaEGM) float64 {
	sum := 0.0
	for _, g := range egmms {
		s, _ := EvaluateVector(X, k, g.EGM.GM)
		sum += s
	}
	return sum
}

func FindBestGaussianMany(X [][]float64, k int, egmms []EmotionGausianMixure) (map[string]int, int) {
	scores := make(map[string]int)
	for _, egmm := range egmms {
		scores[egmm.Emotion] = 0
	}

	sum := 0
	failedAll := 0
	for _, x := range X {
		best, failed := FindBestGaussian(x, k, egmms)
		if failed {
			failedAll++
		} else {
			scores[best]++
		}
		sum++
	}

	if failedAll > 0 {
		fmt.Fprintf(os.Stderr, "Failed (%d/%d)\n", failedAll, sum)
	}
	return scores, sum
}

// EvaluateVector returns the likelihood a given vector
func EvaluateVector(X []float64, k int, g GaussianMixture) (float64, error) {
	return likelihoodFloat(X, k, g)
}

func TestGMM(emotion string, emotions []string, coefficient [][]float64, egmms []EmotionGausianMixure, verbose bool) (int, map[string]int, int) {
	if verbose {
		fmt.Printf("%s\t", emotion)
	}

	k := len(egmms[0].GM)
	failedAll := 0

	counters := make(map[string]int)
	for _, e := range emotions {
		counters[e] = 0
	}

	for _, m := range coefficient {
		best, failed := FindBestGaussian(m, k, egmms)
		if failed {
			failedAll++
			continue
		}
		counters[best]++
	}

	sum := 0
	for _, e := range emotions {
		if verbose {
			fmt.Printf("%d\t", counters[e])
		}
		sum += counters[e]
	}
	if verbose {
		fmt.Printf("\n")
	}

	if failedAll > 0 {
		fmt.Fprintf(os.Stderr, "%s Failed (%d/%d)\n", emotion, failedAll, sum)
	}
	return correct(emotion, counters), counters, sum
}

// func TestGMMFloat(emotion string, emotions []string, coefficient [][]float64, egmms []EmotionGausianMixure, verbose bool) (int, map[string]float64, float64) {
// 	if verbose {
// 		fmt.Printf("%s\t", emotion)
// 	}

// 	k := len(egmms[0].GM)
// 	failedAll := 0

// 	counters := make(map[string]float64)
// 	for _, e := range emotions {
// 		counters[e] = 0
// 	}

// 	for _, m := range coefficient {
// 		best, val, failed := FindBestGaussianFloat(m, k, egmms)
// 		if failed {
// 			failedAll++
// 			continue
// 		}
// 		counters[best] += val
// 	}

// 	sum := 0.0
// 	for _, e := range emotions {
// 		if verbose {
// 			fmt.Printf("%d\t", counters[e])
// 		}
// 		sum += counters[e]
// 	}
// 	if verbose {
// 		fmt.Printf("\n")
// 	}

// 	if failedAll > 0 {
// 		fmt.Fprintf(os.Stderr, "%s Failed (%d/%d)\n", emotion, failedAll, sum)
// 	}
// 	return correctFloat(emotion, counters), counters, sum
// }

func getBest(dict map[string]int) string {
	best := -1
	bestArg := ""
	for k, v := range dict {
		if v > best {
			best = v
			bestArg = k
		}
	}
	return bestArg
}

func TestGMMBoth(emotion string, emotionTypes []string, speechAlphaEGM []AlphaEGM, speechEGM []EmotionGausianMixure, speechFile string, eegAlphaEGM []AlphaEGM, eegEGM []EmotionGausianMixure, eegFile string, bucketSize int) (int, int, int, string) {
	kS := len(speechAlphaEGM[0].EGM.GM)
	kE := len(eegAlphaEGM[0].EGM.GM)

	speechFeatures := GetSpeechFeatureForFile(speechFile)
	eegFeatures := GetEegFeaturesForFile(bucketSize, eegFile)

	speechClassified, sumSpeech := FindBestGaussianMany(speechFeatures, kS, speechEGM)

	eegClassified, sumEEG := FindBestGaussianMany(eegFeatures, kE, eegEGM)

	bestBoth := -1.0
	bothEmotion := emotionTypes[0]
	for _, e := range emotionTypes {
		// bool
		// current := speechAlphaEGM[0].Alpha*float64(bToI(getBest(speechClassified) == e)) +
		// 	eegAlphaEGM[0].Alpha*float64(bToI(getBest(eegClassified) == e))

		current := speechAlphaEGM[0].Alpha*(float64(speechClassified[e])/float64(sumSpeech)) +
			eegAlphaEGM[0].Alpha*(float64(eegClassified[e])/float64(sumEEG))

		if current > bestBoth {
			bestBoth = current
			bothEmotion = e
		}
	}
	// return correct(emotion, counters), counters[emotion], sum
	return bToI(getBest(speechClassified) == emotion), bToI(getBest(eegClassified) == emotion), bToI(bothEmotion == emotion), bothEmotion
}

func TestGMMBothConcat(emotion string, emotionTypes []string, speechAlphaEGM []AlphaEGM, speechFeatures [][]float64, eegAlphaEGM []AlphaEGM, eegFeatures [][]float64) (int, map[string]int, int) {
	kS := len(speechAlphaEGM[0].EGM.GM)
	kE := len(eegAlphaEGM[0].EGM.GM)

	counters := make(map[string]int)
	for _, e := range emotionTypes {
		counters[e] = 0
	}

	for i := 0; i < len(speechFeatures); i++ {
		maxE := ""
		maxV := -1.0

		sumSpeech := sumGaussianAlpha(speechFeatures[i], kS, speechAlphaEGM)
		sumEeg := sumGaussianAlpha(eegFeatures[i], kE, eegAlphaEGM)

		for _, e := range emotionTypes {
			speechChoice := FindBestGaussianAlphaFloat(speechFeatures[i], e, kS, speechAlphaEGM)
			eegChoice := FindBestGaussianAlphaFloat(eegFeatures[i], e, kE, eegAlphaEGM)

			speechChoice /= sumSpeech
			eegChoice /= sumEeg
			// fmt.Fprintf(os.Stderr, "%d %d\n", bToI(speechChoice == emotion), bToI(eegChoice == emotion))

			current := speechAlphaEGM[0].Alpha*speechChoice + eegAlphaEGM[0].Alpha*eegChoice
			if current > maxV {
				maxV = current
				maxE = e
			}
		}
		counters[maxE]++
	}

	sum := 0
	fmt.Printf("%s", emotion)
	for _, e := range emotionTypes {
		fmt.Printf("\t%d", counters[e])
		sum += counters[e]
	}
	fmt.Printf("\n")

	return bToI(getBest(counters) == emotion), counters, sum
}

func bToI(b bool) int {
	if b {
		return 1
	}
	return 0
}

func correct(emotion string, counters map[string]int) int {
	maxV := 0
	maxE := ""
	for e, v := range counters {
		if v > maxV {
			maxV = v
			maxE = e
		}
	}
	if maxE == emotion {
		return 1
	}
	return 0
}

func correctFloat(emotion string, counters map[string]float64) int {
	maxV := 0.0
	maxE := ""
	for e, v := range counters {
		if v > maxV {
			maxV = v
			maxE = e
		}
	}
	if maxE == emotion {
		return 1
	}
	return 0
}

func N(xi []float64, expectation []float64, variance []float64) float64 {
	var exp float64
	for i := 0; i < len(xi); i++ {
		exp += (xi[i] - expectation[i]) * (xi[i] - expectation[i]) / variance[i]
	}
	return -0.5 * (exp + float64(len(xi))*math.Log(2.0*math.Pi) + getLogDeterminant(variance))
	// return log of this
	// return math.Exp(-0.5*exp) / math.Sqrt(math.Pow(2*math.Pi, float64(len(xi)))*getDeterminant(variance))
}

func logLikelihoodFloat(X []float64, k int, g GaussianMixture) float64 {
	sum := 0.0
	for j := 0; j < k; j++ {
		sum += g[j].Phi * math.Exp(N(X, g[j].Expectations, g[j].Variances))
	}
	return math.Log(sum)
}

func likelihoodFloat(X []float64, k int, g GaussianMixture) (float64, error) {
	sum := 0.0
	for j := 0; j < k; j++ {
		sum += g[j].Phi * math.Exp(N(X, g[j].Expectations, g[j].Variances))
	}

	if sum == 0.0 || sum == -0.0 {
		return -1, fmt.Errorf("The likelihood is 0")
	}
	return sum, nil
}

// sum_i log(sum_j phi_j * N(x[i], m[k], s[k]))

func logLikelihood(X []MfccClusterisable, k int, g GaussianMixture) float64 {
	sum := 0.0
	for i := 0; i < len(X); i++ {
		sum += logLikelihoodFloat(X[i].coefficients, k, g)
	}
	return sum
}
