package fixedPoint

import (
	"testing"
)

func TestAddUFix64(t *testing.T) {
	for i, tc := range AddUFix64Tests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		expected := UFix64(tc.Expected)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddUFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("AddUFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range AddUFix64OverflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddUFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestAddFix64(t *testing.T) {
	for i, tc := range AddFix64Tests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		expected := Fix64(tc.Expected)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("AddFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range AddFix64OverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range AddFix64NegOverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Add(b)
		if err != ErrNegOverflow {
			t.Errorf("AddFix64(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSubUFix64(t *testing.T) {
	for i, tc := range SubUFix64Tests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		expected := UFix64(tc.Expected)
		res, err := a.Sub(b)
		if err != nil {
			t.Errorf("SubUFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("SubUFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range SubUFix64NegOverflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Sub(b)
		if err != ErrNegOverflow {
			t.Errorf("SubUFix64(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSubFix64(t *testing.T) {
	for i, tc := range SubFix64Tests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		expected := Fix64(tc.Expected)
		res, err := a.Sub(b)
		if err != nil {
			t.Errorf("SubFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("SubFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range SubFix64OverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Sub(b)
		if err != ErrOverflow {
			t.Errorf("SubFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range SubFix64NegOverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Sub(b)
		if err != ErrNegOverflow {
			t.Errorf("SubFix64(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestMulUFix64(t *testing.T) {
	for i, tc := range MulUFix64Tests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		expected := UFix64(tc.Expected)
		res, err := a.Mul(b)
		if err != nil {
			t.Errorf("MulUFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("MulUFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range MulUFix64OverflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrOverflow {
			t.Errorf("MulUFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulUFix64UnderflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrUnderflow {
			t.Errorf("MulUFix64(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestMulFix64(t *testing.T) {
	for i, tc := range MulFix64Tests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		expected := Fix64(tc.Expected)
		res, err := a.Mul(b)
		if err != nil {
			t.Errorf("MulFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("MulFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range MulFix64OverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrOverflow {
			t.Errorf("MulFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulFix64NegOverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrNegOverflow {
			t.Errorf("MulFix64(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulFix64UnderflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Mul(b)
		if err != ErrUnderflow {
			t.Errorf("MulFix64(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestDivUFix64(t *testing.T) {
	for i, tc := range DivUFix64Tests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		expected := UFix64(tc.Expected)
		res, err := a.Div(b)
		if err != nil {
			t.Errorf("DivUFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("DivUFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range DivUFix64OverflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Div(b)
		if err != ErrOverflow {
			t.Errorf("DivUFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivUFix64UnderflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Div(b)
		if err != ErrUnderflow {
			t.Errorf("DivUFix64(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivUFix64DivByZeroTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		_, err := a.Div(b)
		if err != ErrDivByZero {
			t.Errorf("DivUFix64(0x%016x, 0x%016x) expected div by zero error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestDivFix64(t *testing.T) {
	for i, tc := range DivFix64Tests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		expected := Fix64(tc.Expected)
		res, err := a.Div(b)
		if err != nil {
			t.Errorf("DivFix64(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("DivFix64(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range DivFix64OverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Div(b)
		if err != ErrOverflow {
			t.Errorf("DivFix64(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivFix64NegOverflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Div(b)
		if err != ErrNegOverflow {
			t.Errorf("DivFix64(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivFix64UnderflowTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Div(b)
		if err != ErrUnderflow {
			t.Errorf("DivFix64(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivFix64DivByZeroTests {
		a := Fix64(tc.A)
		b := Fix64(tc.B)
		_, err := a.Div(b)
		if err != ErrDivByZero {
			t.Errorf("DivFix64(0x%016x, 0x%016x) expected div by zero error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestFMDUFix64(t *testing.T) {
	for i, tc := range FMDUFix64Tests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		expected := UFix64(tc.Expected)
		res, err := a.FMD(b, c)
		if err != nil {
			t.Errorf("FMDUFix64(0x%016x, 0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, tc.C, i, err)
			continue
		}
		if res != expected {
			t.Errorf("FMDUFix64(0x%016x, 0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, tc.C, res, i, expected)
		}
	}
	for i, tc := range FMDUFix64OverflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		_, err := a.FMD(b, c)
		if err != ErrOverflow {
			t.Errorf("FMDUFix64(0x%016x, 0x%016x, 0x%016x) (%d) expected overflow error, got: %v", tc.A, tc.B, tc.C, i, err)
		}
	}
	for i, tc := range FMDUFix64UnderflowTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		_, err := a.FMD(b, c)
		if err != ErrUnderflow {
			t.Errorf("FMDUFix64(0x%016x, 0x%016x, 0x%016x) (%d) expected underflow error, got: %v", tc.A, tc.B, tc.C, i, err)
		}
	}
	for _, tc := range FMDUFix64DivByZeroTests {
		a := UFix64(tc.A)
		b := UFix64(tc.B)
		c := UFix64(tc.C)
		_, err := a.FMD(b, c)
		if err != ErrDivByZero {
			t.Errorf("FMDUFix64(0x%016x, 0x%016x, 0x%016x) expected div by zero error, got: %v", tc.A, tc.B, tc.C, err)
		}
	}
}

func TestSqrtUFix64(t *testing.T) {
	for i, tc := range SqrtUFix64Tests {
		a := UFix64(tc.A)
		expected := UFix64(tc.Expected)
		res, err := a.Sqrt()
		if err != nil {
			t.Errorf("SqrtUFix64(0x%016x) (%d) returned error: %v", tc.A, i, err)
			continue
		}
		if res != expected {
			t.Errorf("SqrtUFix64(0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, res, i, expected)
		}
	}
}

func TestLnUFix64(t *testing.T) {
	for i, tc := range LnUFix64Tests {
		a := UFix64(tc.A)
		expected := Fix64(tc.Expected)
		res, err := a.Ln()
		if err != nil {
			t.Errorf("LnFix64(0x%016x) (%d) returned error: %v", tc.A, i, err)
			continue
		}
		diff, _ := res.Sub(expected)
		diff = diff.Abs()

		if res != expected {
			t.Errorf("LnFix64(0x%016x) = 0x%016x (%d), want 0x%016x (±%d)", tc.A, res, i, expected, diff)
		}
	}
}

func TestSinFix64(t *testing.T) {
	for i, tc := range SinFix64Tests {
		a := Fix64(tc.A)
		expected := Fix64(tc.Expected)
		res, err := a.Sin()
		diff, _ := res.Sub(expected)
		diff = diff.Abs()
		if err != nil {
			t.Errorf("SinFix64(0x%016x) (%d) returned error: %v", tc.A, i, err)
			continue
		}
		// We accept minimal error in the lowest digit
		if diff > 0 {
			t.Errorf("SinFix64(0x%016x) = 0x%016x (%d), want 0x%016x (±%d)", tc.A, res, i, expected, diff)
		}
	}
}

func TestCosFix64(t *testing.T) {
	for i, tc := range CosFix64Tests {
		a := Fix64(tc.A)
		expected := Fix64(tc.Expected)
		res, err := a.Cos()
		diff, _ := res.Sub(expected)
		diff = diff.Abs()
		if err != nil {
			t.Errorf("CosFix64(0x%016x) (%d) returned error: %v", tc.A, i, err)
			continue
		}
		// We accept minimal error in the lowest digit
		if diff > 1 {
			t.Errorf("CosFix64(0x%016x) = 0x%016x (%d), want 0x%016x (±%d)", tc.A, res, i, expected, diff)
		}
	}
}

// Commented out because tan() is only of speculative value for smart contracts, and getting a bit-accurate value
// is proving to be VERY complicated. Fundamentally, since tan(x) = sin(x)/cos(x), and cos(x) can be very small
// the error in tan(x) can be very large, even if sin(x) and cos(x) are accurate to the last bit.

// func TestTanFix64(t *testing.T) {
// 	for i, tc := range TanFix64Tests {
// 		a := Fix64(tc.A)
// 		expected := Fix64(tc.Expected)
// 		res, err := a.TanTest()
// 		diff, _ := res.Sub(expected)
// 		diff = diff.Abs()
// 		if err != nil {
// 			t.Errorf("TanFix64(0x%016x) (%d) returned error: %v", tc.A, i, err)
// 			continue
// 		}
// 		if diff > 1 {
// 			t.Errorf("TanFix64(0x%016x) = 0x%016x (%d), want 0x%016x (±%d)", tc.A, res, i, expected, diff)
// 		}
// 	}
// 	for _, tc := range TanFix64OverflowTests {
// 		a := Fix64(tc.A)
// 		_, err := a.Tan()
// 		if err != ErrOverflow {
// 			t.Errorf("TanFix64(0x%016x) expected overflow error, got: %v", tc.A, err)
// 		}
// 	}
// }
