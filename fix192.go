package fixedPoint

// fix192_old is an internal fixed-point type that represents the integer part as a 64-bit int, and the
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
type fix192_old struct {
	i raw64
	f raw128
}

// A 192-bit fixed-point type used for transcendental calculations. It's uses a scale factor of
// 10**24 * 2**64. This means that the top 128 bites are a valid UFix128 value or Fix128 value, with
// the bottom 64 bits being an extension of the fractional part for additional precision. Using the
// 10**24 factor makes converting between the 128-bit types and this type trivial, and without loss
// of precision. The additional 2**64 factor is very easy to handle because a multiplication or
// division by 2**64 can be handled just by selecting the appropriate raw64 components.
type fix192 struct {
	Hi, Mid, Lo raw64
}

func (a fix192) uintDiv(b uint64) fix192 {
	res, _ := div256by64(0, a.Hi, a.Mid, a.Lo, raw64(b))

	return res
}

func (a fix192) powNearOne(b fix192) (fix192, error) {
	// It's hard to get precision for large powers of numbers close to 1, because ln(a) is very
	// small and lacking precision, and any error is magnified a large exponent. Instead, we can
	// directly compute b•ln(a) by using a different expansion of ln(a).
	//
	// We start with the Taylor/Maclaurin series for values close to 1:
	//     ln(1 + ε) ≈ ε - (ε^2)/2 + (ε^3)/3 - ...
	//
	// We combine this with:
	//     a^b = exp(b * ln(a))
	//
	// Instead of computing ln(a) by itself, we compute b * ln(1 + ε) as:
	//     b * ln(1 + ε) ≈ b•ε - b•(ε^2)/2 + b•(ε^3)/3 - ...
	// Note that this works for both positive and negative values of ε (i.e. if a is slightly less
	// than one)
	fix192One := fix192{0x000000000000d3c2, 0x1bcecceda1000000, 0x0000000000000000}
	epsilon := a.sub(fix192One) // ε = a - 1

	epsilon, eSign := epsilon.abs()
	b, bSign := b.abs()

	term, err := epsilon.umul(b)

	if err == ErrNegOverflow {
		// If the product overflows in the negative direction, the exponential
		// would underflow.
		return fix192{}, ErrUnderflow
	} else if err != nil {
		return fix192{}, err
	}

	sum := term
	iter := uint64(1)

	if eSign < 0 {
		// If epsilon is negative, all of the odd terms (including the first!) will
		// be negative, and all of the even terms are negative already. So, we can just
		// do the whole expansion treating all terms as positive and then flip the sign
		// at the end
		for {
			term, _ = term.umul(epsilon)

			if term.isZero() {
				break
			}

			iter++

			sum = sum.add(term.uintDiv(iter))
		}

		bSign = -bSign // Flip the sign of the result
	} else {
		// If epsilon is positive, we need to alternate between adding and subtracting
		// terms. We do this by just unrolling the loop and literally switching between
		// addition and subtraction.
		iter := uint64(1)

		for {
			term, _ = term.umul(epsilon)

			if term.isZero() {
				break
			}

			iter++

			sum = sum.sub(term.uintDiv(iter))

			term, _ = term.umul(epsilon)

			if term.isZero() {
				break
			}

			iter++

			sum = sum.add(term.uintDiv(iter))
		}
	}

	sum, _ = sum.applySign(bSign)

	return sum.exp()
}

func (a fix192) pow(b fix192) (fix192, error) {
	// Use powNearOne() for values close to 1 if the exponent is greater than one
	// if ult64(0xbe95, a.Hi) && ult64(a.Hi, 0xe8ef) { //  && slt64(0xd3c2, b.Hi)
	// 	return a.powNearOne(b)
	// }

	aLn, err := a.ln_test2()

	if err != nil {
		return fix192{}, err
	}

	prod, err := aLn.smul(b)
	fix192One := fix192{0x000000000000d3c2, 0x1bcecceda1000000, 0x0000000000000000}

	if err == ErrUnderflow {
		// If the product is too small, we treat it as zero, and return 1
		return fix192One, nil
	} else if err == ErrNegOverflow {
		// If the product overflows negative, the result is too small to represent, so we return an error.
		return fix192{}, ErrUnderflow
	} else if err != nil {
		// Overflow errors are just overflow errors...
		return fix192{}, err
	}

	// prod = prod.sshiftRightWithRounding(20)

	return prod.exp()
}

func (x fix192) inverse() (fix192, error) {

	if isNeg64(x.Hi) || x.isZero() {
		// If the input is negative or zero, we can't compute the inverse.
		return fix192{}, ErrDomain
	}

	// NOTE: Returns 128 if x == 0
	zeros := leadingZeroBits192(x)

	if zeros >= 96 {
		// It turns out that any value with more than 96 leading zero bits is so small that it's inverse
		// would overflow the 192-bit fixed-point representation.
		return fix192{}, ErrOverflow
	}

	// We use the following recursive formula to compute the inverse:
	// rₙ₊₁ = rₙ (2 - x·rₙ)

	// We start our estimate by taking the input and shifting it "past" the representation of one
	// by the same number of bits as the original distance. This should get us within a factor of 2
	// of the inverse we are looking for.
	fix192Two := fix192{0x00000000000001a784, 0x379d99db42000000, 0x0000000000000000}
	est := x

	if zeros > Fix128OneLeadingZeros {
		// The input has more leading zeros that one. Shift left by twice the difference.
		est = est.shiftLeft(uint64(zeros-Fix128OneLeadingZeros) * 2)
	} else if zeros < Fix128OneLeadingZeros {
		// The input has fewer leading zeros than one. Shift right by twice the difference.
		est = est.ushiftRight(uint64(Fix128OneLeadingZeros-zeros) * 2)
	}

	for i := 0; i < 8; i++ {
		prod, _ := est.umul(x)
		est, _ = est.umul(fix192Two.sub(prod))
	}

	return est, nil
}

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

func (x fix192) ln_test() (fix192, error) {
	res, err := x.ln_test2()

	if err == nil {
		res = res.sshiftRightWithRounding(20)
	}

	return res, err
}

func (x fix192) ln_test2() (fix192, error) {
	if x.isZero() {
		return fix192{}, ErrDomain
	}

	var scaledX fix192

	k := Fix128OneLeadingZeros - int64(leadingZeroBits192(x))

	if k >= 0 {
		scaledX = x.ushiftRight(uint64(k))
	} else if k < 0 {
		scaledX = x.shiftLeft(uint64(-k))
	}

	index := 8
	step := 4

	for step > 0 {
		if scaledX.ult(lnBounds[index]) {
			index -= step
		} else {
			index += step
		}
		step /= 2
	}

	if scaledX.ult(lnBounds[index]) {
		index -= 1
	}

	res := scaledX.chebyPoly(lnChebyCoeffs[index])

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did above. Note that k*ln2 must strictly lie between minLn and maxLn constants
	// and we chose ln2Multiple and ln2Factor so they can't overflow when multiplied
	// by values in that range.
	ln2_192 := fix192{0x92c7, 0x957dcc1d0e60ef10, 0x1f17e2103111cbb2}

	powerCorrection := ln2_192.intMul(k)

	res = res.add(powerCorrection)

	return res, nil
}

