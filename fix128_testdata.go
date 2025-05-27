package fixedPoint

import (
	"math"

	"github.com/ericlagergren/decimal"
)

var decCtx128 = decimal.Context128
var Fix128Scale = decimal.WithPrecision(60).SetMantScale(1, -24)
var TwoToThePowerOf64 = decCtx128.Pow(decimal.WithPrecision(60), deci(2), deci(64))

func decuu(hi uint64, lo uint64) *decimal.Big {
	decHi := decimal.WithPrecision(60).SetUint64(hi)
	decLo := decimal.WithPrecision(60).SetUint64(lo)
	decHi = decHi.Mul(decHi, TwoToThePowerOf64)

	val := decimal.WithPrecision(60).Add(decHi, decLo)
	return val.Quo(val, Fix128Scale)
}

func deciu(hi int64, lo uint64) *decimal.Big {
	decHi := decimal.WithPrecision(60).SetMantScale(hi, 0)
	decLo := decimal.WithPrecision(60).SetUint64(lo)
	decHi = decHi.Mul(decHi, TwoToThePowerOf64)

	val := decimal.WithPrecision(60).Add(decHi, decLo)
	return val.Quo(val, Fix128Scale)
}

func prod(a, b *decimal.Big) *decimal.Big {
	return decimal.WithPrecision(60).Mul(a, b)
}

var MaxUFix128 = decuu(math.MaxUint64, math.MaxUint64)
var HalfMaxUFix128 = decuu(math.MaxUint64/2, math.MaxUint64)

var MaxFix128 = deciu(math.MaxInt64, math.MaxUint64)
var HalfMaxFix128 = deciu(math.MaxInt64/2, math.MaxUint64)
var MinFix128 = deciu(math.MinInt64, 0)
var HalfMinFix128 = deciu(math.MinInt64/2, 0)

