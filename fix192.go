package fixedPoint

// fix192 is an internal fixed-point type that represents the integer part as a 64-bit int, and the
// fractional part as 128-bits. This provides more precision for pow() and exp() calculations.
// The fractional part is always positive, so -1.5 would be represented as (-2, 0.5).
//
// The advantage of this representation is that for values <1, we can ignore the integer part and
// multiplications are super efficient: Just a 128x128->256 multiplication, and use the top 128 bits
// of the result. No division, or even shifting is required. All of the transcendental functions we
// implement have an inner loop that works for values <1 and consists mostly of multiplications.
//
// (Suprisingly, this representation allows us to compute these functions as fast, or even faster
// than working with just 64-bit fixed-point types, since a 64-bit multiplication includes a division.)
type fix192 struct {
	i raw64
	f raw128
}

func (x UFix64) toFix192() fix192 {
	// Convert Fix64 to fix192 by extracting the integer and fractional parts.
	intPart := uint64(x) / uint64(Fix64Scale)
	fracPart := uint64(x) % uint64(Fix64Scale)

	fracPart128, rem := div192by128(raw64(fracPart), 0, 0, raw128{0, raw64(Fix64Scale)})

	if rem.Hi >= 0x8000000000000000 {
		// If the remainder is greater than or equal to 0.5, we round up
		// (Can't overflow since we are working at much higher precision than the input.)
		fracPart128, _ = add128(fracPart128, raw128Zero, 1)
	}

	return fix192{i: raw64(intPart), f: fracPart128}
}

func (x Fix64) toFix192() fix192 {
	xUnsigned, sign := x.Abs()
	res := xUnsigned.toFix192()
	return res.applySign(sign)
}

func (x UFix128) toFix192() fix192 {
	// Convert UFix128 to fix192 by extracting the integer and fractional parts.
	intPart, fracPart := div128(raw128Zero, raw128(x), raw128(Fix128One))
	fracPart128, rem := div128(fracPart, raw128Zero, raw128(Fix128One))

	if rem.Hi >= 0x8000000000000000 {
		// If the remainder is greater than or equal to 0.5, we round up
		// (Can't overflow since we are working at much higher precision than the input.)
		fracPart128, _ = add128(fracPart128, raw128Zero, 1)
	}

	return fix192{i: raw64(intPart.Lo), f: fracPart128}
}

func (x Fix128) toFix192() fix192 {
	xUnsigned, sign := x.Abs()
	res := xUnsigned.toFix192()
	return res.applySign(sign)
}

func (a fix192) toUFix64() (UFix64, error) {
	if isNeg64(a.i) {
		// If the integer part is negative, we can't represent this as a UFix64
		return UFix64Zero, ErrNegOverflow
	}
	// Scale the integer part and add the fractional part.
	hi, i := mul64(a.i, Fix64Scale)
	if !isZero64(hi) {
		// If the high part is non-zero, the integer part overflowed.
		return UFix64Zero, ErrOverflow
	}
	// The fractional part is a 128-bit value, but we only need the most significant 64 bits
	f, lo := mul64(raw64(a.f.Hi), Fix64Scale)

	// Round up if the discarded part is >= 0.5.
	if uint64(lo) >= 0x8000000000000000 {
		f++
	}

	res, carry := add64(i, f, 0)

	if carry != 0 {
		// If there was a carry, the result overflowed.
		return UFix64Zero, ErrOverflow
	}

	if !isZero128(a.f) && isZero64(res) {
		// If the fractional part of the input is non-zero, but the result
		// is zero, we have an underflow.
		return UFix64Zero, ErrUnderflow
	}

	return UFix64(res), nil
}

func (x fix192) toFix64() (Fix64, error) {
	xUnsigned, sign := x.abs()
	res, err := xUnsigned.toUFix64()

	if err != nil {
		return Fix64Zero, err
	}

	return res.ApplySign(sign)
}

