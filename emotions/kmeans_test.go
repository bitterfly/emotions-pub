package emotions

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKMeans(t *testing.T) {
	points1 := [][]float64{
		[]float64{1.435, 1.6354},
		[]float64{1.23, 2.345},
		[]float64{0.2355, 2.123},
		[]float64{15.435, 16.5254},
		[]float64{17.245, 20.2345},
		[]float64{18.234, 19.234},
		[]float64{19.2543, 20.654},
		[]float64{21.324, 21.324},
	}

	clustered, _, _, _ := KMeans(points1, 2)
	smallerPointsCluster := clustered[0].clusterID
	var largerPointsCluster int32
	if smallerPointsCluster == int32(0) {
		largerPointsCluster = int32(1)
	} else {
		largerPointsCluster = int32(0)
	}

	for _, c := range clustered {
		if c.coefficients[0]+c.coefficients[1] < 10.0 {
			assert.Equal(t, smallerPointsCluster, c.clusterID, fmt.Sprintf("Point: %v", c))
		} else {
			assert.Equal(t, largerPointsCluster, c.clusterID, fmt.Sprintf("Point: %v", c))
		}
	}

	points2 := make([][]float64, 0, 400)
	f, err := os.Open("testing/k_means_normal.txt")
	if err != nil {
		t.Log(err)
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		x, _ := strconv.ParseFloat(strings.Split(scanner.Text(), "\t")[0], 64)
		y, _ := strconv.ParseFloat(strings.Split(scanner.Text(), "\t")[1], 64)

		points2 = append(points2, []float64{x, y})
	}

	k := 4
	clustered, _, _, _ = KMeans(points2, k)
	PlotClusters(clustered, k, "testing/kmeans_result.png")
}
