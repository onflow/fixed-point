package fixedPoint

// A 192-bit fixed-point type used for transcendental calculations. It's uses a scale factor of
// 10**24 * 2**64. This means that the top 128 bites are a valid UFix128 value or Fix128 value, with
// the bottom 64 bits being an extension of the fractional part for additional precision. Using the
// 10**24 factor makes converting between the 128-bit types and this type trivial, and without loss
// of precision. The additional 2**64 factor is very easy to handle because a multiplication or
// division by 2**64 can be handled just by selecting the appropriate raw64 components.
type fix192 struct {
	Hi, Mid, Lo raw64
}

// Returns the absolute value of the fix192 value, and a sign integer (1 or -1) matching the original sign
func (x fix192) abs() (fix192, int64) {
	if isNeg64(x.Hi) {
		return x.neg(), -1
	}

	return x, 1
}

// Returns the arithmetic inverse of the fix192 value, i.e. -x.
func (x fix192) neg() fix192 {
	return fix192{}.sub(x)
}

// Apply the sign to the fix192 value, returning an error if the value has a magnintude that is too
// large to represent as a signed value.
func (x fix192) applySign(sign int64) (fix192, error) {
	// If the input is zero, we can just return it as is.
	if x.isZero() {
		return x, nil
	}

	if sign < 0 {
		x = x.neg()

		// If trying to make it negative didn't make it negative, the input is too
		// large to represent as a negative value.
		if !isNeg64(x.Hi) {
			return fix192Zero, ErrNegOverflow
		}
	} else if isNeg64(x.Hi) {
		// If the input reads as negative but wasn't sign flipped, then the input
		// is too big to represent as a signed value.
		return fix192Zero, ErrOverflow
	}

	return x, nil
}

// Returns true if the fix192 value is zero, false otherwise.
func (a fix192) isZero() bool {
	return isZero64(a.Hi) && isZero64(a.Mid) && isZero64(a.Lo)
}

// Returns true if a < b, false otherwise. Interprets both values as unsigned.
func (a fix192) ult(b fix192) bool {
	// It turns out that branchless subtraction (with borrow) is faster than a branching comparison.
	// if a < b, then a-b will be negative, i.e. have a borrow.
	_, borrow := sub64(a.Lo, b.Lo, 0)
	_, borrow = sub64(a.Mid, b.Mid, borrow)
	_, borrow = sub64(a.Hi, b.Hi, borrow)

	return borrow != 0
}

// Converts a UFix64 value to a fix192 value.
func (x UFix64) toFix192() fix192 {
	return x.ToUFix128().toFix192()
}

// Converts a Fix64 value to a fix192 value.
func (x Fix64) toFix192() fix192 {
	return x.ToFix128().toFix192()
}

// Converts a UFix128 value to a fix192 value.
func (x UFix128) toFix192() fix192 {
	return fix192{x.Hi, x.Lo, raw64(0)}
}

// Converts a Fix128 value to a fix192 value.
func (x Fix128) toFix192() fix192 {
	return fix192{x.Hi, x.Lo, raw64(0)}
}

// Converts a fix192 value to a UFix64 value, returning an error if the value can't be represented
// in 64 bits (Note: This includes underflow errors for non-zero values that would round to zero
// when converted, callers should handle ErrUnderflow if they want to treat these values as zero).
func (x fix192) toUFix64() (UFix64, error) {
	x128, err := x.toUFix128()

	if err != nil {
		return UFix64Zero, err
	}

	return x128.ToUFix64()
}

// Converts a fix192 value to a Fix64 value, returning an error if the value can't be represented
// in 64 bits (Note: This includes underflow errors for non-zero values that would round to zero
// when converted, callers should handle ErrUnderflow if they want to treat these values as zero).
func (x fix192) toFix64() (Fix64, error) {
	x128, err := x.toFix128()

	if err != nil {
		return Fix64Zero, err
	}

	return x128.ToFix64()
}

// Converts a fix192 value to a UFix128 value, returning an error if the value can't be represented
// in 128 bits (Note: This includes underflow errors for non-zero values that would round to zero
// when converted, callers should handle ErrUnderflow if they want to treat these values as zero).
func (x fix192) toUFix128() (UFix128, error) {
	_, carry := add64(x.Lo, raw64(0x8000000000000000), 0)

	if isZero64(x.Hi) && isZero64(x.Mid) && !isZero64(x.Lo) && carry == 0 {
		// If the high and mid parts are zero, and the low part is non-zero but less than 0.5, we
		// flag underflow
		return UFix128Zero, ErrUnderflow
	}

	res := raw128{x.Hi, x.Mid}
	res, carry = add128(res, raw128Zero, carry)

	if carry != 0 {
		// The value rounds out to an overflow, so we return an error.
		return UFix128Zero, ErrOverflow
	}

	return UFix128(res), nil
}