func (a fix192) toUFix128() (UFix128, error) {
	if isNeg64(a.i) {
		// If the integer part is negative, we can't convert to UFix128.
		return UFix128Zero, ErrNegOverflow
	}
	// Scale the integer part and add the fractional part.
	hi, i := mul128(raw128{0, a.i}, raw128(Fix128One))
	if !isZero128(hi) {
		// If the high part is non-zero, the integer part overflowed.
		return UFix128Zero, ErrOverflow
	}
	// Scale up the fractional part
	f, lo := mul128(a.f, raw128(Fix128One))
	if uint64(lo.Hi) >= 0x8000000000000000 {
		f, _ = add128(f, raw128Zero, 1)
	}
	res, carry := add128(f, i, 0)
	if carry != 0 {
		// If there was a carry, the result overflowed.
		return UFix128Zero, ErrOverflow
	}

	var HundredTrillion = raw128{0x4b3b4ca85a86c47a, 0x098a224000000000}

	if !ult128(res, HundredTrillion) {
		// 2^64 % 10 = 6, so we multiply the modulus of the high part by 6
		lastDigit := ((res.Hi%10)*6 + (res.Lo % 10)) % 10

		if lastDigit >= 5 {
			// If the last digits are 50 or greater, we round up.
			res, carry = add128(raw128(res), raw128{0, 10 - lastDigit}, 0)

			if carry != 0 {
				// If there was a carry, the result overflowed.
				return UFix128Zero, ErrOverflow
			}
		} else {
			// Round down
			res, _ = sub128(raw128(res), raw128{0, lastDigit}, 0)
		}
	}

	return UFix128(res), nil
}

func (x fix192) toFix128() (Fix128, error) {
	xUnsigned, sign := x.abs()
	res, err := xUnsigned.toUFix128()

	if err != nil {
		return Fix128Zero, err
	}

	return res.ApplySign(sign)
}

func (a fix192) lt(b fix192) bool {
	if isEqual64(a.i, b.i) {
		return ult128(a.f, b.f)
	} else {
		return slt64(a.i, b.i)
	}
}

func (a fix192) eq(b fix192) bool {
	return isEqual64(a.i, b.i) && isEqual128(a.f, b.f)
}

func (x fix192) applySign(sign int64) fix192 {
	if sign < 0 {
		x.i = neg64(x.i)
		if !isZero128(x.f) {
			// If the fractional part is non-zero, we subtract one from the integer part
			// and flip the sign of the fractional part by subtracting it from zero.
			x.i, _ = sub64(x.i, raw64Zero, 1)
			x.f, _ = sub128(raw128Zero, x.f, 0)
		}
	}

	return x
}

func (x fix192) abs() (fix192, int64) {
	if isNeg64(x.i) {
		x.i = neg64(x.i)
		if !isZero128(x.f) {
			// If the fractional part is non-zero, we subtract one from the integer part
			// and flip the sign of the fractional part by subtracting it from zero.
			x.i, _ = sub64(x.i, raw64Zero, 1)
			x.f, _ = sub128(raw128Zero, x.f, 0)
		}

		return x, -1
	}

	return x, 1
}

func (x fix192) shiftLeft(shift uint64) (res fix192) {
	if shift >= 64 {
		panic("fix192 only supports shifts less than 64")
	}

	res.i = shiftLeft64(x.i, shift) | ushiftRight64(x.f.Hi, 64-shift)
	res.f = shiftLeft128(x.f, shift)

	return res
}

func (x fix192) shiftRight(shift uint64) (res fix192) {
	if shift >= 64 {
		panic("fix192 only supports shifts less than 64")
	}

	if isNeg64(x.i) {
		panic("fix192 shiftRight() not implemented for negative values")
	}

	if shift == 0 {
		return x
	}

	res.i = sshiftRight64(x.i, shift)
	res.f = ushiftRight128(x.f, shift)

	// Shift the bottom part of the integer into the top part of the fraction
	res.f.Hi |= shiftLeft64(x.i, 64-shift)

	return res
}

func (a fix192) add(b fix192) (res fix192) {
	var carry uint64

	res.f, carry = add128(a.f, b.f, 0)
	res.i, _ = add64(a.i, b.i, carry)

	return res
}

func (a fix192) sub(b fix192) (res fix192) {
	var borrow uint64

	res.f, borrow = sub128(a.f, b.f, 0)
	res.i, _ = sub64(a.i, b.i, borrow)

	return res
}

