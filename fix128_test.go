package fixedPoint

import (
	"fmt"
	"testing"

	"github.com/ericlagergren/decimal"
)

// Converts a UFix128 to a decimal.Big
func ufix128ToDecimal(f UFix128) *decimal.Big {
	return decuu(f.Hi, f.Lo)
}

// Converts a Fix128 to a decimal.Big
func fix128ToDecimal(f Fix128) *decimal.Big {
	return deciu(int64(f.Hi), f.Lo)
}

func toUFix128(d *decimal.Big) UFix128 {
	scaled := decimal.WithPrecision(60).Mul(d, Fix128Scale).RoundToInt()

	twoToSixtyFour := decimal.WithPrecision(60)
	decCtx128.Pow(twoToSixtyFour, deci(2), deci(64))

	hiDec := decimal.WithPrecision(60)
	loDec := decimal.WithPrecision(60)

	hiDec.QuoRem(scaled, twoToSixtyFour, loDec)

	hi, b := hiDec.Uint64()
	if !b {
		fmt.Printf("toUFix128: value out of range: %s\n", d.String())
		panic("toUFix128: value out of range")
	}

	lo, _ := loDec.Uint64()

	return UFix128{hi, lo}
}

func toFix128(d *decimal.Big) Fix128 {
	isNeg := false

	if d.Sign() < 0 {
		d.Neg(d)
		isNeg = true
	}

	unsignedVal := toUFix128(d)

	if isNeg && unsignedVal.Hi == 0x8000000000000000 && unsignedVal.Lo == 0 {
		// Special case for min value of Fix128
		return Fix128{0x8000000000000000, 0}
	}

	if unsignedVal.Hi > 0x7FFFFFFFFFFFFFFF {
		fmt.Printf("toFix128: value out of range: %s\n", d.String())
	}

	res := Fix128(unsignedVal)

	if isNeg {
		res = res.Neg()
	}

	return res
}

