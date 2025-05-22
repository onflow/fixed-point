package fix64

import (
	"math"
	"testing"
)

// Test Helpers
func toUFix64(f float64) UFix64 {
	return UFix64(f * fix64Scale)
}

func fromUFix64(f UFix64) float64 {
	return float64(f) / fix64Scale
}

func toFix64(f float64) Fix64 {
	return Fix64(f * fix64Scale)
}

func fromFix64(f Fix64) float64 {
	return float64(f) / fix64Scale
}

const maxUFix64 = 184467440737.09551615

func TestAddUFix64(t *testing.T) {
	tests := []struct {
		a, b float64
	}{
		{1.0, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{1.0, 1e8},
		{1.0, 1e8 + 1.0},
		{maxUFix64 - 1.0001, 1.0},
		{math.Floor(maxUFix64 / 2), math.Floor(maxUFix64 / 2)},
	}
	margin := 1e-8
	for _, tc := range tests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		res, err := AddUFix64(a, b)
		expected := tc.a + tc.b
		if err != nil {
			t.Errorf("AddUFix64(%f, %f) returned error: %v", tc.a, tc.b, err)
			continue
		}
		fres := fromUFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("AddUFix64(%f, %f) = %f, want %f (±%f)", tc.a, tc.b, fres, expected, margin)
		}
	}
	// Test overflow cases
	overflowTests := []struct {
		a, b float64
	}{
		{maxUFix64, 1.0},
		{maxUFix64, 0.01},
		{maxUFix64, 0.001},
		{maxUFix64, 0.00001},
		{maxUFix64, 0.0000001},
		{maxUFix64, 0.00000001},
		{maxUFix64, maxUFix64},
		{maxUFix64 / 2, maxUFix64 / 2},
		{maxUFix64 / 2, maxUFix64/2 + 1.0},
	}

	for _, tc := range overflowTests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		_, err := AddUFix64(a, b)
		if err != ErrOverflow {
			t.Errorf("AddUFix64(%f, %f) expected overflow error, got: %v", tc.a, tc.b, err)
		}
	}
}

func TestSubUFix64(t *testing.T) {
	tests := []struct {
		a, b float64
	}{
		{1.0, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{1.0, 1e8},
		{1.0, 1e8 + 1.0},
		{maxUFix64 - 1.0001, 1.0},
		{math.Floor(maxUFix64 / 2), math.Floor(maxUFix64 / 2)},
	}
	margin := 1e-8
	for _, tc := range tests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		res, err := SubUFix64(a, b)
		expected := tc.a - tc.b
		if err != nil {
			t.Errorf("SubUFix64(%f, %f) returned error: %v", tc.a, tc.b, err)
			continue
		}
		fres := fromUFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("SubUFix64(%f, %f) = %f, want %f (±%f)", tc.a, tc.b, fres, expected, margin)
		}
	}
	// Test overflow cases
	overflowTests := []struct {
		a, b float64
	}{
		{maxUFix64, 1.0},
		{maxUFix64, 0.01},
		{maxUFix64, 0.001},
		{maxUFix64, 0.00001},
		{maxUFix64, 0.0000001},
		{maxUFix64, 0.00000001},
		{maxUFix64, maxUFix64},
		{maxUFix64 / 2, maxUFix64 / 2},
		{maxUFix64 / 2, maxUFix64/2 + 1.0},
	}
	for _, tc := range overflowTests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		_, err := SubUFix64(a, b)
		if err != ErrOverflow {
			t.Errorf("SubUFix64(%f, %f) expected overflow error, got: %v", tc.a, tc.b, err)
		}
	}
}

func TestMulUFix64(t *testing.T) {
	tests := []struct {
		a, b float64
	}{
		{1.0, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{1.0, 1e8},
		{1.0, 1e8 + 1.0},
		{maxUFix64 - 1.0001, 1.0},
		{math.Floor(maxUFix64 / 2), math.Floor(maxUFix64 / 2)},
	}
	margin := 1e-8
	for _, tc := range tests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		res, err := MulUFix64(a, b)
		expected := tc.a * tc.b
		if err != nil {
			t.Errorf("MulUFix64(%f, %f) returned error: %v", tc.a, tc.b, err)
			continue
		}
		fres := fromUFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("MulUFix64(%f, %f) = %f, want %f (±%f)", tc.a, tc.b, fres, expected, margin)
		}
	}
	// Test overflow cases
	overflowTests := []struct {
		a, b float64
	}{
		{maxUFix64, 1.0},
		{maxUFix64, 0.01},
		{maxUFix64, 0.001},
		{maxUFix64, 0.00001},
		{maxUFix64, 0.0000001},
		{maxUFix64, maxUFix64},
		{maxUFix64 / 2, maxUFix64 / 2},
		{maxUFix64 / 2, maxUFix64/2 + 1.0},
	}
	for _, tc := range overflowTests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		_, err := MulUFix64(a, b)
		if err != ErrOverflow {
			t.Errorf("MulUFix64(%f, %f) expected overflow error, got: %v", tc.a, tc.b, err)
		}
	}
}