var (
	AddUFix128Tests = []struct{ A, B *decimal.Big }{
		// Simple cases
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0)},
		{decf(0.0), decf(0.0)},
		{decf(0.0), decf(1.0)},
		{decf(1.0), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0)},

		// Random cases
		{decf(123.456), decf(789.012)},
		{decf(456.789), decf(123.456)},
		{decf(0.000123), decf(0.000456)},
		{decf(0.000789), decf(0.000321)},
		{decf(98765.4321), decf(12345.6789)},
		{decf(31415.9265), decf(27182.8182)},
		{decf(27182.8182), decf(31415.9265)},
		{decf(1.23456789), decf(0.98765432)},
		{decf(0.99999999), decf(0.00000001)},

		// Edge cases (upper limit)
		{sum(MaxUFix128, -1.0), decf(1.0)},
		{sum(MaxUFix128, -0.1), decf(0.1)},
		{sum(MaxUFix128, -0.01), decf(0.01)},
		{sum(MaxUFix128, -0.001), decf(0.001)},
		{sum(MaxUFix128, -0.0001), decf(0.0001)},
		{sum(MaxUFix128, -0.00001), decf(0.00001)},
		{sum(MaxUFix128, -0.000001), decf(0.000001)},
		{sum(MaxUFix128, -0.0000001), decf(0.0000001)},
		{sum(MaxUFix128, -0.00000001), decf(0.00000001)},
		{HalfMaxUFix128, HalfMaxUFix128},
		{sum(HalfMaxUFix128, 1e-24), HalfMaxUFix128},
		{HalfMaxUFix128, sum(HalfMaxUFix128, 1e-24)},
	}

	AddUFix128OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxUFix128, decf(1.0)},
		{MaxUFix128, decf(0.01)},
		{MaxUFix128, decf(0.001)},
		{MaxUFix128, decf(0.00001)},
		{MaxUFix128, decf(0.0000001)},
		{MaxUFix128, decf(1e-24)},
		{MaxUFix128, MaxUFix128},
		{sum(HalfMaxUFix128, 1.0), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.1), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.01), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.001), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.0001), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.00001), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.000001), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.0000001), HalfMaxUFix128},
		{sum(HalfMaxUFix128, 1e-24), sum(HalfMaxUFix128, 1e-24)},
	}

	AddFix128Tests = []struct{ A, B *decimal.Big }{
		// Simple cases
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0)},
		{decf(0.0), decf(0.0)},
		{decf(0.0), decf(1.0)},
		{decf(1.0), decf(2.0)},
		{decf(-1.0), decf(2.0)},
		{decf(1.0), decf(-2.0)},
		{decf(-1.0), decf(-2.0)},
		{decf(1.0), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0)},
		{decf(1.0), decf(1e8 - 1.0)},

		// Random cases
		{decf(123.456), decf(789.012)},
		{decf(-456.789), decf(123.456)},
		{decf(0.000123), decf(0.000456)},
		{decf(-0.000789), decf(0.000321)},
		{decf(98765.4321), decf(-12345.6789)},
		{decf(31415.9265), decf(27182.8182)},
		{decf(-27182.8182), decf(-31415.9265)},
		{decf(1.23456789), decf(-0.98765432)},
		{decf(0.99999999), decf(0.00000001)},
		{decf(-0.99999999), decf(-0.00000001)},

		// Edge cases (upper limit)
		{sum(MaxFix128, -1.0), decf(1.0)},
		{sum(MaxFix128, -0.1), decf(0.1)},
		{sum(MaxFix128, -0.01), decf(0.01)},
		{sum(MaxFix128, -0.001), decf(0.001)},
		{sum(MaxFix128, -0.0001), decf(0.0001)},
		{sum(MaxFix128, -0.00001), decf(0.00001)},
		{sum(MaxFix128, -0.000001), decf(0.000001)},
		{sum(MaxFix128, -0.0000001), decf(0.0000001)},
		{sum(MaxFix128, -0.00000001), decf(0.00000001)},
		{HalfMaxFix128, HalfMaxFix128},
		{sum(HalfMaxFix128, 1e-24), HalfMaxFix128},
		{HalfMaxFix128, sum(HalfMaxFix128, 1e-24)},

		{MaxFix128, decf(-1.0)},
		{MaxFix128, decf(0.0)},
		{MaxFix128, decf(-0.1)},
		{MaxFix128, decf(-0.01)},
		{MaxFix128, decf(-0.001)},
		{MaxFix128, decf(-0.0001)},
		{MaxFix128, decf(-0.00001)},
		{MaxFix128, decf(-0.000001)},
		{MaxFix128, decf(-0.0000001)},
		{MaxFix128, decf(-0.00000001)},

		// Edge cases (lower limit)
		{sum(MinFix128, 1.0), decf(-1.0)},
		{sum(MinFix128, 0.1), decf(-0.1)},
		{sum(MinFix128, 0.01), decf(-0.01)},
		{sum(MinFix128, 0.001), decf(-0.001)},
		{sum(MinFix128, 0.0001), decf(-0.0001)},
		{sum(MinFix128, 0.00001), decf(-0.00001)},
		{sum(MinFix128, 0.000001), decf(-0.000001)},
		{sum(MinFix128, 0.0000001), decf(-0.0000001)},
		{sum(MinFix128, 0.00000001), decf(-0.00000001)},
		{sum(MinFix128, 1e-24), decf(-1e-24)},

		{decf(0.0), MinFix128},
		{decf(-0.1), sum(MinFix128, 0.1)},
		{decf(-0.01), sum(MinFix128, 0.01)},
		{decf(-0.001), sum(MinFix128, 0.001)},
		{decf(-0.0001), sum(MinFix128, 0.0001)},
		{decf(-0.00001), sum(MinFix128, 0.00001)},
		{decf(-0.000001), sum(MinFix128, 0.000001)},
		{decf(-0.0000001), sum(MinFix128, 0.0000001)},
		{decf(-0.00000001), sum(MinFix128, 0.00000001)},
		{decf(-1e-24), sum(MinFix128, 1e-24)},

		{HalfMinFix128, HalfMinFix128},
		{sum(HalfMinFix128, 0.00000001), sum(HalfMinFix128, -0.00000001)},
	}

	AddFix128OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxFix128, decf(1.0)},
		{MaxFix128, decf(0.01)},
		{MaxFix128, decf(0.001)},
		{MaxFix128, decf(0.00001)},
		{MaxFix128, decf(0.0000001)},
		{MaxFix128, decf(0.00000001)},
		{MaxFix128, decf(1e-24)},
		{MaxFix128, MaxFix128},
		{sum(HalfMaxFix128, 1.0), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.1), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.01), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.001), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.0001), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.00001), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.000001), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.0000001), HalfMaxFix128},
		{sum(HalfMaxFix128, 0.00000001), sum(HalfMaxFix128, 0.00000001)},
	}

	AddFix128NegOverflowTests = []struct{ A, B *decimal.Big }{
		{decf(-1.0), MinFix128},
		{decf(-0.1), MinFix128},
		{decf(-0.01), MinFix128},
		{decf(-0.001), MinFix128},
		{decf(-0.0001), MinFix128},
		{decf(-0.00001), MinFix128},
		{decf(-0.000001), MinFix128},
		{decf(-0.0000001), MinFix128},
		{decf(-0.00000001), MinFix128},
		{MinFix128, decf(-1.0)},
		{MinFix128, decf(-0.1)},
		{MinFix128, decf(-0.01)},
		{MinFix128, decf(-0.001)},
		{MinFix128, decf(-0.0001)},
		{MinFix128, decf(-0.00001)},
		{MinFix128, decf(-0.000001)},
		{MinFix128, decf(-0.0000001)},
		{MinFix128, decf(-0.00000001)},
		{MinFix128, MinFix128},
		{sum(HalfMinFix128, -0.00000001), HalfMinFix128},
		{HalfMinFix128, sum(HalfMinFix128, -0.00000001)},
	}

	SubUFix128Tests = []struct{ A, B *decimal.Big }{
		// Simple cases
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0)},
		{decf(0.0), decf(0.0)},
		{decf(1.0), decf(0.99999999)},
		{decf(1e8), decf(1e8)},
		{decf(1e8), decf(1e8 - 1.0)},

		// Random cases
		{decf(456.789), decf(123.456)},
		{decf(0.000456), decf(0.000123)},
		{decf(0.000789), decf(0.000321)},
		{decf(98765.4321), decf(12345.6789)},
		{decf(31415.9265), decf(27182.8182)},
		{decf(1.23456789), decf(0.98765432)},
		{decf(0.99999999), decf(0.00000001)},

		// Edge cases (upper limit)
		{MaxUFix128, decf(1.0)},
		{MaxUFix128, decf(0.1)},
		{MaxUFix128, decf(0.01)},
		{MaxUFix128, decf(0.001)},
		{MaxUFix128, decf(0.0001)},
		{MaxUFix128, decf(0.00001)},
		{MaxUFix128, decf(0.000001)},
		{MaxUFix128, decf(0.0000001)},
		{MaxUFix128, decf(0.00000001)},
		{MaxUFix128, decf(0.0)},
		{MaxUFix128, HalfMaxUFix128},
		{HalfMaxUFix128, HalfMaxUFix128},
		{sum(HalfMaxUFix128, 0.00000001), HalfMaxUFix128},
		{HalfMaxUFix128, sum(HalfMaxUFix128, -0.00000001)},

		// Edge cases (lower limit)
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.1)},
		{decf(1.0), decf(0.01)},
		{decf(1.0), decf(0.001)},
		{decf(1.0), decf(0.0001)},
		{decf(1.0), decf(0.00001)},
		{decf(1.0), decf(0.000001)},
		{decf(1.0), decf(0.0000001)},
		{decf(1.0), decf(0.00000001)},

		{decf(1.00000001), decf(1.0)},
		{decf(1.00000001), decf(0.1)},
		{decf(1.00000001), decf(0.01)},
		{decf(1.00000001), decf(0.001)},
		{decf(1.00000001), decf(0.0001)},
		{decf(1.00000001), decf(0.00001)},
		{decf(1.00000001), decf(0.000001)},
		{decf(1.00000001), decf(0.0000001)},
		{decf(1.00000001), decf(0.00000001)},

		{decf(0.1), decf(0.1)},
		{decf(0.01), decf(0.01)},
		{decf(0.001), decf(0.001)},
		{decf(0.0001), decf(0.0001)},
		{decf(0.00001), decf(0.00001)},
		{decf(0.000001), decf(0.000001)},
		{decf(0.0000001), decf(0.0000001)},
		{decf(0.00000001), decf(0.00000001)},
	}

	SubUFix128NegOverflowTests = []struct{ A, B *decimal.Big }{
		{deci(5), deci(7)},

		{deci(0), decf(1.0)},
		{deci(0), decf(0.1)},
		{deci(0), decf(0.01)},
		{deci(0), decf(0.001)},
		{deci(0), decf(0.0001)},
		{deci(0), decf(0.00001)},
		{deci(0), decf(0.000001)},
		{deci(0), decf(0.0000001)},
		{deci(0), decf(0.00000001)},

		{deci(100), sum(deci(100), 0.1)},
		{deci(100), sum(deci(100), 0.01)},
		{deci(100), sum(deci(100), 0.001)},
		{deci(100), sum(deci(100), 0.0001)},
		{deci(100), sum(deci(100), 0.00001)},
		{deci(100), sum(deci(100), 0.000001)},
		{deci(100), sum(deci(100), 0.0000001)},
		{deci(100), sum(deci(100), 0.00000001)},

		{sum(MaxUFix128, -1.0), MaxUFix128},
		{sum(MaxUFix128, -0.1), MaxUFix128},
		{sum(MaxUFix128, -0.01), MaxUFix128},
		{sum(MaxUFix128, -0.001), MaxUFix128},
		{sum(MaxUFix128, -0.0001), MaxUFix128},
		{sum(MaxUFix128, -0.00001), MaxUFix128},
		{sum(MaxUFix128, -0.000001), MaxUFix128},
		{sum(MaxUFix128, -0.0000001), MaxUFix128},
		{sum(MaxUFix128, -0.00000001), MaxUFix128},
	}

	SubFix128Tests = []struct{ A, B *decimal.Big }{
		// Simple cases
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0)},
		{decf(0.0), decf(0.0)},
		{decf(0.0), decf(1.0)},
		{decf(1.0), decf(2.0)},
		{decf(-1.0), decf(2.0)},
		{decf(1.0), decf(-2.0)},
		{decf(-1.0), decf(-2.0)},
		{decf(1.0), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0)},
		{decf(1.0), decf(1e8 - 1.0)},

		// Random cases
		{decf(123.456), decf(789.012)},
		{decf(-456.789), decf(123.456)},
		{decf(0.000123), decf(0.000456)},
		{decf(-0.000789), decf(0.000321)},
		{decf(98765.4321), decf(-12345.6789)},
		{decf(31415.9265), decf(27182.8182)},
		{decf(-27182.8182), decf(-31415.9265)},
		{decf(1.23456789), decf(-0.98765432)},
		{decf(0.99999999), decf(0.00000001)},
		{decf(-0.99999999), decf(-0.00000001)},

		// Edge cases (upper limit)
		{sum(MaxFix128, -1.0), decf(-1.0)},
		{sum(MaxFix128, -0.1), decf(-0.1)},
		{sum(MaxFix128, -0.01), decf(-0.01)},
		{sum(MaxFix128, -0.001), decf(-0.001)},
		{sum(MaxFix128, -0.0001), decf(-0.0001)},
		{sum(MaxFix128, -0.00001), decf(-0.00001)},
		{sum(MaxFix128, -0.000001), decf(-0.000001)},
		{sum(MaxFix128, -0.0000001), decf(-0.0000001)},
		{sum(MaxFix128, -0.00000001), decf(-0.00000001)},
		{HalfMaxFix128, HalfMinFix128},
		{HalfMaxFix128, sum(HalfMinFix128, 0.00000001)},

		{MaxFix128, decf(1.0)},
		{MaxFix128, decf(0.0)},
		{MaxFix128, decf(0.1)},
		{MaxFix128, decf(0.01)},
		{MaxFix128, decf(0.001)},
		{MaxFix128, decf(0.0001)},
		{MaxFix128, decf(0.00001)},
		{MaxFix128, decf(0.000001)},
		{MaxFix128, decf(0.0000001)},
		{MaxFix128, decf(0.00000001)},

		// Edge cases (lower limit)
		{sum(MinFix128, 1.0), decf(1.0)},
		{sum(MinFix128, 0.1), decf(0.1)},
		{sum(MinFix128, 0.01), decf(0.01)},
		{sum(MinFix128, 0.001), decf(0.001)},
		{sum(MinFix128, 0.0001), decf(0.0001)},
		{sum(MinFix128, 0.00001), decf(0.00001)},
		{sum(MinFix128, 0.000001), decf(0.000001)},
		{sum(MinFix128, 0.0000001), decf(0.0000001)},
		{sum(MinFix128, 0.00000001), decf(0.00000001)},

		{decf(0.0), MaxFix128},
		{decf(-0.1), sum(MaxFix128, -0.1)},
		{decf(-0.01), sum(MaxFix128, -0.01)},
		{decf(-0.001), sum(MaxFix128, -0.001)},
		{decf(-0.0001), sum(MaxFix128, -0.0001)},
		{decf(-0.00001), sum(MaxFix128, -0.00001)},
		{decf(-0.000001), sum(MaxFix128, -0.000001)},
		{decf(-0.0000001), sum(MaxFix128, -0.0000001)},
		{decf(-0.00000001), sum(MaxFix128, -0.00000001)},

		{decf(-1.0), MaxFix128},
		{decf(-0.1), MaxFix128},
		{decf(-0.01), MaxFix128},
		{decf(-0.001), MaxFix128},
		{decf(-0.0001), MaxFix128},
		{decf(-0.00001), MaxFix128},
		{decf(-0.000001), MaxFix128},
		{decf(-0.0000001), MaxFix128},
		{decf(-0.00000001), MaxFix128},

		{decf(-0.00000001), MaxFix128},

		{sum(HalfMinFix128, -0.00000001), HalfMaxFix128},
		{HalfMinFix128, sum(HalfMaxFix128, 0.00000001)},

		{HalfMinFix128, HalfMaxFix128},
		{sum(HalfMaxFix128, 0.00000001), sum(HalfMinFix128, 0.00000001)},
		{sum(HalfMinFix128, 0.00000001), sum(HalfMaxFix128, -0.00000001)},
	}

	// {sum(HalfMaxFix128, 0.00000001), HalfMinFix128}, // NEG OVERFLOW

	SubFix128OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxFix128, decf(-1.0)},
		{MaxFix128, decf(-0.01)},
		{MaxFix128, decf(-0.001)},
		{MaxFix128, decf(-0.00001)},
		{MaxFix128, decf(-0.0000001)},
		{MaxFix128, decf(-0.00000001)},
		{MaxFix128, MinFix128},
		{sum(HalfMaxFix128, 1.0), HalfMinFix128},
		{sum(HalfMaxFix128, 0.1), HalfMinFix128},
		{sum(HalfMaxFix128, 0.01), HalfMinFix128},
		{sum(HalfMaxFix128, 0.001), HalfMinFix128},
		{sum(HalfMaxFix128, 0.0001), HalfMinFix128},
		{sum(HalfMaxFix128, 0.00001), HalfMinFix128},
		{sum(HalfMaxFix128, 0.000001), HalfMinFix128},
		{sum(HalfMaxFix128, 0.0000001), HalfMinFix128},
	}

	SubFix128NegOverflowTests = []struct{ A, B *decimal.Big }{
		{decf(-2.0), MaxFix128},
		{decf(-1.1), MaxFix128},
		{decf(-1.01), MaxFix128},
		{decf(-1.001), MaxFix128},
		{decf(-1.0001), MaxFix128},
		{decf(-1.00001), MaxFix128},
		{decf(-1.000001), MaxFix128},
		{decf(-1.0000001), MaxFix128},
		{decf(-1.00000001), MaxFix128},
		{MinFix128, decf(1.0)},
		{MinFix128, decf(0.1)},
		{MinFix128, decf(0.01)},
		{MinFix128, decf(0.001)},
		{MinFix128, decf(0.0001)},
		{MinFix128, decf(0.00001)},
		{MinFix128, decf(0.000001)},
		{MinFix128, decf(0.0000001)},
		{MinFix128, decf(0.00000001)},
		{MinFix128, MaxFix128},
		{sum(HalfMinFix128, -0.00000001), sum(HalfMaxFix128, 0.00000001)},
	}

	MulUFix128Tests = []struct{ A, B *decimal.Big }{
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0)},
		{decf(0.0), decf(0.0)},
		{decf(1.0), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0)},

		// The prime factors of UINT64_MAX are 3, 5, 17, 257, 641, 65537, 6700417, and 18446744073709551617
		{scale(decf(3 * 5 * 17 * 257 * 641 * 65537)), decu(6700417)},
		{decf(3 * 5 * 17 * 257 * 641), scale(decu(65537 * 6700417))},
		{decf(3 * 5 * 17 * 257), scale(decu(641 * 65537 * 6700417))},
		{decf(3 * 5 * 17), scale(decu(257 * 641 * 65537 * 6700417))},
		{decf(3 * 5), scale(decu(17 * 257 * 641 * 65537 * 6700417))},
		{decf(3), scale(decu(5 * 17 * 257 * 641 * 65537 * 6700417))},

		// SLIGHTLY LESS than the square root of max UFix128
		{decf(18446744.07370954), decf(18446744.07370954)},

		{MaxUFix128, decu(1)},
		{MaxUFix128, decf(0.1)},
		{MaxUFix128, decf(0.01)},
		{MaxUFix128, decf(0.001)},
		{MaxUFix128, decf(0.0001)},
		{MaxUFix128, decf(0.00001)},
		{MaxUFix128, decf(0.000001)},
		{MaxUFix128, decf(0.0000001)},
		{MaxUFix128, decf(0.00000001)},
		{MaxUFix128, decf(1e-24)},
		{MaxUFix128, decu(0)},
		{sum(MaxUFix128, -1.0), decu(1)},
		{HalfMaxUFix128, decu(2)},

		// Things that multiply to the smallest UFix128
		{decf(1e0), decf(1e-24)},
		{decf(1e-1), decf(1e-23)},
		{decf(1e-2), decf(1e-22)},
		{decf(1e-3), decf(1e-21)},
		{decf(1e-4), decf(1e-20)},
		{decf(1e-5), decf(1e-19)},
		{decf(1e-6), decf(1e-18)},
		{decf(1e-7), decf(1e-17)},
		{decf(1e-8), decf(1e-16)},
		{decf(1e-9), decf(1e-15)},
		{decf(1e-10), decf(1e-14)},
		{decf(1e-11), decf(1e-13)},
		{decf(1e-12), decf(1e-12)},
		{decf(1e-13), decf(1e-11)},
		{decf(1e-14), decf(1e-10)},
		{decf(1e-15), decf(1e-9)},
		{decf(1e-16), decf(1e-8)},
		{decf(1e-17), decf(1e-7)},
		{decf(1e-18), decf(1e-6)},
		{decf(1e-19), decf(1e-5)},
		{decf(1e-20), decf(1e-4)},
		{decf(1e-21), decf(1e-3)},
		{decf(1e-22), decf(1e-2)},
		{decf(1e-23), decf(1e-1)},
		{decf(1e-24), decf(1e0)},

		{decf(0.00000005), decf(0.2)},
		{decf(0.00000002), decf(0.5)},
	}

	MulUFix128OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxUFix128, decf(1.1)},
		{MaxUFix128, decf(1.01)},
		{MaxUFix128, decf(1.001)},
		{MaxUFix128, decf(1.00001)},
		{MaxUFix128, decf(1.0000001)},
		{MaxUFix128, sum(decu(1), 1e-24)},
		{MaxUFix128, MaxUFix128},
		{HalfMaxFix128, HalfMaxFix128},

		// Square root of 2^64
		{decf(18446744.07370955), decf(18446744.07370955)},

		// {sum(scale(decf(3*5*17*257*641*65537)), 0.00000001), decu(6700417)},
		// {sum(decf(3*5*17*257*641), 0.00000001), scale(decu(65537 * 6700417))},
		// {sum(decf(3*5*17*257), 0.00000001), scale(decu(641 * 65537 * 6700417))},
		// {sum(decf(3*5*17), 0.00000001), scale(decu(257 * 641 * 65537 * 6700417))},
		// {sum(decf(3*5), 0.00000001), scale(decu(17 * 257 * 641 * 65537 * 6700417))},
		// {sum(decf(3), 0.00000001), scale(decu(5 * 17 * 257 * 641 * 65537 * 6700417))},
	}

	MulUFix128UnderflowTests = []struct{ A, B *decimal.Big }{
		{decf(0.00000001), decf(0.0000001)},
		{decf(0.0000001), decf(0.00000001)},
		{decf(0.000001), decf(0.000001)},

		{decf(0.01), decf(0.0000001)},
		{decf(0.001), decf(0.000001)},
		{decf(0.0001), decf(0.00001)},
		{decf(0.00001), decf(0.0001)},
		{decf(0.000001), decf(0.001)},
		{decf(0.0000001), decf(0.01)},
		{decf(0.0000001), decf(0.01)},

		{decf(0.99999999), decf(0.00000001)},
		{decf(0.09999999), decf(0.0000001)},
		{decf(0.00999999), decf(0.000001)},
		{decf(0.00099999), decf(0.00001)},
		{decf(0.00009999), decf(0.0001)},
		{decf(0.00000999), decf(0.001)},
		{decf(0.00000099), decf(0.01)},
		{decf(0.00000009), decf(0.1)},

		{decf(0.00000005), decf(0.19999999)},
		{decf(0.00000002), decf(0.49999999)},
	}

	DivUFix128Tests = []struct{ A, B *decimal.Big }{
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(1e8)},
		{decf(10.0), decf(1e8 + 1.0)},
		{decf(1e8), decf(1e8)},
		{decf(1e8), decf(1e8 - 1.0)},
		{decf(1e8), decf(1e8 + 1.0)},
		{decu(5), decu(1)},
		{decu(5), decu(2)},
		{decu(5), decu(3)},
		{decu(5), decu(4)},
		{decu(5), decu(5)},
		{decu(5), decu(6)},
		{decu(5), decu(7)},
		{decu(5), decu(8)},
		{decu(5), decu(9)},
		{decu(5), decu(10)},

		// The prime factors of UINT64_MAX are 3, 5, 17, 257, 641, 65537, and 6700417
		{MaxUFix128, scale(decf(3 * 5 * 17 * 257 * 641 * 65537))},
		{MaxUFix128, decf(3 * 5 * 17 * 257 * 641)},
		{MaxUFix128, decf(3 * 5 * 17 * 257)},
		{MaxUFix128, decf(3 * 5 * 17)},
		{MaxUFix128, decf(3 * 5)},
		{MaxUFix128, decf(3)},

		// Near the square root of 2^64
		{MaxUFix128, decf(429496.7296)},
		{MaxUFix128, decf(429496.7295)},
		{MaxUFix128, decf(429496.72959999)},

		{MaxUFix128, decu(1)},
		{MaxUFix128, decu(10)},
		{MaxUFix128, decu(100)},
		{MaxUFix128, decu(1000)},
		{MaxUFix128, decu(10000)},
		{MaxUFix128, decu(100000)},
		{MaxUFix128, decu(1000000)},
		{MaxUFix128, decu(10000000)},
		{MaxUFix128, decu(100000000)},
		{MaxUFix128, decu(1000000000)},
		{MaxUFix128, decu(10000000000)},
		{MaxUFix128, decu(100000000000)},

		{sum(MaxUFix128, -1.0), decf(1.0)},
		{HalfMaxUFix128, decf(0.5)},

		// Things that divide to the smallest UFix128
		{decf(0.00000001), decu(1)},
		{decf(0.0000001), decu(10)},
		{decf(0.000001), decu(100)},
		{decf(0.00001), decu(1000)},
		{decf(0.0001), decu(10000)},
		{decf(0.001), decu(100000)},
		{decf(0.01), decu(1000000)},
		{decf(0.1), decu(10000000)},
		{decf(1.0), decu(100000000)},

		{decf(0.00000001), decf(0.99999999)},
		{decf(0.0000001), decf(9.99999999)},
		{decf(0.000001), decf(99.99999999)},
		{decf(0.00001), decf(999.99999999)},
		{decf(0.0001), decf(9999.99999999)},
		{decf(0.001), decf(99999.99999999)},
		{decf(0.01), decf(999999.99999999)},
		{decf(0.1), decf(9999999.99999999)},
		{decf(1.0), decf(99999999.99999999)},

		{scale(decu(18446744073709551615)), decf(1.0)},
		{scale(decu(1844674407370955161)), decf(0.1)},
		{scale(decu(184467440737095516)), decf(0.01)},
		{scale(decu(18446744073709551)), decf(0.001)},
		{scale(decu(1844674407370955)), decf(0.0001)},
		{scale(decu(184467440737095)), decf(0.00001)},
		{scale(decu(18446744073709)), decf(0.000001)},
		{scale(decu(1844674407370)), decf(0.0000001)},
		{scale(decu(184467440737)), decf(0.00000001)},

		{decf(0.00000005), decu(5)},
		{decf(0.00000002), decu(2)},
	}

	DivUFix128OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxUFix128, decf(0.99999999)},
		{MaxUFix128, decf(0.01)},
		{MaxUFix128, decf(0.001)},
		{MaxUFix128, decf(0.00001)},
		{MaxUFix128, decf(0.0000001)},

		// {scale(decu(18446744073709551615)), decf(0.99999999)},
		// {scale(decu(1844674407370955161)), decf(0.09999999)},
		// {scale(decu(184467440737095516)), decf(0.00999999)},
		// {scale(decu(18446744073709551)), decf(0.00099999)},
		// {scale(decu(1844674407370955)), decf(0.00009999)},
		// {scale(decu(184467440737095)), decf(0.00000999)},
		// {scale(decu(18446744073709)), decf(0.00000099)},
		// {scale(decu(1844674407370)), decf(0.00000009)},
	}

	DivUFix128UnderflowTests = []struct{ A, B *decimal.Big }{
		{decf(0.1), decu(100000000)},
		{decf(0.01), decu(10000000)},
		{decf(0.001), decu(1000000)},
		{decf(0.0001), decu(100000)},
		{decf(0.00001), decu(10000)},
		{decf(0.000001), decu(1000)},
		{decf(0.0000001), decu(100)},
		{decf(0.00000001), decu(10)},

		{decf(0.00000002), decu(3)},
		{decf(0.00000002), decf(2.1)},
		{decf(0.00000002), decf(2.01)},
		{decf(0.00000002), decf(2.001)},
		{decf(0.00000002), decf(2.0001)},
		{decf(0.00000002), decf(2.00001)},
		{decf(0.00000002), decf(2.000001)},
		{decf(0.00000002), decf(2.0000001)},
		{decf(0.00000002), decf(2.00000001)},
	}

	FMDUFix128Tests = []struct{ A, B, C *decimal.Big }{
		{decf(1.0), decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0), decf(1.0)},
		{decf(0.0), decf(1.0), decf(2.0)},
		{decf(1.0), decf(1e8), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0), decf(1e8)},
		{MaxUFix128, decf(1.0), decf(1.0)},
	}

	FMDUFix128OverflowTests = []struct{ A, B, C *decimal.Big }{
		{MaxUFix128, decf(1.1), decf(1.0)},
	}
	SqrtUFix128Tests = []*decimal.Big{
		decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0), decf(6.0), decf(7.0), decf(8.0), decf(9.0),
		decf(10.0), decf(16.0), decf(25.0), decf(49.0), decf(64.0), decf(81.0), decf(100.0), decf(1000.0),
		decf(10000.0), decf(100000.0), decf(1000000.0), decf(10000000.0), decf(100000000.0), decf(1000000000.0),
		MaxUFix128, decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.00000001),
	}
	Ln128Tests = []*decimal.Big{
		decf(2.7182818), decf(1.0), decf(1.1), decf(1.01), decf(1.001), decf(1.0001), decf(1.00001),
		decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001), decf(0.0000001),
		decf(0.00000001), decf(0.9), decf(0.99), decf(0.999), decf(0.9999), decf(0.99999), decf(0.5),
		decf(10.0), decf(20.0), decf(50.0), decf(100.0), decf(500.0), decf(1000.0), decf(5000.0),
		decf(10000.0), decf(3.1415927), decf(7.3890561), decf(15.0), decf(25.0), decf(75.0), decf(250.0),
		decf(750.0), decf(2500.0), decf(7500.0), decf(100000.0), decf(1000000.0), decf(10000000.0),
		decf(100000000.0), decf(1000000000.0), MaxUFix128,
	}
	Exp128Tests = []*decimal.Big{
		decf(0.0), decf(1.0), decf(2.0), decf(5.0), decf(7.9), decf(7.99), decf(8.0), decf(8.01),
		decf(8.1), decf(10.0), decf(15.0), decf(15.9), decf(15.99), decf(16.0), decf(16.01),
		decf(16.1), decf(17.0), decf(20.0), decf(25.0), decf(25.2), decf(-1.0), decf(-2.0),
		decf(-5.0), decf(-6.0), decf(-7.0), decf(-8.0), decf(-9.0), decf(-10.0), decf(-11.0),
		decf(-12.0), decf(-13.0), decf(-14.0), decf(-15.0), decf(-15.1), decf(-15.2), decf(-15.3),
		decf(-15.4), decf(-15.5), decf(-15.6), decf(-15.7), decf(-15.8), decf(-15.9), decf(-16.0), decf(-17.0), decf(-18.0),
	}
	Pow128Tests = []struct{ A, B *decimal.Big }{
		{decf(2.0), decf(3.0)}, {decf(9.0), decf(0.5)}, {decf(27.0), decf(1.0 / 3.0)},
		{decf(5.0), decf(0.0)}, {decf(0.0), decf(5.0)},
	}
	Sin128Tests = []*decimal.Big{
		decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.0000001), decf(0.00000001), decf(0.00391486), decf(0.00391486 + 1e-8), decf(0.00391486 - 1e-8),
		decf(0.2), decf(0.28761102), decf(0.3), decf(0.4), decf(0.5), decf(0.6), decf(0.7), decf(0.8),
		decf(0.9), decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0), decf(6.0), decf(7.0),
		decf(2 - math.Pi/2), decf(2 + 3*math.Pi/2), decf(math.Pi / 2), decf(math.Pi), decf(3 * math.Pi / 2),
		decf(2 * math.Pi), decf(-math.Pi / 2), decf(-math.Pi), decf(-3 * math.Pi / 2), decf(-2 * math.Pi),
	}
	Cos128Tests = []*decimal.Big{
		decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.0000001), decf(0.00000001), decf(0.2), decf(0.28761102), decf(0.3), decf(0.4), decf(0.5),
		decf(0.6), decf(0.7), decf(0.8), decf(0.9), decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0),
		decf(6.0), decf(7.0), decf(math.Pi / 2), decf(math.Pi), decf(3 * math.Pi / 2), decf(2 * math.Pi),
		decf(-math.Pi / 2), decf(-math.Pi), decf(-3 * math.Pi / 2), decf(-2 * math.Pi),
	}
	Tan128Tests = []*decimal.Big{
		decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.0000001), decf(0.00000001), decf(0.2), decf(0.28761102), decf(0.3), decf(0.4), decf(0.5),
		decf(0.6), decf(0.7), decf(0.8), decf(0.9), decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0),
		decf(6.0), decf(7.0), decf(math.Pi / 4), decf(math.Pi / 3), decf(math.Pi), decf(2 * math.Pi),
		decf(-math.Pi / 4), decf(-math.Pi), decf(-2 * math.Pi),
	}
)