// Converts a fix192 value to a Fix128 value, returning an error if the value can't be represented
// in 128 bits (Note: This includes underflow errors for non-zero values that would round to zero
// when converted, callers should handle ErrUnderflow if they want to treat these values as zero).
func (x fix192) toFix128() (Fix128, error) {
	unsignedX, sign := x.abs()

	unsignedRes, err := unsignedX.toUFix128()

	if err != nil {
		return Fix128Zero, err
	}

	return unsignedRes.ApplySign(sign)
}

// Adds two fix192 values together, does not handle overflow, so only use internally where overflow
// can't happen. Works for both signed and unsigned values.
func (a fix192) add(b fix192) (res fix192) {
	res, _ = add192(a, b, 0)

	return res
}

// Subtracts two fix192 values, does not handle overflow, so only use internally where underflow
// can't happen. Works for both signed and unsigned values.
func (a fix192) sub(b fix192) (res fix192) {
	var borrow uint64

	res.Lo, borrow = sub64(a.Lo, b.Lo, 0)
	res.Mid, borrow = sub64(a.Mid, b.Mid, borrow)
	res.Hi, _ = sub64(a.Hi, b.Hi, borrow)

	return res
}

// Multiplies two fix192 values together, treating both as unsigned values. Does not flag underflow
// and will simply return zero if the product is too small to represent.
func (a fix192) umul(b fix192) (fix192, error) {
	// The basic logic here is the same as the logic in mul128(), so check that code for more
	// details. We start by computing each "row" of the long-form multiplicaiton that
	// you would do by hand. We then add all of these results together at the end.
	//
	// The overall multiplication is a 192x192 multiplication that would produce a 384-bit result,
	// except that we immediately throw out the least significant 64 bits of the result because we
	// need to divide the result by 2**64 at the end anyway. We collect the remaining 320 bits into
	// a 192-bit value for the low three words, and a 128-bit value that holds the high two words.

	// The low parts of the result, one for each "row" (the lower three words)
	var r1lo, r2lo, r3lo fix192

	// The high part of the result (the upper two words). Note that row 1 doesn't contribute to the
	// high part directly, so we only need variables for rows 2 and 3.
	var r2hi, r3hi raw128

	// Compute each row.
	r1lo.Hi, r1lo.Mid, r1lo.Lo, _ = mul192by64(a, b.Lo)
	r2hi.Lo, r2lo.Hi, r2lo.Mid, r2lo.Lo = mul192by64(a, b.Mid)
	r3hi.Hi, r3hi.Lo, r3lo.Hi, r3lo.Mid = mul192by64(a, b.Hi)

	// The final sum is spread over these two variables.
	var rawProductLo fix192
	var rawProductHi raw128

	// Carry flags for addition
	var carry1, carry2 uint64

	rawProductLo, carry1 = add192(r1lo, r2lo, 0)
	rawProductLo, carry2 = add192(rawProductLo, r3lo, 0)
	rawProductHi, _ = add128(r2hi, r3hi, carry1)
	rawProductHi, _ = add128(rawProductHi, raw128Zero, carry2)

	// By dropping the lowest word of the first row, we've already efficiently divided the result by
	// 2**64. We now need to scale this result down by 10**24, which will either fit into 192 bits,
	// or we mark the multiplication as an overflow.
	//
	// We do this in two steps:
	// 1. We shift everything down by 24 bits. This is equivalent to dividing the result by 2**24,
	//    which is leaves us still needing us to divide by 5**24. However, this is enough to
	//    shrink the divisor to fit in 64-bits, dramatically simplifying the division. We can also
	//    check for overflow after the shift since the result will only fit in 192 if the part that
	//    extends beyond the bottom 192 bits is less than 5**24 before the division.
	// 2. After shifting down and checking for overflow, we divide the result by 5**24 and return.

	// This check is here to detect any changes the Fix128Scale constant. It should compile to
	// a no-op if the constant is matches our expectation, and will panic if it doesn't.
	if Fix128Scale != 1e24 {
		panic("fix192 assumes Fix128Scale equals 10e24")
	}

	rawProductLo = rawProductLo.ushiftRight(24)
	rawProductLo.Hi |= shiftLeft64(rawProductHi.Lo, 40)
	rawProductHi = ushiftRight128(rawProductHi, 24)

	// The final result will overflow unless the top word (rawProductHi.Hi) is zero, and the
	// second highest word (rawProductHi.Lo) is less than fiveToThe24th.
	if !isZero64(rawProductHi.Hi) || !ult64(rawProductHi.Lo, fiveToThe24) {
		return fix192{}, ErrOverflow
	}

	var quo fix192
	var rem raw64

	quo.Hi, rem = div64(rawProductHi.Lo, rawProductLo.Hi, fiveToThe24)
	quo.Mid, rem = div64(rem, rawProductLo.Mid, fiveToThe24)
	quo.Lo, rem = div64(rem, rawProductLo.Lo, fiveToThe24)

	if ushouldRound64(rem, fiveToThe24) {
		quo, _ = add192(quo, fix192Zero, 1)
	}

	return quo, nil
}