func TestDivUFix64(t *testing.T) {
	tests := []struct {
		a, b float64
	}{
		{1.0, 1.0},
		{1.0, 0.0},
		{0.0, 0.0},
		{1.0, 1e8},
		{1.0, 1e8 + 1.0},
	}
	margin := 1e-8
	for _, tc := range tests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		res, err := DivUFix64(a, b)
		expected := tc.a / tc.b
		if err != nil {
			t.Errorf("DivUFix64(%f, %f) returned error: %v", tc.a, tc.b, err)
			continue
		}
		fres := fromUFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("DivUFix64(%f, %f) = %f, want %f (±%f)", tc.a, tc.b, fres, expected, margin)
		}
	}
	// Test overflow cases
	overflowTests := []struct {
		a, b float64
	}{
		{maxUFix64, 1.0},
		{maxUFix64, 0.01},
		{maxUFix64, 0.001},
		{maxUFix64, 0.00001},
		{maxUFix64, 0.0000001},
		{maxUFix64, maxUFix64},
		{maxUFix64 / 2, maxUFix64 / 2},
		{maxUFix64 / 2, maxUFix64/2 + 1.0},
	}
	for _, tc := range overflowTests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		_, err := DivUFix64(a, b)
		if err != ErrOverflow {
			t.Errorf("DivUFix64(%f, %f) expected overflow error, got: %v", tc.a, tc.b, err)
		}
	}
}

func TestFMDUFix64(t *testing.T) {
	tests := []struct {
		a, b, c float64
	}{
		{1.0, 1.0, 1.0},
		{1.0, 0.0, 1.0},
		{0.0, 1.0, 0.0},
		{1.0, 1e8, 1e-8},
		{1.0, 1e8 + 1.0, 1e-8},
		{maxUFix64 - 1.0001, 1.0, 1.0},
		{math.Floor(maxUFix64 / 2), math.Floor(maxUFix64 / 2), 1.0},
	}
	margin := 1e-8
	for _, tc := range tests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		c := toUFix64(tc.c)
		res, err := FMDUFix64(a, b, c)
		expected := tc.a * tc.b / tc.c
		if err != nil {
			t.Errorf("FMDUFix64(%f, %f, %f) returned error: %v", tc.a, tc.b, tc.c, err)
			continue
		}
		fres := fromUFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("FMDUFix64(%f, %f, %f) = %f, want %f (±%f)", tc.a, tc.b, tc.c, fres, expected, margin)
		}
	}
	// Test overflow cases
	overflowTests := []struct {
		a, b float64
		c    float64
	}{
		{maxUFix64, 1.0, 1.0},
		{maxUFix64, 0.01, 1.0},
		{maxUFix64, 0.001, 1.0},
		{maxUFix64, 0.00001, 1.0},
		{maxUFix64, 0.0000001, 1.0},
		{maxUFix64, maxUFix64, 1.0},
		{maxUFix64 / 2, maxUFix64 / 2, 1.0},
		{maxUFix64 / 2, maxUFix64/2 + 1.0, 1.0},
	}
	for _, tc := range overflowTests {
		a := toUFix64(tc.a)
		b := toUFix64(tc.b)
		c := toUFix64(tc.c)
		_, err := FMDUFix64(a, b, c)
		if err != ErrOverflow {
			t.Errorf("FMDUFix64(%f, %f, %f) expected overflow error, got: %v", tc.a, tc.b, tc.c, err)
		}
	}
}

