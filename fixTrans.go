package fixedPoint

import (
	"math/bits"
)

type FixedPoint[T any] interface {
	// Public methods
	Add(other T) (T, error)
	Sub(other T) (T, error)
	Div(other T) (T, error)
	Mul(other T) (T, error)
	Cmp(other T) int

	// Internal methods used by the transcendental functions for efficiency
	// and/or to allow generics.
	intDiv(other int64) T // Integer division
	intMul(other int64) T // Integer multiplication
	isZero() bool
	zero() T
	one() T

	Fix64 | UFix64 | Fix128 | UFix128 | fix64_extra | fix128_extra
}

// Generic Cmp implementation for types that are just integer wrappers.
// This can be used for Fix64, UFix64, and fix64_extra.
func cmpInt[T ~int64 | ~uint64](a, b T) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// Use the generic cmpInt for each concrete type.
func (a Fix64) Cmp(b Fix64) int             { return cmpInt(a, b) }
func (a UFix64) Cmp(b UFix64) int           { return cmpInt(a, b) }
func (a fix64_extra) Cmp(b fix64_extra) int { return cmpInt(a, b) }

func (a Fix64) zero() Fix64             { return 0 }
func (a UFix64) zero() UFix64           { return 0 }
func (a fix64_extra) zero() fix64_extra { return 0 }

func (a Fix64) isZero() bool       { return a == 0 }
func (a UFix64) isZero() bool      { return a == 0 }
func (a fix64_extra) isZero() bool { return a == 0 }

func (a Fix64) one() Fix64             { return Fix64(fix64One) }
func (a UFix64) one() UFix64           { return UFix64(fix64One) }
func (a fix64_extra) one() fix64_extra { return fix64_extra(fix64_ExtraOne) }

// An internal signed fixed-point type that provides a bit more precision
// than the standard Fix64 type. It is used internally for transcendental calculations
// that require higher precision; it gives us an extra few bits of precision, which
// is enough to avoid accumulated rounding errors in our various expansion functions.
const extraBits = 8

type fix64_extra int64
type fix128_extra raw128

const fix64_ExtraOne = fix64_extra(uint64(fix64One) << extraBits)

var fix128_ExtraOne = Fix128(fix128Scale).shiftLeft(extraBits)

func (a fix64_extra) Add(b fix64_extra) (fix64_extra, error) {
	sum, e := Fix64(a).Add(Fix64(b))
	return fix64_extra(sum), e
}

func (a fix64_extra) Sub(b fix64_extra) (fix64_extra, error) {
	sum, e := Fix64(a).Sub(Fix64(b))
	return fix64_extra(sum), e
}

func (a fix64_extra) Mul(b fix64_extra) (fix64_extra, error) {
	product, e := Fix64(a).Mul(Fix64(b))
	return fix64_extra(product >> extraBits), e
}

func (a fix64_extra) Div(b fix64_extra) (fix64_extra, error) {
	quo, e := Fix64(a << extraBits).Div(Fix64(b))
	return fix64_extra(quo), e
}

func (a fix64_extra) intMul(other int64) fix64_extra {
	return fix64_extra(int64(a) * other)
}

func (a fix64_extra) intDiv(other int64) fix64_extra {
	return fix64_extra(int64(a) / other)
}

func fixToExtra(x Fix64) fix64_extra {
	return fix64_extra(x << extraBits)
}

func extraToFix(x fix64_extra) Fix64 {
	// Theoretically, we should just be able to shift right to convert back
	// to Fix64, but shift right always rounds down, while we want to round towards zero.
	// (i.e. Negative numbers can end up being one step more negative than they should be)
	if x > 0 {
		return Fix64(x >> extraBits)
	} else {
		return Fix64(x / (1 << extraBits))
	}
}

func (a fix128_extra) Add(b fix128_extra) (fix128_extra, error) {
	sum, e := Fix128(a).Add(Fix128(b))
	return fix128_extra(sum), e
}

func (a fix128_extra) Sub(b fix128_extra) (fix128_extra, error) {
	sum, e := Fix128(a).Sub(Fix128(b))
	return fix128_extra(sum), e
}

func (a fix128_extra) Mul(b fix128_extra) (fix128_extra, error) {
	product, e := Fix128(a).Mul(Fix128(b))
	return fix128_extra(product.shiftRight(extraBits)), e
}

func (a fix128_extra) Div(b fix128_extra) (fix128_extra, error) {
	quo, e := Fix128(Fix128(a).shiftLeft(extraBits)).Div(Fix128(b))
	return fix128_extra(quo), e
}

func (a fix128_extra) intMul(other int64) fix128_extra {
	return fix128_extra(Fix128(a).intMul(other))
}

func (a fix128_extra) intDiv(other int64) fix128_extra {
	return fix128_extra(Fix128(a).intDiv(other))
}

func (a fix128_extra) isZero() bool {
	return a.Hi == 0 && a.Lo == 0
}

func (a fix128_extra) zero() fix128_extra {
	return fix128_extra{0, 0}
}

func (a fix128_extra) one() fix128_extra {
	return fix128_extra(fix128_ExtraOne)
}