func (x fix192) ln() (fix192, error) {
	if x.isZero() {
		return fix192{}, ErrDomain
	}

	// The Taylor expansion of ln(x) takes powers of the ratio z = (x - 1) / (x + 1). We can use the
	// "chebyMul" method for multiplication if we ensure that z is positive, and the series will
	// converge more quickly for smaller values. (The value of (x - 1) / (x + 1) is always < 1, but
	// closer to zero is still better than farther.)
	//
	// At the same time, ln(x•2^k) = ln(x) + ln(2)•k. We can scale out powers of two from the input
	// just by shifting. If we scale x so that it is between 1 and 2, we will have a value of z
	// between 0 and 1/3. We can do this very efficiently just by counting the number of leading
	// zero bits in x, and shifting it so that it has the same number of leading zero bits as the
	// fixed point representation of two. We can then compare the value against two, and if it is
	// higher, we can shift it one more bit over. This will always give us a value between 1 and 2.

	// We use the value k to keep track of which power of 2 we used to scale the input. Negative
	// values mean we scaled the input UP, and need to subtract a multiple of ln(2) at the end.
	var scaledX fix192

	k := (Fix128OneLeadingZeros - 1) - int64(leadingZeroBits192(x))

	if k >= 0 {
		scaledX = x.ushiftRight(uint64(k))
	} else if k < 0 {
		scaledX = x.shiftLeft(uint64(-k))
	}

	fix192One := fix192{0x000000000000d3c2, 0x1bcecceda1000000, 0x0000000000000000}
	fix192Two := fix192{0x000000000001a784, 0x379d99db42000000, 0x0000000000000000}

	if !scaledX.ult(fix192Two) {
		// If the scaled value is greater than or equal to two, we need to shift it one more bit to the right
		scaledX = scaledX.ushiftRight(1)
		k += 1
	}

	// We now compute the natural log of the scaled value using the expansion:
	// ln(x) = 2 * (z + z^3/3 + z^5/5 + z^7/7 + ...)
	// where z = (x - 1) / (x + 1)
	num := scaledX.sub(fix192One) // x - 1
	den := scaledX.add(fix192One) // x + 1
	dInv, _ := den.inverse()      // 1 / (x + 1)

	z, _ := num.umul(dInv) // z = (x - 1) / (x + 1)

	// Precompute z^2 to avoid recomputing it in the loop.
	z2, _ := z.umul(z)

	term := z
	sum := z
	iter := uint64(1)

	// Keep iterating until "term" and/or "next" rounds to zero
	for {
		term, _ = term.umul(z2)

		next, _ := div256by64(0, term.Hi, term.Mid, term.Lo, raw64(iter*2+1))

		if next.isZero() {
			break
		}

		sum = sum.add(next)
		iter += 1
	}

	res := sum.shiftLeft(1)

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did above. Note that k*ln2 must strictly lie between minLn and maxLn constants
	// and we chose ln2Multiple and ln2Factor so they can't overflow when multiplied
	// by values in that range.
	ln2_192 := fix192{0x92c7, 0x957dcc1d0e60ef10, 0x1f17e2103111cbb2}
	powerCorrection := ln2_192.intMul(k)

	res = res.add(powerCorrection)

	return res, nil
}

func (x fix192) exp() (fix192, error) {
	xUnsigned, sign := x.abs()

	// We compute exp(x) by using the identity:
	//     exp(x) = exp(i + f) = exp(i) * exp(f)
	// where i is the integer part and f is the fractional part of x.
	//
	// The easist way to do this would be to divide x by the value of one in fix192, using the
	// quotient as the integer part and the remainder as the fractional part. However, we don't have
	// a 192x192 division, and the value of one in fix192 extends over all 3 words. However, we can
	// use the fact that the value of one in fix192 is 10**24 * 2**64, which is equivalent to
	// 5^24 * 2^24 * 2^64. If we divide both the numerator and denominator by the same
	// value, the quotient will be the same, and the remainder will be scaled down by that value.
	//
	// So, we scale x by 2^24 * 2^64, which is equivalent to dropping the last word, and shifting the result
	// by 24. We then divide the result by 5^24, which is a 64-bit value.
	xTop := raw128{xUnsigned.Hi, xUnsigned.Mid}
	xTop = ushiftRight128(xTop, 24)
	i, rem := div64(xTop.Hi, xTop.Lo, raw64(0xd3c21bcecceda1)) // 5**24

	// Our remainder is now the fractional part, but, it's been scaled down by 2^24•2^64, AND we are missing
	// the bits that got shifted out. However, because those bits are zero in the denominator, we can just
	// copy those bits from the original input into the fractional part.
	f := fix192{rem >> 40, rem<<24 | xUnsigned.Mid&0xffffff, xUnsigned.Lo}
	fIsNonZero := !f.isZero()

	// We now have the integer part, i, and the fractional part, f of abs(x). If sign is negative, we need to
	// flip the sign of the result. Additionally, if the fractional part is non-zero, we subtract it from one
	// so that f is added to i, not subtracted from it.
	fix192One := fix192{0x000000000000d3c2, 0x1bcecceda1000000, 0x0000000000000000}

	if sign < 0 {
		i = -i
		if fIsNonZero {
			i = i - 1
			f = fix192One.sub(f)
		}
	}

	intPowerIndex := int64(i) - smallestExpIntPower

	if intPowerIndex < 0 {
		return fix192{}, ErrUnderflow
	} else if intPowerIndex >= int64(len(expIntPowers)) {
		return fix192{}, ErrOverflow
	}

	res := expIntPowers[intPowerIndex]
	var err error

	if fIsNonZero {
		res, err = res.umul(f.chebyPoly(expChebyCoeffs))
	}

	return res, err
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

	halfPi192 := fix192{0x14ca1, 0x0a1cd055eb965859, 0xb10f4d810eb09a8d}
	// threeHalfPi192 := fix192{0x3e5e3, 0x1e567101c2c3090d, 0x132de8832c11cfa8}

	if clampedX.ult(halfPi192) {
		// cos(x) = sin(π/2 - x)
		y = halfPi192.sub(clampedX)
	} else {
		// cos(x) = -sin(x - π/2)
		y = clampedX.sub(halfPi192)
		sign *= -1
	}

	res := y.chebyPoly(sinChebyCoeffs)

	return res.applySign(sign)
}

// func (accum fix192) chebyMul(x fix192) fix192 {
// 	a, aSign := accum.abs()

// 	var res fix192
// 	var carry1, carry2 uint64
// 	var r1, r2, r3 fix192
// 	var xhi raw64
// 	var round1, round2 raw64

// 	r1.Mid, r1.Lo, round1, _ = mul192by64_new(a, x.Lo)

// 	// midHi, _ := mul64(a.Mid, x.Lo)
// 	// hiHi, hiLo := mul64(a.Hi, x.Lo)

// 	// r1.Lo, carry1 = add64(midHi, hiLo, 0)
// 	// r1.Mid, _ = add64(hiHi, raw64Zero, carry1)

// 	r2.Hi, r2.Mid, r2.Lo, round2 = mul192by64_new(a, x.Mid)
// 	xhi, r3.Hi, r3.Mid, r3.Lo = mul192by64_new(a, x.Hi)

// 	_, carry1 = add64(round1, round2, 0)

// 	res, carry1 = add192(r1, r2, carry1)
// 	res, carry2 = add192(res, r3, 0)
// 	xhi, _ = add64(xhi, raw64(carry1), carry2)

// 	if uint64(xhi) >= (1 << 0) {
// 		panic("xhi overflowed in chebyMul")
// 	}

// 	// res = res.ushiftRight(54)
// 	// res.Hi |= xhi << 10

// 	res, err := res.applySign(aSign)

// 	if err != nil {
// 		panic("chebyMul: " + err.Error())
// 	}

// 	return res
// }

func (x fix192) sshiftRightWithRounding(bits uint64) fix192 {
	// We start by adding 0.5 (or -0.5, in the case of negative numbers) to the
	// value before shifting. This will ensure that the result is rounded
	// correctly when we shift right.
	signSmear := sshiftRight64(x.Hi, 63) // All 1s when x is negative, all 0s when x is positive.
	roundingValue := fix192{signSmear, signSmear, (signSmear | 1) << (bits - 1)}
	res := x.add(roundingValue).sshiftRight(bits)

	return res
}

// func (x fix192) chebyPoly(coeffs []fix192) fix192 {
// 	// Compute the Chebyshev polynomial using Horner's method.
// 	accum := coeffs[0]

// 	// x = x.shiftLeft(30)
// 	// xPrescaled, _ := div256by64(x.Hi, x.Mid, x.Lo, 0, raw64(0xd3c21bcecceda1))

// 	for i := 1; i < len(coeffs); i++ {
// 		accum = accum.chebyMul(x)
// 		accum = accum.add(coeffs[i])
// 	}

// 	return accum
// }

// func (x fix192) sin() (fix192, error) {
// 	clampedX, sign := x.clampAngle()

// 	halfPi192 := fix192{0x14ca1, 0x0a1cd055eb965859, 0xb10f4d810eb09a8d}
// 	pi192 := fix192{0x29942, 0x1439a0abd72cb0b3, 0x621e9b021d61351a}
// 	// twoPi192 := fix192{0x53284, 0x28734157ae596166, 0xc43d36043ac26a35}

// 	// Leverage the identity sin(x) = sin(π - x) to keep the input angle in the range [0, π/2]
// 	if halfPi192.ult(clampedX) {
// 		clampedX = pi192.sub(clampedX)
// 	}

