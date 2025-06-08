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
	product, e := Fix64(a).Mul(Fix64(b))
	return fix64_extra(product >> extraBits), e
}

func (a fix64_extra) Div(b fix64_extra) (fix64_extra, error) {
	quo, e := Fix64(a << extraBits).Div(Fix64(b))
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
	const ln2_extra int64 = ln2_e12 / (1e18 / (int64(Fix64One) << extraBits)) // ln(2) in extra precision

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

// func (x Fix64) Exp() (Fix64, error) {
// 	var err error

// 	// If x is 0, return 1.
// 	if x == 0 {
// 		return fix64One, nil
// 	}

// 	// TODO: These bounds could be tightened...
// 	// Values over e^26 are too large to represent in a Fix64, and values under
// 	// e^-18 are too small.
// 	if x >= 26*fix64One {
// 		return 0, ErrOverflow
// 	} else if x < -18*fix64One {
// 		return 0, ErrUnderflow
// 	}

// 	// If the input is negative, we could compute it direcly using the Taylor
// 	// series, but negative numbers DIVERGE before converging, so we leverage the
// 	// fact that e^x is equal to 1/e^-x and positive exponents converge relatively
// 	// quickly.
// 	inverted := false

// 	if x < 0 {
// 		inverted = true
// 		x = -x
// 	}

// 	// Use the higher precision fix64_12 type for the Taylor series approximation.
// 	x_12 := fixToTwelve(x)

// 	// We scale the input down somewhat so we can fit the *result* inside our
// 	// internal fix64_12 type. For each power of 2 that we shift the input down,
// 	// we will compensate at the end by squaring.
// 	scaleFactor := 0

// 	// NOTE: Because the highest value that can get this far is 26 (see above)
// 	//       this loop will run at most twice. As above, this bound could be
// 	//       tightened.
// 	for x_12 > 8*fix64_12One {
// 		x_12 = x_12 >> 1
// 		scaleFactor += 1
// 	}

// 	term := fix64_12One
// 	sum := fix64_12One

// 	// Use the Taylor series to compute e^x. The series is:
// 	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...
// 	// This loop tends to converge in 20-40 iterations, but we hardcap to 50
// 	// to avoid runaway.
// 	for i := int64(1); i < 50; i++ {
// 		// Multiply in another power of x to the term.
// 		term, err = mulFix64_12(term, x_12)
// 		if err != nil {
// 			return 0, err
// 		}

// 		// Divide by the iteration number to account for the factorial.
// 		// We can use simple division here because we know that we are
// 		// dividing by an integer constant that can't overflow.
// 		term = fix64_12(int64(term) / i)

// 		// Break out of the loop when the term is too small to change the sum.
// 		// TODO: We should probably break out of this loop when term < 1000
// 		// (which is the smallest fix64_12 value representable in Fix64)
// 		if term == 0 {
// 			break
// 		}

// 		// Add the current term to the sum, we can use basic arithmetic
// 		// because we know we are adding converging terms that can't overflow.
// 		sum += term
// 	}

// 	// Convert the result back to the lower precision Fix64 type.
// 	result := twelveToFix(sum)

// 	// Unwind our scaling factor by squaring the result as necessary.
// 	for scaleFactor > 0 {
// 		result, err = result.Mul(result)

// 		if err != nil {
// 			return 0, err
// 		}
// 		scaleFactor -= 1
// 	}

// 	// If we inverted the result on the way in, return 1/result.
// 	if inverted {
// 		result, err = fix64One.Div(result)
// 		if err != nil {
// 			return 0, err
// 		}
// 	}

// 	return result, nil
// }

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
	// if x.Cmp(pi.intDiv(4)) > 0 {
	// 	// recursively call sin(x/2) and sin(x/4), this should only recurse once since we
	// 	// reduced the input to the range [0, π/2] at the beginning of this function.
	// 	sin_half := internalSin(x.intDiv(2), pi, iota)
	// 	sin_quarter := internalSin(x.intDiv(4), pi, iota)
	// 	sin_quarter_squared, _ := sin_quarter.Mul(sin_quarter)
	// 	cosTerm, _ := x.one().Sub(sin_quarter_squared.intMul(2)) // cos(x/2) = 1 - 2•sin²(x/4)
	// 	res, _ := sin_half.intMul(2).Mul(cosTerm)                // sin(x) = 2•sin(x/2)•cos(x/2)
	// 	return res
	// }

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

func clampAngle(x Fix64) (res Fix64, isNeg bool) {
	// The goal of this function is to normalize the input angle x to the range [0, pi], with
	// a seperate flag to indicate if the result should be interpreted as negative. (Seperating
	// out the sign is actually pretty convenient for the calling functions.)

	if x < 0 {
		isNeg = true
		x = -x
	}

	// Now that we have a positive value, we take it modulo 2*pi. Unfortunately, Fix64 doesn't
	// have enough precision to accurately represent pi, and for large values of x,
	// the result of x % (2•pi) will be very inaccurate.
	//
	// Fortunately for us, we have access to a 128/64 bit division function, that provides a
	// remainder. We can scale up the input to a much higher precision and divide it by a constant
	// that is ALSO a multiple of the fixed-point representation of pi.
	//
	// The constant above, called fix64_TwoPiShifted33, is equal to 2•pi•10^8•2^33. We take the input
	// value (which is x•10^8) and scale it up by 2^33. We can then divide it by this constant to get:
	//
	//   x  •  10^8•2^33
	//  -------------
	//  2•pi • 10^8•2^33
	//
	// The remainder of this division will be in the range the remainder of x/2•pi, scaled up by 10^8•2^36.
	// Since we WANT the remainder to be scaled up by 10^8, we can just shift the result right by 33 bits
	// to get a result in the range [0, 2•pi]. We then subtract an additional 2•pi (if necessary) to
	// bring the result into the range [-pi, pi], as is required by the taylor expansion for sin(x).
	tempHi := uint64(x) >> 31
	tempLo := uint64(x) << 33

	_, scaledRem := bits.Div64(tempHi, tempLo, fix64_TwoPiShifted33)

	res = Fix64(scaledRem >> 33) // Shift right by 33 bits to get the remainder in the range [-2•pi, 2•pi]

	// If the angle is greater than pi, it from 2*pi to bring it
	// into the range [0, pi] and flip the sign flag.

	if ((fix64_2Pi & 1) != 0) && (res == Fix64_Pi+1) {
		// Weird edge case: if res is exactly Fix64_Pi + 1, (where 1 there is lowest bit of precision)
		// AND the value of 2•pi is odd when converted to a Fix64, then subtracting res (which is NOT pi)
		// from 2•pi will result in exactly pi! We handle this case specifically and return fix64_Pi - 1 instead.
		res = Fix64_Pi - 1
		isNeg = !isNeg
	} else if res > Fix64_Pi {
		res = fix64_2Pi - res
		isNeg = !isNeg
	}

	return res, isNeg
}

func (x Fix64) Sin() (Fix64, error) {

	x, isNeg := clampAngle(x)

	if x == 0 {
		return 0, nil
	}

	x_extra := fixToExtra(x)

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
	} else {
		x, _ := clampAngle(x)

		// cos(x) = sin(π/2 - x)
		y, _ := fix64_PiOver2.Sub(x)

		return y.Sin()
	}
}

