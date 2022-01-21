package emotions

import (
	"fmt"
	"math"
)

const EPS = 0.00001

// Complex is a representation of complex numbers, because I didn't want to use the built in one
type Complex struct {
	Re float64
	Im float64
}

func zeroComplex() Complex {
	return Complex{
		Re: 0.0,
		Im: 0.0,
	}
}

func (c Complex) String() string {
	// a
	if c.Im < EPS && c.Im > 0-EPS {
		if c.Re < EPS && c.Re > 0-EPS {
			return fmt.Sprintf("0")
		}
		return fmt.Sprintf("%f", c.Re)
	}

	//(a - bi)
	if c.Im < 0.0 {
		if c.Re < EPS && c.Re > 0-EPS {
			return fmt.Sprintf("0 - i(%f)", math.Abs(c.Im))
		}
		return fmt.Sprintf("%f - i(%f)", c.Re, math.Abs(c.Im))
	}
	// (a + bi)
	if c.Re < EPS && c.Re > 0-EPS {
		return fmt.Sprintf("0 + i(%f)", c.Im)
	}
	return fmt.Sprintf("%f + i(%f)", c.Re, c.Im)
}

func (c *Complex) swapped() {
	temp := c.Re
	c.Re = c.Im
	c.Im = temp
}

func (c Complex) divide(x float64) Complex {
	return Complex{
		Re: c.Re / x,
		Im: c.Im / x,
	}
}

func (c *Complex) divided(x float64) {
	c.Re = c.Re / x
	c.Im = c.Im / x
}

func (c Complex) conjugate() Complex {
	return Complex{Re: c.Re, Im: -c.Im}
}

func (c *Complex) added(o Complex) {
	c.Re = c.Re + o.Re
	c.Im = c.Im + o.Im
}

func (c Complex) add(o Complex) Complex {
	return Complex{
		Re: c.Re + o.Re,
		Im: c.Im + o.Im,
	}
}

// e^(⁻² ᵖᶦ ᶦᵏʲ/ᴺ)
func e(k, j int, n int) Complex {
	return Complex{
		Re: math.Cos(2 * math.Pi * float64(k) * float64(j) / float64(n)),
		Im: math.Sin(2 * math.Pi * float64(k) * float64(j) / float64(n)),
	}
}

func dot(c1, c2 Complex) Complex {
	// c1 = (a + bi)
	// c2 = (c + di)
	// ac - bd + i(bc + ad)

	return Complex{
		Re: c1.Re*c2.Re - c1.Im*c2.Im,
		Im: c1.Im*c2.Re + c1.Re*c2.Im,
	}
}

//Edot returns the dot product of the k-th and the s-th roots of unity in the
// n dimentional space
func Edot(k, s, n int) Complex {
	sum := zeroComplex()
	for i := 0; i < n; i++ {
		sum.added(dot(e(k, i, n), e(s, i, n)))
	}
	sum.divided(float64(n))
	return sum
}

// a_k = (? 1/N) j=0^N-1 (b_j.e ^(-2πijk/N))
func a(k int, b []Complex, n int) Complex {
	sum := zeroComplex()
	for j := 0; j < n; j++ {
		d := dot(b[j], e(k, j, n))
		sum.added(d)
	}

	return sum.divide(float64(n))
}

// b_k = (? 1/N) * Σ_j=0^N-1 (a_j * e ^(-2πijk/N))
// follows this formula, but with a_j - real
func b_real(k int, x []float64) Complex {
	sum := zeroComplex()

	for j := 0; j < len(x); j++ {
		d := Complex{}
		d.Re = x[j] * math.Cos(2*math.Pi*float64(k)*float64(j)/float64(len(x)))
		d.Im = -x[j] * math.Sin(2*math.Pi*float64(k)*float64(j)/float64(len(x)))
		sum.added(d)
	}

	// This is for normalisation if you want to keep the energy
	// sum.divided(float64(len(x)))
	return sum
}

// b_k = (1/N) * Σ_j=0^N-1 (a_j * e ^(-2πijk/N))
func b(k int, x []Complex) Complex {
	sum := zeroComplex()
	for j := 0; j < len(x); j++ {
		d := dot(x[j], e(k, j, len(x)).conjugate())
		sum.added(d)
	}

	// Normalisation: This is for normalisation if you want to keep the energy
	// sum.divided(float64(len(x)))
	return sum
}

// Power returns the power of the given complex number
// a + ib
// a^2 + b^2
func Power(c Complex) float64 {
	return (c.Re*c.Re + c.Im*c.Im)
}

// Magnitude returns the magnitude of the given complex number
// a + ib
// √(a^2 + b^2)
func Magnitude(c Complex) float64 {
	return math.Sqrt(c.Re*c.Re + c.Im*c.Im)
}

// Idft returns the inverse descrete fourier transform
func Idft(c []Complex) []Complex {
	n := len(c)
	x := make([]Complex, n, n)

	for k := 0; k < n; k++ {
		x[k] = a(k, c, n)
	}

	return x
}

func b_k_fast(x []Complex, W []Complex, depth int, first int, step int, len int) Complex {
	if len == 1 {
		return x[first]
	}

	// The idea is to split the signal on even and odd part so we store the begining of the array and the step
	// the term represents the even part and the right - the odd part (scaled with e^...)
	return b_k_fast(x, W, depth+1, first, step*2, len/2).add(dot(W[depth], b_k_fast(x, W, depth+1, first+step, step*2, len/2)))

	// Normalisation: The /2 is for normalisation if you want to keep the energy
	// return b_k_fast(x, W, depth+1, first, step*2, len/2).add(dot(W[depth], b_k_fast(x, W, depth+1, first+step, step*2, len/2))).divide(2)
}