// 	res := clampedX.chebyPoly(sinChebyCoeffs)
// 	// res = res.sshiftRightWithRounding(20)

// 	return res.applySign(sign)
// }

func (accum fix192) chebyMul(x fix192) fix192 {
	a, aSign := accum.abs()

	var res fix192
	var carry1, carry2 uint64
	var r1, r2, r3 fix192
	var xhi raw64
	var round1, round2 raw64

	r1.Mid, r1.Lo, round1, _ = mul192by64_new(a, x.Lo)

	// midHi, _ := mul64(a.Mid, x.Lo)
	// hiHi, hiLo := mul64(a.Hi, x.Lo)

	// r1.Lo, carry1 = add64(midHi, hiLo, 0)
	// r1.Mid, _ = add64(hiHi, raw64Zero, carry1)

	r2.Hi, r2.Mid, r2.Lo, round2 = mul192by64_new(a, x.Mid)
	xhi, r3.Hi, r3.Mid, r3.Lo = mul192by64_new(a, x.Hi)

	_, carry1 = add64(round1, round2, 0)

	res, carry1 = add192(r1, r2, carry1)
	res, carry2 = add192(res, r3, 0)
	xhi, _ = add64(xhi, raw64(carry1), carry2)

	if uint64(xhi) >= (1 << 48) {
		panic("xhi overflowed in chebyMul")
	}

	res = res.ushiftRight(16)
	res.Hi |= xhi << 48

	res, err := res.applySign(aSign)

	if err != nil {
		panic("chebyMul: " + err.Error())
	}

	return res
}

func (x fix192) chebyPoly(coeffs []fix192) fix192 {
	// Compute the Chebyshev polynomial using Horner's method.
	accum := coeffs[0]

	// x = x.shiftLeft(30)
	// xPrescaled, _ := div256by64(x.Hi, x.Mid, x.Lo, 0, raw64(0xd3c21bcecceda1))

	for i := 1; i < len(coeffs); i++ {
		accum = accum.chebyMul(x)
		accum = accum.add(coeffs[i])
	}

	return accum
}

func (x fix192) sin() (fix192, error) {
	clampedX, sign := x.clampAngle()

	halfPi192 := fix192{0x14ca1, 0x0a1cd055eb965859, 0xb10f4d810eb09a8d}
	pi192 := fix192{0x29942, 0x1439a0abd72cb0b3, 0x621e9b021d61351a}
	// twoPi192 := fix192{0x53284, 0x28734157ae596166, 0xc43d36043ac26a35}

	// Leverage the identity sin(x) = sin(π - x) to keep the input angle in the range [0, π/2]
	if halfPi192.ult(clampedX) {
		clampedX = pi192.sub(clampedX)
	}

	res := clampedX.chebyPoly(sinChebyCoeffs)
	// res = res.sshiftRightWithRounding(20)

	return res.applySign(sign)
}

func (x fix192) clampAngle() (res fix192, sign int64) {
	var xUnsigned fix192
	xUnsigned, sign = x.abs()

	pi192 := fix192{0x29942, 0x1439a0abd72cb0b3, 0x621e9b021d61351a}
	twoPi192 := fix192{0x53284, 0x28734157ae596166, 0xc43d36043ac26a35}
	twoPiErr := raw64(4886854831257614257)

	if xUnsigned.ult(pi192) {
		// If the input is already less than π, we can just return it as is.
		return xUnsigned, sign
	} else if twoPi192.ult(xUnsigned) {
		// If the input is larger than 2π, we we want to find the angle modulo 2π. The "easiest" way
		// to do this is to do a 192x192 division by 2π, and to look at the remainder. However
		// implementing that division is non-trivial and (somewhat surprisingly!) we don't need
		// 192x192 division for any other operations. Given that we know the divisor is _always_ 2π,
		// we can craft a speical modulus operation that is quite efficient.
		//
		// We start with this multiple of 2π that fits in 64-bits. It is chosen to be a large
		// multiple of 2π that fits in 64-bits AND has a fractional part that is very small (less
		// that 1e-10 in this case). This reduces the error we have when doing the division.
		twoPiMultiple := raw64(0xe1573806bdfd57db)

		// We now take the input and divide it by the multiple of 2π. This gives us a strong
		// estimate of the number of full 2π cycles in the input angle. (Note that this division
		// operation requires that the result fits in 128-bits. This will always be true if num.Hi <
		// den. We know that this is true because the input to this function was signed, and our
		// twoPiMultiple is larger than first 64-bits of the largest possible signed value.)
		q, _ := div192by64(xUnsigned.Hi, xUnsigned.Mid, xUnsigned.Lo, twoPiMultiple)

		// Our input, x, has a factor of 2**64 * 10**24, and our twoPiMultiple has a factor of
		// 21264757054. This means that our quotient is too large by a factor of 2**64 * 10**24 /
		// 21264757054. It seems like it might be daunting to divide by that ugly fraction, but
		// multiplying by the reciprocal is equivalent... and completely trivial! We multiply by
		// 21264757054 and then divide by 2**64 * 10**24, the former is just dropping the last
		// 64-bit word, and the latter is just shifting the result right by 24 bits and then dividing
		// by 5**24 (which fits into 64 bits).
		prodHi, prodLo := mul128By64(q, raw64(25842797545))
		temp := raw128{prodHi.Lo, prodLo.Hi}
		temp = ushiftRight128(temp, 16)
		realQuo, _ := div64(temp.Hi, temp.Lo, raw64(0x2386f26fc1)) // 5**16

		// This quotient is either exactly right (which it will be 99% of the time), or it could
		// be off by one. We take the input and subtract a bunch of 2π multiples from it and see
		// if the result is negative.
		twoPiProd := twoPi192.uintMul(uint64(realQuo))
		errorTerm, _ := mul64(twoPiErr, realQuo)
		twoPiProd = twoPiProd.add(fix192{0, 0, errorTerm})
		res = xUnsigned.sub(twoPiProd)

		if isNeg64(res.Hi) {
			// If the result is negative, we overshot by one 2π multiple, so we add back one 2π
			// multiple to get the correct remainder.
			var carry uint64
			res, carry = add192(res, twoPi192, 0)

			if carry != 0 {
				panic("clampAngle: carry should never happen when adding twoPiMultiple")
			}
		}
	} else {
		res = xUnsigned
	}

	// res is now scaled down below 2π. If the angle is greater than π, subtract it from 2π to bring it
	// into the range [0, π] and flip the sign flag.
	if pi192.ult(res) {
		res = twoPi192.sub(res)
		sign *= -1
	}

	return res, sign
}

// func (x fix192) branchless_abs() (fix192, int64) {
// 	// This is a branchless version of abs() that uses bitwise operations to avoid branching.
// 	// It returns the absolute value of x and a sign multiplier.
// 	//
// 	// If x is negative, it returns -x and -1, otherwise it returns x and 1.
// 	//
// 	// This is useful for performance-sensitive code where branching can be expensive.

// 	mask := raw64(x.Hi) >> 63
// 	sign := 1 - 2*int64(mask) // sign will be -1 if x is negative, 1 if x is positive
// 	res := fix192{
// 		x.Hi ^ mask,
// 		x.Mid ^ mask,
// 		x.Lo ^ mask,
// 	}.add(fix192{0, 0, mask & 1}) // Add 1 to the low part if x is negative

// 	return res, sign
// }

func (x fix192) abs() (fix192, int64) {
	if isNeg64(x.Hi) {
		return x.neg(), -1
	}

	return x, 1
}

func (x fix192) neg() fix192 {
	return fix192{}.sub(x)
}

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
			return fix192{}, ErrNegOverflow
		}
	} else if isNeg64(x.Hi) {
		// If the input reads as negative but wasn't sign flipped, then the input
		// is too big to represent as a signed value.
		return fix192{}, ErrOverflow
	}

	return x, nil
}

func (a fix192) add(b fix192) (res fix192) {
	var carry uint64

	res.Lo, carry = add64(a.Lo, b.Lo, 0)
	res.Mid, carry = add64(a.Mid, b.Mid, carry)
	res.Hi, _ = add64(a.Hi, b.Hi, carry)

	return res
}

