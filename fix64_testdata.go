package fixedPoint

import (
	"math"

	"github.com/ericlagergren/decimal"
)

// Test data for fix64_test.go

// decf is a helper to create *decimal.Big from float64
func decf(f float64) *decimal.Big {

	// Take "one step" away from zero to avoid issues with rounding down
	// when we convert decimal.Big to fixed point types.
	if f < 0.0 {
		f = math.Nextafter(f, math.Inf(-1))
	} else {
		f = math.Nextafter(f, math.Inf(1))
	}

	return decimal.WithPrecision(60).SetFloat64(f)
}

func decu(i uint64) *decimal.Big {
	return decimal.WithPrecision(60).SetUint64(i)
}

func deci(i int64) *decimal.Big {
	return decimal.WithPrecision(60).SetMantScale(i, 0)
}

func sum(a *decimal.Big, b float64) *decimal.Big {
	return decimal.WithPrecision(60).Add(a, decf(b))
}

func scale(a *decimal.Big) *decimal.Big {
	return decimal.WithPrecision(60).Quo(a, decf(1e8))
}

var MaxUFix64 = decimal.WithPrecision(30).Quo(decu(math.MaxUint64), decu(1e8))
var HalfMaxUFix64 = decimal.WithPrecision(30).Quo(decu(math.MaxUint64/2), decu(1e8))

var MaxFix64 = decimal.WithPrecision(30).Quo(deci(math.MaxInt64), deci(1e8))
var HalfMaxFix64 = decimal.WithPrecision(30).Quo(deci(math.MaxInt64/2), deci(1e8))
var MinFix64 = decimal.WithPrecision(30).Quo(deci(math.MinInt64), deci(1e8))
var HalfMinFix64 = decimal.WithPrecision(30).Quo(deci(math.MinInt64/2), deci(1e8))