// Performs multiplication of two fix192 values, treating both as signed values.
func (a fix192) smul(b fix192) (fix192, error) {
	aUnsigned, aSign := a.abs()
	bUnsigned, bSign := b.abs()
	rSign := aSign * bSign

	resUnsigned, err := aUnsigned.umul(bUnsigned)

	if err != nil {
		return fix192Zero, applySign(err, rSign)
	}

	// Apply the sign to the result
	return resUnsigned.applySign(rSign)
}

// Perform integer multiplication of a fix192 value by a uint64 value, treating a as an unsigned
// value. Does NOT handle overflow, so only use internally where overflow can't happen.
func (a fix192) uintMul(b uint64) fix192 {
	hi, mid, lo := mul128By64(raw128{a.Mid, a.Lo}, raw64(b))
	_, t4 := mul64(a.Hi, raw64(b))

	sum, _ := add64(hi, t4, 0)

	return fix192{sum, mid, lo}
}

// Perform integer multiplication of a fix192 value by a int64 value, treating a as a signed
// value. Does NOT handle overflow, so only use internally where overflow can't happen.
func (a fix192) intMul(b int64) fix192 {
	aUnsigned, sign := a.abs()

	if b < 0 {
		sign *= -1
		b = -b
	}
	res := aUnsigned.uintMul(uint64(b))
	res, _ = res.applySign(sign)

	return res
}

// Performs a left shift on a fix192 value, shifting the bits to the left by the specified amount.
func (x fix192) shiftLeft(shift uint64) (res fix192) {
	if shift == 0 {
		return x
	} else if shift >= 128 {
		shift -= 128

		res.Hi = shiftLeft64(x.Lo, shift)
		res.Mid = 0
		res.Lo = 0

		return res
	} else if shift >= 64 {
		shift -= 64

		res.Hi = shiftLeft64(x.Mid, shift)
		res.Hi |= ushiftRight64(x.Lo, 64-shift)
		res.Mid = shiftLeft64(x.Lo, shift)
		res.Lo = 0

		return res
	} else {
		res.Hi = shiftLeft64(x.Hi, shift)
		res.Hi |= ushiftRight64(x.Mid, 64-shift)
		res.Mid = shiftLeft64(x.Mid, shift)
		res.Mid |= ushiftRight64(x.Lo, 64-shift)
		res.Lo = shiftLeft64(x.Lo, shift)

		return res
	}
}

// Performs an unsigned right shift on a fix192 value, shifting the bits to the right by the
// specified amount.
func (x fix192) ushiftRight(shift uint64) (res fix192) {
	if shift == 0 {
		return x
	} else if shift >= 128 {
		shift -= 128

		res.Lo = ushiftRight64(x.Hi, shift)
		res.Mid = 0
		res.Hi = 0

		return res
	} else if shift >= 64 {
		shift -= 64

		res.Lo = ushiftRight64(x.Mid, shift)
		res.Lo |= shiftLeft64(x.Hi, 64-shift)
		res.Mid = ushiftRight64(x.Hi, shift)
		res.Hi = 0

		return res
	} else {
		res.Lo = ushiftRight64(x.Lo, shift)
		res.Lo |= shiftLeft64(x.Mid, 64-shift)
		res.Mid = ushiftRight64(x.Mid, shift)
		res.Mid |= shiftLeft64(x.Hi, 64-shift)
		res.Hi = ushiftRight64(x.Hi, shift)

		return res
	}
}

// Computes the natural logarithm of an unsigned fix192 value, returning an error if the input is zero.
// Note that the input is treated as an UNSIGNED value, but the output should be interpreted as a
// SIGNED value.
func (x fix192) ln() (fix192, error) {
	if x.isZero() {
		return fix192Zero, ErrDomain
	}

	var scaledX fix192

	// The Chebyshev polynomials for ln(x) are defined in the range where the input value has the
	// same number of zeros as the representation of "one". Since the upper two words of a fix192
	// simply "are" a valid Fix128 value, we can use the same constant for "the number of leading
	// zero bits of the representation of 1".
	k := Fix128OneLeadingZeros - int64(leadingZeroBits192(x))

	// Scale the input value by 2^k, so that it falls into the range where the Chebyshev polynomials
	// are defined.
	if k >= 0 {
		scaledX = x.ushiftRight(uint64(k))
	} else if k < 0 {
		scaledX = x.shiftLeft(uint64(-k))
	}

	// Binary search to find the largest index where lnBounds[index] <= scaledX
	left := 0
	right := len(lnBounds) - 1

	for left < right {
		mid := left + (right-left+1)/2 // Use upper mid to avoid infinite loop
		if scaledX.ult(lnBounds[mid]) {
			right = mid - 1
		} else {
			left = mid
		}
	}

	res := scaledX.chebyPoly(lnChebyCoeffs[left])

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did above.
	powerCorrection := fix192Ln2.intMul(k)
	res = res.add(powerCorrection)

	return res, nil
}

