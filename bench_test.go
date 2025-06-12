package fixedPoint

import (
	"math/big"
	"testing"
)

func BenchmarkAddUFix64(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	for i := 0; i < b.N; i++ {
		_, _ = a.Add(c)
	}
}

func BenchmarkAddUFix64_Ref(b *testing.B) {
	a := 123456789
	c := 987654321
	for i := 0; i < b.N; i++ {
		_ = a + c
	}
}

func BenchmarkSubUFix64(b *testing.B) {
	a := UFix64(987654321)
	c := UFix64(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = a.Sub(c)
	}
}

func BenchmarkSubUFix64_Ref(b *testing.B) {
	a := UFix64(987654321)
	c := UFix64(123456789)
	for i := 0; i < b.N; i++ {
		_ = a - c
	}
}

func BenchmarkMulUFix64(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	for i := 0; i < b.N; i++ {
		_, _ = a.Mul(c)
	}
}

func BenchmarkMulUFix64_Ref(b *testing.B) {
	a := UFix64(123456789)
	c := UFix64(987654321)
	scale := new(big.Int).SetUint64(1e8)

	for i := 0; i < b.N; i++ {
		aB := new(big.Int).SetUint64(uint64(a))
		cB := new(big.Int).SetUint64(uint64(c))
		result := new(big.Int).Mul(aB, cB)
		result.Div(result, scale)
	}
}

func BenchmarkDivUFix64(b *testing.B) {
	a := UFix64(987654321)
	c := UFix64(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = a.Div(c)
	}
}

func BenchmarkDivUFix64_Ref(b *testing.B) {
	a := UFix64(987654321)
	c := UFix64(123456789)
	scale := new(big.Int).SetUint64(1e8)

	for i := 0; i < b.N; i++ {
		aB := new(big.Int).SetUint64(uint64(a))
		aB = aB.Mul(aB, scale)
		cB := new(big.Int).SetUint64(uint64(c))
		_ = new(big.Int).Div(aB, cB)
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

func BenchmarkSqrt64(b *testing.B) {
	a := UFix64(1234567890000)
	for i := 0; i < b.N; i++ {
		_, _ = a.Sqrt()
	}
}

func BenchmarkSqrt128(b *testing.B) {
	a := UFix128{234334, 123456789}
	for i := 0; i < b.N; i++ {
		_, _ = a.SqrtTest()
	}
}

func BenchmarkLnUFix64(b *testing.B) {
	a := UFix64(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = a.Ln()
	}
}

func BenchmarkExpFix64(b *testing.B) {
	a := Fix64(123456789)
	for i := 0; i < b.N; i++ {
		_, _ = a.Exp()
	}
}

func BenchmarkLnUFix128(b *testing.B) {
	a := UFix128{123456789, 987654321}
	for i := 0; i < b.N; i++ {
		_, _ = a.Ln()
	}
}

func BenchmarkSinFix64(b *testing.B) {
	a := Fix64(40075472476)
	for i := 0; i < b.N; i++ {
		_, _ = a.Sin()
	}
}