func (a fix192) umul(b fix192) (fix192, error) {
	// a = a.i + a.f, b = b.i + b.f
	// a•b = (a.i + a.f)•(b.i + b.f)
	//      = a.i•b.i + a.i•b.f + a.f•b.i + a.f•b.f
	if isNeg64(a.i) || isNeg64(b.i) {
		panic("mul() not implemented for negative fix192 values")
	}

	// Compute the integer part of term one
	hi, i1 := mul64(a.i, b.i)
	if !isZero64(hi) {
		return fix192{}, ErrOverflow
	}

	// Compute the cross terms
	i2, f2 := mul128By64(a.f, b.i)
	i3, f3 := mul128By64(b.f, a.i)

	// Compute the fractional part of term four
	f4, lo := mul128(a.f, b.f)

	if lo.Hi >= 0x8000000000000000 {
		// If the low part is greater than or equal to 0.5, we round up
		f4, _ = add128(f4, raw128Zero, 1)
	}

	var res fix192
	var carry1, carry2 uint64

	// Sum up the fractional parts, holding on to the carries
	res.f, carry1 = add128(f2, f3, 0)
	res.f, carry2 = add128(res.f, f4, 0)

	// Sum up the integer parts, using the carries from the fractional parts
	res.i, carry1 = add64(i1, raw64(i2.Lo), carry1)
	res.i, carry2 = add64(res.i, raw64(i3.Lo), carry2)

	// If either add operation produced a carry, we have an overflow.
	if carry1 != 0 || carry2 != 0 {
		return fix192{}, ErrOverflow
	}

	return res, nil
}

func (a fix192) smul(b fix192) (fix192, error) {
	aUnsigned, aSign := a.abs()
	bUnsigned, bSign := b.abs()

	resUnsigned, err := aUnsigned.umul(bUnsigned)

	if err != nil {
		return fix192{}, err
	}

	// Apply the sign to the result
	return resUnsigned.applySign(aSign * bSign), nil
}

// Multiplies the fractional parts of fix192 values, can't overflow, but it CAN underflow
// which this method does not check for. Callers that care should check for:
//
//	a != 0 && b != 0 && res == 0
func (a raw128) mulFrac(b raw128) (res raw128) {
	var lo raw128
	res, lo = mul128(a, b)

	if lo.Hi >= 0x8000000000000000 {
		// Round up if the lo part is greater than or equal to 0.5.
		// Can't overflow since both inputs are <1
		res, _ = add128(res, raw128Zero, 1)
	}

	return res
}

// Squares a fix192 value, assuming the integer part is non-negative. Slightly
// more efficient than multiplying the value by itself (1 less multiplication).
func (a fix192) squared() (fix192, error) {
	if isNeg64(a.i) {
		panic("squared() not implemented for negative fix192 values")
	}

	// a = i + f
	// a^2 = (i + f)^2
	//      = i^2 + 2*i*f + f^2

	// Square the integer part
	hi, iSquared := mul64(a.i, a.i)

	if !isZero64(hi) {
		// If the high part is non-zero, the integer squared overflowed.
		return fix192{}, ErrOverflow
	}

	// Double the product of the integer and fractional parts
	ifProductHi, ifProductLo := mul128By64(a.f, shiftLeft64(a.i, 1))

	// Square the fractional part
	fSquared := a.f.mulFrac(a.f)

	var res fix192
	var carry uint64

	res.f, carry = add128(fSquared, ifProductLo, 0)
	res.i, carry = add64(iSquared, raw64(ifProductHi.Lo), carry)

	if carry != 0 {
		return fix192{}, ErrOverflow
	}

	return res, nil
}