// Computes the exponential of a fix192 value (e^x), returning an error if the input is too large or
// too small to be represented as a fix192 value. The input is treated as a SIGNED value, but the
// output should be interpreted as an UNSIGNED value.
func (x fix192) exp() (fix192, error) {
	xUnsigned, sign := x.abs()

	// We compute exp(x) by using the identity:
	//     exp(x) = exp(i + f) = exp(i) * exp(f)
	// where i is the integer part and f is the fractional part of x.
	//
	// The easist way to do this would be to divide x by the value of one in fix192, using the
	// quotient as the integer part and the remainder as the fractional part. However, we don't have
	// a 192x192 division, and the value of one in fix192 extends over all 3 words. However, we can
	// use the fact that the value of one in fix192 is 10**24 * 2**64, which is equivalent to 5^24 *
	// 2^24 * 2^64. If we divide both the numerator and denominator by the same value, the quotient
	// will be the same, and the remainder will be scaled down by that value.
	//
	// So, we scale x by 2^24 * 2^64, which is equivalent to dropping the last word, and shifting
	// the result by 24. We then divide the result by 5^24, which is a 64-bit value.

	// The value of x, after dropping the lower 64 bits
	xTop := raw128{xUnsigned.Hi, xUnsigned.Mid}

	// Shift by 24 bits, which is equivalent to dividing by 2^24
	xTop = ushiftRight128(xTop, 24)

	// Divide out the 5^24 factor.
	i, rem := div64(xTop.Hi, xTop.Lo, fiveToThe24)

	// Our remainder is now the fractional part, but, it's been scaled down by 2^24•2^64, AND we are
	// missing the bits that got shifted out. However, because those bits are zero in the
	// denominator (which is 1.0 remember!), we can just copy those bits from the original input
	// into the fractional part.
	f := fix192{rem >> 40, rem<<24 | xUnsigned.Mid&0xffffff, xUnsigned.Lo}
	fIsNonZero := !f.isZero()

	// We now have the integer part, i, and the fractional part, f of abs(x). If sign is negative, we need to
	// flip the sign of the result. Additionally, if the fractional part is non-zero, we subtract it from one
	// so that f is added to i, not subtracted from it.
	if sign < 0 {
		i = -i
		if fIsNonZero {
			i = i - 1
			f = fix192One.sub(f)
		}
	}

	// Determine the index of the exponential power in our table of e^x values.
	intPowerIndex := int64(i) - smallestExpIntPower

	// If the integer points to a value outside the range of the lookup table, we know that value
	// either overflows or underflows (since the table covers all possible values within the range
	// of Fix128)
	if intPowerIndex < 0 {
		return fix192Zero, ErrUnderflow
	} else if intPowerIndex >= int64(len(expIntPowers)) {
		return fix192Zero, ErrOverflow
	}

	res := expIntPowers[intPowerIndex]
	var err error = nil

	if fIsNonZero {
		// Calculate e^f using the Chebyshev polynomial, which is defined in the range [0, 1].
		fracExp := f.chebyPoly(expChebyCoeffs)

		// Multiply the fractional part by the integer part to get the final result
		res, err = res.umul(fracExp)
	}

	return res, err
}

// Computes the power of a fix192 value raised to another fix192 value, returning an error if the
// input can't be represented as a fix192 value. The base (a) is treated as an UNSIGNED value, and
// the exponent (b) is treated as a SIGNED value. The result must also treated as an UNSIGNED value.
func (a fix192) pow(b fix192) (fix192, error) {
	aLn, err := a.ln()

	if err != nil {
		return fix192{}, err
	}

	prod, err := aLn.smul(b)

	if err == ErrUnderflow {
		// If the product is too small, we treat it as zero, and return 1
		return fix192One, nil
	} else if err == ErrNegOverflow {
		// If the product overflows negative, the result is too small to represent, so we return an error.
		return fix192Zero, ErrUnderflow
	} else if err != nil {
		// Overflow errors are just overflow errors...
		return fix192Zero, err
	}

	return prod.exp()
}