func TestSqrtUFix64(t *testing.T) {
	tests := []struct {
		a float64
	}{
		{1.0},
		{2.0},
		{3.0},
		{4.0},
		{5.0},
		{6.0},
		{7.0},
		{8.0},
		{9.0},
		{10.0},
		{16.0},
		{25.0},
		{49.0},
		{64.0},
		{81.0},
		{100.0},
		{1000.0},
		{10000.0},
		{100000.0},
		{1000000.0},
		{10000000.0},
		{100000000.0},
		{1000000000.0},
		{maxUFix64},
		{0.0},
		{0.1},
		{0.01},
		{0.001},
		{0.0001},
		{0.00001},
		{0.000001},
	}
	margin := 2e-8
	for _, tc := range tests {
		fx := toUFix64(tc.a)
		res, err := SqrtUFix64(fx)
		expected := math.Sqrt(tc.a)
		if err != nil {
			t.Errorf("SqrtUFix64(%f) returned error: %v", tc.a, err)
			continue
		}
		fres := fromUFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("SqrtUFix64(%f) = %f, want %f (±%f)", tc.a, fres, expected, margin)
		}
	}
}

func TestLn(t *testing.T) {
	tests := []float64{
		2.7182818,
		1.0,
		1.1,
		1.01,
		1.001,
		1.0001,
		1.00001,
		// 1.000001,
		// 1.0000001,
		// 1.00000001,
		0.1,
		0.01,
		0.001,
		0.0001,
		0.00001,
		0.000001,
		0.0000001,
		0.00000001,
		0.9,
		0.99,
		0.999,
		0.9999,
		0.99999,
		// 0.999999,
		// 0.9999999,
		// 0.99999999,
		0.5,
		10.0,
		20.0,
		50.0,
		100.0,
		500.0,
		1000.0,
		5000.0,
		10000.0,
		3.1415927,
		7.3890561,
		15.0,
		25.0,
		75.0,
		250.0,
		750.0,
		2500.0,
		7500.0,
		100000.0,
		1000000.0,
		10000000.0,
		100000000.0,
		1000000000.0,
		math.MaxInt64 / 1e8,
	}
	for _, tc := range tests {
		fx := toUFix64(tc)
		res, err := Ln(fx)
		expected := math.Log(tc)
		margin := 2e-8
		if err != nil {
			t.Errorf("Ln(%.8f) returned error: %v", tc, err)
			continue
		}
		fres := fromFix64(res)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("Ln(%.8f) = %.8f, want %.8f (±%.8f)", tc, fres, expected, margin)
		}
	}
}

func TestExp(t *testing.T) {
	tests := []struct {
		input float64
	}{
		{0.0},
		{1.0},
		{2.0},
		{5.0},
		{7.9},
		{7.99},
		{8.0},
		{8.01},
		{8.1},
		{10.0},
		{15.0},
		{15.9},
		{15.99},
		{16.0},
		{16.01},
		{16.1},
		{17.0},
		{20.0},
		{25.0},
		{25.2},
		{-1.0},
		{-2.0},
		{-5.0},
		{-6.0},
		{-7.0},
		{-8.0},
		{-9.0},
		{-10.0},
		{-11.0},
		{-12.0},
		{-13.0},
		{-14.0},
		{-15.0},
		{-15.1},
		{-15.2},
		{-15.3},
		{-15.4},
		{-15.5},
		{-15.6},
		{-15.7},
		{-15.8},
		{-15.9},
		{-16.0},
		{-17.0},
		{-18.0},
	}
	margin := 2e-8

	for _, tc := range tests {
		fx := toFix64(tc.input)
		res, err := Exp(fx)
		if err != nil {
			t.Errorf("Exp(%f) returned error: %v", tc.input, err)
			continue
		}
		fres := fromFix64(res)
		expected := math.Exp(tc.input)
		scaledMargin := margin
		if expected > 100 {
			scaledMargin = expected / 1e9
		}
		if diff := math.Abs(fres - expected); diff > scaledMargin {
			t.Errorf("Exp(%.8f) = %.8f, want %.8f (±%.8f)", tc, fres, expected, math.Abs(diff))
		}
	}
}

func TestPow(t *testing.T) {
	tests := []struct {
		a, b float64
	}{
		{2.0, 3.0},
		{9.0, 0.5},
		{27.0, 1.0 / 3.0},
		{5.0, 0.0},
		{0.0, 5.0},
	}
	margin := 2e-8

	for _, tc := range tests {
		fxa := toFix64(tc.a)
		fxb := toFix64(tc.b)
		res, err := Pow(fxa, fxb)
		if err != nil {
			t.Errorf("Pow(%f, %f) returned error: %v", tc.a, tc.b, err)
			continue
		}
		fres := fromFix64(res)
		expected := math.Pow(tc.a, tc.b)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("Pow(%.8f, %.8f) = %.8f, want %.8f (±%.8f)", tc.a, tc.b, fres, expected, diff)
		}
	}
}

