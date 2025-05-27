package fixedPoint

import (
	"math/bits"
)

type raw128 struct {
	Hi uint64
	Lo uint64
}

type UFix128 raw128
type Fix128 raw128

var (
	raw128Zero  = raw128{0, 0}
	UFix128Zero = UFix128(raw128Zero)
	Fix128Zero  = Fix128(raw128Zero)
)

// A raw128 value that represents the scale factor for UFix128 and Fix128 (1e24).
var fix128Scale = raw128{Hi: 54210, Lo: 2003764205206896640}

func (a UFix128) isZero() bool {
	return raw128(a).isZero()
}

func (a Fix128) isZero() bool {
	return raw128(a).isZero()
}

func (a Fix128) Neg() Fix128 {
	res, _ := sub128(raw128Zero, raw128(a), 0)

	return Fix128(res)
}

func (a UFix128) Add(b UFix128) (UFix128, error) {
	rawSum, carry := add128(raw128(a), raw128(b), 0)

	if carry != 0 {
		return UFix128Zero, ErrOverflow
	}

	return UFix128(rawSum), nil
}

func (a Fix128) Add(b Fix128) (Fix128, error) {
	// Add as unsigned values
	sum, _ := add128(raw128(a), raw128(b), 0)

	// Check for overflow by checking the sign bits of the operands and the result.
	if int64(a.Hi) >= 0 && int64(b.Hi) >= 0 && int64(sum.Hi) < 0 {
		return Fix128Zero, ErrOverflow
	} else if int64(a.Hi) < 0 && int64(b.Hi) < 0 && int64(sum.Hi) >= 0 {
		return Fix128Zero, ErrNegOverflow
	}

	return Fix128(sum), nil
}

func (a UFix128) Sub(b UFix128) (UFix128, error) {
	rawDiff, borrow := sub128(raw128(a), raw128(b), 0)

	if borrow != 0 {
		return UFix128Zero, ErrUnderflow
	}

	return UFix128(rawDiff), nil
}

func (a Fix128) Sub(b Fix128) (Fix128, error) {
	rawDiff, _ := sub128(raw128(a), raw128(b), 0)

	if int64(a.Hi) >= 0 && int64(b.Hi) < 0 && int64(rawDiff.Hi) < 0 {
		return Fix128Zero, ErrOverflow
	} else if int64(a.Hi) < 0 && int64(b.Hi) >= 0 && int64(rawDiff.Hi) >= 0 {
		return Fix128Zero, ErrNegOverflow
	}

	return Fix128(rawDiff), nil
}

func (a UFix128) Mul(b UFix128) (UFix128, error) {
	hi, lo := mul128(raw128(a), raw128(b))

	// If the high part of the result is larger than or equal to the scale factor,
	// it means the result will be too large to fit in UFix128 after scaling.
	if ucmp128(hi, fix128Scale) >= 0 {
		return UFix128Zero, ErrOverflow
	}

	// If the product is non-zero, but less than the scale factor, it means the result is
	// too small to represent as a UFix128.
	if hi.isZero() && !lo.isZero() && ucmp128(lo, fix128Scale) < 0 {
		return UFix128Zero, ErrUnderflow
	}

	quo := div128(hi, lo, fix128Scale)

	return UFix128(quo), nil
}

func (a UFix128) Div(b UFix128) (UFix128, error) {
	if a.isZero() {
		return UFix128Zero, nil
	}

	if b.isZero() {
		return UFix128Zero, ErrDivByZero
	}

	// We can apply the scale factor BEFORE we divide
	//
	// We're starting with (a * scale) and (b * scale) and we want to end
	// up with (a / b) * scale. The concatended hi-lo values here are equivalent
	// to be equal to (a * scale * scale). When we divide by (b * scale) we'll
	// get our desired result.
	hi, lo := mul128(raw128(a), fix128Scale)

	// If the high part of the dividend is greater than the divisor, the
	// result won't fit into 64 bits.
	if ucmp128(hi, raw128(b)) >= 0 {
		return UFix128Zero, ErrOverflow
	}

	quo := div128(hi, lo, raw128(b))

	// We can't get here if a == 0 because we checked that first. So,
	// a quotient of 0 means the result is too small to represent, i.e. underflow.
	if quo.isZero() {
		return UFix128Zero, ErrUnderflow
	}

	return UFix128(quo), nil
}

func (a raw128) isZero() bool {
	return a.Hi == 0 && a.Lo == 0
}

func ucmp128(a, b raw128) int {
	if a.Hi < b.Hi {
		return -1
	} else if a.Hi > b.Hi {
		return 1
	}

	if a.Lo < b.Lo {
		return -1
	} else if a.Lo > b.Lo {
		return 1
	}

	return 0
}

