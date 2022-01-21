package emotions

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

// MfccClusterisable - because every group of mfccs coefficient belongs to exacly one cluster
// we store the cluster Id along with the coefficients
type MfccClusterisable struct {
	coefficients []float64
	clusterID    int32
}

// GetCluster returns the cluster the mfcc vector was sorted into
func (m MfccClusterisable) GetCluster() int32 {
	return m.clusterID
}

// KMeans takes all the mfccs for a file
// which are of size nx39 (where n is file_len / 10ms)
// and separates them into k clusters
// then returns the means and variance of each cluster
func KMeans(mfccsFloats [][]float64, k int) ([]MfccClusterisable, [][]float64, [][]float64, []int) {
	e, variances := getμAndσ(mfccsFloats)
	f, _ := os.Create("/tmp/danni")
	defer f.Close()

	fmt.Fprintf(f, "===EXP=====\n")
	for i, ee := range e {
		fmt.Fprintf(f, "i: %d e: %v\n", i, ee)
	}
	fmt.Fprintf(f, "===VAR=====\n")
	for i, vv := range variances {
		fmt.Fprintf(f, "i: %d v: %v\n", i, vv)
	}

	mfccs := make([]MfccClusterisable, len(mfccsFloats), len(mfccsFloats))
	for i, mfcc := range mfccsFloats {
		mfccs[i] = MfccClusterisable{
			coefficients: mfcc,
			clusterID:    -1,
		}
	}

	fmt.Fprintf(os.Stderr, "\n================kMeans=================\n")

	μ, σ, numInCluster := kMeans(mfccs, variances, k)
	return mfccs, μ, σ, numInCluster
}

func initialiseRandom(mfccs []MfccClusterisable, k int, variances []float64) map[int32]struct{} {
	centroidIndices := make(map[int32]struct{})

	// First choose randomly the firts centroids
	// Keep generating a new random number until there are k keys in centroidIndices
	rand.Seed(time.Now().UTC().UnixNano())
	for len(centroidIndices) < k {
		ind := rand.Int31n(int32(len(mfccs)))
		if _, ok := centroidIndices[ind]; !ok {
			centroidIndices[ind] = struct{}{}
		}
	}

	return centroidIndices
}

func initialiseKPP(mfccs []MfccClusterisable, k int, variances []float64) map[int32]struct{} {
	centroidIndices := make(map[int32]struct{})
	rand.Seed(time.Now().UTC().UnixNano())
	centroidIndices[rand.Int31n(int32(len(mfccs)))] = struct{}{}

	for len(centroidIndices) < k {
		max := math.Inf(-42)
		argmax := -1
		for i, d := range mfccs {
			curr := findClosestCentroidFromPoints(mfccs, centroidIndices, d.coefficients, variances)
			if curr > max {
				max = curr
				argmax = i
			}
		}

		if _, ok := centroidIndices[int32(argmax)]; !ok {
			centroidIndices[int32(argmax)] = struct{}{}
		}
	}

	return centroidIndices
}

func kMeans(mfccs []MfccClusterisable, variances []float64, k int) ([][]float64, [][]float64, []int) {
	centroids := make([][]float64, 0, k)
	for i := range initialiseRandom(mfccs, k, variances) {
		centroids = append(centroids, mfccs[i].coefficients)
	}

	iterations := 100
	rsss := make([]float64, 0, iterations)
	// Group the documents in clusters and recalculate the new centroid of the cluster
	for times := 0; times < iterations; times++ {
		for i := range mfccs {
			mfccs[i].clusterID = findClosestCentroid(centroids, mfccs[i].coefficients, variances)
		}

		centroids = findNewCentroids(mfccs, k)
		rsss = append(rsss, getRss(mfccs, centroids, variances))

		// break if there is no difference between new and old centroids
		if times > 1 && math.Abs(rsss[times-1]-rsss[times]) < 0.0000001 {
			fmt.Fprintf(os.Stderr, "Break on iteration %d RSS: %f\n", times, rsss[times])
			break
		}
	}

	for i := range mfccs {
		mfccs[i].clusterID = findClosestCentroid(centroids, mfccs[i].coefficients, variances)
	}

	return getClustersμσcount(mfccs, k)
}

