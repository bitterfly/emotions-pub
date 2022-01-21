package emotions

import (
	"fmt"
	"math"
)

func Zero(x *[]float64) {
	zero(x)
}

func zero(x *[]float64) {
	for i := 0; i < len(*x); i++ {
		(*x)[i] = 0.0
	}
}

func Divide(x *[]float64, n float64) {
	divide(x, n)
}

func divide(x *[]float64, n float64) {
	for i := 0; i < len(*x); i++ {
		(*x)[i] /= n
	}
}

func Add(x *[]float64, y []float64) {
	add(x, y)
}

func add(x *[]float64, y []float64) {
	for i := 0; i < len(y); i++ {
		(*x)[i] += y[i]
	}
}

func multiply(x *[]float64, y float64) {
	for i := 0; i < len(*x); i++ {
		(*x)[i] *= y
	}
}

func multiplied(x []float64, y float64) []float64 {
	z := make([]float64, len(x), len(x))
	for i := 0; i < len(z); i++ {
		z[i] = x[i] * y
	}

	return z
}

func inverse(x *[]float64) {
	for i := 0; i < len(*x); i++ {
		(*x)[i] = float64(1) / (*x)[i]
	}
}

func square(x *[]float64) {
	for i := 0; i < len(*x); i++ {
		(*x)[i] = (*x)[i] * (*x)[i]
	}
}

func eps(x *[]float64, epsilon float64) {
	for i := 0; i < len(*x); i++ {
		if (*x)[i] < epsilon {
			(*x)[i] = epsilon
		}
	}
}

func minused(x []float64, y []float64) []float64 {
	z := make([]float64, len(x), len(x))
	for i := 0; i < len(x); i++ {
		z[i] = x[i] - y[i]
	}
	return z
}

func getSqrt(x *[]float64) {
	for i := 0; i < len(*x); i++ {
		(*x)[i] = math.Sqrt((*x)[i])
	}
}

func combineSlices(a, b []string) []string {
	c := make([]string, len(a)+len(b), len(a)+len(b))
	for i := 0; i < len(a); i++ {
		c[i] = a[i]
	}

	for j := 0; j < len(b); j++ {
		c[j+len(a)] = b[j]
	}

	return c
}

func Average(x [][]float64) []float64 {
	result := make([]float64, len(x[0]), len(x[0]))
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(x[0]); j++ {
			result[j] += x[i][j]
		}
	}

	for j := 0; j < len(x[0]); j++ {
		result[j] /= float64(len(x))
	}

	return result
}

// Concat concatenates the tuples of vectors of x and y
func Concat(x [][]float64, y [][]float64) [][]float64 {
	z := make([][]float64, len(x), len(x))
	if len(x) != len(y) {
		panic(fmt.Sprintf("When using concat, sizes of the arrays should match: %d != %d\n", len(x), len(y)))
	}
	for i := 0; i < len(x); i++ {
		z[i] = make([]float64, len(x[i])+len(y[i]))
		t := 0
		for j := 0; j < len(x[i]); j++ {
			z[i][t] = x[i][j]
			t++
		}
		for j := 0; j < len(y[i]); j++ {
			z[i][t] = y[i][j]
			t++
		}
	}

	return z
}

// AverageSlice accumulates every average elements of the array x
func AverageSlice(x [][]float64, average int) [][]float64 {
	averagedSlice := make([][]float64, 0, len(x)/average)
	for i := 0; i+average <= len(x); i += average {
		averagedSlice = append(averagedSlice, Average(x[i:i+average]))
	}

	if len(averagedSlice) != len(x)/average {
		panic(fmt.Sprintf("Len: %d should be: %d", len(averagedSlice), len(x)/average))
	}

	return averagedSlice
}

var ElectrodeCouples [][2]int = [][2]int{
	[2]int{0, 1},
	[2]int{2, 3},
	[2]int{5, 6},
	[2]int{8, 9},
	[2]int{11, 12},
	[2]int{13, 14},
	[2]int{15, 16},
	[2]int{17, 18},
}

func GetDE(data [][]float64) [][]float64 {
	result := make([][]float64, len(data), len(data))
	n := len(ElectrodeCouples) * len(waveRanges)

	for i := 0; i < len(result); i++ {
		result[i] = make([]float64, n+3*len(waveRanges), n+3*len(waveRanges))
		for j, c := range ElectrodeCouples {
			for k := 0; k < len(waveRanges); k++ {
				// fmt.Printf("result[%d][%d] = data[%d][%d] - data[%d][%d]\n", i, k+j*len(waveRanges), i, k+(c[0]*len(waveRanges)), i, k+(c[1]*len(waveRanges)))
				result[i][k+j*len(waveRanges)] = math.Abs(data[i][k+(c[0]*len(waveRanges))] - data[i][k+(c[1]*len(waveRanges))])
			}
		}
		for k := 0; k < len(waveRanges); k++ {
			result[i][n+k] = data[i][k+4*len(waveRanges)]
			result[i][n+len(waveRanges)+k] = data[i][k+7*len(waveRanges)]
			result[i][n+2*len(waveRanges)+k] = data[i][k+10*len(waveRanges)]
		}
	}
	return result
}