func scmp128(a, b raw128) int {
	if int64(a.Hi) < int64(b.Hi) {
		return -1
	} else if int64(a.Hi) > int64(b.Hi) {
		return 1
	}

	if a.Lo < b.Lo {
		return -1
	} else if a.Lo > b.Lo {
		return 1
	}

	return 0
}

func add128(a, b raw128, carry uint64) (sum raw128, carryOut uint64) {
	sum.Lo, carry = bits.Add64(a.Lo, b.Lo, carry)
	sum.Hi, carryOut = bits.Add64(a.Hi, b.Hi, carry)
	return
}

func sub128(a, b raw128, borrow uint64) (diff raw128, borrowOut uint64) {
	diff.Lo, borrow = bits.Sub64(a.Lo, b.Lo, borrow)
	diff.Hi, borrowOut = bits.Sub64(a.Hi, b.Hi, borrow)
	return
}

// A utility function used in the 128x128 multiplication algorithm to efficiently
// handle multiplications where one of the operands fits in 64 bits.
func mul128By64(a raw128, b uint64) (hi, lo raw128) {
	if b == 0 || a.isZero() {
		return raw128Zero, raw128Zero
	}

	// Perform multiplication using bits.Mul64. You can think about this as
	// long multiplication where our "base" is 2^64.
	//      aH  aL
	// x         b
	// -----------
	//       w   x
	// + y   z
	// -----------
	//   q   r   s
	// where:
	//   aH = high part of a (most significant 64 bits)
	//   aL = low part of a (least significant 64 bits)
	//   b  = the 64-bit multiplier
	//   w  = high part of b•aL
	//   x  = low part of b•aL
	//   y  = high part of b•aH
	//   z  = low part of b•aH
	//   q  = high part of the result, note that result fits in 192 bits, so this
	//        is is actually hi.Lo (the low part of the high 128 bits that we return)
	//   r  = mid part of the result, (lo.Hi in the return value)
	//   s  = low part of the result, (lo.Lo in the return value)
	//   (Please note that s == x)

	var w, z, carry uint64
	w, lo.Lo = bits.Mul64(a.Lo, b)
	hi.Lo, z = bits.Mul64(a.Hi, b)

	lo.Hi, carry = bits.Add64(w, z, 0)

	// Can't overflow, since that would imply a 128 x 64 multiplication
	// overflowed 192 bits, which is not possible.
	hi.Lo += carry

	return hi, lo
}

// A utility function to perform 128x128 multiplication with a 256-bit result.
func mul128(a, b raw128) (hi, lo raw128) {

	// If either operand fits into 64 bits, we can use a simpler multiplication.
	// This also handles the case where one of the operands is zero.
	if a.Hi == 0 {
		return mul128By64(b, a.Lo)
	} else if b.Hi == 0 {
		return mul128By64(a, b.Lo)
	}

	// Observe that:
	//   a = aH•B + aL and b = bH•B + bL (where B = 2^64)
	//   a * b = (aH * bH) * B^2 + ((aH * bL) + (aL * bH)) * B + (aL * bL)
	//
	// Note that we DO NOT use Karatsuba multiplication here, because we have
	// access to efficient 64-bit multiplication, and the "Karatusba product"
	// operates on sums that could overflow 64 bits and require edge-case handling.

	// u is aH * bH
	// v is (aH * bL) + (aL * bH)
	// w is aL * bL
	var u, v1, v2 raw128
	var wHi uint64
	u.Hi, u.Lo = bits.Mul64(a.Hi, b.Hi)
	v1.Hi, v1.Lo = bits.Mul64(a.Hi, b.Lo)
	v2.Hi, v2.Lo = bits.Mul64(a.Lo, b.Hi)
	v, vCarry := add128(v1, v2, 0)
	wHi, lo.Lo = bits.Mul64(a.Lo, b.Lo)

	// The lowest word of the result (lo.Lo) was directly set when we computed w above

	// We now sum up lo.Hi, which is the low part of v plus the high part of w
	var midCarry, hiCarry uint64
	lo.Hi, midCarry = bits.Add64(v.Lo, wHi, 0)

	// The hi.Lo is the sum of the low part of u with the high part of v plus any carry
	// from the previous sum.
	hi.Lo, hiCarry = bits.Add64(u.Lo, v.Hi, midCarry)

	// hi.Hi is the high part of u plus any carry from the previous sum (and any carry from
	// computing v).
	hi.Hi, _ = bits.Add64(u.Hi, vCarry, hiCarry)

	return
}