func fft(x []Complex, W [][]Complex) []Complex {
	n := len(x)
	coefficients := make([]Complex, n, n)
	for k := 0; k < n; k++ {
		coefficients[k] = b_k_fast(x, W[k], 0, 0, 1, len(x))
	}

	return coefficients
}

// Fft computes the fast fourier transform on the given signal with len N
// it returns N complex coefficients
func Fft(x []Complex) []Complex {
	n := len(x)

	W := make([][]Complex, n, n)
	for k := 0; k < n; k++ {
		W[k] = make([]Complex, n, n)
		j := 0
		for m := n; m != 0; m /= 2 {
			W[k][j] = e(1, k, m).conjugate()
			j++
		}
	}
	return fft(x, W)
}

// Ifft returns the inverse fast fourier transform of the given fourier coefficients
func Ifft(x []Complex) []Complex {
	n := len(x)

	W := make([][]Complex, n, n)
	for k := 0; k < n; k++ {
		W[k] = make([]Complex, n, n)
		j := 0
		for m := n; m != 0; m /= 2 {
			W[k][j] = e(1, k, m).conjugate()
			j++
		}
	}

	for i := 0; i < n; i++ {
		x[i].swapped()
	}

	coefficients := fft(x, W)
	for i := 0; i < n; i++ {
		coefficients[i].swapped()
		coefficients[i].divided(float64(len(x)))
	}

	return coefficients
}

// FftReal returns the fourier coefficients for the given signal of len N
// it returns N/2 + 1 coefficients and works only with powers of two
func FftReal(x []float64) ([]Complex, float64) {
	if !IsPowerOfTwo(len(x)) {
		panic("FFT expects the len of the data to be a power of 2")
	}

	// split the signal on even and odd part
	n := len(x)
	even := make([]float64, n/2, n/2)
	odd := make([]float64, n/2, n/2)

	X := make([]Complex, n/2+1, n/2+1)
	var energy float64

	for k := 0; k < n/2; k++ {
		energy += x[2*k]*x[2*k] + x[2*k+1]*x[2*k+1]
		// calculate the special N/2 coefficient
		//since it's the sum of the signal with alternating signs
		X[n/2].Re += x[2*k] - x[2*k+1]
		even[k] = x[2*k]
		odd[k] = x[2*k+1]
	}

	Even, Odd := DoubleReal(even, odd)

	for k := 0; k < n/2; k++ {
		X[k] = Even[k].add(dot(e(k, 1, n).conjugate(), Odd[k]))
	}

	//  Normalisation: This is again for normalisation because the first and last coefficients don't have conjugates, so the energy shouldn't be doubled.
	// X[0].divided(2.0)
	// X[n/2].divided(float64(n))

	return X, energy
}

// FftWav returns the fourier coefficients for the given wav file of len N
// It returns N/2 + 1 coefficients
func FftWav(f WavFile) ([]Complex, float64) {
	if !IsPowerOfTwo(len(f.data)) {
		panic("FFT expects the len of the data to be a power of 2")
	}

	return FftReal(f.data)
}

// DoubleReal returns the fourier coefficients for the two given real signals of len N
// it composes a new complex signal - its real part is the first signal and the imaginary part is the second signals
// It then runs Fft fot the newly composed complex signal and then separates the coefficients which are with len N
func DoubleReal(x, y []float64) ([]Complex, []Complex) {
	n := len(x)

	xCoefficients := make([]Complex, n, n)
	yCoefficients := make([]Complex, n, n)

	z := make([]Complex, n, n)
	for i := 0; i < n; i++ {
		xCoefficients[0].Re += x[i]
		yCoefficients[0].Re += y[i]

		z[i].Re = x[i]
		z[i].Im = y[i]
	}

	// Normalisation: Because the first coefficient is calculated as a special case, this is for normalisation
	// xCoefficients[0].Re = xCoefficients[0].Re / float64(n)
	// yCoefficients[0].Re = yCoefficients[0].Re / float64(n)

	zCoefficients := Fft(z)
	for k := 1; k < n; k++ {
		xCoefficients[k].Re = (zCoefficients[k].Re + zCoefficients[n-k].Re) / 2.0
		xCoefficients[k].Im = (zCoefficients[k].Im - zCoefficients[n-k].Im) / 2.0

		yCoefficients[k].Re = (zCoefficients[k].Im + zCoefficients[n-k].Im) / 2.0
		yCoefficients[k].Im = -(zCoefficients[k].Re - zCoefficients[n-k].Re) / 2.0
	}

	return xCoefficients, yCoefficients
}

//Dft returns the discrete fourier transform for complex signal with len N
// it return N coefficients [0...N-1]
func Dft(x []Complex) []Complex {
	coefficients := make([]Complex, len(x), len(x))
	for k := 0; k < len(x); k++ {
		coefficients[k] = b(k, x)
	}
	return coefficients
}

//Dft returns the discrete fourier transform for real signal with len N
// it returns N/2 + 1 coefficients [0....N/2]
func DftReal(x []float64) []Complex {
	coefficients := make([]Complex, len(x)/2+1, len(x)/2+1)
	for k := 0; k < len(x)/2+1; k++ {
		coefficients[k] = b_real(k, x)

		// Energy: This is for saving energy
		// if k > 0 && k < len(x)/2 {
		// 	coefficients[k].divided(0.5)
		// }

	}
	return coefficients
}
