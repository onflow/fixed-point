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
	a := Fix64(0xffffffff00000000)
	for i := 0; i < b.N; i++ {
		_, _ = a.Abs()
	}
}

func BenchmarkSqrt64(b *testing.B) {
	a := UFix64(1234567890000)
	for i := 0; i < b.N; i++ {
		_, _ = a.Sqrt()
	}
}

func BenchmarkSqrt128(b *testing.B) {
	a := UFix128{123456789123456789, 123456789}
	for i := 0; i < b.N; i++ {
		_, _ = a.Sqrt()
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

func BenchmarkExpFix128(b *testing.B) {
	a := Fix128{0x1bad6e, 987654321}
	for i := 0; i < b.N; i++ {
		_, _ = a.Exp()
	}
}

func BenchmarkSinFix64(b *testing.B) {
	a := Fix64(0x1dcd6500) // 5.0
	// a := Fix64(0x2faf080) // 0.5
	for i := 0; i < b.N; i++ {
		_, _ = a.Sin()
	}
}

func BenchmarkSinTestFix64(b *testing.B) {
	// a := Fix64(0x5f5e100)
	a := Fix64(0x2faf080) // 0.5
	for i := 0; i < b.N; i++ {
		_, _ = a.Sin()
	}
}

func BenchmarkCosFix64(b *testing.B) {
	a := Fix64(10000)
	for i := 0; i < b.N; i++ {
		_, _ = a.Cos()
	}
}

func BenchmarkSinFix128(b *testing.B) {
	a := Fix128{123456789, 123456789}
	for i := 0; i < b.N; i++ {
		_, _ = a.Sin()
	}
}

func BenchmarkCosFix128(b *testing.B) {
	a := Fix128{123456789, 123456789}
	for i := 0; i < b.N; i++ {
		_, _ = a.Cos()
	}
}

func BenchmarkChebySin(b *testing.B) {
	a := raw128{123456789, 123456789}
	for i := 0; i < b.N; i++ {
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
		_, _ = mul128(a, a)
		_, _ = add128(a, a, 0)
	}
}
