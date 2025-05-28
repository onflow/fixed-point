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

func BenchmarkAddUFix64(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	for i := 0; i < b.N; i++ {
		_, _ = a.Add(c)
	}
}

func BenchmarkAddUFix64_Reference(b *testing.B) {
	a := 123456789
	c := 987654321
	for i := 0; i < b.N; i++ {
		_ = a + c
	}
}

// func BenchmarkSubUFix64(b *testing.B) {
// 	a := UFix64(987654321)
// 	c := UFix64(123456789)
// 	for i := 0; i < b.N; i++ {
// 		_, _ = a.Sub(c)
// 	}
// }

// func BenchmarkMulUFix64(b *testing.B) {
// 	a := UFix64(123456789)
// 	c := UFix64(987654321)
// 	for i := 0; i < b.N; i++ {
// 		_, _ = a.Mul(c)
// 	}
// }

// func BenchmarkDivUFix64(b *testing.B) {
// 	a := UFix64(987654321)
// 	c := UFix64(123456789)
// 	for i := 0; i < b.N; i++ {
// 		_, _ = a.Div(c)
// 	}
// }

// func BenchmarkFMDUFix64(b *testing.B) {
// 	a := UFix64(123456789)
// 	c := UFix64(987654321)
// 	d := UFix64(55555555)
// 	for i := 0; i < b.N; i++ {
// 		_, _ = a.FMD(c, d)
// 	}
// }

// func BenchmarkAbsFix64(b *testing.B) {
// 	a := Fix64(-123456789)
// 	for i := 0; i < b.N; i++ {
// 		_ = a.Abs()
// 	}
// }