// Computes the sine of a fix192 value, returns an error for symmetry with other functions, but
// can't actually fail...
func (x fix192) sin() (fix192, error) {
	// Normalize the input angle to the range [0, π], with a flag indicating
	// if the result should be interpreted as negative.
	clampedX, sign := x.clampAngle()

	// Leverage the identity sin(x) = sin(π - x) to keep the input angle in the range [0, π/2]
	if fix192HalfPi.ult(clampedX) {
		clampedX = fix192Pi.sub(clampedX)
	}

	res := clampedX.chebyPoly(sinChebyCoeffs)

	return res.applySign(sign)
}

// Computes the cosine of a fix192 value, returns an error for symmetry with other functions, but
// can't actually fail...
func (x fix192) cos() (fix192, error) {
	// Normalize the input angle to the range [0, π], with a flag indicating
	// if the result should be interpreted as negative.
	clampedX, _ := x.clampAngle()
	sign := int64(1)

	// We use the following identities to compute cos(x):
	//     cos(x) = sin(π/2 - x)
	//     cos(x) = -sin(x - π/2)
	// If x is is less than or equal to π/2, we can use the first identity, if x is greater than
	// π/2, we use the second identity. In both cases, we end up with a value in the range [0, π/2],
	// to use with the Chebyshev polynomial.
	var y fix192

	if clampedX.ult(fix192HalfPi) {
		// cos(x) = sin(π/2 - x)
		y = fix192HalfPi.sub(clampedX)
	} else {
		// cos(x) = -sin(x - π/2)
		y = clampedX.sub(fix192HalfPi)
		sign *= -1
	}

	res := y.chebyPoly(sinChebyCoeffs)

	return res.applySign(sign)
}

// Counts the number of leading zero bits in a fix192 value, returning the count as an unsigned integer.
func leadingZeroBits192(a fix192) uint64 {
	// Count the number of leading zero bits in a fix192 value.
	if a.Hi == 0 {
		if a.Mid == 0 {
			return leadingZeroBits64(a.Lo) + 128
		} else {
			return leadingZeroBits64(a.Mid) + 64
		}
	} else {
		return leadingZeroBits64(a.Hi)
	}
}

// A special multiplication function that used for Chebyshev polynomials. It makes a number of
// assumptions that only hold for our specific Chebyshev coefficients:
//  1. The input x is always positive (accum can be negative).
//  2. The coefficients are prescaled such that this multiplication can scale down by 2**145 (which
//     is a simple shift) instead of 2**64 * 10**24 (which involves a division).
func (accum fix192) chebyMul(x fix192) fix192 {
	a, aSign := accum.abs()

	var res fix192
	var carry1, carry2 uint64
	var r1, r2, r3 fix192
	var xhi raw64

	// Similar multiplication as umul(), but throwing out the bottom TWO words (instead of just one)
	r1.Mid, r1.Lo, _, _ = mul192by64(a, x.Lo)
	r2.Hi, r2.Mid, r2.Lo, _ = mul192by64(a, x.Mid)
	xhi, r3.Hi, r3.Mid, r3.Lo = mul192by64(a, x.Hi)

	res, carry1 = add192(r1, r2, 0)
	res, carry2 = add192(res, r3, 0)
	xhi, _ = add64(xhi, raw64(carry1), carry2)

	if uint64(xhi) >= (1 << 48) {
		// TODO: Remove this check in production code. We check that this multiplication can't
		// overflow when we compute the Chebyshev coefficients.
		panic("xhi overflowed in chebyMul")
	}

	// Because we threw out the bottom two words above (which effectively divided the result by
	// 2**128), to scaled down by the expected 2**145, we need to shift the result down by 17 more
	// bits.
	res = res.ushiftRight(17)
	res.Hi |= xhi << 47

	res, err := res.applySign(aSign)

	if err != nil {
		// TODO: Remove this check in production code. We check that this multiplication can't
		// overflow when we compute the Chebyshev coefficients.
		panic("chebyMul: " + err.Error())
	}

	return res
}

// Computes a Chebyshev polynomial using the coefficients provided.
func (x fix192) chebyPoly(coeffs []fix192) fix192 {
	// Compute the Chebyshev polynomial using Horner's method.
	accum := coeffs[0]

	for i := 1; i < len(coeffs); i++ {
		accum = accum.chebyMul(x)
		accum = accum.add(coeffs[i])
	}

	return accum
}