// Commented out because tan() is only of speculative value for smart contracts, and getting a bit-accurate value
// is proving to be VERY complicated. Fundamentally, since tan(x) = sin(x)/cos(x), and cos(x) can be very small
// the error in tan(x) can be very large, even if sin(x) and cos(x) are accurate to the last bit.

// func (x Fix64) Tan() (Fix64, error) {

// 	sin_x, err := x.Sin()
// 	if err != nil {
// 		return 0, err
// 	}
// 	cos_x, err := x.Cos()
// 	if err != nil {
// 		return 0, err
// 	}

// 	return sin_x.Div(cos_x)
// }

// func (x Fix64) TanTest() (Fix64, error) {
// 	// tan(x) = sin(x) / cos(x)
// 	// We could just use the Sin() and Cos() methods directly, but we'll get better
// 	// precision if we call the internalSin() function direclty and divide the result.
// 	x, isNeg := clampAngle(x)

// 	// Let's check fo values that are very close to π/2 and -π/2, these values should
// 	// return an overflow error since tan(x) is undefined for these values.
// 	if x == fix64_PiOver2 {
// 		if !isNeg {
// 			// Remember that fix64_PiOver2 is actually an approximation of π/2, and slightly
// 			// less. So we return a positive overflow error
// 			return 0, ErrOverflow
// 		} else {
// 			// In this case, we're slightly more positively than -π/2, so we return a negative
// 			// overflow error
// 			return 0, ErrNegOverflow
// 		}
// 	}

// 	y, _ := fix64_PiOver2.Sub(x)
// 	if y < 0 {
// 		y = -y
// 		isNeg = !isNeg
// 	}

// 	x_extra := fixToExtra(x)
// 	y_extra := fixToExtra(y)

// 	sinx_extra := internalSin(x_extra, fix64_extra_Pi, fix64_extra_Iota)
// 	cosx_extra := internalSin(y_extra, fix64_extra_Pi, fix64_extra_Iota)

// 	// cos(x) = 1 - 2•sin²(x/2)
// 	// sinx_half := internalSin(x_extra.intDiv(2))
// 	// sinx_squared, _ := sinx_half.Mul(sinx_half)
// 	// cosx_extra, _ := fix64_ExtraOne.Sub(sinx_squared.intMul(2))

// 	res_extra, err := sinx_extra.Div(cosx_extra)

// 	if err != nil {
// 		return 0, err
// 	}

// 	res := extraToFix(res_extra)

// 	if isNeg {
// 		// tan(-x) = -tan(x)
// 		res = res.Neg()
// 	}

// 	return res, nil
// }

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
// 	sinx_extra := internalSin(x_extra, fix64_extra_Pi, fix64_extra_Iota)

// 	// cos(x) = 1 - 2•sin²(x/2)
// 	sinx_half := internalSin(x_extra.intDiv(2), fix64_extra_Pi, fix64_extra_Iota)
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