func getClustersμσcount(mfccs []MfccClusterisable, k int) ([][]float64, [][]float64, []int) {
	expectations := make([][]float64, k, k)
	expectationsSquared := make([][]float64, k, k)
	variances := make([][]float64, k, k)
	numInCluster := make([]int, k, k)

	for i := 0; i < k; i++ {
		expectations[i] = make([]float64, len(mfccs[0].coefficients))
		expectationsSquared[i] = make([]float64, len(mfccs[0].coefficients))
		variances[i] = make([]float64, len(mfccs[0].coefficients))
	}

	for _, mfcc := range mfccs {
		numInCluster[mfcc.clusterID]++
		for i, c := range mfcc.coefficients {
			expectations[mfcc.clusterID][i] += c
			expectationsSquared[mfcc.clusterID][i] += c * c
		}
	}

	for i := 0; i < k; i++ {
		for j := 0; j < len(mfccs[0].coefficients); j++ {
			expectations[i][j] /= float64(numInCluster[i])
			expectationsSquared[i][j] /= float64(numInCluster[i])
			variances[i][j] = expectationsSquared[i][j] - expectations[i][j]*expectations[i][j]
		}
	}

	fmt.Printf("numInCluster: %v\n", numInCluster)
	for i := 0; i < len(numInCluster); i++ {
		if numInCluster[i] == 1 {
			fmt.Printf("%d\n", i)
		}
	}

	return expectations, variances, numInCluster
}

func Getσ(mfccs [][]float64) []float64 {
	_, variances := getμAndσ(mfccs)
	return variances
}

func getμAndσ(mfccs [][]float64) ([]float64, []float64) {
	fmt.Printf("Len mfccs: %d\n", len(mfccs))
	variances := make([]float64, len(mfccs[0]), len(mfccs[0]))

	expectation := make([]float64, len(mfccs[0]), len(mfccs[0]))
	expectationSquared := make([]float64, len(mfccs[0]), len(mfccs[0]))
	for j := 0; j < len(mfccs[0]); j++ {
		for i := 0; i < len(mfccs); i++ {
			expectation[j] += mfccs[i][j]
			expectationSquared[j] += expectation[j] * expectation[j]
		}
	}

	for j := 0; j < len(mfccs[0]); j++ {
		expectation[j] /= float64(len(mfccs))
		expectationSquared[j] /= float64(len(mfccs))
		variances[j] = expectationSquared[j] - expectation[j]*expectation[j]
		if variances[j] < EPS {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("variance is EPS: %d %f", j, variances[j]))
			variances[j] = EPS
		}
	}

	return expectation, variances
}

func getRss(mfccs []MfccClusterisable, centroids [][]float64, variances []float64) float64 {
	var rss float64

	for _, mfcc := range mfccs {
		distance := mahalanobisDistance(mfcc.coefficients, centroids[mfcc.clusterID], variances)
		rss += distance * distance
	}

	return rss
}

func findNewCentroids(mfccs []MfccClusterisable, k int) [][]float64 {
	centroids := make([][]float64, k, k)

	for i := range centroids {
		centroids[i] = make([]float64, len(mfccs[0].coefficients), len(mfccs[0].coefficients))
	}
	mfccsInCluster := make([]int, k, k)

	for _, mfcc := range mfccs {

		mfccsInCluster[mfcc.clusterID]++
		add(&centroids[mfcc.clusterID], mfcc.coefficients)
	}

	for i := range centroids {
		divide(&centroids[i], float64(mfccsInCluster[i]))
	}

	return centroids
}

func findClosestCentroidFromPoints(mfccs []MfccClusterisable, centroidIds map[int32]struct{}, point []float64, variances []float64) float64 {
	// Returns positive infty if argument is >=0
	min := math.Inf(42)

	for centroidID := range centroidIds {
		currentDistance := mahalanobisDistance(mfccs[centroidID].coefficients, point, variances)
		if currentDistance < min {
			min = currentDistance
		}
	}
	return min
}

func findClosestCentroid(centroids [][]float64, mfcc []float64, variances []float64) int32 {
	// Returns positive infty if argument is >=0
	min := math.Inf(42)
	argmin := int32(-1)

	for i, centroid := range centroids {
		currentDistance := mahalanobisDistance(centroid, mfcc, variances)
		if currentDistance < min {
			min = currentDistance
			argmin = int32(i)
		}
	}
	return argmin
}

func mahalanobisDistance(x []float64, y []float64, variances []float64) float64 {
	sum := 0.0
	for i := 0; i < len(x); i++ {
		sum += ((x[i] - y[i]) * (x[i] - y[i])) / variances[i]
	}

	return sum
}

func euclidianDistance(x []float64, y []float64, variances []float64) float64 {
	sum := 0.0
	for i := 0; i < len(x); i++ {
		sum += (x[i] - y[i]) * (x[i] - y[i])
	}

	return sum
}
