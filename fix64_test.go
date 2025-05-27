package fixedPoint

import (
	"fmt"
	"math"
	"testing"

	"github.com/ericlagergren/decimal"
)

var decCtx = decimal.Context64

// Converts a UFix64 to a decimal.Big
func ufix64ToDecimal(f UFix64) *decimal.Big {
	// UFix64 is an integer representing value * fix64Scale
	// So: value = float64(f) / fix64Scale
	// To avoid float conversion, use integer math: f / fix64Scale
	// But decimal.Big can take string, so use string formatting
	return new(decimal.Big).Quo(
		new(decimal.Big).SetUint64(uint64(f)),
		new(decimal.Big).SetUint64(uint64(fix64Scale)),
	)
}

// Converts a Fix64 to a decimal.Big
func fix64ToDecimal(f Fix64) *decimal.Big {
	neg := false
	uf := uint64(f)
	if f < 0 {
		neg = true
		uf = uint64(-f)
	}
	dec := new(decimal.Big).Quo(
		new(decimal.Big).SetUint64(uf),
		new(decimal.Big).SetUint64(uint64(fix64Scale)),
	)
	if neg {
		dec.Neg(dec)
	}
	return dec
}

func toUFix64(d *decimal.Big) UFix64 {
	scaled := decimal.WithPrecision(30).Mul(d, new(decimal.Big).SetUint64(uint64(fix64Scale))).RoundToInt()

	f, b := scaled.Uint64()
	if !b {
		// Print out an error message with the value that caused the panic
		fmt.Printf("toUFix64: value out of range: %s\n", d.String())
		panic("toUFix64: value out of range")
	}

	return UFix64(f)
}

func toFix64(d *decimal.Big) Fix64 {
	scaled := decimal.WithPrecision(30).Mul(d, new(decimal.Big).SetUint64(uint64(fix64Scale)))

	// There is a limitation in the decimal library that it treats large negative integers
	// as being unrepresentable as Int64, even when they are. So, we do the conversion
	// ourselves if the value is negative.
	if scaled.Sign() < 0 {
		// Flip to positive
		scaled.Neg(scaled)

		f, b := scaled.Uint64()
		if !b {
			// Print out an error message with the value that caused the panic
			fmt.Printf("toFix64: value out of range: %s\n", d.String())
			panic("toFix64: value out of range")
		}

		// NOTE: The most negative value has an absolute value one greater than the
		// most positive value!
		if f == (1 << 63) {
			// The input is "the most negative value"
			return Fix64(math.MinInt64)
		} else if f > (1 << 63) {
			// The number is too negative to fit in a signed int64
			fmt.Printf("toFix64: value out of range: %s\n", d.String())
			panic("toFix64: value out of range")
		} else {
			// Just flip the sign back
			return Fix64(-f)
		}
	} else {
		f, b := scaled.Int64()
		if !b {
			// Print out an error message with the value that caused the panic
			fmt.Printf("toFix64: value out of range: %s\n", d.String())
			panic("toFix64: value out of range")
		}

		return Fix64(f)
	}
}

func TestAddUFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, tc := range AddUFix64Tests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddUFix64(%.8f, %.8f) returned error: %v", tc.A, tc.B, err)
			continue
		}
		expected := new(decimal.Big).Add(tc.A, tc.B)
		decRes := ufix64ToDecimal(res)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("AddUFix64(%.8f, %.8f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range AddUFix64OverflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddUFix64(%.8f, %.8f) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestAddFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, tc := range AddFix64Tests {
		a := toFix64(tc.A)
		b := toFix64(tc.B)
		res, err := a.Add(b)
		if err != nil && tc.A.Sign() >= 0 && tc.B.Sign() >= 0 {
			t.Errorf("AddFix64(%s, %s) unexpected error: %v", tc.A, tc.B, err)
			continue
		}
		if err == nil {
			expected := new(decimal.Big).Add(tc.A, tc.B)
			decRes := fix64ToDecimal(res)
			fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
			if fdiff.Cmp(margin) >= 0 {
				t.Errorf("AddFix64(%s, %s) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
			}
		}
	}
	// Overflow tests for AddFix64
	for _, tc := range AddFix64OverflowTests {
		a := toFix64(tc.A)
		b := toFix64(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddFix64(%s, %s) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	// Negative overflow tests for AddFix64
	for _, tc := range AddFix64NegOverflowTests {
		a := toFix64(tc.A)
		b := toFix64(tc.B)
		_, err := a.Add(b)
		if err != ErrNegOverflow {
			t.Errorf("AddFix64(%s, %s) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSubUFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for i, tc := range SubUFix64Tests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		res, err := a.Sub(b)
		if err != nil {
			t.Errorf("SubUFix64(%s, %s) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		decRes := ufix64ToDecimal(res)
		expected := new(decimal.Big).Sub(tc.A, tc.B)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("SubUFix64(%s, %s) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range SubUFix64NegOverflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		_, err := a.Sub(b)
		if err != ErrNegOverflow {
			t.Errorf("SubUFix64(%s, %s) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSubFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for i, tc := range SubFix64Tests {
		a := toFix64(tc.A)
		b := toFix64(tc.B)
		res, err := a.Sub(b)
		if err != nil && !(tc.A.Sign() < 0 && tc.B.Sign() > 0) {
			t.Errorf("SubFix64(%s, %s) (%d) unexpected error: %v", tc.A, tc.B, i, err)
			continue
		}
		if err == nil {
			expected := new(decimal.Big).Sub(tc.A, tc.B)
			decRes := fix64ToDecimal(res)
			fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
			if fdiff.Cmp(margin) >= 0 {
				t.Errorf("SubFix64(%s, %s) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
			}
		}
	}
	// Overflow tests for SubFix64
	for i, tc := range SubFix64OverflowTests {
		a := toFix64(tc.A)
		b := toFix64(tc.B)
		_, err := a.Sub(b)
		if err != ErrOverflow {
			t.Errorf("SubFix64(%s, %s) (%d) expected overflow error, got: %v", tc.A, tc.B, i, err)
		}
	}
	// Negative overflow tests for SubFix64
	for i, tc := range SubFix64NegOverflowTests {
		a := toFix64(tc.A)
		b := toFix64(tc.B)
		_, err := a.Sub(b)
		if err != ErrNegOverflow {
			t.Errorf("SubFix64(%s, %s) (%d) expected negative overflow error, got: %v", tc.A, tc.B, i, err)
		}
	}
}

func TestMulUFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for i, tc := range MulUFix64Tests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		res, err := a.Mul(b)
		if err != nil {
			t.Errorf("MulUFix64(%s, %s) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		decRes := ufix64ToDecimal(res)
		expected := new(decimal.Big).Mul(tc.A, tc.B)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("MulUFix64(%s, %s) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range MulUFix64OverflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrOverflow {
			t.Errorf("MulUFix64(%s, %s) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulUFix64UnderflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrUnderflow {
			t.Errorf("MulUFix64(%s, %s) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestDivUFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, tc := range DivUFix64Tests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		res, err := a.Div(b)
		if err != nil {
			t.Errorf("DivUFix64(%s, %s) returned error: %v", tc.A, tc.B, err)
			continue
		}
		decRes := ufix64ToDecimal(res)
		if tc.B.Cmp(decimal.New(0, 0)) == 0 {
			t.Errorf("DivUFix64(%s, %s) division by zero", tc.A, tc.B)
			continue
		}
		expected := new(decimal.Big).Quo(tc.A, tc.B)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("DivUFix64(%s, %s) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range DivUFix64OverflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		_, err := a.Div(b)
		if err != ErrOverflow {
			t.Errorf("DivUFix64(%s, %s) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for i, tc := range DivUFix64UnderflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		_, err := a.Div(b)
		if err != ErrUnderflow {
			t.Errorf("DivUFix64(%s, %s) (%d) expected underflow error, got: %v", tc.A, tc.B, i, err)
		}
	}
}

func TestFMDUFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, tc := range FMDUFix64Tests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		c := toUFix64(tc.C)
		res, err := a.FMD(b, c)
		if err != nil {
			t.Errorf("FMDUFix64(%s, %s, %s) returned error: %v", tc.A, tc.B, tc.C, err)
			continue
		}
		decRes := ufix64ToDecimal(res)
		if tc.C.Cmp(decimal.New(0, 0)) == 0 {
			t.Errorf("FMDUFix64(%s, %s, %s) division by zero", tc.A, tc.B, tc.C)
			continue
		}
		expected := new(decimal.Big).Quo(new(decimal.Big).Mul(tc.A, tc.B), tc.C)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("FMDUFix64(%s, %s, %s) = %s, want %s (±%s)", tc.A, tc.B, tc.C, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range FMDUFix64OverflowTests {
		a := toUFix64(tc.A)
		b := toUFix64(tc.B)
		c := toUFix64(tc.C)
		_, err := a.FMD(b, c)
		if err != ErrOverflow {
			t.Errorf("FMDUFix64(%s, %s, %s) expected overflow error, got: %v", tc.A, tc.B, tc.C, err)
		}
	}
}

func TestSqrtUFix64(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, a := range SqrtUFix64Tests {
		fx := toUFix64(a)
		res, err := fx.Sqrt()
		if err != nil {
			t.Errorf("SqrtUFix64(%s) returned error: %v", a.String(), err)
			continue
		}
		decRes := ufix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Sqrt(expected, a)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("SqrtUFix64(%s) = %s, want %s (±%s)", a.String(), decRes.String(), expected.String(), margin.String())
		}
	}
}

func TestLn(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, tc := range LnTests {
		fx := toUFix64(tc)
		res, err := fx.Ln()
		if err != nil {
			t.Errorf("Ln(%s) returned error: %v", tc.String(), err)
			continue
		}
		decRes := fix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Log(expected, tc)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("Ln(%s) = %s, want %s (±%s)", tc.String(), decRes.String(), expected.String(), margin.String())
		}
	}
}

func TestExp(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, input := range ExpTests {
		fx := toFix64(input)
		res, err := fx.Exp()
		if err != nil {
			t.Errorf("Exp(%s) returned error: %v", input.String(), err)
			continue
		}
		decRes := fix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Exp(expected, input)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		scaledMargin := margin
		if expected.Cmp(decimal.New(100, 0)) > 0 {
			scaledMargin = new(decimal.Big).Quo(expected, decimal.New(1e9, 0))
		}
		if fdiff.Cmp(scaledMargin) > 0 {
			t.Errorf("Exp(%s) = %s, want %s (±%s)", input.String(), decRes.String(), expected.String(), scaledMargin.String())
		}
	}
}

func TestPow(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, tc := range PowTests {
		fxa := toFix64(tc.A)
		fxb := toFix64(tc.B)
		res, err := fxa.Pow(fxb)
		if err != nil {
			t.Errorf("Pow(%s, %s) returned error: %v", tc.A.String(), tc.B.String(), err)
			continue
		}
		decRes := fix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Pow(expected, tc.A, tc.B)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("Pow(%s, %s) = %s, want %s (±%s)", tc.A.String(), tc.B.String(), decRes.String(), expected.String(), margin.String())
		}
	}
}

func TestSin(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, input := range SinTests {
		fx := toFix64(input)
		res, err := fx.Sin()
		if err != nil {
			t.Errorf("Sin(%s) returned error: %v", input.String(), err)
			continue
		}
		decRes := fix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Sin(expected, input)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("Sin(%s) = %s, want %s (±%s)", input.String(), decRes.String(), expected.String(), margin.String())
		}
	}
}

func TestCos(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, input := range CosTests {
		fx := toFix64(input)
		res, err := fx.Cos()
		if err != nil {
			t.Errorf("Cos(%s) returned error: %v", input.String(), err)
			continue
		}
		decRes := fix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Cos(expected, input)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("Cos(%s) = %s, want %s (±%s)", input.String(), decRes.String(), expected.String(), margin.String())
		}
	}
}

func TestTan(t *testing.T) {
	margin := new(decimal.Big).SetFloat64(1e-8)
	for _, input := range TanTests {
		fx := toFix64(input)
		res, err := fx.Tan()
		if err != nil {
			t.Errorf("Tan(%s) returned error: %v", input.String(), err)
			continue
		}
		decRes := fix64ToDecimal(res)
		expected := new(decimal.Big)
		decCtx.Tan(expected, input)
		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
		if fdiff.Cmp(margin) > 0 {
			t.Errorf("Tan(%s) = %s, want %s (±%s)", input.String(), decRes.String(), expected.String(), margin.String())
		}
	}
}

func BenchmarkAddUFix64(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	for i := 0; i < b.N; i++ {
		_, _ = a.Add(c)
	}
}

func BenchmarkSubUFix64(b *testing.B) {
	a := UFix64(987654321)
	c := UFix64(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = a.Sub(c)
	}
}

func BenchmarkMulUFix64(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	for i := 0; i < b.N; i++ {
		_, _ = a.Mul(c)
	}
}

func BenchmarkDivUFix64(b *testing.B) {
	a := UFix64(987654321)
	c := UFix64(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = a.Div(c)
	}
}

func BenchmarkFMDUFix64(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	d := UFix64(55555555)
	for i := 0; i < b.N; i++ {
		_, _ = a.FMD(c, d)
	}
}

func BenchmarkAbsFix64(b *testing.B) {
	a := Fix64(-123456789)
	for i := 0; i < b.N; i++ {
		_ = a.Abs()
	}
}