func div192by128(hi, mid, lo uint64, y raw128) (quot raw128, rem raw128) {
	// We assume this function is only ever called when y is >= 2^64 (i.e. y.Hi != 0).
	shift := bits.LeadingZeros64(y.Hi)

	// We take the 64 leading, non-zero bits of the denominator and shift it
	// into a uint64. We shift the top bits of the numerator the same amount
	// (filling in with bits from the middle value) and divide them to get
	// an estimate of the quotient.
	estY := (y.Hi << shift) | (y.Lo >> (64 - shift))
	estHi := hi >> (64 - shift)
	estLo := (hi << shift) | (mid >> (64 - shift))

	quot.Hi, _ = bits.Div64(estHi, estLo, estY)

	// We multiply our estimate by the denominator and subtract it from the
	// original numerator to get an intermediate remainder. Note that if our
	// estimate is too high, this will result in a negative remainder, that
	// we'll have to adjust afterwards
	pHi, pLo := mul128By64(y, quot.Hi)

	// TODO: I think that pHi is always zero here, but I'm not 100% sure.

	// Subtract out the product from the top two parts (hi and mid) of the numerator
	// to get an interim result.
	var interimHi, interimMid, borrow uint64
	interimMid, borrow = bits.Sub64(mid, pLo.Lo, 0)
	interimHi, borrow = bits.Sub64(hi, pLo.Hi, borrow)

	if pHi.Lo != 0 || borrow != 0 {
		// If we borrowed or pHi is non-zero, it means that our estimate was too
		// high, so we need to decrement it by 1.
		quot.Hi--

		// Add back in another copy of the denominator to get the right interim remainder.
		var carry uint64
		interimMid, carry = bits.Add64(interimMid, y.Lo, 0)
		interimHi, _ = bits.Add64(interimHi, y.Hi, carry)
	}

	// The interim remainder is a 128-bit value but we know it's less than y. The next step
	// is to tack on the final 64 bits of the numerator (the low part) to the interim remainder
	// and then divide it by the denominator to get the final quotient and remainder.
	// It might look like we're right back where we started; we have a 192-bit numerator
	// (interimHi, interimMid, lo) and a 128-bit denominator (y), but we can use the fact that
	// we know that interim < y to predict that the result of this final division will fit
	// into 64 bits. We can shift the interim remainder down by (64 - shift), which is guaranteed
	// to fit 128 bits, and use the shifted y we used for the first estimate to get our final result
	finalHi := (interimHi << shift) | (interimMid >> (64 - shift))
	finalLo := (interimMid << shift) | (lo >> (64 - shift))

	quot.Lo, _ = bits.Div64(finalHi, finalLo, estY)

	// Now we just need to compute the final remainder
	pHi, pLo = mul128By64(y, quot.Lo)

	rem.Lo, borrow = bits.Sub64(lo, pLo.Lo, 0)
	rem.Hi, _ = bits.Sub64(interimMid, pLo.Hi, borrow)

	return
}

// A helper function to perform unsigned long division of a 256-bit numerator
// by a 128-bit denominator. Used as an analogue of bits.Div64 for 128-bit fixed-point division.
func div128(hi, lo, y raw128) (quot raw128) {
	if y.isZero() {
		panic("div128: division by zero")
	}

	// Special case: denominator fits in 64 bits
	if y.Hi == 0 {
		// If the denominator fits in 64 bits, we know that the EITHER the numerator
		// fits in 192 bits (hi.Hi == 0) OR the division will result in a value that overflows
		// 128 bits. We're mostly trying to "emulate" bits.Div64, which would handle the
		// analogous case by truncating its result, but returning a valid remainder.
		// Since we don't return a remainder, there's no point in computing it, and
		// we can count on our caller to not even call this function in the case where the
		// quotient would overflow, so we just panic here in that case.
		if hi.Hi != 0 {
			panic("div128: overflow")
		}

		qHi, rem := bits.Div64(hi.Lo, lo.Hi, y.Lo)
		qLo, _ := bits.Div64(rem, lo.Lo, y.Lo)
		return raw128{qHi, qLo}
	}

	// We use the "divide and conquer" approach to compute the quotient of a 256-bit numerator
	// by a 128-bit denominator. It involves two calls to a 192 over 128 division algorithm,
	// ("3 by 2" division)
	qHi, rHi := div192by128(hi.Hi, hi.Lo, lo.Hi, y)
	qLo, _ := div192by128(rHi.Hi, rHi.Lo, lo.Lo, y)

	// Effectively multiple qHi by 2^64, assuming that qHi is under 2^64 to start with
	qHi.Hi = qHi.Lo
	qHi.Lo = 0

	quo, _ := add128(qHi, qLo, 0)

	return quo
}