func (a fix128_extra) Cmp(b fix128_extra) int {
	return Fix128(a).Cmp(Fix128(b))
}

func fixToExtra128(x Fix128) fix128_extra {
	return fix128_extra(x.shiftLeft(extraBits))
}

func extraToFix128(x fix128_extra) Fix128 {
	// Theoretically, we should just be able to shift right to convert back
	// to Fix64, but shift right always rounds down, while we want to round towards zero.
	// (i.e. Negative numbers can end up being one step more negative than they should be)
	if x.Hi > 0 {
		return Fix128(x).shiftRight(extraBits)
	} else {
		return Fix128(x).intDiv(1 << extraBits)
	}
}

func (a UFix64) intDiv(other int64) UFix64 {
	return UFix64(uint64(a) / uint64(other))
}

func (a Fix64) intDiv(other int64) Fix64 {
	return Fix64(int64(a) / other)
}

func (a UFix64) intMul(other int64) UFix64 {
	return UFix64(uint64(a) * uint64(other))
}

func (a Fix64) intMul(other int64) Fix64 {
	return Fix64(int64(a) * other)
}

func (a UFix128) isZero() bool {
	return a.Hi == 0 && a.Lo == 0
}

func (a UFix128) zero() UFix128 {
	return UFix128Zero
}

func (a UFix128) one() UFix128 {
	return UFix128(fix128Scale)
}

func (a UFix128) intDiv(other int64) UFix128 {
	qHi, rem := bits.Div64(0, a.Hi, uint64(other))
	qLo, _ := bits.Div64(rem, a.Lo, uint64(other))
	return UFix128{qHi, qLo}
}

func (a UFix128) intMul(other int64) UFix128 {
	m1, pLo := bits.Mul64(a.Lo, uint64(other))
	_, m2 := bits.Mul64(a.Hi, uint64(other))
	pHi, _ := bits.Add64(m1, m2, 0)

	return UFix128{pHi, pLo}
}

func (a Fix128) intDiv(other int64) Fix128 {
	return Fix128(UFix128(a).intDiv(other))
}

func (a Fix128) intMul(other int64) Fix128 {
	return Fix128(UFix128(a).intMul(other))
}

func (a Fix128) shiftRight(shift uint) Fix128 {
	if shift >= 64 {
		// NOTE: We need to copy the sign bit into the high part
		return Fix128{Hi: uint64(int64(a.Hi) >> 63), Lo: uint64(int64(a.Hi) >> (shift - 64))}
	}

	return Fix128{Hi: a.Hi >> shift, Lo: (a.Lo >> shift) | (a.Hi << (64 - shift))}
}

func (a Fix128) shiftLeft(shift uint) Fix128 {
	if shift >= 64 {
		return Fix128{Hi: a.Lo << (shift - 64), Lo: 0}
	}

	return Fix128{Hi: (a.Hi << shift) | (a.Lo >> (64 - shift)), Lo: a.Lo << shift}
}

func (a UFix128) shiftRight(shift uint) UFix128 {
	if shift >= 64 {
		// No sign bit this time...
		return UFix128{Hi: 0, Lo: a.Hi >> (shift - 64)}
	}

	return UFix128{Hi: a.Hi >> shift, Lo: (a.Lo >> shift) | (a.Hi << (64 - shift))}
}

func (a UFix128) shiftLeft(shift uint) UFix128 {
	if shift >= 64 {
		return UFix128{Hi: a.Lo << (shift - 64), Lo: 0}
	}

	return UFix128{Hi: (a.Hi << shift) | (a.Lo >> (64 - shift)), Lo: a.Lo << shift}
}

func internalSqrt[T FixedPoint[T]](x T, est T) (T, error) {
	for {
		quo, err := x.Div(est)

		if err != nil {
			return x.zero(), err
		}

		if est.Cmp(quo) > 0 {
			quo, est = est, quo
		}

		diff, err := quo.Sub(est)

		if err != nil {
			return x.zero(), err
		}

		diff = diff.intDiv(2)

		// We've converged!
		if diff.isZero() {
			break
		}

		est, err = est.Add(diff)

		if err != nil {
			return x.zero(), err
		}
	}

	return est, nil
}

func (x UFix64) Sqrt() (UFix64, error) {
	if x == 0 {
		return 0, nil
	}

	n := bits.Len64(uint64(x))

	// We start our estimate with a number that has a bit length
	// halfway between the original number and the fixed-point representation
	// of 1. This will be of the same order of magnitude as the square root.
	n = (n + 27) / 2
	est := UFix64(1 << n)

	return internalSqrt(x, est)
}

func (x UFix128) Sqrt() (UFix128, error) {
	if x.IsZero() {
		return UFix128Zero, nil
	}

	n := bits.Len64(uint64(x.Hi))
	if n == 0 {
		n = bits.Len64(uint64(x.Lo))
	} else {
		n += 64
	}

	// We start our estimate with a number that has a bit length
	// halfway between the original number and the fixed-point representation
	// of 1. This will be of the same order of magnitude as the square root.
	n = (n + 80) / 2

	var est UFix128
	if n >= 64 {
		est = UFix128{Hi: 1 << (n - 64), Lo: 0}
	} else {
		est = UFix128{Hi: 0, Lo: 1 << n}
	}

	return internalSqrt(x, est)
}