func (x fix192) ln(prescale uint64) (fix192, error) {
	if isZero64(x.i) && isZero128(x.f) {
		return fix192{}, ErrDomain
	}

	if isNeg64(x.i) {
		return fix192{}, ErrDomain
	}

	// The Taylor expansion of ln(x) takes powers of the ratio (x - 1) / (x + 1). Since
	// we know we can efficiently multiply values between 0-1, we start by scaling X
	// by a power of two to be between 1-2. We can do this very efficiently just by
	// counting the number of leading zero bits in x. If x is less than 1, we shift the
	// input to the left until it is between 1 and 2, and if x is >= 2, we shift to the right.

	// We use the value k to keep track of which power of 2 we used to scale the input. Negative
	// values mean we scaled the input UP, and need to subtract a multiple of ln(2) at the end.
	var scaledX fix192
	var k uint64
	scaledUp := false

	if isZero64(x.i) {
		// If the integer part is zero, we can only have a fractional part. See how much
		// we need to scale it up by counting the zero bits, and adding one (so the top
		// set bit "flows over" into the integer part)
		fracScale := leadingZeroBits128(x.f) + 1

		scaledX = fix192{1, shiftLeft128(x.f, fracScale)}
		k = fracScale + prescale
		scaledUp = true
	} else {
		if prescale != 0 {
			panic("prescale can only be used for values < 1")
		}
		// If the integer part is non-zero, we can just count the leading zero bits
		// in the integer part to determine how much we need to scale it up or down.
		intScale := 63 - leadingZeroBits64(raw64(x.i))

		scaledX = x.shiftRight(intScale)
		k = intScale
	}

	// We now compute the natural log of the scaled value using the expansion:
	// ln(x) = 2 * (z + z^3/3 + z^5/5 + z^7/7 + ...)
	// where z = (x - 1) / (x + 1)
	num := fix192{0, scaledX.f}
	den := fix192{2, scaledX.f}

	// We don't have a division operation for fix192, but if we divide top and bottom by
	// four (using shiftRight), we can just divide the fractional parts.
	num = num.shiftRight(2)
	den = den.shiftRight(2)

	z, rem := div128(num.f, raw128Zero, den.f)

	if rem.Hi >= 0x8000000000000000 {
		// If the remainder is greater than or equal to 0.5, we round up
		z, _ = add128(z, raw128Zero, 1)
	}

	// Precompute z^2 to avoid recomputing it in the loop.
	z2 := z.mulFrac(z)

	term := z
	sum := z
	iter := uint64(1)

	// Keep iterating until "term" and/or "next" rounds to zero
	for {
		term = term.mulFrac(z2)

		next := uintDiv128(term, iter*2+1)

		if isZero128(next) {
			break
		}

		sum, _ = add128(sum, next, 0)
		iter += 1
	}

	// Apply the global scaling factor of 2 to the result.
	res := fix192{0, sum}
	res = res.shiftLeft(1)

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did above. Note that k*ln2 must strictly lie between minLn and maxLn constants
	// and we chose ln2Multiple and ln2Factor so they can't overflow when multiplied
	// by values in that range.
	ln2_192 := fix192{0, raw128{0xb17217f7d1cf79ab, 0xc9e3b39803f2f6af}}
	powerCorrection, _ := ln2_192.umul(fix192{raw64(k), raw128Zero})

	if scaledUp {
		res = res.sub(powerCorrection)
	} else {
		res = res.add(powerCorrection)
	}

	return res, nil
}

func (x fix192) exp() (fix192, error) {

	powerIndex := int64(x.i) - smallestExpIntPower

	if powerIndex < 0 {
		return fix192{}, ErrUnderflow
	} else if powerIndex >= int64(len(expIntPowers)) {
		return fix192{}, ErrOverflow
	}

	intExp192 := expIntPowers[powerIndex]
	fracExp192 := x.f.fracExp()

	return intExp192.umul(fracExp192)
}

func (a fix192) pow(b fix192, prescale uint64) (fix192, error) {
	aLn, err := a.ln(prescale)

	if err != nil {
		return fix192{}, err
	}

	prod, err := aLn.smul(b)

	if err != nil {
		return fix192{}, err
	}

	return prod.exp()
}

func (x raw128) fracExp() fix192 {
	// We compute the fractional exp using the typical Taylor series:
	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...
	// However, we can make some simplifying assumptions that speed up the calculations:
	// 1. If the input is larger than ln(2), we divide the input by 2 and then square the result
	// 2. We ignore the first term in the expansion (the 1); all other terms are guaranteed to sum up
	//    to less than 1 (since the final result is <2 from the previous step), so we can JUST do our
	//    computations on the fractional part.
	// 3. Multiplication of two fractional 128-bit values is very easy, we just do a 128x128-> 256 bit
	//    multiplication, and just use the top 128 bits of the result.
	term := x
	sum := term
	iter := uint64(1)
	i := raw64(1)

	for {
		term = term.mulFrac(x)

		iter++
		term = uintDiv128(term, iter)

		if isZero128(term) {
			break
		}

		var carry uint64
		sum, carry = add128(sum, term, 0)
		i, _ = add64(i, raw64Zero, carry)
	}

	return fix192{i, sum}
}