func (a fix192) sub(b fix192) (res fix192) {
	var borrow uint64

	res.Lo, borrow = sub64(a.Lo, b.Lo, 0)
	res.Mid, borrow = sub64(a.Mid, b.Mid, borrow)
	res.Hi, _ = sub64(a.Hi, b.Hi, borrow)

	return res
}

func (a fix192) ult(b fix192) bool {
	// It turns out that branchless subtraction (with borrow) is faster than a branching comparison.
	// if a < b, then a-b will be negative, i.e. have a borrow.
	_, borrow := sub64(a.Lo, b.Lo, 0)
	_, borrow = sub64(a.Mid, b.Mid, borrow)
	_, borrow = sub64(a.Hi, b.Hi, borrow)

	return borrow != 0
}

// Perform integer multiplication of a fix192 value by a uint64 value, treating a as an unsigned integer.
// Does NOT handle overflow, so only use internally where overflow can't happen.
func (a fix192) uintMul(b uint64) fix192 {
	t1, t2 := mul128By64(raw128{a.Mid, a.Lo}, raw64(b))
	_, t4 := mul64(a.Hi, raw64(b))

	sum, _ := add64(t1.Lo, t4, 0)

	return fix192{sum, t2.Hi, t2.Lo}
}

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

func (x UFix128) toFix192() fix192 {
	return fix192{x.Hi, x.Lo, raw64(0)}
}

func (x Fix128) toFix192() fix192 {
	return fix192{x.Hi, x.Lo, raw64(0)}
}

func (x fix192) toUFix128() (UFix128, error) {
	carry := raw64(0)

	if !ult64(x.Lo, 0x8000000000000000) {
		carry = 1
	}

	if isZero64(x.Hi) && isZero64(x.Mid) && !isZero64(x.Lo) && carry == 0 {
		// If the high and mid parts are zero, and the low part is non-zero but less than 0.5, we
		// flag underflow
		return UFix128Zero, ErrUnderflow
	}

	return UFix128{x.Hi, x.Mid}.Add(UFix128{0, carry})
}

func (x fix192) toFix128() (Fix128, error) {
	unsignedX, sign := x.abs()

	unsignedRes, err := unsignedX.toUFix128()

	if err != nil {
		return Fix128Zero, err
	}

	return unsignedRes.ApplySign(sign)
}

func (a fix192) isZero() bool {
	return isZero64(a.Hi) && isZero64(a.Mid) && isZero64(a.Lo)
}

func add192(a, b fix192, carryIn uint64) (res fix192, carryOut uint64) {
	res.Lo, carryOut = add64(a.Lo, b.Lo, carryIn)
	res.Mid, carryOut = add64(a.Mid, b.Mid, carryOut)
	res.Hi, carryOut = add64(a.Hi, b.Hi, carryOut)

	return
}

// func (a UFix128) Mul192(b UFix128) (UFix128, error) {
// 	a192 := UFix128(a).toFix192()
// 	b192 := UFix128(b).toFix192()
// 	r192, err := a192.umul(b192)

// 	if err != nil {
// 		return UFix128Zero, err
// 	}

// 	if !a.IsZero() && !b.IsZero() && r192.isZero() {
// 		// fixed 192 multiplication doesn't check for underflow, so we do it here
// 		return UFix128Zero, ErrUnderflow
// 	}

// 	return r192.toUFix128()
// }

// // A faster helper for umul() for when a fits in 128 bits and b fits in 64 bits
// func (a fix192) umulSmall(b fix192) (fix192, error) {
// 	// The multiplication can be a simple 128x64 multiplication, and we know the result will fit
// 	// in 64 bits because we are multiplying two fractions, so the product will be less than
// 	// either input.
// 	hi, lo := mul128By64(raw128{a.Mid, a.Lo}, b.Lo)

// 	// We have to drop the low 64 bits, and then divide by 10**24
// 	temp := raw128{hi.Lo, lo.Hi}
// 	temp = ushiftRight128(temp, 24)                            // equivalent to dividing by 2**24
// 	res, _ := div64(temp.Hi, temp.Lo, raw64(0xd3c21bcecceda1)) // dividing by 5**24

// 	return fix192{0, 0, res}, nil
// }

func (a fix192) umul(b fix192) (fix192, error) {
	// The basic logic here is the same as the logic in mul128(), so check that code for more
	// details. Each section below computes terms of "one row" of the long-form multiplicaiton that
	// you would do by hand. We then add all of these results together at the end. We compute the
	// rows into their own variables (instead of using) some kind of accumulator pattern) in the
	// hopes that the compiler/CPU can pipeline and/or parallelize the operations better.
	//
	// The overal multiplication is a 192x192 multiplication that would produce a 384-bit result,
	// except that we immediately throw out the least significant 64 bits of the result because
	// we need to divide the result by 2**64 at the end anyway. We collect the remaining 320 bits
	// into two 192-bit values, one that holds the bottom 192 bits of the result, and one that
	// holds the top 128 bits of the result.
	var c uint64

	// Row 1 only has values in the low part of the result
	var r1Lo fix192

	// We're multiplying the entire 192-bit a value by the low bits of b. This is a 192x64
	// multiplication that produces a 256-bit result. However, we discard the low 64-bits
	// since we are multiplying two values that are each scaled by a factor of 2**64 and would
	// just need to divide by 2**64 at the end to get the right result.
	r1_t1, r1_t2 := mul128By64(raw128{a.Mid, a.Lo}, b.Lo)
	r1_t3, r1_t4 := mul64(a.Hi, b.Lo)

	r1Lo.Lo = r1_t2.Hi
	r1Lo.Mid, c = add64(r1_t1.Lo, r1_t4, 0)
	r1Lo.Hi, _ = add64(r1_t3, 0, c)

	// See if we need to round up the low part of the result by adding 1 if the low part is >= 0.5
	_, carry := add64(r1_t2.Lo, 0x8000000000000000, 0)

	// Row 2 contributes to the top and bottom parts of the result
	var r2Hi, r2Lo fix192

	r2_t1, r2_t2 := mul128By64(raw128{a.Mid, a.Lo}, b.Mid)
	r2_t3, r2_t4 := mul64(a.Hi, b.Mid)

	r2Lo.Lo = r2_t2.Lo
	r2Lo.Mid = r2_t2.Hi
	r2Lo.Hi, c = add64(r2_t1.Lo, r2_t4, 0)
	r2Hi.Lo, _ = add64(r2_t3, 0, c)

	// Row 3 contributes to the top and bottom parts of the result
	var r3Hi, r3Lo fix192

	r3_t1, r3_t2 := mul128By64(raw128{a.Mid, a.Lo}, b.Hi)
	r3_t3, r3_t4 := mul64(a.Hi, b.Hi)

	r3Lo.Mid = r3_t2.Lo
	r3Lo.Hi = r3_t2.Hi
	r3Hi.Lo, c = add64(r3_t1.Lo, r3_t4, 0)
	r3Hi.Mid, _ = add64(r3_t3, 0, c)

	// Add up the three rows to get an interim result (that still needs to be scaled down by 10**24)
	var interimLo, interimHi fix192
	var carry1, carry2 uint64
	interimLo, carry1 = add192(r1Lo, r2Lo, carry)
	interimLo, carry2 = add192(interimLo, r3Lo, 0)

	interimHi, _ = add192(r2Hi, r3Hi, carry1)
	interimHi, _ = add192(interimHi, fix192{}, carry2)

	// We now need to scale this result down by 10**24, which either fits into 192 bits, or
	// overflows the type. We do this in two steps:
	// 1. We shift everything down by 24 bits. This is equivalent to dividing the result by 2**24,
	//    which is leaves us still needing us to divide by 5**24. However, this is enough to
	//    shrink the divisor to fit in 64-bits, dramatically simplifying the division. We can also
	//    check for overflow after the shift since the result will only fit in 192 if the part that
	//    extends beyond the bottom 192 bits is less than 5**24 before the division.
	// 2. After shifting down and checking for overflow, we divide the result by 5**24 and return.

	// This check is here in case someone changes the Fix128Scale constant. It should compile to
	// a no-op if the constant is correct.
	if Fix128Scale != 1e24 {
		panic("fix192 assumes Fix128Scale equals 10e24")
	}

	interimLo = interimLo.ushiftRight(24)
	interimLo.Hi |= shiftLeft64(interimHi.Lo, 40)
	interimHi = interimHi.ushiftRight(24)

	fiveToThe24th := raw64(0xd3c21bcecceda1)

	// We know that interimHi.Hi is zero, since we never assigned any value to it, and the only
	// possible way it could be have values in it is from overflow during the addition. But that
	// was before we shifted right by 24 bits, which is much larger than any conceivable overflow
	// from addition! However, we could still have a value in interimHi.Mid, and we also need
	// interimHi.Lo to be less than fiveToThe24th, so that after the division, the result fits
	// in 192 bits.
	if !isZero64(interimHi.Mid) || !ult64(interimHi.Lo, fiveToThe24th) {
		return fix192{}, ErrOverflow
	}

	// quo, rem := magicDivision(interimHi.Lo, interimLo.Hi, interimLo.Mid, interimLo.Lo)
	quo, rem := div256by64(interimHi.Lo, interimLo.Hi, interimLo.Mid, interimLo.Lo, fiveToThe24th)

	// if quoT != quo && remT == rem {
	// 	panic("magicDivision returned different quotient than div256by64")
	// }

	if ushouldRound64(rem, fiveToThe24th) {
		// If the remainder is greater than or equal to 0.5, we round up.
		quo, _ = add192(quo, fix192{}, 1)
	}
	_ = rem

	return quo, nil
}

