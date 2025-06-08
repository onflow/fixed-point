package fixedPoint

import (
	"testing"
)

func TestAddUFix128(t *testing.T) {
	for i, tc := range AddUFix128Tests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		expected := UFix128(tc.Expected)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddUFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("AddUFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range AddUFix128OverflowTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddUFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestAddFix128(t *testing.T) {
	for i, tc := range AddFix128Tests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		expected := Fix128(tc.Expected)
		res, err := a.Add(b)
		if err != nil {
			t.Errorf("AddFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("AddFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range AddFix128OverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Add(b)
		if err != ErrOverflow {
			t.Errorf("AddFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range AddFix128NegOverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Add(b)
		if err != ErrNegOverflow {
			t.Errorf("AddFix128(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSubUFix128(t *testing.T) {
	for i, tc := range SubUFix128Tests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		expected := UFix128(tc.Expected)
		res, err := a.Sub(b)
		if err != nil {
			t.Errorf("SubUFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("SubUFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range SubUFix128NegOverflowTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Sub(b)
		if err != ErrNegOverflow {
			t.Errorf("SubUFix128(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSubFix128(t *testing.T) {
	for i, tc := range SubFix128Tests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		expected := Fix128(tc.Expected)
		res, err := a.Sub(b)
		if err != nil {
			t.Errorf("SubFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("SubFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range SubFix128OverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Sub(b)
		if err != ErrOverflow {
			t.Errorf("SubFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range SubFix128NegOverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Sub(b)
		if err != ErrNegOverflow {
			t.Errorf("SubFix128(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestMulUFix128(t *testing.T) {
	for i, tc := range MulUFix128Tests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		expected := UFix128(tc.Expected)
		res, err := a.Mul(b)
		if err != nil {
			t.Errorf("MulUFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("MulUFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range MulUFix128OverflowTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Mul(b)
		if err != ErrOverflow {
			t.Errorf("MulUFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulUFix128UnderflowTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Mul(b)
		if err != ErrUnderflow {
			t.Errorf("MulUFix128(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestMulFix128(t *testing.T) {
	for i, tc := range MulFix128Tests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		expected := Fix128(tc.Expected)
		res, err := a.Mul(b)
		if err != nil {
			t.Errorf("MulFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("MulFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range MulFix128OverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Mul(b)
		if err != ErrOverflow {
			t.Errorf("MulFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulFix128NegOverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Mul(b)
		if err != ErrNegOverflow {
			t.Errorf("MulFix128(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range MulFix128UnderflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Mul(b)
		if err != ErrUnderflow {
			t.Errorf("MulFix128(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestDivUFix128(t *testing.T) {
	for i, tc := range DivUFix128Tests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		expected := UFix128(tc.Expected)
		res, err := a.Div(b)
		if err != nil {
			t.Errorf("DivUFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("DivUFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range DivUFix128OverflowTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Div(b)
		if err != ErrOverflow {
			t.Errorf("DivUFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivUFix128UnderflowTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Div(b)
		if err != ErrUnderflow {
			t.Errorf("DivUFix128(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivUFix128DivByZeroTests {
		a := UFix128(tc.A)
		b := UFix128(tc.B)
		_, err := a.Div(b)
		if err != ErrDivByZero {
			t.Errorf("DivUFix128(0x%016x, 0x%016x) expected div by zero error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestDivFix128(t *testing.T) {
	for i, tc := range DivFix128Tests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		expected := Fix128(tc.Expected)
		res, err := a.Div(b)
		if err != nil {
			t.Errorf("DivFix128(0x%016x, 0x%016x) (%d) returned error: %v", tc.A, tc.B, i, err)
			continue
		}
		if res != expected {
			t.Errorf("DivFix128(0x%016x, 0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, tc.B, res, i, expected)
		}
	}
	for _, tc := range DivFix128OverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Div(b)
		if err != ErrOverflow {
			t.Errorf("DivFix128(0x%016x, 0x%016x) expected overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivFix128NegOverflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Div(b)
		if err != ErrNegOverflow {
			t.Errorf("DivFix128(0x%016x, 0x%016x) expected negative overflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivFix128UnderflowTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Div(b)
		if err != ErrUnderflow {
			t.Errorf("DivFix128(0x%016x, 0x%016x) expected underflow error, got: %v", tc.A, tc.B, err)
		}
	}
	for _, tc := range DivFix128DivByZeroTests {
		a := Fix128(tc.A)
		b := Fix128(tc.B)
		_, err := a.Div(b)
		if err != ErrDivByZero {
			t.Errorf("DivFix128(0x%016x, 0x%016x) expected div by zero error, got: %v", tc.A, tc.B, err)
		}
	}
}

func TestSqrtUFix128(t *testing.T) {
	for i, tc := range SqrtUFix64Tests {
		a := UFix128{0, tc.A}
		expected := UFix128{0, tc.Expected}
		a = a.intMul(1e16)
		expected = expected.intMul(1e16)
		res, err := a.SqrtTest()
		if err != nil {
			t.Errorf("SqrtUFix128(0x%016x) (%d) returned error: %v", tc.A, i, err)
			continue
		}
		if res != expected {
			t.Errorf("SqrtUFix128(0x%016x) = 0x%016x (%d), want 0x%016x", tc.A, res, i, expected)
		}
	}
}

// func TestLnUFix128(t *testing.T) {
// 	for i, tc := range LnUFix128Tests {
// 		a := UFix128(tc.A)
// 		expected := Fix128(tc.Expected)
// 		res, err := a.Ln()
// 		if err != nil {
// 			t.Errorf("LnFix128(0x%016x) (%d) returned error: %v", tc.A, i, err)
// 			continue
// 		}
// 		diff, _ := res.Sub(expected)
// 		diff = diff.Abs()

// 		if res != expected {
// 			t.Errorf("LnFix128(0x%016x) = 0x%016x (%d), want 0x%016x (±%d)", tc.A, res, i, expected, diff)
// 		}
// 	}
// }