func (x fix192) sin() (fix192, error) {
	clampedX, sign := x.clampAngle()

	res := clampedX.chebySin()

	return res.applySign(sign), nil
}

func (x fix192) cos() (fix192, error) {
	// Normalize the input angle to the range [0, π], with a flag indicating
	// if the result should be interpreted as negative.
	clampedX, _ := x.clampAngle()
	sign := int64(1)

	// We use the following identities to compute cos(x):
	//     cos(x) = sin(π/2 - x)
	//     cos(x) = -sin(3π/2 − x)
	// If x is is less than or equal to π/2, we can use the first identity,
	// if x is greater than π/2, we use the second identity.
	// In both cases, we end up with a value in the range [0, π], to pass
	// to chebySin().
	var y fix192

	if clampedX.lt(fix192HalfPi) {
		// cos(x) = sin(π/2 - x)
		y = fix192HalfPi.sub(clampedX)
	} else {
		// cos(x) = -sin(3π/2 − x)
		y = fix192ThreeHalfPi.sub(clampedX)
		sign *= -1
	}

	res := y.chebySin()

	return res.applySign(sign), nil
}

func (x fix192) tan() (fix192, error) {
	// tan(x) = sin(x) / cos(x)
	// We don't want to just call the sin() and cos() methods directly since we will
	// just double-up the call to clampAngle().

	// Normalize the input angle to the range [0, π]
	clampedX, sign := x.clampAngle()

	if clampedX.lt(fix192{0, raw128{0x1000000000000000, 0}}) {
		// If the value is less than 1/8, we can direcly use our chebyTan() calculation
		res := fix192{0, clampedX.f.chebyTan()}
		return res.applySign(sign), nil
	}

	var y fix192

	if clampedX.eq(fix192HalfPi) {
		// In practice, this will probably never happen since the input types have too little precision
		// to _exactly equal_ π/2 at fix192 precision. However! We handle it just in case...
		return fix192{}, ErrOverflow
	} else if clampedX.lt(fix192HalfPi) {
		// This y value will be passed to sin() to compute cos(x), unless we are close enough to π/2
		// to use the chebyTan() method.
		y = fix192HalfPi.sub(clampedX)

		// See if x is close enough to π/2 that we can use the tan(π/2 - x) identity.
		if y.lt(fix192{0, raw128{0x1000000000000000, 0}}) {
			// tan(π/2 - x) = 1 / tan(x)
			// We compute tan(x) using the chebyTan() method, and then take the inverse.
			inverseTan := y.f.chebyTan()
			res, err := inverseTan.inverse()

			if err != nil {
				if sign < 0 {
					return fix192{}, ErrNegOverflow
				} else {
					return fix192{}, ErrOverflow
				}
			}

			return res.applySign(sign), nil
		}
	} else {
		// The input is greater than π/2, see if it's close enough to use the chebyTan() method.
		if isEqual64(clampedX.i, fix192HalfPi.i) {
			// We don't bother doing the subtraction if the integer parts aren't equal...
			halfPiDiff, _ := sub128(clampedX.f, fix192HalfPi.f, 0)

			if halfPiDiff.Hi <= 0x1000000000000000 {
				inverseTan := halfPiDiff.chebyTan()
				// tan(x) = -tan(π/2 - x)
				sign *= -1
				res, err := inverseTan.inverse()

				if err != nil {
					if sign < 0 {
						return fix192{}, ErrNegOverflow
					} else {
						return fix192{}, ErrOverflow
					}
				}

				return res.applySign(sign), nil
			}
		}

		// if the input is not close enough to π/2, we'll need to compute
		// cos(x) = -sin(3π/2 − x)
		y = fix192ThreeHalfPi.sub(clampedX)
		sign *= -1
	}

	sinX := clampedX.chebySin()
	cosX := y.chebySin()

	if cosX.i == 1 {
		// If cosX is 1, we can just return sinX as the result.
		return sinX.applySign(sign), nil
	} else if isZero128(cosX.f) || isIota128(cosX.f) {
		// If cosX is zero or iota, we treat it as overflow
		if sign < 0 {
			return fix192{}, ErrNegOverflow
		} else {
			return fix192{}, ErrOverflow
		}
	} else {
		// Divide the result, if cos is small, we scale up both the numerator and denominator
		// to get more precision. (Note that the largest possible value for sinX is 1, so we
		// can scale it up by 2^63 without overflowing.)
		shift := leadingZeroBits64(raw64(cosX.f.Hi))
		if shift > 63 {
			shift = 63
		}

		sinX = sinX.shiftLeft(shift)
		cosX = cosX.shiftLeft(shift)

		// We know have a 192-bit numerator for sin, and a 128-bit denominator for cos.
		// Remember, though, that sinX and cosX are the true values each multiplied by 2**128
		// (and the extra shift). So if we just divide them, we'll get the result _as an integer_,
		// which is a lot less precision that we'd like! So, we do two divisions. A division of
		// to get the integer part, and then a second division (using the remainder) to get the fractional part.
		//
		// Note this division can only overflow if the high part of the numerator is greater than or equal to the
		// denominator. Since the largest value of sinX.i is 1, this can only happen if cosX.f is that or equal to 1.
		// (i.e. isIota128(cosX.f) or isZero128(cosX.f)). We checked for this earlier and returned an error.

		quoHi, rem := div128(raw128{0, sinX.i}, sinX.f, cosX.f)
		quoLo, _ := div128(rem, raw128Zero, cosX.f)

		res := fix192{i: quoHi.Lo, f: quoLo}

		return res.applySign(sign), nil
	}
}