// Clamps the input angle to the range [-π, π] by removing multiples of 2π. The result is the positive
// clamped value and a sign integer (1 or -1).
func (x fix192) clampAngle() (fix192, int64) {
	xUnsigned, sign := x.abs()

	if xUnsigned.ult(fix192Pi) {
		// If the input is already less than π, we can just return it as is.
		return xUnsigned, sign
	} else if fix192TwoPi.ult(xUnsigned) {
		// If the input is larger than 2π, we we want to find the angle modulo 2π. The most obvious
		// way to do this would be to do a 192x192 division by a fix192 representation of 2π, and to
		// look at the remainder. However implementing that division is non-trivial and (somewhat
		// surprisingly!) we don't need 192x192 division for any other operations. Given that we
		// know the divisor is _always_ 2π, we can craft a special modulus operation that is quite
		// efficient.
		//
		// Instead of computing the remainder directly, we will compute the quotient (using some
		// efficient tricks), and then determine the remainder by subtracting that quotient times 2π
		// from the input. Once we have an accurate integer quotient, we can cheaply and accurately
		// multiply it by 2π and then subtract that product from the input to get the residual
		// value.
		//
		// To compute this quotient, we start with a carefully selected multiple of 2π that fits in
		// 64-bits and divide the input by that value. By choosing a 64-bit divisor, we can do a
		// single 192x64 division (which is MUCH cheaper than 192x192). Furthermore, we can choose
		// this 64-bit value such that this integer value, divided by the multiple used to compute
		// it, is especially close to 2π, ensuring that the error introduced by the approximation is
		// quite small. (Think about the approximations of π like 22/7 and 335/113, we're kind of
		// doing the same thing, but with a denominator much larger than 7 or 113.)
		//
		// Let's call that "magic multiple" m. So m = k•2π, where k is some large integer. Our
		// input, x, is conceptually a real number, but what we actually have is x * 10^24 * 2^64
		// (the scale of fix192). So, we want to compute floor(x / 2π), but we're starting with the
		// scaled value of x and m, so, it would be straightforward for us to compute the value s
		// such that, s = (x • 10^24 • 2•64) / m, which is (by expanding m = k•2π):
		//
		//       (x • 10^24 • 2^64)
		//  s =  ------------------
		//            (k • 2π)
		//
		// Once we have s, we can rearrange this equation to solve for the value we REALLY want:
		//
		//        s•k          x
		//  -------------- = ----
		//  (10^24 • 2^64)    2π
		//
		// What this tells us is once we compute s (via the 192x64 bit division operation), we can
		// just multiply that value by k (the scale factor used for m) and then divide by the scale
		// factor of fix192 (10^24 • 2^64). Provided m / k is a close approximation of 2π, we should
		// have the integer part of x / 2π which we can then use to compute the remainder.
		//
		// Dividing by 10^24 • 2^64 can be partially managed by a shift by 24 + 64, and then a
		// division by 5^24. This is another 192x64 division (which isn't terribly expensive), but
		// we can get rid of it entirely if choose a value for k that is itself a multiple of 5^24!
		// If we chose k = j•5^24 for some integer j, then we can replace the equation above with:
		//
		//     s•j•5^24        x
		//  -------------- = ----
		//  (10^24 • 2^64)    2π
		//
		// If we then expand 10^24 into 5^24 • 2^24:
		//
		//        s•j•5^24           x
		//  -------------------- = ----
		//  (5^24 • 2^24 • 2^64)    2π
		//
		// Now the 5^24 terms cancel:
		//
		//       s•j          x
		//  ------------- = ----
		//  (2^24 • 2^64)    2π
		//
		// (Note that dividing by 2^64 is just dropping a 64-bit word.)
		//
		// So we're now able to compute x/2π by just dividing the fixed-point representation of x by
		// m (called clampAngleTwoPiMultiple in code), giving us a value for s. Multiply that s
		// value by j (called clampAngleTwoPiFactor in code), drop the bottom word (functionally
		// dividing by 2^64), and shift the result right by 24 to perform the final division by
		// 2^24.
		//
		// One last little optimization: if we ensure that m = k•2π is larger that the high-word of
		// any 192-bit input, we know that the the result of dividing all 192-bit inputs by m will
		// be a 128-bit value. Since the input to this function is a signed 192-bit value, and we
		// have taken its absolute value, any value of m larger than 2^63 will ensure that the
		// quotient of all possible input values will fit in 128 bits, simplifying the other
		// multiplication and shifting operations.

		// Compute s = x / m
		s, _ := div192by64(xUnsigned.Hi, xUnsigned.Mid, xUnsigned.Lo, clampAngleTwoPiMultiple)

		// Multiply s by j, dropping the bottom word to effectively divide by 2^64
		var temp raw128
		temp.Hi, temp.Lo, _ = mul128By64(s, clampAngleTwoPiFactor)

		// Shift the result down by 24 bits to effectively divide by 2^24
		temp = ushiftRight128(temp, 24)

		// We know that our real quotient will fit in 64 bits, so we can just take the lower
		// 64 bits of the result and cast it to a uint64.
		q := uint64(temp.Lo)

		// Now we use a simple product to compute the sum of all of the "whole" multiples of 2π that
		// we need to subtract from the input. (NOTE: We just a value for 2π that is specific to
		// clampAngle, to ensure it was rounded down so we can add the error term below.)
		twoPiProd := clampAngleTwoPi.uintMul(q)

		// If q is large (and it can be as big as 1e13), the error from using "only" 192-bits of
		// precision for 2π could be significant. To correct for this, we can take the "next"
		// 64-bits of the representation of 2π (clampAngleTwoPiResidual) and multiply that by q to
		// get an error term that we can add to the product. Since the error has an additional
		// factor of 2^64 in the denominator, we only need to add the hi part to our product.
		errorTermHi, _ := mul64(raw64(q), clampAngleTwoPiResidual)
		twoPiProd = twoPiProd.add(fix192{0, 0, errorTermHi})
		res := xUnsigned.sub(twoPiProd)

		// In all of my testing, the q value has ended up being accurate for all inputs, without
		// further adjustment. However, we computed the value of m by rounding down, so if there IS
		// an error, it will result in a value for q that is too large by one. In that case, the
		// subtraction would result in a negative value, and we can correct by adding back one
		// multiple of 2π.
		if isNeg64(res.Hi) {
			// If the result is negative, we overshot by one 2π multiple, so we add back one 2π
			// multiple to get the correct remainder.
			var carry uint64
			res, carry = add192(res, clampAngleTwoPi, 0)

			if carry == 0 {
				// If we do the correct, we should be flipping a negative value into a positive,
				// which would result in the carry flag being set. If that flag isn't set, then
				// something has gone horribly wrong!
				panic("clampAngle: residual adjustment failed")
			}
		}

		// res is now scaled down below 2π. If the angle is greater than π, subtract it from 2π to bring it
		// into the range [0, π] and flip the sign flag.
		if fix192Pi.ult(res) {
			res = fix192TwoPi.sub(res)
			sign *= -1
		}

		return res, sign
	} else {
		// If we get here, the original input was between π and 2π, subtract it from 2π to bring it
		// into the range [0, π] and flip the sign flag.
		res := fix192TwoPi.sub(xUnsigned)
		sign *= -1

		return res, sign
	}
}