// This is an internal function that assumes that the input is already very close to 1,
// callers to this function should multiply/divide the input by a power of two to ensure
// this is the case, and can add/subtract a multiple of ln(2) to the result
// to account for the scaling.
func internalLn[T FixedPoint[T]](x T) (T, error) {
	// We will compute ln(x) using the approximation:
	// ln(x) = 2 * (z + z^3/3 + z^5/5 + z^7/7 + ...)
	// where z = (x - 1) / (x + 1)
	num, _ := x.Sub(x.one())
	den, _ := x.Add(x.one())
	z, err := num.Div(den)
	if err != nil {
		return x.zero(), err
	}

	// If we just use z directly, we don't have enough precision to get
	// an accurate result
	z2, err := z.Mul(z)
	if err != nil {
		return x.zero(), err
	}
	term := z
	sum := z
	iter := int64(1)

	// Keep interating until "term" and/or "next" rounds to zero (at the precision
	// of a Fix64).
	for {
		term, err = term.Mul(z2)

		if err == ErrUnderflow {
			break
		} else if err != nil {
			return x.zero(), err
		}

		// We can use basic arithmetic here since we are dividing by a
		// an integer constant
		next := term.intDiv(iter*2 + 1)

		if next.isZero() {
			break
		}

		sum, _ = sum.Add(next)
		iter += 1
	}

	return sum.intMul(2), nil
}

func (x UFix64) Ln() (Fix64, error) {
	if x == 0 {
		return 0, ErrDomain
	}

	// The Taylor expansion of ln(x) converges faster for values closer to 1.
	// If we scale the input to have exactly 37 leading zero bits, the input will
	// be a number in the range (0.67108864, 1.33584262). (0.67108864 = 1 << 36 / scale)
	// For every power of two removed (or added) by this shift, we add (or subtract)
	// a multiple of ln(2) at the end of the function.
	leadingZeros := bits.LeadingZeros64(uint64(x))
	k := int64(37 - leadingZeros)

	shift := k - extraBits

	if shift > 0 {
		x = x >> shift
	} else if shift < 0 {
		x = x << -shift
	}

	// We directly cast here instead of using the conversion function, we take into
	// account the extra bits of precision by modifying the shift value used above.
	// This lets us use the bits from x that woudl "fall off" the right side if we
	// just shifted right, and shifted back left again when moving to the extra precision type.
	x_extra := fix64_extra(x)

	res_extra, err := internalLn(x_extra)

	if err != nil {
		return 0, err
	}

	// Multiply by 2 to account for the global 2x at the begining of the Tayler
	// expansion, and then add/subtract as many ln(2)s as required to account
	// for the scaling by 2^k we did at the beginning.
	const ln2_e12 int64 = 693147180559945309                                  // ln(2) * 1e18, fixed-point representation
	const ln2_extra int64 = ln2_e12 / (1e18 / (int64(fix64One) << extraBits)) // ln(2) in extra precision

	powerCorrection := fix64_extra(ln2_extra * k)
	res_extra, _ = res_extra.Add(powerCorrection)

	// Convert the result back to the lower precision Fix64 type.
	return extraToFix(res_extra), nil
}

func (x UFix128) Ln() (Fix128, error) {
	if x.IsZero() {
		return Fix128Zero, ErrDomain
	}

	// The Taylor expansion of ln(x) converges faster for values closer to 1.
	// If we scale the input to have exactly 48 leading zero bits, the input will
	// be a number in the range (0.60446412, 1.20892824).
	// For every power of two removed (or added) by this shift, we add (or subtract)
	// a multiple of ln(2) at the end of the function.
	leadingZeros := bits.LeadingZeros64(x.Hi)
	if leadingZeros == 64 {
		leadingZeros += bits.LeadingZeros64(x.Lo)
	}

	k := int64(48 - leadingZeros)

	shift := k - extraBits

	if shift > 0 {
		x = x.shiftRight(uint(shift))
	} else if shift < 0 {
		x = x.shiftLeft(uint(-shift))
	}

	// We directly cast here instead of using the conversion function, we take into
	// account the extra bits of precision by modifying the shift value used above.
	// This lets us use the bits from x that woudl "fall off" the right side if we
	// just shifted right, and shifted back left again when moving to the extra precision type.
	x_extra := fix128_extra(x)

	res_extra, err := internalLn(x_extra)

	if err != nil {
		return Fix128Zero, err
	}

	// Multiply by 2 to account for the global 2x at the begining of the Tayler
	// expansion, and then add/subtract as many ln(2)s as required to account
	// for the scaling by 2^k we did at the beginning.
	var ln2_extra = fix128_extra{0x92c795, 0x7dcc1d0e60ef101f} // ln(2) in extra precision

	res_extra, _ = res_extra.Add(ln2_extra.intMul(k))

	// Convert the result back to the lower precision Fix64 type.
	return extraToFix128(res_extra), nil
}