func (x fix192) clampAngle() (fix192, int64) {
	xUnsigned, sign := x.abs()

	if xUnsigned.i < 3 {
		// If the input is less than 3, we can just return it as is.
		return xUnsigned, sign
	}

	reducedX := xUnsigned.shiftRight(3)

	_, rem := div192by128(reducedX.i, reducedX.f.Hi, reducedX.f.Lo,
		raw128{0xc90fdaa22168c234, 0xc4c6628b80dc1cd1})

	// rem is now the input angle, modulo 2π, divided by 8 (to keep the value
	// less than 1, so we can ignore the integer part). If the angle is greater
	// than π, subtract it from 2π to bring it into the range [0, π] and flip
	// the sign flag.
	if ult128(raw128{0x6487ed5110b4611a, 0x62633145c06e0e68}, rem) {
		rem, _ = sub128(raw128{0xc90fdaa22168c234, 0xc4c6628b80dc1cd1}, rem, 0)
		sign *= -1
	}

	res := fix192{0, rem}
	res = res.shiftLeft(3)

	return res, sign
}

// Computes a Chebyshev polynomial at a particular x value. This method
// assumes that the input, coefficients AND result are all strictly less
// than 1.0, with the input and output also being positive.
func (x raw128) chebyPoly(coeffs []coeff) raw128 {
	// Because we have very efficient math functions for fractional values,
	// we can't use Horner's method for polynomial expansion here. Some of the
	// interim values can end up outside the range (0, 1). However, if we do a more
	// naive implementation we have a few extra multiplications, but we know that
	// we'll never end up with an intermediate value outside of a simple fractional
	// value (with the functions we are using, anyway!)

	// Start with the constant term (usually zero)
	accum := coeffs[0].value

	// The current power of x we are computing. Starting with x^1.
	pow := x
	term := pow.mulFrac(coeffs[1].value)
	accum, _ = add128(accum, term, 0)

	for i := 2; i < len(coeffs); i++ {
		pow = pow.mulFrac(x)
		term = pow.mulFrac(coeffs[i].value)
		if coeffs[i].isNeg {
			accum, _ = sub128(accum, term, 0)
		} else {
			accum, _ = add128(accum, term, 0)
		}
	}

	return accum
}

