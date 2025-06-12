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
	Neg() T

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

func (a Fix64) one() Fix64             { return Fix64(Fix64One) }
func (a UFix64) one() UFix64           { return UFix64(Fix64One) }
func (a fix64_extra) one() fix64_extra { return fix64_extra(fix64_ExtraOne) }

// An internal signed fixed-point type that provides a bit more precision
// than the standard Fix64 type. It is used internally for transcendental calculations
// that require higher precision; it gives us an extra few bits of precision, which
// is enough to avoid accumulated rounding errors in our various expansion functions.

const fix64_ExtraOne = fix64_extra(uint64(Fix64One) << extraBits)

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
	product, e := Fix64(a).FMD(Fix64(b), Fix64(fix64_ExtraOne))
	return fix64_extra(product), e
}

func (a fix64_extra) Div(b fix64_extra) (fix64_extra, error) {
	quo, e := Fix64(a).FMD(Fix64(fix64_ExtraOne), Fix64(b))
	return fix64_extra(quo), e
}

func (a fix64_extra) Neg() fix64_extra {
	return fix64_extra(-int64(a))
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
	// Adding a half before shifting right to round to closest integer
	x_plusHalf := int64(x) + (1 << (extraBits - 1))

	if x < 0 {
		// Crazy edge case: if x is negative, we want to add SLIGHTLY LESS
		// Than half to get the correct rounding behavior.
		x_plusHalf -= 1
	}

	return Fix64(x_plusHalf >> extraBits)
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

func (a fix128_extra) Neg() fix128_extra {
	return a.intMul(-1)
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

func (a Fix64) Neg() Fix64 {
	return Fix64(-int64(a))
}

func (a UFix64) Neg() UFix64 {
	panic("Neg called on unsigned type UFix64")
}

func (a UFix128) Neg() UFix128 {
	panic("Neg called on unsigned type UFix128")
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

func (a raw128) intMul(other int64) raw128 {
	m1, pLo := bits.Mul64(a.Lo, uint64(other))
	_, m2 := bits.Mul64(a.Hi, uint64(other))
	pHi, _ := bits.Add64(m1, m2, 0)

	return raw128{pHi, pLo}
}

func (a UFix128) intMul(other int64) UFix128 {
	return UFix128(raw128(a).intMul(other))
}

func (a Fix128) intDiv(other int64) Fix128 {
	return Fix128(UFix128(a).intDiv(other))
}

func (a Fix128) intMul(other int64) Fix128 {
	return Fix128(raw128(a).intMul(other))
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

func (x UFix128) SqrtTest() (UFix128, error) {
	if x.isZero() {
		return UFix128Zero, nil
	}

	// Count the number of set bits in x, this is a cheap way of estimating
	// the order of magnitude of the input.
	n := bits.Len64(uint64(x.Hi))
	if n == 0 {
		n = bits.Len64(uint64(x.Lo))
	} else {
		n += 64
	}

	// We start our estimate with a number that has a bit length
	// halfway between the original number and the fixed-point representation
	// of 1. This will be of the same order of magnitude as the square root, allowing
	// our Newton-Raphson loop below to converge quickly.
	n = (n + 80) / 2

	var est raw128
	if n >= 64 {
		est = raw128{Hi: 1 << (n - 64), Lo: 0}
	} else {
		est = raw128{Hi: 0, Lo: 1 << n}
	}

	// The inner loop here will frequently divide the input by the current estimate,
	// so instead of using the Fix128.Div method, we expand the numerator once outside
	// the loop, and then directly call div128 in the loop.
	xHi, xLo := mul128(raw128(x), raw128(Fix128One))

	for {
		// This division can't fail: est is always a positive value somewhere between
		// x and 1, so it est will also be between x and 1.
		quo, rem := div128(xHi, xLo, est)

		if ucmp128(rem.intMul(2), est) > 0 {
			// If the remainder is at least half the divisor, we round up.
			quo = add128To64(quo, 1)
		}

		// This swap originally started as an easy way not use unsigned arithmetic
		// to compare the two values, but empircal testing indicated that by always
		// approaching the true value "from below", the algorithm converges to the
		// correct value _including rounding_. When this loop used signed arithmetic
		// it would sometimes overshoot and end up with a value that was off by one.
		if ucmp128(est, quo) > 0 {
			est, quo = quo, est
		}

		// We take the difference using basic arithmetic, since we know that quo
		// and est are close to each other and far away from zero, so the difference
		// will never overflow or underflow a signed int (although it can be negative).
		diff, _ := sub128(quo, est, 0)

		// Effectively divide by 2...
		diff = raw128{Hi: diff.Hi >> 1, Lo: (diff.Lo >> 1) | (diff.Hi << 63)}

		// See if we've converged
		if diff.isZero() {
			break
		}

		// Again, we can add the diff using simple signed arithmetic, since we know that
		// the diff value is small relative to est.
		est, _ = add128(est, diff, 0)
	}

	return UFix128(est), nil
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
	if err == ErrUnderflow {
		// If z is too small to represent, we just return 0
		return x.zero(), nil
	} else if err != nil {
		return x.zero(), err
	}

	// Precompute z^2 to avoid recomputing it in the loop.
	z2, err := z.Mul(z)

	if err == ErrUnderflow {
		// If z^2 is too small, we just return 2*z, which is the first term
		// in the series.
		return z.intMul(2), nil
	} else if err != nil {
		return x.zero(), err
	}
	term := z
	sum := z
	iter := int64(1)

	// Keep interating until "term" and/or "next" rounds to zero
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

// func (a Fix64) Pow(b Fix64) (Fix64, error) {
// 	if a == 0 {
// 		return 0, nil
// 	}
// 	if b == 0 {
// 		return fix64One, nil
// 	}
// 	if b == fix64One {
// 		return a, nil
// 	}

// 	lnA, err := UFix64(a).Ln()
// 	if err != nil {
// 		return 0, err
// 	}
// 	lnA_times_b, err := lnA.Mul(b)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return lnA_times_b.Exp()
// }

// This is an internal function that assumes that the input is already in the range [0, π],
// it returns sin(x), and requires two extra parameters:
//   - pi: the value of π in the same fixed-point representation as x
//   - iota: the largest value in the type T such that the value of (iota^3/3!) rounds to zero
//     (i.e. underflows). This allows us to skip the approximation root entirely for values
//     close to zero, since sin(x) is approximately x for very small x.
func internalSin[T FixedPoint[T]](x T, pi T, iota T) T {
	// Leverage the identity sin(x) = sin(π - x) to keep the input angle
	// in the range [0, π/2]. This is useful because the Taylor expansion for sin(x)
	// converges faster for smaller values of x
	if x.Cmp(pi.intDiv(2)) > 0 {
		x, _ = pi.Sub(x)
	}

	if x.Cmp(iota) <= 0 {
		// If x is very small, we can just return x since sin(x) is approximately x for small x.
		return x
	}

	// NOTE: This code is commented out because it's actually slower to use this recursion
	// instead of just running through more loops of the Taylor expansion. HOWEVER, if we
	// switched the loop below to use a minimax approximation for sin(x), this code would
	// still be useful.
	// The expansion we use below converges very quickly in the range (0, π/4).
	// We can use the following identity to reduce the input angle to this range:
	//     sin(x) = 2•sin(x/2)•cos(x/2)
	// At the same time, cos(y) = 1-2•sin²(y/2), so we can further expand this to:
	//     sin(x) = 2•sin(x/2)•(1 - 2•sin²(x/4))
	if x.Cmp(pi.intDiv(4)) > 0 {
		// recursively call sin(x/2) and sin(x/4), this should only recurse once since we
		// reduced the input to the range [0, π/2] at the beginning of this function.
		sin_half := internalSin(x.intDiv(2), pi, iota)
		sin_quarter := internalSin(x.intDiv(4), pi, iota)
		sin_quarter_squared, _ := sin_quarter.Mul(sin_quarter)
		cosTerm, _ := x.one().Sub(sin_quarter_squared.intMul(2)) // cos(x/2) = 1 - 2•sin²(x/4)
		res, _ := sin_half.intMul(2).Mul(cosTerm)                // sin(x) = 2•sin(x/2)•cos(x/2)
		return res
	}

	// sin(x) = x - x^3/3! + x^5/5! - x^7/7! + ...
	x_squared, _ := x.Mul(x)

	sum := x
	term := x
	iter := int64(1)

	for {
		var err error

		term = term.Neg() // Alternate the sign of the term

		term, err = term.Mul(x_squared)

		if err == ErrUnderflow {
			break
		}

		term = term.intDiv((iter + 1) * (iter + 2))
		if term.isZero() {
			break
		}

		iter += 2

		// Both sum and term are always close to zero, so we don't need to worry
		// about overflow
		sum, _ = sum.Add(term)
	}

	// fmt.Printf("sin(%v) %d iters\n", x, iter)

	return sum
}

// Normalizes the input angle x to the range [0, π], and returns a flag
// indicating if the result should be interpreted as negative.
func clampAngle(x Fix64) (res fix64_extra, isNeg bool) {
	// The goal of this function is to normalize the input angle x to the range [0, pi], with
	// a separate flag to indicate if the result should be interpreted as negative. (Separating
	// out the sign is actually pretty convenient for the calling functions.)
	var unsignedX uint64

	if x >= 0 {
		unsignedX = uint64(x)
		isNeg = false
	} else {
		unsignedX = uint64(-x)
		isNeg = true
	}

	// If the input is outside the range [0, 2π], we need to normalize it.
	if unsignedX <= uint64(fix64_2Pi) {
		// Input is already within the range [0, 2π], we can just shift it
		// left by the extra bits to convert it to fix64_extra.
		res = fix64_extra(unsignedX << extraBits)
	} else {
		// We know we have a positive value, and need take it modulo 2π. Unfortunately, Fix64 doesn't
		// have enough precision to accurately represent pi, and for large values of x,
		// the result of x % 2π will be very inaccurate.

		// This first step is a cheap way to remove a large number of multiples of 2π from the input.
		// The constant fix64_TwoPiMultiple is a multiple of 2π that is chosen to be VERY
		// accurate in the space of a Fix64. By chosing a multiple of 2π that has a series
		// of zeros in the 9th place and beyond, we can have a value that can be stored with
		// 8 decimal places of precision, but which is accurate to several more decimal places.
		//
		// Even better, this constant is less than the maximum value representable as a fix64_extra,
		// so we can then convert the input to a fix64_extra without worrying about overflow.
		unsignedX = unsignedX % 646448019151968420
		unsignedX = unsignedX % uint64(fix64_TwoPiMultiple)

		x_extra := unsignedX << extraBits

		// Now that we are in the range of of a fix64_extra, we can use ANOTHER modulus (similar
		// to the one above) that minimizes the error at the precision of a fix64_extra. This
		// constant isn't as obviously "pretty" as the one above (since fix64_extra isn't a neat
		// multiple of 10), but the same logic applies.
		x_extra = x_extra % 342709898545259501
		x_extra = x_extra % uint64(fix64_extra_TwoPiMultiple)

		// Fortunately for us, we have access to a 128/64 bit division function that provides a
		// remainder. We can scale up the input to a much higher precision and divide it by a constant
		// that is ALSO a multiple of the fixed-point representation of pi.
		//
		// The constant fix64_TwoPiShifted33 is equal to 2π • 10^8 • 2^33. We take the input
		// value (which is x • 10^8) and scale it up by 2^33 (which is no problem if we extend
		// the input temporarily to 128 bits). We can then divide this scaled up value our constant to get:
		//
		//   x  •  10^8 • 2^33
		//  -------------------
		//  2•pi • 10^8 • 2^33
		//
		// The remainder of this division will be in the range the remainder of x/2π, scaled up by 10^8 • 2^33.
		// Since we WANT the remainder to be scaled up by 10^8, we can just shift the result right by 33 bits
		// to get a result in the range [0, 2π].
		tempHi := x_extra >> (64 - 21)
		tempLo := x_extra << 21

		_, scaledRem := bits.Div64(tempHi, tempLo, uint64(fix64_TwoPiShifted33))

		// Shifting right by 33 bits would get the remainder in the range [-2π, 2π] as a Fix64.
		// However, since we are returning fix64_extra anyway, we can scale it down by
		// extraBits *fewer* places so keep some additional precision.
		res = fix64_extra((scaledRem + (1 << 20)) >> 21)
	}

	// If the angle is greater than π, subract it from 2π to bring it
	// into the range [0, π] and flip the sign flag.

	if ((fix64_extra_2Pi & 1) != 0) && (res == fix64_extra_Pi+1) {
		// Weird edge case: if res is exactly fix64_extra_Pi + 1, (where 1 is lowest bit of precision)
		// AND the value of 2π is odd when converted to a fix64_extra, then subtracting res
		// (which IS NOT π) from 2π will result in exactly π! We check for this case specifically
		// and return fix64_Pi - 1 instead.
		res = fix64_extra_Pi - 1
		isNeg = !isNeg
	} else if res > fix64_extra_Pi {
		res = fix64_extra_2Pi - res
		isNeg = !isNeg
	}

	return res, isNeg
}

func (x Fix64) Sin() (Fix64, error) {

	// Normalize the input angle to the range [0, π], with a flag indicating
	// if the result should be interpreted as negative.
	x_extra, isNeg := clampAngle(x)

	if x == 0 {
		return 0, nil
	}

	res := extraToFix(internalSin(x_extra, fix64_extra_Pi, fix64_extra_sinIota))

	// sin(-x) = -sin(x)
	if isNeg {
		res = res.Neg()
	}

	return res, nil
}

func (x Fix64) Cos() (Fix64, error) {
	if x == 0 {
		return Fix64One, nil
	}

	// Ignore the sign since cos(-x) = cos(x)
	x_extra, _ := clampAngle(x)

	// We use the following identities to compute cos(x):
	//     cos(x) = sin(π/2 - x)
	//     cos(x) = -sin(3π/2 − x)
	// If x is is less than or equal to π/2, we can use the first identity,
	// if x is greater than π/2, we use the second identity.
	// In both cases, we end up with a value in the range [0, π], to pass
	// to internalSin().
	var y_extra fix64_extra
	var is_neg bool

	if x_extra <= fix64_extra_PiOver2 {
		// cos(x) = sin(π/2 - x)
		y_extra = fix64_extra_PiOver2 - x_extra
		is_neg = false
	} else {
		// cos(x) = -sin(3π/2 − x)
		y_extra = fix64_extra_3PiOver2 - x_extra
		is_neg = true
	}

	res := extraToFix(internalSin(y_extra, fix64_extra_Pi, fix64_extra_sinIota))

	if is_neg {
		res = res.Neg()
	}

	return res, nil
}

// Commented out because tan() is only of speculative value for smart contracts, and getting a bit-accurate value
// is proving to be VERY complicated. Fundamentally, since tan(x) = sin(x)/cos(x), and cos(x) can be very small
// the error in tan(x) can be very large, even if sin(x) and cos(x) are accurate to the last bit.

func (x Fix64) Tan() (Fix64, error) {

	sin_x, err := x.Sin()
	if err != nil {
		return 0, err
	}
	cos_x, err := x.Cos()
	if err != nil {
		return 0, err
	}

	return sin_x.Div(cos_x)
}

func (x Fix64) TanTest() (Fix64, error) {
	// tan(x) = sin(x) / cos(x)
	// We could just use the Sin() and Cos() methods directly, but we'll get better
	// precision if we call the internalSin() function direclty and divide the result.

	// Normalize the input angle to the range [0, π]
	x_extra, xNeg := clampAngle(x)

	if x_extra == 0 {
		return 0, nil
	}

	if x_extra == fix64_extra_PiOver2 {
		if !xNeg {
			// Remember that fix64_extra_PiOver2 is actually an approximation of π/2, and slightly
			// less. So we return a positive overflow error
			return 0, ErrOverflow
		} else {
			// In this case, we're slightly more positively than -π/2, so we return a negative
			// overflow error
			return 0, ErrNegOverflow
		}
	}

	// We compute y the same way we did in the cos() function above.
	var y_extra fix64_extra
	var is_neg bool

	if x_extra <= fix64_extra_PiOver2 {
		// cos(x) = sin(π/2 - x)
		y_extra = fix64_extra_PiOver2 - x_extra
		is_neg = false
	} else {
		// cos(x) = -sin(3π/2 − x)
		y_extra = fix64_extra_3PiOver2 - x_extra
		is_neg = true
	}

	sinX := internalSin(x_extra, fix64_extra_Pi, fix64_extra_sinIota)
	cosX := internalSin(y_extra, fix64_extra_Pi, fix64_extra_sinIota)

	res_extra, err := sinX.Div(cosX)

	if err != nil {
		return 0, err
	}

	if is_neg {
		res_extra = res_extra.Neg()
	}

	return extraToFix(res_extra), nil
}

// func (x Fix64) TanTest2() (Fix64, error) {
// 	// tan(x) = sin(x) / cos(x)
// 	// We could just use the Sin() and Cos() methods directly, but we'll get better
// 	// precision if we call the internalSin() function direclty and divide the result.
// 	x, xNeg := clampAngle(x)

// 	// Let's check fo values that are very close to π/2 and -π/2, these values should
// 	// return an overflow error since tan(x) is undefined for these values.
// 	if x == fix64_PiOver2 {
// 		if !xNeg {
// 			// Remember that fix64_PiOver2 is actually an approximation of π/2, and slightly
// 			// less. So we return a positive overflow error
// 			return 0, ErrOverflow
// 		} else {
// 			// In this case, we're slightly more positively than -π/2, so we return a negative
// 			// overflow error
// 			return 0, ErrNegOverflow
// 		}
// 	}

// 	x_extra := fixToExtra(x)
// 	sinx_extra := internalSin(x_extra, fix64_extra_Pi, fix64_extra_sinIota)

// 	// cos(x) = 1 - 2•sin²(x/2)
// 	sinx_half := internalSin(x_extra.intDiv(2), fix64_extra_Pi, fix64_extra_sinIota)
// 	sinx_squared, _ := sinx_half.Mul(sinx_half)
// 	cosx_extra, _ := fix64_ExtraOne.Sub(sinx_squared.intMul(2))

// 	res_extra, err := sinx_extra.Div(cosx_extra)

// 	if err != nil {
// 		return 0, err
// 	}

// 	res := extraToFix(res_extra)

// 	if xNeg {
// 		// tan(-x) = -tan(x)
// 		res = res.Neg()
// 	}

// 	return res, nil
// }