type raw256 struct {
	Hi, Lo raw128
}

func mul256by128(a raw256, b raw128) (hi, lo raw256) {
	var w, z raw128
	var carry uint64

	w, lo.Lo = mul128(a.Lo, b)
	hi.Lo, z = mul128(a.Hi, b)

	lo.Hi, carry = add128(w, z, 0)

	// Can't overflow, since that would imply a 128 x 64 multiplication
	// overflowed 192 bits, which is not possible.
	hi.Lo, _ = add128(hi.Lo, raw128Zero, carry)

	return hi, lo
}

func mul192by64_new(a fix192, b raw64) (xhi, hi, mid, lo raw64) {
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

func mul192by64(a raw256, b raw64) (res raw256) {
	var w raw128

	w, res.Lo = mul128By64(a.Lo, b)
	z1, z2 := mul64(a.Hi.Lo, b)

	res.Hi, _ = add128(w, raw128{z1, z2}, 0)

	return res
}

func add256(a, b raw256, carryIn uint64) (res raw256, carry uint64) {
	res.Lo, carry = add128(a.Lo, b.Lo, carryIn)
	res.Hi, _ = add128(a.Hi, b.Hi, carry)
	return res, carry
}

func mul256(a, b raw256) (hi, lo raw256) {
	var u, v1, v2 raw256
	var wHi raw128
	u.Hi, u.Lo = mul128(a.Hi, b.Hi)
	v1.Hi, v1.Lo = mul128(a.Hi, b.Lo)
	v2.Hi, v2.Lo = mul128(a.Lo, b.Hi)
	v, vCarry := add256(v1, v2, 0)
	wHi, lo.Lo = mul128(a.Lo, b.Lo)

	// The lowest word of the result (lo.Lo) was directly set when we computed w above

	// We now sum up lo.Hi, which is the low part of v plus the high part of w
	var midCarry, hiCarry uint64
	lo.Hi, midCarry = add128(v.Lo, wHi, 0)

	// The hi.Lo is the sum of the low part of u with the high part of v plus any carry
	// from the previous sum.
	hi.Lo, hiCarry = add128(u.Lo, v.Hi, midCarry)

	// hi.Hi is the high part of u plus any carry from the previous sum (and any carry from
	// computing v).
	hi.Hi, _ = add128(u.Hi, raw128{0, raw64(vCarry)}, hiCarry)

	return
}

// func magicDivision(xhi, hi, mid, lo raw64) (fix192, raw64) {
// 	magicConstant1 := raw128{0x135, 0x7c299a88ea76a589}

// 	quo := raw256{}
// 	input := raw256{raw128{xhi, hi}, raw128{mid, lo}}
// 	residual := input

// 	for {
// 		hi, lo := mul256by128(residual, magicConstant1)
// 		// We need to shift the result right by 64 bits, which is equivalent to dividing by 2**64.
// 		delta := raw256{hi.Lo, lo.Hi}
// 		quo, _ = add256(quo, delta)

// 		prod := mul192by64(quo, 0xd3c21bcecceda1)

// 		rLo, borrow := sub128(input.Lo, prod.Lo, 0)
// 		rHi, borrow := sub128(input.Hi, prod.Hi, borrow)

// 		if borrow != 0 {
// 			panic("magicDivision: borrow should never happen when subtracting pLo from residual")
// 		}

// 		if isZero128(rHi) && isZero64(rLo.Hi) {
// 			if ult64(rLo.Lo, 0xd3c21bcecceda1) {
// 				res := fix192{quo.Hi.Lo, quo.Lo.Hi, quo.Lo.Lo}
// 				return res, rLo.Lo
// 			} else if isEqual64(rLo.Lo, 0xd3c21bcecceda1) {
// 				res := fix192{quo.Hi.Lo, quo.Lo.Hi, quo.Lo.Lo}
// 				res = res.add(fix192{0, 0, 1})
// 				return res, 0
// 			}
// 		}

// 		residual = raw256{rHi, rLo}
// 	}
// }

func (a fix192) smul(b fix192) (fix192, error) {
	aUnsigned, aSign := a.abs()
	bUnsigned, bSign := b.abs()
	rSign := aSign * bSign

	resUnsigned, err := aUnsigned.umul(bUnsigned)

	if err == ErrOverflow {
		if rSign < 0 {
			return fix192{}, ErrNegOverflow
		} else {
			return fix192{}, ErrOverflow
		}
	} else if err != nil {
		return fix192{}, err
	}

	// Apply the sign to the result
	return resUnsigned.applySign(rSign)
}

func div256by64(xhi, hi, mid, lo raw64, y raw64) (quo fix192, rem raw64) {
	if xhi == 0 && hi < y {
		rem = hi
	} else {
		quo.Hi, rem = div64(xhi, hi, y)
	}

	if rem == 0 && mid < y {
		rem = mid
	} else {
		quo.Mid, rem = div64(rem, mid, y)
	}

	if rem == 0 && lo < y {
		rem = lo
	} else {
		quo.Lo, rem = div64(rem, lo, y)
	}

	return quo, rem
}

func (x fix192) shiftLeft(shift uint64) (res fix192) {
	if shift == 0 {
		return x
	}

	if shift >= 128 {
		shift -= 128

		res.Hi = shiftLeft64(x.Lo, shift)
		res.Mid = 0
		res.Lo = 0

		return res
	}

	if shift >= 64 {
		shift -= 64

		res.Hi = shiftLeft64(x.Mid, shift)
		res.Hi |= ushiftRight64(x.Lo, 64-shift)
		res.Mid = shiftLeft64(x.Lo, shift)
		res.Lo = 0

		return res
	}

	res.Hi = shiftLeft64(x.Hi, shift)
	res.Hi |= ushiftRight64(x.Mid, 64-shift)
	res.Mid = shiftLeft64(x.Mid, shift)
	res.Mid |= ushiftRight64(x.Lo, 64-shift)
	res.Lo = shiftLeft64(x.Lo, shift)

	return res
}

func (x fix192) ushiftRight(shift uint64) (res fix192) {
	if shift == 0 {
		return x
	}

	if shift >= 128 {
		shift -= 128

		res.Lo = ushiftRight64(x.Hi, shift)
		res.Mid = 0
		res.Hi = 0

		return res
	}

	if shift >= 64 {
		shift -= 64

		res.Lo = ushiftRight64(x.Mid, shift)
		res.Lo |= shiftLeft64(x.Hi, 64-shift)
		res.Mid = ushiftRight64(x.Hi, shift)
		res.Hi = 0

		return res
	}

	res.Lo = ushiftRight64(x.Lo, shift)
	res.Lo |= shiftLeft64(x.Mid, 64-shift)
	res.Mid = ushiftRight64(x.Mid, shift)
	res.Mid |= shiftLeft64(x.Hi, 64-shift)
	res.Hi = ushiftRight64(x.Hi, shift)

	return res
}

func (x fix192) sshiftRight(shift uint64) (res fix192) {
	if shift == 0 {
		return x
	}

	if shift >= 128 {
		shift -= 128

		res.Lo = sshiftRight64(x.Hi, shift)
		res.Mid = sshiftRight64(x.Hi, 63)
		res.Hi = sshiftRight64(x.Hi, 63)

		return res
	}

	if shift >= 64 {
		shift -= 64

		res.Lo = ushiftRight64(x.Mid, shift)
		res.Lo |= shiftLeft64(x.Hi, 64-shift)
		res.Mid = sshiftRight64(x.Hi, shift)
		res.Hi = sshiftRight64(x.Hi, 63)

		return res
	}

	res.Lo = ushiftRight64(x.Lo, shift)
	res.Lo |= shiftLeft64(x.Mid, 64-shift)
	res.Mid = ushiftRight64(x.Mid, shift)
	res.Mid |= shiftLeft64(x.Hi, 64-shift)
	res.Hi = sshiftRight64(x.Hi, shift)

	return res
}

func (x UFix64) toFix192_old() fix192_old {
	// Convert Fix64 to fix192 by extracting the integer and fractional parts.
	intPart := uint64(x) / uint64(Fix64Scale)
	fracPart := uint64(x) % uint64(Fix64Scale)

	fracPart128, rem := div192by128(raw64(fracPart), 0, 0, raw128{0, raw64(Fix64Scale)})

	if rem.Hi >= 0x8000000000000000 {
		// If the remainder is greater than or equal to 0.5, we round up
		// (Can't overflow since we are working at much higher precision than the input.)
		fracPart128, _ = add128(fracPart128, raw128Zero, 1)
	}

	return fix192_old{i: raw64(intPart), f: fracPart128}
}

func (x Fix64) toFix192_old() fix192_old {
	xUnsigned, sign := x.Abs()
	res := xUnsigned.toFix192_old()
	res, _ = res.applySign(sign) // can't fail, input is well within the range of fix192
	return res
}

func (x UFix128) toFix192_old() fix192_old {
	// Convert UFix128 to fix192 by extracting the integer and fractional parts.
	intPart, fracPart := div128(raw128Zero, raw128(x), raw128(Fix128One))
	fracPart128, rem := div128(fracPart, raw128Zero, raw128(Fix128One))

	if ushouldRound128(rem, raw128(Fix128One)) {
		// If the remainder is greater than or equal to 0.5, we round up
		// (Can't overflow since we are working at much higher precision than the input.)
		fracPart128, _ = add128(fracPart128, raw128Zero, 1)
	}

	return fix192_old{i: intPart.Lo, f: fracPart128}
}

func (x Fix128) toFix192_old() fix192_old {
	xUnsigned, sign := x.Abs()
	res := xUnsigned.toFix192_old()
	res, _ = res.applySign(sign) // can't fail, input is well within the range of fix192
	return res
}

func (a fix192_old) toUFix64() (UFix64, error) {
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

func (x fix192_old) toFix64() (Fix64, error) {
	xUnsigned, sign := x.abs()
	res, err := xUnsigned.toUFix64()

	if err != nil {
		return Fix64Zero, err
	}

	return res.ApplySign(sign)
}

func (a fix192_old) toUFix128() (UFix128, error) {
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

	// var HundredTrillion = raw128{0x4b3b4ca85a86c47a, 0x098a224000000000}

	// if !ult128(res, HundredTrillion) {
	// 	// 2^64 % 10 = 6, so we multiply the modulus of the high part by 6
	// 	lastDigit := ((res.Hi%10)*6 + (res.Lo % 10)) % 10

	// 	if lastDigit >= 5 {
	// 		// If the last digits are 50 or greater, we round up.
	// 		res, carry = add128(raw128(res), raw128{0, 10 - lastDigit}, 0)

	// 		if carry != 0 {
	// 			// If there was a carry, the result overflowed.
	// 			return UFix128Zero, ErrOverflow
	// 		}
	// 	} else {
	// 		// Round down
	// 		res, _ = sub128(raw128(res), raw128{0, lastDigit}, 0)
	// 	}
	// }

	return UFix128(res), nil
}

func (x fix192_old) toFix128() (Fix128, error) {
	xUnsigned, sign := x.abs()
	res, err := xUnsigned.toUFix128()

	if err != nil {
		return Fix128Zero, err
	}

	return res.ApplySign(sign)
}

func (a fix192_old) lt(b fix192_old) bool {
	if isEqual64(a.i, b.i) {
		return ult128(a.f, b.f)
	} else {
		return slt64(a.i, b.i)
	}
}

func (a fix192_old) eq(b fix192_old) bool {
	return isEqual64(a.i, b.i) && isEqual128(a.f, b.f)
}

func (x fix192_old) applySign(sign int64) (fix192_old, error) {
	if isZero64(x.i) && isZero128(x.f) {
		// If the input is zero, we can return it as is, regardless of the sign.
		return x, nil
	}

	if sign < 0 {
		x = x.neg()

		// If trying to make it negative didn't make it negative, the input is too
		// large to represent as a negative value.
		if !isNeg64(x.i) {
			return fix192_old{}, ErrNegOverflow
		}
	} else if isNeg64(x.i) {
		// If the input looks negative to start with, it's too big to represent
		return fix192_old{}, ErrOverflow
	}

	return x, nil
}

func (x fix192_old) abs() (fix192_old, int64) {
	if isNeg64(x.i) {
		return x.neg(), -1
	}

	return x, 1
}

func (x fix192_old) neg() fix192_old {
	x.i = neg64(x.i)

	if !isZero128(x.f) {
		// If the fractional part is non-zero, we subtract one from the integer part
		// and flip the sign of the fractional part by subtracting it from zero.
		x.i, _ = sub64(x.i, raw64Zero, 1)
		x.f, _ = sub128(raw128Zero, x.f, 0)
	}

	return x
}

func (x fix192_old) shiftLeft(shift uint64) (res fix192_old) {
	if shift >= 64 {
		panic("fix192 only supports shifts less than 64")
	}

	res.i = shiftLeft64(x.i, shift) | ushiftRight64(x.f.Hi, 64-shift)
	res.f = shiftLeft128(x.f, shift)

	return res
}

func (x fix192_old) shiftRight(shift uint64) (res fix192_old) {
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

func (a fix192_old) add(b fix192_old) (res fix192_old) {
	var carry uint64

	res.f, carry = add128(a.f, b.f, 0)
	res.i, _ = add64(a.i, b.i, carry)

	return res
}

func (a fix192_old) sub(b fix192_old) (res fix192_old) {
	var borrow uint64

	res.f, borrow = sub128(a.f, b.f, 0)
	res.i, _ = sub64(a.i, b.i, borrow)

	return res
}

func (a fix192_old) umulFull(b fix192_old) (fix192_old, error) {
	// Compute the integer part of term one
	hi, i1 := mul64(a.i, b.i)
	if !isZero64(hi) {
		return fix192_old{}, ErrOverflow
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

	var res fix192_old
	var carry1, carry2 uint64

	// Sum up the fractional parts, holding on to the carries
	res.f, carry1 = add128(f2, f3, 0)
	res.f, carry2 = add128(res.f, f4, 0)

	// Sum up the integer parts, using the carries from the fractional parts
	res.i, carry1 = add64(i1, raw64(i2.Lo), carry1)
	res.i, carry2 = add64(res.i, raw64(i3.Lo), carry2)

	// If either add operation produced a carry, we have an overflow.
	if carry1 != 0 || carry2 != 0 {
		return fix192_old{}, ErrOverflow
	}

	return res, nil
}

func (a fix192_old) umul(b fix192_old) (fix192_old, error) {
	// a = a.i + a.f, b = b.i + b.f
	// a•b = (a.i + a.f)•(b.i + b.f)
	//      = a.i•b.i + a.i•b.f + a.f•b.i + a.f•b.f
	if isZero64(a.i) {
		if isZero64(b.i) {
			return fix192_old{0, a.f.mulFrac(b.f)}, nil
		} else {
			return b.mulByFraction(a.f), nil
		}
	} else if isZero64(b.i) {
		return a.mulByFraction(b.f), nil
	} else {
		return a.umulFull(b)
	}
}

func (a fix192_old) smul(b fix192_old) (fix192_old, error) {
	aUnsigned, aSign := a.abs()
	bUnsigned, bSign := b.abs()
	rSign := aSign * bSign

	resUnsigned, err := aUnsigned.umul(bUnsigned)

	if err == ErrOverflow {
		if rSign < 0 {
			return fix192_old{}, ErrNegOverflow
		} else {
			return fix192_old{}, ErrOverflow
		}
	} else if err != nil {
		return fix192_old{}, err
	}

	// Apply the sign to the result
	return resUnsigned.applySign(rSign)
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
func (a fix192_old) squared() (fix192_old, error) {
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
		return fix192_old{}, ErrOverflow
	}

	// Double the product of the integer and fractional parts
	ifProductHi, ifProductLo := mul128By64(a.f, shiftLeft64(a.i, 1))

	// Square the fractional part
	fSquared := a.f.mulFrac(a.f)

	var res fix192_old
	var carry uint64

	res.f, carry = add128(fSquared, ifProductLo, 0)
	res.i, carry = add64(iSquared, raw64(ifProductHi.Lo), carry)

	if carry != 0 {
		return fix192_old{}, ErrOverflow
	}

	return res, nil
}

func (x fix192_old) ln(prescale uint64) (fix192_old, error) {
	if isZero64(x.i) && isZero128(x.f) {
		return fix192_old{}, ErrDomain
	}

	if isNeg64(x.i) {
		return fix192_old{}, ErrDomain
	}

	// The Taylor expansion of ln(x) takes powers of the ratio (x - 1) / (x + 1). Since
	// we know we can efficiently multiply values between 0-1, we start by scaling X
	// by a power of two to be between 1-2. We can do this very efficiently just by
	// counting the number of leading zero bits in x. If x is less than 1, we shift the
	// input to the left until it is between 1 and 2, and if x is >= 2, we shift to the right.

	// We use the value k to keep track of which power of 2 we used to scale the input. Negative
	// values mean we scaled the input UP, and need to subtract a multiple of ln(2) at the end.
	var scaledX fix192_old
	var k uint64
	scaledUp := false

	if isZero64(x.i) {
		// If the integer part is zero, we can only have a fractional part. See how much
		// we need to scale it up by counting the zero bits, and adding one (so the top
		// set bit "flows over" into the integer part)
		fracScale := leadingZeroBits128(x.f) + 1

		scaledX = fix192_old{1, shiftLeft128(x.f, fracScale)}
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
	num := fix192_old{0, scaledX.f}
	den := fix192_old{2, scaledX.f}

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
	res := fix192_old{0, sum}
	res = res.shiftLeft(1)

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did above. Note that k*ln2 must strictly lie between minLn and maxLn constants
	// and we chose ln2Multiple and ln2Factor so they can't overflow when multiplied
	// by values in that range.
	ln2_192 := fix192_old{0, raw128{0xb17217f7d1cf79ab, 0xc9e3b39803f2f6af}}
	powerCorrection, _ := ln2_192.umul(fix192_old{raw64(k), raw128Zero})

	if scaledUp {
		res = res.sub(powerCorrection)
	} else {
		res = res.add(powerCorrection)
	}

	return res, nil
}

func (x fix192_old) exp() (fix192_old, error) {

	powerIndex := int64(x.i) - smallestExpIntPower

	if powerIndex < 0 {
		return fix192_old{}, ErrUnderflow
	} else if powerIndex >= int64(len(expIntPowers)) {
		return fix192_old{}, ErrOverflow
	}

	intExp192 := fix192_old{} //expIntPowers[powerIndex]
	fracExp192 := x.f.fracExp()

	return intExp192.umul(fracExp192)
}

func (a fix192_old) pow(b fix192_old, prescale uint64) (fix192_old, error) {
	aBitLessThanOne := fix192_old{0, raw128{0xe000000000000000, 0}}
	aBitMoreThanOne := fix192_old{1, raw128{0x2000000000000000, 0}}

	// Use powNearOne() for values close to 1 if the exponent is large
	if slt64(raw64(10), b.i) && prescale == 0 && aBitLessThanOne.lt(a) && a.lt(aBitMoreThanOne) {
		return a.powNearOne(b)
	}

	aLn, err := a.ln(prescale)

	if err != nil {
		return fix192_old{}, err
	}

	prod, err := aLn.smul(b)

	if err != nil {
		return fix192_old{}, err
	}

	return prod.exp()
}

func (x raw128) fracExp() fix192_old {
	return x.chebyPoly([]fix192_old{})

	// We compute the fractional exp using the typical Taylor series:
	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...
	// However, we can make some simplifying assumptions that speed up the calculations:
	// 1. If the input is larger than ln(2), we divide the input by 2 and then square the result
	// 2. We ignore the first term in the expansion (the 1); all other terms are guaranteed to sum up
	//    to less than 1 (since the final result is <2 from the previous step), so we can JUST do our
	//    computations on the fractional part.
	// 3. Multiplication of two fractional 128-bit values is very easy, we just do a 128x128-> 256 bit
	//    multiplication, and just use the top 128 bits of the result.
	// term := x
	// sum := term
	// iter := uint64(1)
	// i := raw64(1)

	// for {
	// 	term = term.mulFrac(x)

	// 	iter++
	// 	term = uintDiv128(term, iter)

	// 	if isZero128(term) {
	// 		break
	// 	}

	// 	var carry uint64
	// 	sum, carry = add128(sum, term, 0)
	// 	i, _ = add64(i, raw64Zero, carry)
	// }

	// return fix192{i, sum}
}

func (x fix192_old) sin() (fix192_old, error) {
	clampedX, sign := x.clampAngle()

	res := clampedX.chebySin()

	return res.applySign(sign)
}

func (x fix192_old) cos() (fix192_old, error) {
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
	var y fix192_old

	if clampedX.lt(fix192HalfPi) {
		// cos(x) = sin(π/2 - x)
		y = fix192HalfPi.sub(clampedX)
	} else {
		// cos(x) = -sin(3π/2 − x)
		y = fix192ThreeHalfPi.sub(clampedX)
		sign *= -1
	}

	res := y.chebySin()

	return res.applySign(sign)
}

func (x fix192_old) tan() (fix192_old, error) {
	// tan(x) = sin(x) / cos(x)
	// We don't want to just call the sin() and cos() methods directly since we will
	// just double-up the call to clampAngle().

	// Normalize the input angle to the range [0, π]
	clampedX, sign := x.clampAngle()

	if clampedX.lt(fix192_old{0, raw128{0x1000000000000000, 0}}) {
		// If the value is less than 1/8, we can direcly use our chebyTan() calculation
		res := clampedX.f.chebyTan()
		return res.applySign(sign)
	}

	var y fix192_old

	if clampedX.eq(fix192HalfPi) {
		// In practice, this will probably never happen since the input types have too little precision
		// to _exactly equal_ π/2 at fix192 precision. However! We handle it just in case...
		return fix192_old{}, ErrOverflow
	} else if clampedX.lt(fix192HalfPi) {
		// This y value will be passed to sin() to compute cos(x), unless we are close enough to π/2
		// to use the chebyTan() method.
		y = fix192HalfPi.sub(clampedX)

		// See if x is close enough to π/2 that we can use the tan(π/2 - x) identity.
		if y.lt(fix192_old{0, raw128{0x1000000000000000, 0}}) {
			// tan(π/2 - x) = 1 / tan(x)
			// We compute tan(x) using the chebyTan() method, and then take the inverse.
			inverseTan := y.f.chebyTan()
			res, err := inverseTan.f.inverse()

			if err != nil {
				if sign < 0 {
					return fix192_old{}, ErrNegOverflow
				} else {
					return fix192_old{}, ErrOverflow
				}
			}

			return res.applySign(sign)
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
				res, err := inverseTan.f.inverse()

				if err != nil {
					if sign < 0 {
						return fix192_old{}, ErrNegOverflow
					} else {
						return fix192_old{}, ErrOverflow
					}
				}

				return res.applySign(sign)
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
		return sinX.applySign(sign)
	} else if isZero128(cosX.f) || isIota128(cosX.f) {
		// If cosX is zero or iota, we treat it as overflow
		if sign < 0 {
			return fix192_old{}, ErrNegOverflow
		} else {
			return fix192_old{}, ErrOverflow
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

		res := fix192_old{i: quoHi.Lo, f: quoLo}

		return res.applySign(sign)
	}
}

func (x fix192_old) clampAngle() (fix192_old, int64) {
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

	res := fix192_old{0, rem}
	res = res.shiftLeft(3)

	return res, sign
}

// Computes a Chebyshev polynomial at a particular x value. This method
// assumes that the input, coefficients AND result are all strictly less
// than 1.0, with the input and output also being positive.
// func (x raw128) chebyPolyOld(coeffs []coeff) raw128 {
// 	// Because we have very efficient math functions for fractional values,
// 	// we can't use Horner's method for polynomial expansion here. Some of the
// 	// interim values can end up outside the range (0, 1). However, if we do a more
// 	// naive implementation we have a few extra multiplications, but we know that
// 	// we'll never end up with an intermediate value outside of a simple fractional
// 	// value (with the functions we are using, anyway!)

// 	// Start with the constant term (usually zero)
// 	accum := coeffs[0].value

// 	// The current power of x we are computing. Starting with x^1.
// 	pow := x
// 	term := pow.mulFrac(coeffs[1].value)
// 	accum, _ = add128(accum, term, 0)

// 	for i := 2; i < len(coeffs); i++ {
// 		pow = pow.mulFrac(x)
// 		term = pow.mulFrac(coeffs[i].value)
// 		if coeffs[i].isNeg {
// 			accum, _ = sub128(accum, term, 0)
// 		} else {
// 			accum, _ = add128(accum, term, 0)
// 		}
// 	}

// 	return accum
// }

// Computes a Chebyshev polynomial at a particular x value. This method
// assumes that the input, coefficients AND result are all strictly less
// than 1.0, with the input and output also being positive.
func (x raw128) chebyPoly(coeffs []fix192_old) fix192_old {
	// Compute the Chebyshev polynomial using Horner's method.
	accum := coeffs[0]

	for i := 1; i < len(coeffs); i++ {
		accum, _ = accum.smul(fix192_old{raw64Zero, x})
		accum = accum.add(coeffs[i])
	}

	return accum // .intDiv(1048576)
}

func (x fix192_old) chebySin() fix192_old {
	if int64(x.i) >= 4 || isNeg64(x.i) {
		// TODO: Remove this check after testing is complete.
		panic("chebySin requires the input to be in the range (0, π)")
	}

	// Leverage the identity sin(x) = sin(π - x) to keep the input angle
	// in the range [0, 2] (required for the rest of this function).
	if x.i >= 2 {
		pi192 := fix192_old{3, raw128{0x243f6a8885a308d3, 0x13198a2e03707344}}
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
		sinXHalf := fix192_old{0, halfXFrac}.chebySin()
		quarterXFrac := ushiftRight128(halfXFrac, 1)
		sinXQuarter := fix192_old{0, quarterXFrac}.chebySin()

		// cos(x/2) = 1 - 2•sin²(x/4)
		// Note: We subtract from zero because the true value of one can't fit in 128-bits
		// but the result is the same because of twos-complement.
		sinXQuarterSquared := sinXQuarter.f.mulFrac(sinXQuarter.f)
		cosFrac, _ := sub128(raw128Zero, shiftLeft128(sinXQuarterSquared, 1), 0)

		res := sinXHalf.f.mulFrac(cosFrac)

		// If the product of the sign and cos terms is 0.5, the final result (which should
		// be multiplied by 2) will be exactly 1.0, so we return that.)
		if res.Hi == 0x8000000000000000 {
			return fix192_old{1, raw128{0, 0}}
		}

		return fix192_old{0, shiftLeft128(res, 1)}
	}

	if ult128(x.f, raw128{0x4942ff, 0x86c20bd3e6e3533a}) {
		// If x is very small, we can just return x since sin(x) is linear for small x.
		return x
	}

	// Start with the constant term of the chebyshev polynomial (often just zero!)
	sum := x.f.chebyPoly([]fix192_old{})

	// This polynomial expansion will never result in 1.0. The only positive inputs for sin()
	// that result in 1.0 are larger than 1 (specifically, π/2), and we only use the polynomial
	// expansion for values less than 1.0. (See the if x.i == 1 block above for more details.)
	return sum
}

// Returns tan(x) for very small x values, using the Chebyshev polynomial. Input and output
// must be strictly less than 1.0, and positive.
func (x raw128) chebyTan() fix192_old {
	return x.chebyPoly([]fix192_old{})
}

// A multiplication funcion that multiplies a complete fix192 value
// by a raw128 fractional value, assumes a positive fix192 value.
func (a fix192_old) mulByFraction(b raw128) (res fix192_old) {
	// a = a.i + a.f, b = b.f
	// a•b = (a.i + a.f)•b
	//      = a.i•b + a.f•b
	if isNeg64(a.i) {
		panic("mulByFraction() not implemented for negative fix192 values")
	}

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
func (x raw128) inverse() (fix192_old, error) {

	// NOTE: Returns 128 if x == 0
	fracZeros := leadingZeroBits128(x)

	if fracZeros >= 63 {
		// If the input is less than 2^-63, we can't represent it as a fix192 value.
		return fix192_old{}, ErrOverflow
	}

	// We use the following recursive formula to compute the inverse:
	// rₙ₊₁ = rₙ (2 - x·rₙ)

	// We start our estimate with a value that is 2^fracZeros, which is the
	// largest power of two that is less than or equal to 1/x. This ensures
	// our first estimate is on the same order of magnitude as the final result.
	two := fix192_old{2, raw128Zero}
	est := fix192_old{raw64(1 << fracZeros), raw128Zero}

	for i := 0; i < 8; i++ {
		prod := est.mulByFraction(x)
		est, _ = est.umul(two.sub(prod))
	}

	return est, nil
}

func (a fix192_old) intDiv(b uint64) (res fix192_old) {
	q := uint64(a.i) / b
	r := uint64(a.i) % b

	var rem raw128
	res.i = raw64(q)
	res.f, rem = div192by128(raw64(r), a.f.Hi, a.f.Lo, unscaledRaw128(b))

	if ushouldRound64(rem.Lo, raw64(b)) {
		res.f, _ = add128(res.f, raw128Zero, 1)
	}

	return res
}

func (a fix192_old) powNearOne(b fix192_old) (fix192_old, error) {
	// It's hard to get precision for large powers of numbers close to 1, because ln(a)
	// is very small and lacking precision, and any error is magnified a large exponent.
	// Instead, we can directly compute b•ln(a) by using a different expansion of ln(a).
	//
	// We start with the Taylor/Maclaurin series for values close to 1:
	//     ln(1 + ε) ≈ ε - (ε^2)/2 + (ε^3)/3 - ...
	//
	// We combine this with:
	//     a^b = exp(b * ln(a))
	//
	// Instead of computing ln(a) by itself, we compute b * ln(1 + ε) as:
	//     b * ln(1 + ε) ≈ b•ε - b•(ε^2)/2 + b•(ε^3)/3 - ...
	// Note that this works for both positive and negative values of ε or b.

	epsilon := a.sub(fix192_old{1, raw128Zero})

	epsilon, eSign := epsilon.abs()
	b, bSign := b.abs()

	term, err := epsilon.umul(b)

	if err == ErrNegOverflow {
		// If the product overflows in the negative direction, the exponential
		// would underflow.
		return fix192_old{}, ErrUnderflow
	} else if err != nil {
		return fix192_old{}, err
	}

	sum := term
	iter := uint64(1)

	if eSign < 0 {
		// If epsilon is negative, all of the odd terms (including the first!) will
		// be negative, and all of the even terms are negative already. So, we can just
		// do the whole expansion treating all terms as positive and then flip the sign
		// at the end (which we can do by just flipping the sign of b!)
		bSign *= -1

		for {
			term = term.mulByFraction(epsilon.f)
			if isZero64(term.i) && isZero128(term.f) {
				break
			}
			iter++
			sum = sum.add(term.intDiv(iter))
		}
	} else {
		// If epsilon is positive, we need to alternate between adding and subtracting
		// terms. We do this by just unrolling the loop and literally switching between
		// addition and subtraction.
		iter := uint64(1)

		for {
			term = term.mulByFraction(epsilon.f)
			if isZero64(term.i) && isZero128(term.f) {
				break
			}
			iter++
			sum = sum.sub(term.intDiv(iter))

			term = term.mulByFraction(epsilon.f)
			if isZero64(term.i) && isZero128(term.f) {
				break
			}
			iter++
			sum = sum.add(term.intDiv(iter))
		}
	}

	// We know that sum is strictly less than b, so applySign can't fail
	expValue, _ := sum.applySign(bSign)

	return expValue.exp()
}