func (x fix192) chebySin() fix192 {
	if int64(x.i) >= 4 || isNeg64(x.i) {
		// TODO: Remove this check after testing is complete.
		panic("chebySin requires the input to be in the range (0, π)")
	}

	// Leverage the identity sin(x) = sin(π - x) to keep the input angle
	// in the range [0, 2] (required for the rest of this function).
	if x.i >= 2 {
		pi192 := fix192{3, raw128{0x243f6a8885a308d3, 0x13198a2e03707344}}
		x = pi192.sub(x)
	}

	// At this point, x is <2, we further reduce inputs >= 1 by half to
	// ensure the input to the polynomial is strictly less than 1. This
	// ensures that all powers of x will be less than 1, which allows us
	// to use 128-bit multiplication without the possiblity of overflow.
	if x.i == 1 {
		// We can use the following identity to reduce the input angle by half:
		//     sin(x) = 2•sin(x/2)•cos(x/2)
		// We don't have a direct implementation of cos(), so we leverage the fact
		// that cos(y) = 1-2•sin²(y/2), so we can further expand this to:
		//     sin(x) = 2•sin(x/2)•(1 - 2•sin²(x/4))
		halfXFrac := ushiftRight128(x.f, 1)
		halfXFrac.Hi += 0x8000000000000000 // Add 0.5 to the fractional part (can't overflow)
		sinXHalf := fix192{0, halfXFrac}.chebySin()
		quarterXFrac := ushiftRight128(halfXFrac, 1)
		sinXQuarter := fix192{0, quarterXFrac}.chebySin()

		// cos(x/2) = 1 - 2•sin²(x/4)
		// Note: We subtract from zero because the true value of one can't fit in 128-bits
		// but the result is the same because of twos-complement.
		sinXQuarterSquared := sinXQuarter.f.mulFrac(sinXQuarter.f)
		cosFrac, _ := sub128(raw128Zero, shiftLeft128(sinXQuarterSquared, 1), 0)

		res := sinXHalf.f.mulFrac(cosFrac)

		// If the product of the sign and cos terms is 0.5, the final result (which should
		// be multiplied by 2) will be exactly 1.0, so we return that.)
		if res.Hi == 0x8000000000000000 {
			return fix192{1, raw128{0, 0}}
		}

		return fix192{0, shiftLeft128(res, 1)}
	}

	if ult128(x.f, raw128{0x4942ff, 0x86c20bd3e6e3533a}) {
		// If x is very small, we can just return x since sin(x) is linear for small x.
		return x
	}

	// Start with the constant term of the chebyshev polynomial (often just zero!)
	sum := x.f.chebyPoly(sinChebyCoeffs)

	// This polynomial expansion will never result in 1.0. The only positive inputs for sin()
	// that result in 1.0 are larger than 1 (specifically, π/2), and we only use the polynomial
	// expansion for values less than 1.0. (See the if x.i == 1 block above for more details.)
	return fix192{0, sum}
}

// Returns tan(x) for very small x values, using the Chebyshev polynomial. Input and output
// must be strictly less than 1.0, and positive.
func (x raw128) chebyTan() raw128 {
	return x.chebyPoly(tanChebyCoeffs)
}

// A multiplication funcion that multiplies a complete fix192 value
// by a raw128 fractional value.
func (a fix192) mulByFraction(b raw128) (res fix192) {
	// a = a.i + a.f, b = b.f
	// a•b = (a.i + a.f)•b
	//      = a.i•b + a.f•b

	i1, f1 := mul128By64(b, a.i)
	f2, lo := mul128(a.f, b)

	if lo.Hi >= 0x8000000000000000 {
		// If the low part is greater than or equal to 0.5, we round up
		f2, _ = add128(f2, raw128Zero, 1)
	}

	// Sum up the fractional parts, holding on to the carries
	var carry uint64

	res.f, carry = add128(f1, f2, 0)
	res.i, _ = add64(i1.Lo, raw64Zero, carry)

	return res
}

// Computes the geometric inverse of a fractional value (1/x). Uses
// Newton-Raphson reciprocal iteration.
func (x raw128) inverse() (fix192, error) {

	// NOTE: Returns 128 if x == 0
	fracZeros := leadingZeroBits128(x)

	if fracZeros >= 63 {
		// If the input is less than 2^-63, we can't represent it as a fix192 value.
		return fix192{}, ErrOverflow
	}

	// We use the following recursive formula to compute the inverse:
	// rₙ₊₁ = rₙ (2 - x·rₙ)

	// We start our estimate with a value that is 2^fracZeros, which is the
	// largest power of two that is less than or equal to 1/x. This ensures
	// our first estimate is on the same order of magnitude as the final result.
	two := fix192{2, raw128Zero}
	est := fix192{raw64(1 << fracZeros), raw128Zero}

	for i := 0; i < 8; i++ {
		prod := est.mulByFraction(x)
		est, _ = est.umul(two.sub(prod))
	}

	return est, nil
}