var (
	AddUFix64Tests = []struct{ A, B *decimal.Big }{
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
		{sum(MaxUFix64, -1.0), decf(1.0)},
		{sum(MaxUFix64, -0.1), decf(0.1)},
		{sum(MaxUFix64, -0.01), decf(0.01)},
		{sum(MaxUFix64, -0.001), decf(0.001)},
		{sum(MaxUFix64, -0.0001), decf(0.0001)},
		{sum(MaxUFix64, -0.00001), decf(0.00001)},
		{sum(MaxUFix64, -0.000001), decf(0.000001)},
		{sum(MaxUFix64, -0.0000001), decf(0.0000001)},
		{sum(MaxUFix64, -0.00000001), decf(0.00000001)},
		{HalfMaxUFix64, HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.00000001), HalfMaxUFix64},
		{HalfMaxUFix64, sum(HalfMaxUFix64, 0.00000001)},
	}

	AddUFix64OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxUFix64, decf(1.0)},
		{MaxUFix64, decf(0.01)},
		{MaxUFix64, decf(0.001)},
		{MaxUFix64, decf(0.00001)},
		{MaxUFix64, decf(0.0000001)},
		{MaxUFix64, decf(0.00000001)},
		{MaxUFix64, MaxUFix64},
		{sum(HalfMaxUFix64, 1.0), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.1), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.01), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.001), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.0001), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.00001), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.000001), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.0000001), HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.00000001), sum(HalfMaxUFix64, 0.00000001)},
	}

	AddFix64Tests = []struct{ A, B *decimal.Big }{
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
		{sum(MaxFix64, -1.0), decf(1.0)},
		{sum(MaxFix64, -0.1), decf(0.1)},
		{sum(MaxFix64, -0.01), decf(0.01)},
		{sum(MaxFix64, -0.001), decf(0.001)},
		{sum(MaxFix64, -0.0001), decf(0.0001)},
		{sum(MaxFix64, -0.00001), decf(0.00001)},
		{sum(MaxFix64, -0.000001), decf(0.000001)},
		{sum(MaxFix64, -0.0000001), decf(0.0000001)},
		{sum(MaxFix64, -0.00000001), decf(0.00000001)},
		{HalfMaxFix64, HalfMaxFix64},
		{sum(HalfMaxFix64, 0.00000001), HalfMaxFix64},
		{HalfMaxFix64, sum(HalfMaxFix64, 0.00000001)},

		{MaxFix64, decf(-1.0)},
		{MaxFix64, decf(0.0)},
		{MaxFix64, decf(-0.1)},
		{MaxFix64, decf(-0.01)},
		{MaxFix64, decf(-0.001)},
		{MaxFix64, decf(-0.0001)},
		{MaxFix64, decf(-0.00001)},
		{MaxFix64, decf(-0.000001)},
		{MaxFix64, decf(-0.0000001)},
		{MaxFix64, decf(-0.00000001)},

		// Edge cases (lower limit)
		{sum(MinFix64, 1.0), decf(-1.0)},
		{sum(MinFix64, 0.1), decf(-0.1)},
		{sum(MinFix64, 0.01), decf(-0.01)},
		{sum(MinFix64, 0.001), decf(-0.001)},
		{sum(MinFix64, 0.0001), decf(-0.0001)},
		{sum(MinFix64, 0.00001), decf(-0.00001)},
		{sum(MinFix64, 0.000001), decf(-0.000001)},
		{sum(MinFix64, 0.0000001), decf(-0.0000001)},
		{sum(MinFix64, 0.00000001), decf(-0.00000001)},

		{decf(0.0), MinFix64},
		{decf(-0.1), sum(MinFix64, 0.1)},
		{decf(-0.01), sum(MinFix64, 0.01)},
		{decf(-0.001), sum(MinFix64, 0.001)},
		{decf(-0.0001), sum(MinFix64, 0.0001)},
		{decf(-0.00001), sum(MinFix64, 0.00001)},
		{decf(-0.000001), sum(MinFix64, 0.000001)},
		{decf(-0.0000001), sum(MinFix64, 0.0000001)},
		{decf(-0.00000001), sum(MinFix64, 0.00000001)},

		{HalfMinFix64, HalfMinFix64},
		{sum(HalfMinFix64, 0.00000001), sum(HalfMinFix64, -0.00000001)},
	}

	AddFix64OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxFix64, decf(1.0)},
		{MaxFix64, decf(0.01)},
		{MaxFix64, decf(0.001)},
		{MaxFix64, decf(0.00001)},
		{MaxFix64, decf(0.0000001)},
		{MaxFix64, decf(0.00000001)},
		{MaxFix64, MaxFix64},
		{sum(HalfMaxFix64, 1.0), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.1), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.01), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.001), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.0001), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.00001), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.000001), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.0000001), HalfMaxFix64},
		{sum(HalfMaxFix64, 0.00000001), sum(HalfMaxFix64, 0.00000001)},
	}

	AddFix64NegOverflowTests = []struct{ A, B *decimal.Big }{
		{decf(-1.0), MinFix64},
		{decf(-0.1), MinFix64},
		{decf(-0.01), MinFix64},
		{decf(-0.001), MinFix64},
		{decf(-0.0001), MinFix64},
		{decf(-0.00001), MinFix64},
		{decf(-0.000001), MinFix64},
		{decf(-0.0000001), MinFix64},
		{decf(-0.00000001), MinFix64},
		{MinFix64, decf(-1.0)},
		{MinFix64, decf(-0.1)},
		{MinFix64, decf(-0.01)},
		{MinFix64, decf(-0.001)},
		{MinFix64, decf(-0.0001)},
		{MinFix64, decf(-0.00001)},
		{MinFix64, decf(-0.000001)},
		{MinFix64, decf(-0.0000001)},
		{MinFix64, decf(-0.00000001)},
		{MinFix64, MinFix64},
		{sum(HalfMinFix64, -0.00000001), HalfMinFix64},
		{HalfMinFix64, sum(HalfMinFix64, -0.00000001)},
	}

	SubUFix64Tests = []struct{ A, B *decimal.Big }{
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
		{MaxUFix64, decf(1.0)},
		{MaxUFix64, decf(0.1)},
		{MaxUFix64, decf(0.01)},
		{MaxUFix64, decf(0.001)},
		{MaxUFix64, decf(0.0001)},
		{MaxUFix64, decf(0.00001)},
		{MaxUFix64, decf(0.000001)},
		{MaxUFix64, decf(0.0000001)},
		{MaxUFix64, decf(0.00000001)},
		{MaxUFix64, decf(0.0)},
		{MaxUFix64, HalfMaxUFix64},
		{HalfMaxUFix64, HalfMaxUFix64},
		{sum(HalfMaxUFix64, 0.00000001), HalfMaxUFix64},
		{HalfMaxUFix64, sum(HalfMaxUFix64, -0.00000001)},

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

	SubUFix64NegOverflowTests = []struct{ A, B *decimal.Big }{
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

		{sum(MaxUFix64, -1.0), MaxUFix64},
		{sum(MaxUFix64, -0.1), MaxUFix64},
		{sum(MaxUFix64, -0.01), MaxUFix64},
		{sum(MaxUFix64, -0.001), MaxUFix64},
		{sum(MaxUFix64, -0.0001), MaxUFix64},
		{sum(MaxUFix64, -0.00001), MaxUFix64},
		{sum(MaxUFix64, -0.000001), MaxUFix64},
		{sum(MaxUFix64, -0.0000001), MaxUFix64},
		{sum(MaxUFix64, -0.00000001), MaxUFix64},
	}

	SubFix64Tests = []struct{ A, B *decimal.Big }{
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
		{sum(MaxFix64, -1.0), decf(-1.0)},
		{sum(MaxFix64, -0.1), decf(-0.1)},
		{sum(MaxFix64, -0.01), decf(-0.01)},
		{sum(MaxFix64, -0.001), decf(-0.001)},
		{sum(MaxFix64, -0.0001), decf(-0.0001)},
		{sum(MaxFix64, -0.00001), decf(-0.00001)},
		{sum(MaxFix64, -0.000001), decf(-0.000001)},
		{sum(MaxFix64, -0.0000001), decf(-0.0000001)},
		{sum(MaxFix64, -0.00000001), decf(-0.00000001)},
		{HalfMaxFix64, HalfMinFix64},
		{HalfMaxFix64, sum(HalfMinFix64, 0.00000001)},

		{MaxFix64, decf(1.0)},
		{MaxFix64, decf(0.0)},
		{MaxFix64, decf(0.1)},
		{MaxFix64, decf(0.01)},
		{MaxFix64, decf(0.001)},
		{MaxFix64, decf(0.0001)},
		{MaxFix64, decf(0.00001)},
		{MaxFix64, decf(0.000001)},
		{MaxFix64, decf(0.0000001)},
		{MaxFix64, decf(0.00000001)},

		// Edge cases (lower limit)
		{sum(MinFix64, 1.0), decf(1.0)},
		{sum(MinFix64, 0.1), decf(0.1)},
		{sum(MinFix64, 0.01), decf(0.01)},
		{sum(MinFix64, 0.001), decf(0.001)},
		{sum(MinFix64, 0.0001), decf(0.0001)},
		{sum(MinFix64, 0.00001), decf(0.00001)},
		{sum(MinFix64, 0.000001), decf(0.000001)},
		{sum(MinFix64, 0.0000001), decf(0.0000001)},
		{sum(MinFix64, 0.00000001), decf(0.00000001)},

		{decf(0.0), MaxFix64},
		{decf(-0.1), sum(MaxFix64, -0.1)},
		{decf(-0.01), sum(MaxFix64, -0.01)},
		{decf(-0.001), sum(MaxFix64, -0.001)},
		{decf(-0.0001), sum(MaxFix64, -0.0001)},
		{decf(-0.00001), sum(MaxFix64, -0.00001)},
		{decf(-0.000001), sum(MaxFix64, -0.000001)},
		{decf(-0.0000001), sum(MaxFix64, -0.0000001)},
		{decf(-0.00000001), sum(MaxFix64, -0.00000001)},

		{decf(-1.0), MaxFix64},
		{decf(-0.1), MaxFix64},
		{decf(-0.01), MaxFix64},
		{decf(-0.001), MaxFix64},
		{decf(-0.0001), MaxFix64},
		{decf(-0.00001), MaxFix64},
		{decf(-0.000001), MaxFix64},
		{decf(-0.0000001), MaxFix64},
		{decf(-0.00000001), MaxFix64},

		{decf(-0.00000001), MaxFix64},

		{sum(HalfMinFix64, -0.00000001), HalfMaxFix64},
		{HalfMinFix64, sum(HalfMaxFix64, 0.00000001)},

		{HalfMinFix64, HalfMaxFix64},
		{sum(HalfMaxFix64, 0.00000001), sum(HalfMinFix64, 0.00000001)},
		{sum(HalfMinFix64, 0.00000001), sum(HalfMaxFix64, -0.00000001)},
	}

	// {sum(HalfMaxFix64, 0.00000001), HalfMinFix64}, // NEG OVERFLOW

	SubFix64OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxFix64, decf(-1.0)},
		{MaxFix64, decf(-0.01)},
		{MaxFix64, decf(-0.001)},
		{MaxFix64, decf(-0.00001)},
		{MaxFix64, decf(-0.0000001)},
		{MaxFix64, decf(-0.00000001)},
		{MaxFix64, MinFix64},
		{sum(HalfMaxFix64, 1.0), HalfMinFix64},
		{sum(HalfMaxFix64, 0.1), HalfMinFix64},
		{sum(HalfMaxFix64, 0.01), HalfMinFix64},
		{sum(HalfMaxFix64, 0.001), HalfMinFix64},
		{sum(HalfMaxFix64, 0.0001), HalfMinFix64},
		{sum(HalfMaxFix64, 0.00001), HalfMinFix64},
		{sum(HalfMaxFix64, 0.000001), HalfMinFix64},
		{sum(HalfMaxFix64, 0.0000001), HalfMinFix64},
	}

	SubFix64NegOverflowTests = []struct{ A, B *decimal.Big }{
		{decf(-2.0), MaxFix64},
		{decf(-1.1), MaxFix64},
		{decf(-1.01), MaxFix64},
		{decf(-1.001), MaxFix64},
		{decf(-1.0001), MaxFix64},
		{decf(-1.00001), MaxFix64},
		{decf(-1.000001), MaxFix64},
		{decf(-1.0000001), MaxFix64},
		{decf(-1.00000001), MaxFix64},
		{MinFix64, decf(1.0)},
		{MinFix64, decf(0.1)},
		{MinFix64, decf(0.01)},
		{MinFix64, decf(0.001)},
		{MinFix64, decf(0.0001)},
		{MinFix64, decf(0.00001)},
		{MinFix64, decf(0.000001)},
		{MinFix64, decf(0.0000001)},
		{MinFix64, decf(0.00000001)},
		{MinFix64, MaxFix64},
		{sum(HalfMinFix64, -0.00000001), sum(HalfMaxFix64, 0.00000001)},
	}

	MulUFix64Tests = []struct{ A, B *decimal.Big }{
		{decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0)},
		{decf(0.0), decf(0.0)},
		{decf(1.0), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0)},

		// The prime factors of UINT64_MAX are 3, 5, 17, 257, 641, 65537, and 6700417
		{scale(decu(3 * 5 * 17 * 257 * 641 * 65537)), decu(6700417)},
		{decu(3 * 5 * 17 * 257 * 641), scale(decu(65537 * 6700417))},
		{decu(3 * 5 * 17 * 257), scale(decu(641 * 65537 * 6700417))},
		{decu(3 * 5 * 17), scale(decu(257 * 641 * 65537 * 6700417))},
		{decu(3 * 5), scale(decu(17 * 257 * 641 * 65537 * 6700417))},
		{decu(3), scale(decu(5 * 17 * 257 * 641 * 65537 * 6700417))},

		// SLIGHTLY LESS than the square root of 2^64
		{decf(429496.7295), decf(429496.7295)},
		{decf(429496.72959999), decf(429496.72959999)},

		{MaxUFix64, decf(1.0)},
		{MaxUFix64, decf(0.1)},
		{MaxUFix64, decf(0.01)},
		{MaxUFix64, decf(0.001)},
		{MaxUFix64, decf(0.0001)},
		{MaxUFix64, decf(0.00001)},
		{MaxUFix64, decf(0.000001)},
		{MaxUFix64, decf(0.0000001)},
		{MaxUFix64, decf(0.00000001)},
		{MaxUFix64, decf(0.0)},
		{sum(MaxUFix64, -1.0), decf(1.0)},
		{HalfMaxUFix64, decf(2.0)},

		// Things that multiply to the smallest UFix64
		{decf(0.1), decf(0.0000001)},
		{decf(0.01), decf(0.000001)},
		{decf(0.001), decf(0.00001)},
		{decf(0.0001), decf(0.0001)},
		{decf(0.00001), decf(0.001)},
		{decf(0.000001), decf(0.01)},
		{decf(0.0000001), decf(0.1)},

		{decf(0.00000005), decf(0.2)},
		{decf(0.00000002), decf(0.5)},
	}

	MulUFix64OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxUFix64, decf(1.1)},
		{MaxUFix64, decf(1.01)},
		{MaxUFix64, decf(1.001)},
		{MaxUFix64, decf(1.00001)},
		{MaxUFix64, decf(1.0000001)},
		{MaxUFix64, MaxUFix64},
		{HalfMaxFix64, HalfMaxFix64},

		// Square root of 2^64
		{decf(429496.7296), decf(429496.7296)},

		{sum(scale(decf(3*5*17*257*641*65537)), 0.00000001), decu(6700417)},
		{sum(decf(3*5*17*257*641), 0.00000001), scale(decu(65537 * 6700417))},
		{sum(decf(3*5*17*257), 0.00000001), scale(decu(641 * 65537 * 6700417))},
		{sum(decf(3*5*17), 0.00000001), scale(decu(257 * 641 * 65537 * 6700417))},
		{sum(decf(3*5), 0.00000001), scale(decu(17 * 257 * 641 * 65537 * 6700417))},
		{sum(decf(3), 0.00000001), scale(decu(5 * 17 * 257 * 641 * 65537 * 6700417))},
	}

	MulUFix64UnderflowTests = []struct{ A, B *decimal.Big }{
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

	DivUFix64Tests = []struct{ A, B *decimal.Big }{
		{decf(1.0), decf(1.0)},
		{decu(1.0), decu(1e8)},
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
		{MaxUFix64, scale(decf(3 * 5 * 17 * 257 * 641 * 65537))},
		{MaxUFix64, decf(3 * 5 * 17 * 257 * 641)},
		{MaxUFix64, decf(3 * 5 * 17 * 257)},
		{MaxUFix64, decf(3 * 5 * 17)},
		{MaxUFix64, decf(3 * 5)},
		{MaxUFix64, decf(3)},

		// Near the square root of 2^64
		{MaxUFix64, decf(429496.7296)},
		{MaxUFix64, decf(429496.7295)},
		{MaxUFix64, decf(429496.72959999)},

		{MaxUFix64, decu(1)},
		{MaxUFix64, decu(10)},
		{MaxUFix64, decu(100)},
		{MaxUFix64, decu(1000)},
		{MaxUFix64, decu(10000)},
		{MaxUFix64, decu(100000)},
		{MaxUFix64, decu(1000000)},
		{MaxUFix64, decu(10000000)},
		{MaxUFix64, decu(100000000)},
		{MaxUFix64, decu(1000000000)},
		{MaxUFix64, decu(10000000000)},
		{MaxUFix64, decu(100000000000)},

		{sum(MaxUFix64, -1.0), decf(1.0)},
		{HalfMaxUFix64, decf(0.5)},

		// Things that divide to the smallest UFix64
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

	DivUFix64OverflowTests = []struct{ A, B *decimal.Big }{
		{MaxUFix64, decf(0.99999999)},
		{MaxUFix64, decf(0.01)},
		{MaxUFix64, decf(0.001)},
		{MaxUFix64, decf(0.00001)},
		{MaxUFix64, decf(0.0000001)},

		{scale(decu(18446744073709551615)), decf(0.99999999)},
		{scale(decu(1844674407370955161)), decf(0.09999999)},
		{scale(decu(184467440737095516)), decf(0.00999999)},
		{scale(decu(18446744073709551)), decf(0.00099999)},
		{scale(decu(1844674407370955)), decf(0.00009999)},
		{scale(decu(184467440737095)), decf(0.00000999)},
		{scale(decu(18446744073709)), decf(0.00000099)},
		{scale(decu(1844674407370)), decf(0.00000009)},
	}

	DivUFix64UnderflowTests = []struct{ A, B *decimal.Big }{
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

	FMDUFix64Tests = []struct{ A, B, C *decimal.Big }{
		{decf(1.0), decf(1.0), decf(1.0)},
		{decf(1.0), decf(0.0), decf(1.0)},
		{decf(0.0), decf(1.0), decf(2.0)},
		{decf(1.0), decf(1e8), decf(1e8)},
		{decf(1.0), decf(1e8 + 1.0), decf(1e8)},
		{MaxUFix64, decf(1.0), decf(1.0)},
	}

	FMDUFix64OverflowTests = []struct{ A, B, C *decimal.Big }{
		{MaxUFix64, decf(1.1), decf(1.0)},
	}
	SqrtUFix64Tests = []*decimal.Big{
		decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0), decf(6.0), decf(7.0), decf(8.0), decf(9.0),
		decf(10.0), decf(16.0), decf(25.0), decf(49.0), decf(64.0), decf(81.0), decf(100.0), decf(1000.0),
		decf(10000.0), decf(100000.0), decf(1000000.0), decf(10000000.0), decf(100000000.0), decf(1000000000.0),
		MaxUFix64, decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.00000001),
	}
	LnTests = []*decimal.Big{
		decf(2.7182818), decf(1.0), decf(1.1), decf(1.01), decf(1.001), decf(1.0001), decf(1.00001),
		decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001), decf(0.0000001),
		decf(0.00000001), decf(0.9), decf(0.99), decf(0.999), decf(0.9999), decf(0.99999), decf(0.5),
		decf(10.0), decf(20.0), decf(50.0), decf(100.0), decf(500.0), decf(1000.0), decf(5000.0),
		decf(10000.0), decf(3.1415927), decf(7.3890561), decf(15.0), decf(25.0), decf(75.0), decf(250.0),
		decf(750.0), decf(2500.0), decf(7500.0), decf(100000.0), decf(1000000.0), decf(10000000.0),
		decf(100000000.0), decf(1000000000.0), MaxUFix64,
	}
	ExpTests = []*decimal.Big{
		decf(0.0), decf(1.0), decf(2.0), decf(5.0), decf(7.9), decf(7.99), decf(8.0), decf(8.01),
		decf(8.1), decf(10.0), decf(15.0), decf(15.9), decf(15.99), decf(16.0), decf(16.01),
		decf(16.1), decf(17.0), decf(20.0), decf(25.0), decf(25.2), decf(-1.0), decf(-2.0),
		decf(-5.0), decf(-6.0), decf(-7.0), decf(-8.0), decf(-9.0), decf(-10.0), decf(-11.0),
		decf(-12.0), decf(-13.0), decf(-14.0), decf(-15.0), decf(-15.1), decf(-15.2), decf(-15.3),
		decf(-15.4), decf(-15.5), decf(-15.6), decf(-15.7), decf(-15.8), decf(-15.9), decf(-16.0), decf(-17.0), decf(-18.0),
	}
	PowTests = []struct{ A, B *decimal.Big }{
		{decf(2.0), decf(3.0)}, {decf(9.0), decf(0.5)}, {decf(27.0), decf(1.0 / 3.0)},
		{decf(5.0), decf(0.0)}, {decf(0.0), decf(5.0)},
	}
	SinTests = []*decimal.Big{
		decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.0000001), decf(0.00000001), decf(0.00391486), decf(0.00391486 + 1e-8), decf(0.00391486 - 1e-8),
		decf(0.2), decf(0.28761102), decf(0.3), decf(0.4), decf(0.5), decf(0.6), decf(0.7), decf(0.8),
		decf(0.9), decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0), decf(6.0), decf(7.0),
		decf(2 - math.Pi/2), decf(2 + 3*math.Pi/2), decf(math.Pi / 2), decf(math.Pi), decf(3 * math.Pi / 2),
		decf(2 * math.Pi), decf(-math.Pi / 2), decf(-math.Pi), decf(-3 * math.Pi / 2), decf(-2 * math.Pi),
	}
	CosTests = []*decimal.Big{
		decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.0000001), decf(0.00000001), decf(0.2), decf(0.28761102), decf(0.3), decf(0.4), decf(0.5),
		decf(0.6), decf(0.7), decf(0.8), decf(0.9), decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0),
		decf(6.0), decf(7.0), decf(math.Pi / 2), decf(math.Pi), decf(3 * math.Pi / 2), decf(2 * math.Pi),
		decf(-math.Pi / 2), decf(-math.Pi), decf(-3 * math.Pi / 2), decf(-2 * math.Pi),
	}
	TanTests = []*decimal.Big{
		decf(0.0), decf(0.1), decf(0.01), decf(0.001), decf(0.0001), decf(0.00001), decf(0.000001),
		decf(0.0000001), decf(0.00000001), decf(0.2), decf(0.28761102), decf(0.3), decf(0.4), decf(0.5),
		decf(0.6), decf(0.7), decf(0.8), decf(0.9), decf(1.0), decf(2.0), decf(3.0), decf(4.0), decf(5.0),
		decf(6.0), decf(7.0), decf(math.Pi / 4), decf(math.Pi / 3), decf(math.Pi), decf(2 * math.Pi),
		decf(-math.Pi / 4), decf(-math.Pi), decf(-2 * math.Pi),
	}
)