func TestAddUFix128(t *testing.T) {
	margin := decimal.WithPrecision(60).SetMantScale(1, 24)
	for i, tc := range AddUFix128Tests {
		a := toUFix128(tc.A)
		b := toUFix128(tc.B)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddUFix128(%.24f, %.24f) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		expected := decimal.WithPrecision(60).Add(tc.A, tc.B)
		decRes := ufix128ToDecimal(res)
		fdiff := decimal.WithPrecision(60).Sub(decRes, expected).Abs(decimal.WithPrecision(60))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("AddUFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range AddUFix128OverflowTests {
		a := toUFix128(tc.A)
		b := toUFix128(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddUFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestAddFix128(t *testing.T) {
	margin := decimal.WithPrecision(60).SetMantScale(1, 24)
	for i, tc := range AddFix128Tests {
		a := toFix128(tc.A)
		b := toFix128(tc.B)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddFix128(%.24f, %.24f) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		expected := decimal.WithPrecision(60).Add(tc.A, tc.B)
		decRes := fix128ToDecimal(res)
		fdiff := decimal.WithPrecision(60).Sub(decRes, expected).Abs(decimal.WithPrecision(60))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("AddFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for i, tc := range AddFix128OverflowTests {
		a := toFix128(tc.A)
		b := toFix128(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddFix128(%.24f, %.24f) (%d) expected overflow error, got: %v", tc.A, tc.B, i, err)
		}
	}
}

// func TestSubUFix128(t *testing.T) {
// 	margin := new(decimal.Big).SetFloat64(1e-24)
// 	for _, tc := range SubUFix128Tests {
// 		a := toUFix128(tc.A)
// 		b := toUFix128(tc.B)
// 		res, err := a.Sub(b)
// 		if err != nil {
// 			t.Errorf("SubUFix128(%.24f, %.24f) returned error: %v", tc.A, tc.B, err)
// 			continue
// 		}
// 		expected := new(decimal.Big).Sub(tc.A, tc.B)
// 		decRes := ufix128ToDecimal(res)
// 		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
// 		if fdiff.Cmp(margin) >= 0 {
// 			t.Errorf("SubUFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
// 		}
// 	}
// 	for _, tc := range SubUFix128OverflowTests {
// 		a := toUFix128(tc.A)
// 		b := toUFix128(tc.B)
// 		_, err := a.Sub(b)
// 		if err != ErrOverflow {
// 			t.Errorf("SubUFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
// 		}
// 	}
// }

func TestMulUFix128(t *testing.T) {
	margin := decimal.WithPrecision(60).SetMantScale(1, 24)
	for i, tc := range MulUFix128Tests {
		a := toUFix128(tc.A)
		b := toUFix128(tc.B)
		res, err := a.Mul(b)
		if err != nil {
			t.Errorf("MulUFix128(%.24f, %.24f) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		expected := decimal.WithPrecision(60).Mul(tc.A, tc.B)
		decRes := ufix128ToDecimal(res)
		fdiff := decimal.WithPrecision(60).Sub(decRes, expected).Abs(decimal.WithPrecision(60))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("MulUFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range MulUFix128OverflowTests {
		a := toUFix128(tc.A)
		b := toUFix128(tc.B)
		_, err := a.Mul(b)
		if err != ErrOverflow {
			t.Errorf("MulUFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestDivUFix128(t *testing.T) {
	margin := decimal.WithPrecision(60).SetMantScale(1, 24)
	for _, tc := range DivUFix128Tests {
		a := toUFix128(tc.A)
		b := toUFix128(tc.B)
		res, err := a.Div(b)
		if err != nil {
			t.Errorf("DivUFix128(%.24f, %.24f) returned error: %v", tc.A, tc.B, err)
			continue
		}
		expected := decimal.WithPrecision(60).Quo(tc.A, tc.B)
		decRes := ufix128ToDecimal(res)
		fdiff := decimal.WithPrecision(60).Sub(decRes, expected).Abs(decimal.WithPrecision(60))
		if fdiff.Cmp(margin) >= 0 {
			t.Errorf("DivUFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
		}
	}
	for _, tc := range DivUFix128OverflowTests {
		a := toUFix128(tc.A)
		b := toUFix128(tc.B)
		_, err := a.Div(b)
		if err != ErrOverflow {
			t.Errorf("DivUFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

// func TestSubFix128(t *testing.T) {
// 	margin := new(decimal.Big).SetFloat64(1e-24)
// 	for _, tc := range SubFix128Tests {
// 		a := toFix128(tc.A)
// 		b := toFix128(tc.B)
// 		res, err := a.Sub(b)
// 		if err != nil {
// 			t.Errorf("SubFix128(%.24f, %.24f) returned error: %v", tc.A, tc.B, err)
// 			continue
// 		}
// 		expected := new(decimal.Big).Sub(tc.A, tc.B)
// 		decRes := fix128ToDecimal(res)
// 		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
// 		if fdiff.Cmp(margin) >= 0 {
// 			t.Errorf("SubFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
// 		}
// 	}
// 	for _, tc := range SubFix128OverflowTests {
// 		a := toFix128(tc.A)
// 		b := toFix128(tc.B)
// 		_, err := a.Sub(b)
// 		if err != ErrOverflow {
// 			t.Errorf("SubFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
// 		}
// 	}
// }

// func TestMulFix128(t *testing.T) {
// 	margin := new(decimal.Big).SetFloat64(1e-24)
// 	for _, tc := range MulFix128Tests {
// 		a := toFix128(tc.A)
// 		b := toFix128(tc.B)
// 		res, err := a.Mul(b)
// 		if err != nil {
// 			t.Errorf("MulFix128(%.24f, %.24f) returned error: %v", tc.A, tc.B, err)
// 			continue
// 		}
// 		expected := new(decimal.Big).Mul(tc.A, tc.B)
// 		decRes := fix128ToDecimal(res)
// 		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
// 		if fdiff.Cmp(margin) >= 0 {
// 			t.Errorf("MulFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
// 		}
// 	}
// 	for _, tc := range MulFix128OverflowTests {
// 		a := toFix128(tc.A)
// 		b := toFix128(tc.B)
// 		_, err := a.Mul(b)
// 		if err != ErrOverflow {
// 			t.Errorf("MulFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
// 		}
// 	}
// }

// func TestDivFix128(t *testing.T) {
// 	margin := new(decimal.Big).SetFloat64(1e-24)
// 	for _, tc := range DivFix128Tests {
// 		a := toFix128(tc.A)
// 		b := toFix128(tc.B)
// 		res, err := a.Div(b)
// 		if err != nil {
// 			t.Errorf("DivFix128(%.24f, %.24f) returned error: %v", tc.A, tc.B, err)
// 			continue
// 		}
// 		expected := new(decimal.Big).Quo(tc.A, tc.B)
// 		decRes := fix128ToDecimal(res)
// 		fdiff := new(decimal.Big).Sub(decRes, expected).Abs(new(decimal.Big))
// 		if fdiff.Cmp(margin) >= 0 {
// 			t.Errorf("DivFix128(%.24f, %.24f) = %s, want %s (±%s)", tc.A, tc.B, decRes.String(), expected.String(), margin.String())
// 		}
// 	}
// 	for _, tc := range DivFix128OverflowTests {
// 		a := toFix128(tc.A)
// 		b := toFix128(tc.B)
// 		_, err := a.Div(b)
// 		if err != ErrOverflow {
// 			t.Errorf("DivFix128(%.24f, %.24f) expected overflow error, got: %v", tc.A, tc.B, err)
// 		}
// 	}
// }