func TestSin(t *testing.T) {
	tests := []struct {
		input float64
	}{
		{0.0},
		{0.1},
		{0.01},
		{0.001},
		{0.0001},
		{0.00001},
		{0.000001},
		{0.0000001},
		{0.00000001},
		{0.00391486},
		{0.00391486 + 1e-8},
		{0.00391486 - 1e-8},
		{0.2},
		{0.28761102},
		{0.3},
		{0.4},
		{0.5},
		{0.6},
		{0.7},
		{0.8},
		{0.9},
		{1.0},
		{2.0},
		{3.0},
		{4.0},
		{5.0},
		{6.0},
		{7.0},
		{2 - math.Pi/2},
		{2 + 3*math.Pi/2},
		{math.Pi / 2},
		{math.Pi},
		{3 * math.Pi / 2},
		{2 * math.Pi},
		{-math.Pi / 2},
		{-math.Pi},
		{-3 * math.Pi / 2},
		{-2 * math.Pi},
	}
	margin := 2e-8

	for _, tc := range tests {
		fx := toFix64(tc.input)
		res, err := Sin(fx)
		if err != nil {
			t.Errorf("Sin(%f) returned error: %v", tc.input, err)
			continue
		}
		fres := fromFix64(res)
		expected := math.Sin(tc.input)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("Sin(%.8f) = %.8f, want %.8f (±%.8f)", tc.input, fres, expected, diff)
		}
	}
}

func TestCos(t *testing.T) {
	tests := []struct {
		input float64
	}{
		{0.0},
		{0.1},
		{0.01},
		{0.001},
		{0.0001},
		{0.00001},
		{0.000001},
		{0.0000001},
		{0.00000001},
		{0.2},
		{0.28761102},
		{0.3},
		{0.4},
		{0.5},
		{0.6},
		{0.7},
		{0.8},
		{0.9},
		{1.0},
		{2.0},
		{3.0},
		{4.0},
		{5.0},
		{6.0},
		{7.0},
		{math.Pi / 2},
		{math.Pi},
		{3 * math.Pi / 2},
		{2 * math.Pi},
		{-math.Pi / 2},
		{-math.Pi},
		{-3 * math.Pi / 2},
		{-2 * math.Pi},
	}
	margin := 2e-8

	for _, tc := range tests {
		fx := toFix64(tc.input)
		res, err := Cos(fx)
		if err != nil {
			t.Errorf("Cos(%f) returned error: %v", tc.input, err)
			continue
		}
		fres := fromFix64(res)
		expected := math.Cos(tc.input)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("Cos(%.8f) = %.8f, want %.8f (±%.8f)", tc.input, fres, expected, diff)
		}
	}
}

func TestTan(t *testing.T) {
	tests := []struct {
		input float64
	}{
		{0.0},
		{0.1},
		{0.01},
		{0.001},
		{0.0001},
		{0.00001},
		{0.000001},
		{0.0000001},
		{0.00000001},
		{0.2},
		{0.28761102},
		{0.3},
		{0.4},
		{0.5},
		{0.6},
		{0.7},
		{0.8},
		{0.9},
		{1.0},
		{2.0},
		{3.0},
		{4.0},
		{5.0},
		{6.0},
		{7.0},
		{math.Pi / 4},
		{math.Pi / 3},
		{math.Pi},
		// {3 * math.Pi / 2},
		{2 * math.Pi},
		{-math.Pi / 4},
		// {-math.Pi / 2},
		{-math.Pi},
		// {-3 * math.Pi / 2},
		{-2 * math.Pi},
	}
	margin := 4e-8

	for _, tc := range tests {
		fx := toFix64(tc.input)
		res, err := Tan(fx)
		if err != nil {
			t.Errorf("Tan(%f) returned error: %v", tc.input, err)
			continue
		}
		fres := fromFix64(res)
		expected := math.Tan(tc.input)
		if diff := math.Abs(fres - expected); diff > margin {
			t.Errorf("Tan(%.8f) = %.8f, want %.8f (±%.8f)", tc.input, fres, expected, diff)
		}
	}
}