func add192(a, b fix192, carryIn uint64) (res fix192, carryOut uint64) {
	res.Lo, carryOut = add64(a.Lo, b.Lo, carryIn)
	res.Mid, carryOut = add64(a.Mid, b.Mid, carryOut)
	res.Hi, carryOut = add64(a.Hi, b.Hi, carryOut)

	return
}

func mul192by64(a fix192, b raw64) (xhi, hi, mid, lo raw64) {
	var carry uint64

	loHi, loLo := mul64(a.Lo, b)
	midHi, midLo := mul64(a.Mid, b)
	hiHi, hiLo := mul64(a.Hi, b)

	lo = loLo
	mid, carry = add64(loHi, midLo, 0)
	hi, carry = add64(midHi, hiLo, carry)
	xhi, _ = add64(hiHi, raw64Zero, carry)

	return
}

func div256by64(xhi, hi, mid, lo raw64, y raw64) (quo fix192, rem raw64) {
	quo.Hi, rem = div64(xhi, hi, y)
	quo.Mid, rem = div64(rem, mid, y)
	quo.Lo, rem = div64(rem, lo, y)

	return quo, rem
}

// func (x fix192) tan() (fix192, error) {
// 	// tan(x) = sin(x) / cos(x)
// 	// We don't want to just call the sin() and cos() methods directly since we will
// 	// just double-up the call to clampAngle().

// 	// Normalize the input angle to the range [0, π]
// 	clampedX, sign := x.clampAngle()

// 	if clampedX.lt(fix192_old{0, raw128{0x1000000000000000, 0}}) {
// 		// If the value is less than 1/8, we can direcly use our chebyTan() calculation
// 		res := clampedX.f.chebyTan()
// 		return res.applySign(sign)
// 	}

// 	var y fix192_old

// 	if clampedX.eq(fix192HalfPi) {
// 		// In practice, this will probably never happen since the input types have too little precision
// 		// to _exactly equal_ π/2 at fix192 precision. However! We handle it just in case...
// 		return fix192_old{}, ErrOverflow
// 	} else if clampedX.lt(fix192HalfPi) {
// 		// This y value will be passed to sin() to compute cos(x), unless we are close enough to π/2
// 		// to use the chebyTan() method.
// 		y = fix192HalfPi.sub(clampedX)

// 		// See if x is close enough to π/2 that we can use the tan(π/2 - x) identity.
// 		if y.lt(fix192_old{0, raw128{0x1000000000000000, 0}}) {
// 			// tan(π/2 - x) = 1 / tan(x)
// 			// We compute tan(x) using the chebyTan() method, and then take the inverse.
// 			inverseTan := y.f.chebyTan()
// 			res, err := inverseTan.f.inverse()

// 			if err != nil {
// 				if sign < 0 {
// 					return fix192_old{}, ErrNegOverflow
// 				} else {
// 					return fix192_old{}, ErrOverflow
// 				}
// 			}

// 			return res.applySign(sign)
// 		}
// 	} else {
// 		// The input is greater than π/2, see if it's close enough to use the chebyTan() method.
// 		if isEqual64(clampedX.i, fix192HalfPi.i) {
// 			// We don't bother doing the subtraction if the integer parts aren't equal...
// 			halfPiDiff, _ := sub128(clampedX.f, fix192HalfPi.f, 0)

// 			if halfPiDiff.Hi <= 0x1000000000000000 {
// 				inverseTan := halfPiDiff.chebyTan()
// 				// tan(x) = -tan(π/2 - x)
// 				sign *= -1
// 				res, err := inverseTan.f.inverse()

// 				if err != nil {
// 					if sign < 0 {
// 						return fix192_old{}, ErrNegOverflow
// 					} else {
// 						return fix192_old{}, ErrOverflow
// 					}
// 				}

// 				return res.applySign(sign)
// 			}
// 		}

// 		// if the input is not close enough to π/2, we'll need to compute
// 		// cos(x) = -sin(3π/2 − x)
// 		y = fix192ThreeHalfPi.sub(clampedX)
// 		sign *= -1
// 	}

// 	sinX := clampedX.chebySin()
// 	cosX := y.chebySin()

// 	if cosX.i == 1 {
// 		// If cosX is 1, we can just return sinX as the result.
// 		return sinX.applySign(sign)
// 	} else if isZero128(cosX.f) || isIota128(cosX.f) {
// 		// If cosX is zero or iota, we treat it as overflow
// 		if sign < 0 {
// 			return fix192_old{}, ErrNegOverflow
// 		} else {
// 			return fix192_old{}, ErrOverflow
// 		}
// 	} else {
// 		// Divide the result, if cos is small, we scale up both the numerator and denominator
// 		// to get more precision. (Note that the largest possible value for sinX is 1, so we
// 		// can scale it up by 2^63 without overflowing.)
// 		shift := leadingZeroBits64(raw64(cosX.f.Hi))
// 		if shift > 63 {
// 			shift = 63
// 		}

// 		sinX = sinX.shiftLeft(shift)
// 		cosX = cosX.shiftLeft(shift)

// 		// We know have a 192-bit numerator for sin, and a 128-bit denominator for cos.
// 		// Remember, though, that sinX and cosX are the true values each multiplied by 2**128
// 		// (and the extra shift). So if we just divide them, we'll get the result _as an integer_,
// 		// which is a lot less precision that we'd like! So, we do two divisions. A division of
// 		// to get the integer part, and then a second division (using the remainder) to get the fractional part.
// 		//
// 		// Note this division can only overflow if the high part of the numerator is greater than or equal to the
// 		// denominator. Since the largest value of sinX.i is 1, this can only happen if cosX.f is that or equal to 1.
// 		// (i.e. isIota128(cosX.f) or isZero128(cosX.f)). We checked for this earlier and returned an error.

// 		quoHi, rem := div128(raw128{0, sinX.i}, sinX.f, cosX.f)
// 		quoLo, _ := div128(rem, raw128Zero, cosX.f)

// 		res := fix192_old{i: quoHi.Lo, f: quoLo}

// 		return res.applySign(sign)
// 	}
// }

// // Computes the geometric inverse of a fractional value (1/x). Uses
// // Newton-Raphson reciprocal iteration.
// func (x raw128) inverse() (fix192_old, error) {

// 	// NOTE: Returns 128 if x == 0
// 	fracZeros := leadingZeroBits128(x)

// 	if fracZeros >= 63 {
// 		// If the input is less than 2^-63, we can't represent it as a fix192 value.
// 		return fix192_old{}, ErrOverflow
// 	}

// 	// We use the following recursive formula to compute the inverse:
// 	// rₙ₊₁ = rₙ (2 - x·rₙ)

// 	// We start our estimate with a value that is 2^fracZeros, which is the
// 	// largest power of two that is less than or equal to 1/x. This ensures
// 	// our first estimate is on the same order of magnitude as the final result.
// 	two := fix192_old{2, raw128Zero}
// 	est := fix192_old{raw64(1 << fracZeros), raw128Zero}

// 	for i := 0; i < 8; i++ {
// 		prod := est.mulByFraction(x)
// 		est, _ = est.umul(two.sub(prod))
// 	}

// 	return est, nil
// }
