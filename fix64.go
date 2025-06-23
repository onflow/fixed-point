package fixedPoint

// This file implements most of the fixed-point arithmetic operations for UFix64 and Fix64 types.
// HOWEVER, it is carefully written so that a simple script can generate a new file with
// the same methods for the 128-bit fixed-point types, UFix128 and Fix128. If there are
// any changes to the logic here, that script should be re-run to generate the new file.
//
// If you see things in this file that seem awkward, that is probably why.

// == Comparison Operators ==

// Eq returns true if a and b are equal.
func (a UFix64) Eq(b UFix64) bool { return isEqual64(raw64(a), raw64(b)) }
func (a Fix64) Eq(b Fix64) bool   { return isEqual64(raw64(a), raw64(b)) }

// Lt returns true if a is less than b.
func (a UFix64) Lt(b UFix64) bool { return ult64(raw64(a), raw64(b)) }
func (a Fix64) Lt(b Fix64) bool   { return slt64(raw64(a), raw64(b)) }

// Gt returns true if a is greater than b.
func (a UFix64) Gt(b UFix64) bool { return b.Lt(a) }
func (a Fix64) Gt(b Fix64) bool   { return b.Lt(a) }

// Lte returns true if a is less than or equal to b.
func (a UFix64) Lte(b UFix64) bool { return !a.Gt(b) }
func (a Fix64) Lte(b Fix64) bool   { return !a.Gt(b) }

// Gte returns true if a is greater than or equal to b.
func (a UFix64) Gte(b UFix64) bool { return !a.Lt(b) }
func (a Fix64) Gte(b Fix64) bool   { return !a.Lt(b) }

// IsNeg returns true if a is negative.
func (a Fix64) IsNeg() bool { return isNeg64(raw64(a)) }

// IsZero returns true if a is zero.
func (a UFix64) IsZero() bool { return isZero64(raw64(a)) }
func (a Fix64) IsZero() bool  { return isZero64(raw64(a)) }

// These methods DO NOT check for overflow or underflow, so we keep them internal; they
// save time and error checking when we are doing computations with known bounds
// (i.e. in our transcendental functions).

// intMul multiplies a by b *without* checking for overflow or underflow.
func (a UFix64) intMul(b uint64) UFix64 { return UFix64(uintMul64(raw64(a), b)) }
func (a Fix64) intMul(b int64) Fix64    { return Fix64(sintMul64(raw64(a), b)) }

// intDiv divides a by b *without* checking for overflow or underflow.
func (a UFix64) intDiv(b uint64) UFix64 { return UFix64(uintDiv64(raw64(a), b)) }
func (a Fix64) intDiv(b int64) Fix64    { return Fix64(sintDiv64(raw64(a), b)) }

// shiftLeft shifts a left by n bits *without* checking for overflow.
func (a UFix64) shiftLeft(n uint64) UFix64 { return UFix64(shiftLeft64(raw64(a), n)) }
func (a Fix64) shiftLeft(n uint64) Fix64   { return Fix64(shiftLeft64(raw64(a), n)) }

// shiftRight shifts a right by n bits *without* checking for underflow.
func (a UFix64) shiftRight(n uint64) UFix64 { return UFix64(ushiftRight64(raw64(a), n)) }
func (a Fix64) shiftRight(n uint64) Fix64   { return Fix64(sshiftRight64(raw64(a), n)) }

// == Arithmetic Operators ==

// Add returns the sum of a and b, or an error on overflow.
func (a UFix64) Add(b UFix64) (UFix64, error) {
	sum, carry := add64(raw64(a), raw64(b), 0)
	if carry != 0 {
		return UFix64Zero, ErrOverflow
	}
	return UFix64(sum), nil
}

// Add returns the sum of a and b, or an error on overflow or negative overflow.
func (a Fix64) Add(b Fix64) (Fix64, error) {
	sum, _ := add64(raw64(a), raw64(b), 0)

	res := Fix64(sum)

	// Check for overflow by checking the sign bits of the operands and the result.
	if !a.IsNeg() && !b.IsNeg() && res.IsNeg() {
		return Fix64Zero, ErrOverflow
	} else if a.IsNeg() && b.IsNeg() && !res.IsNeg() {
		return Fix64Zero, ErrNegOverflow
	}

	return res, nil
}

// Sub returns the difference of a and b, or an error on negative overflow.
func (a UFix64) Sub(b UFix64) (UFix64, error) {
	diff, borrow := sub64(raw64(a), raw64(b), 0)
	if borrow != 0 {
		return UFix64Zero, ErrNegOverflow
	}
	return UFix64(diff), nil
}

// Sub returns the difference of a and b, or an error on overflow or negative overflow.
func (a Fix64) Sub(b Fix64) (Fix64, error) {
	diff, _ := sub64(raw64(a), raw64(b), 0)

	res := Fix64(diff)

	// Overflow occurs when:
	// 1. Subtracting a positive from a non-positive results in a positive
	// 2. Subtracting a negative from a non-negative results in a negative
	// Subtracting two non-zero values with the same sign can't overflow in a signed int64
	if !a.IsNeg() && b.IsNeg() && res.IsNeg() {
		return Fix64Zero, ErrOverflow
	} else if a.IsNeg() && !b.IsNeg() && !res.IsNeg() {
		return Fix64Zero, ErrNegOverflow
	}

	return res, nil
}

// Abs returns the absolute value of a as an unsigned value, with a sign value as an int64.
// Note that this method works properly for Fix64Min, which can NOT be represented as a positive Fix64.
func (a Fix64) Abs() (UFix64, int64) {
	if a.IsNeg() {
		// In twos-complement, multiplying "min value" (0x80000...) by -1 is a no-op!
		// And, the correct UNSIGNED value of the absolute value of min value is 0x80000...
		return UFix64(a.intMul(-1)), -1
	}

	return UFix64(a), 1
}

// Mul returns the product of a and b, or an error on overflow or underflow.
func (a UFix64) Mul(b UFix64) (UFix64, error) {
	// It might seem strange to implement multiplication in terms of fused multiply-divide,
	// but it turns out that a simple mulitiplication fixed-point operation needs to both
	// multiply and divide anyway. (Multiply the inputs, and then divide by the scale factor.)

	// Additionally, the logic for handling rounding is REALLY not trivial, so
	// having that in one location is a big win. In the end, the only real cost
	// is the overhead of an extra function call, which might be inlined anyway.
	return a.FMD(b, UFix64One)
}

// Mul returns the product of a and b, or an error on overflow or underflow.
func (a Fix64) Mul(b Fix64) (Fix64, error) {
	// Same rationale as above for UFix64.Mul, but even more critical because handling the
	// signs correctly is ALSO not trivial.
	return a.FMD(b, Fix64One)
}

// Div returns the quotient of a and b, or an error on division by zero, overflow, or underflow.
func (a UFix64) Div(b UFix64) (UFix64, error) {
	// Same rationale for using FMD as for UFix64.Mul
	return a.FMD(UFix64One, b)
}

// Div returns the quotient of a and b, or an error on division by zero, overflow, or underflow.
func (a Fix64) Div(b Fix64) (Fix64, error) {
	// Same rationale as above...
	return a.FMD(Fix64One, b)
}

// FMD returns a*b/c without intermediate rounding, or an error on division by zero, overflow, or underflow.
func (a UFix64) FMD(b, c UFix64) (UFix64, error) {
	// Must come before the check for a or b == 0 so we flag 0.0/0.0 as an error.
	if c.IsZero() {
		return UFix64Zero, ErrDivByZero
	}

	if a.IsZero() || b.IsZero() {
		return UFix64Zero, nil
	}

	hi, lo := mul64(raw64(a), raw64(b))

	// If the hi part is >= the divisor the result can't fit in 64 bits.
	if UFix64(hi).Gte(c) {
		return UFix64Zero, ErrOverflow
	}

	quo, rem := div64(hi, lo, raw64(c))

	if ushouldRound64(rem, raw64(c)) {
		var carry uint64
		quo, carry = add64(quo, raw64Zero, 1)

		// Make sure we don't "round up" to a value outside of the range of UFix64!
		if carry != 0 {
			return UFix64Zero, ErrOverflow
		}
	}

	// We can't get here if a == 0 or b == 0 because we checked that first. So,
	// a quotient of 0 means the result is too small to represent, i.e. underflow.
	// Note that we check this AFTER rounding.
	if isZero64(quo) {
		return UFix64Zero, ErrUnderflow
	}

	return UFix64(quo), nil
}

// FMD returns a*b/c without intermediate rounding, or an error on division by zero, overflow, or underflow.
func (a Fix64) FMD(b, c Fix64) (Fix64, error) {
	// Must come before the check for a or b == 0 so we flag 0.0/0.0 as an error.
	if c.IsZero() {
		return Fix64Zero, ErrDivByZero
	}

	if a.IsZero() || b.IsZero() {
		return Fix64Zero, nil
	}

	// Determine the sign of the result based on the signs of a, b, and c.
	sign := int64(1)

	aUnsigned, signMul := a.Abs()
	sign *= signMul
	bUnsigned, signMul := b.Abs()
	sign *= signMul
	cUnsigned, signMul := c.Abs()
	sign *= signMul

	// Compute the result using unsigned arithmetic.
	res, err := aUnsigned.FMD(bUnsigned, cUnsigned)

	if err != nil {
		if err == ErrOverflow && sign < 0 {
			return Fix64Zero, ErrNegOverflow
		} else {
			return Fix64Zero, err
		}
	}

	// Special case: if the result's sign should be negative and the product is the minimum
	// representable value, we can just return the minimum value. We need to do this because
	// the comparison against FixMax will fail below, even thought would be a valid result.
	if sign < 0 && isEqual64(raw64(res), raw64(Fix64Min)) {
		return Fix64Min, nil
	}

	if UFix64(res).Gt(UFix64(Fix64Max)) {
		if sign < 0 {
			return Fix64Zero, ErrNegOverflow
		} else {
			return Fix64Zero, ErrOverflow
		}
	}

	return Fix64(res).intMul(sign), nil
}

// Sqrt returns the square root of x as a UFix64, or an error if the result cannot be represented.
// Uses a Newton-Raphson iteration for fast convergence. Returns UFix64Zero for input zero.
func (x UFix64) Sqrt() (UFix64, error) {
	if x.IsZero() {
		return UFix64Zero, nil
	}

	// Count the number of leading zero bits in x, this is a cheap way of estimating
	// the order of magnitude of the input.
	n := leadingZeroBits64(raw64(x))

	// The loop below needs to start with some kind of estimate for the square root.
	// The closer it is to correct, the faster the loop will converge. We'll start
	// with a number that has a number of leading zero bits halfway between the number
	// of leading zero bits of x and the number of leading zero bits of the fixed-point
	// representation of 1. This will be of the same order of magnitude as the square
	// root, allowing our Newton-Raphson loop below to converge quickly.

	est := raw64(x)

	if n < Fix64OneLeadingZeros {
		// If the input has fewer leading zeros than FixOne, we'll start with an input
		// estimate that is shifted right by half the difference
		est = ushiftRight64(est, (Fix64OneLeadingZeros-n)/2)
	} else {
		// If the input has more leading zeros than FixOne, we shift left.
		est = shiftLeft64(est, (n-Fix64OneLeadingZeros)/2)
	}

	// The inner loop here will frequently divide the input by the current estimate,
	// so instead of using the Fix64.Div method, we expand the numerator once outside
	// the loop, and then directly call div64 in the loop.
	xHi, xLo := mul64(raw64(x), raw64(Fix64One))

	for {
		// This division can't fail: est is always a positive value somewhere between
		// x and 1, so it est will also be between x and 1.
		quo, rem := div64(xHi, xLo, est)

		if ushouldRound64(rem, est) {
			quo, _ = add64(quo, raw64Zero, 1)
		}

		// We take the difference using basic arithmetic, since we know that quo
		// and est are close to each other and far away from zero, so the difference
		// will never overflow or underflow a signed int (although it can be negative).
		diff, _ := sub64(quo, est, 0)

		// If the difference is zero, we've converged cleanly.
		if isZero64(diff) {
			break
		}

		// If the difference is ±iota, we know that the correct answer is either
		// quo or est, but we can't be sure which one is closer! The easiest way to
		// be sure is to just square the two values and see which one is closer to
		// the original input.
		if isIota64(diff) {
			// Diff is positive, so quo is larger than est, and quo^2 will be larger than x

			// Note that ignoring the hi part of this multiplication, and the borrow bit of
			// the subtraction are both effectively doing math modulo 2^64. Since we know that
			// the error is less than 2^64, we just ignore those potential "overflows" and
			// accept that the result will be correct modulo 2^64.
			_, estLo := mul64(raw64(est), raw64(est))
			estError, _ := sub64(xLo, estLo, 0)

			_, quoLo := mul64(raw64(quo), raw64(quo))
			quoError, _ := sub64(quoLo, xLo, 0)

			if ult64(quoError, estError) {
				// If quo has a lower error, use that instead of est.
				est = quo
			}
			break
		} else if isNegIota64(diff) {
			// Same logic as above, except diff is negative, so quo is smaller
			_, estLo := mul64(raw64(est), raw64(est))
			estError, _ := sub64(estLo, xLo, 0)

			_, quoLo := mul64(raw64(quo), raw64(quo))
			quoError, _ := sub64(xLo, quoLo, 0)

			if ult64(quoError, estError) {
				// If the estimate is further away, we can just use quo.
				est = quo
			}
			// If quo has a lower error, use that instead of est.
			break
		}

		diff = sshiftRight64(diff, 1)

		est, _ = add64(est, raw64(diff), 0)
	}

	return UFix64(est), nil
}

// The inner loop of the Ln() method, factored out into its own function to simplify
// the cases where the input is small enough that we don't need to actually go through
// the loop.
func lnInnerLoop64(xScaled UFix64) Fix64 {
	// We will compute ln(x) using the approximation:
	// ln(x) = 2 * (z + z^3/3 + z^5/5 + z^7/7 + ...)
	// where z = (x - 1) / (x + 1)

	num, _ := Fix64(xScaled).Sub(fix64LnScale)
	den, _ := Fix64(xScaled).Add(fix64LnScale)
	z, err := num.FMD(fix64LnScale, den)

	if err == ErrUnderflow {
		// If z is too small to represent, we just return 0
		return Fix64Zero
	}

	// Precompute z^2 to avoid recomputing it in the loop.
	z2, err := z.FMD(z, fix64LnScale)

	if err == ErrUnderflow {
		// If z^2 is too small, we just return 2*z, which is the first term
		// in the series.
		return z.intMul(2)
	}

	term := z
	sum := z
	iter := int64(1)

	// Keep iterating until "term" and/or "next" rounds to zero
	for {
		// The input to this function is restricted such that this can't overflow
		term, err = term.FMD(z2, fix64LnScale)

		if err == ErrUnderflow {
			break
		}

		next := term.intDiv(iter*2 + 1)

		// intDiv doesn't check for underflow, but a result of zero means
		// the same thing here
		if next.IsZero() {
			break
		}

		sum, _ = sum.Add(next)
		iter += 1
	}

	return sum.intMul(2)
}

// Ln returns the natural logarithm of x, or an error if x is zero or negative.
func (x UFix64) Ln() (Fix64, error) {
	if x.IsZero() {
		return Fix64Zero, ErrDomain
	}

	// The Taylor expansion of ln(x) converges faster for values closer to 1.
	// If we scale the input to have exactly the same number of leading zero bits
	// as FixOne's representation, the input will be between 0.5 and 2 (for the
	// two fixed-point types we use, it's actually in the range (0.6 - 1.3)).
	//
	// For every power of two removed (or added) by this shift, we can add (or subtract)
	// a multiple of ln(2) at the end to get the final ln() value.
	leadingZeros := int64(leadingZeroBits64(raw64(x)))
	k := Fix64OneLeadingZeros - leadingZeros

	var xScaled UFix64

	if k <= 0 {
		// If k is 0 or negative we want to shift the input to the left
		// (i.e multiply). No worries about losing precision or overflowing
		// here: k can only be negative if the input is less than one. We just
		// shift the input left by -k bits and multiply by fix64LnMultiplier
		shifted := UFix64(x).shiftLeft(uint64(-k))
		xScaled = shifted.intMul(fix64LnMultiplier)
	} else {
		// If k is positive, we want to divide the input by 2^k, but if we just
		// shift x to the right, we could lose precision. Instead, we use our
		// FMD function to scale the input up by fix64LnMultiplier at the same time
		// as scaling it down by 2^k.
		xScaled, _ = x.FMD(UFix64(unscaledRaw64(fix64LnMultiplier)), UFix64Iota.shiftLeft(uint64(k)))
	}

	resScaled := lnInnerLoop64(xScaled)

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did above. Note that k*ln2 must strictly lie between minLn and maxLn constants
	// and we chose ln2Multiple and ln2Factor so they can't overflow when multiplied
	// by values in that range.
	powerCorrection := fix64Ln2Scaled.intMul(k)

	resScaled, err := resScaled.Add(powerCorrection)

	if err != nil {
		return Fix64Zero, err
	}

	// Divide out lnScale to get the final result.
	return resScaled.Div(fix64LnScale)
}

// clampAngle normalizes the input angle x to the range [0, π], returning the normalized value
// and a flag indicating if the result should be interpreted as negative. The output is scaled
// up by fix64TrigScale. Used by Sin(), Cos() and Tan().
func clampAngle64(x Fix64) (Fix64, bool) {
	// The goal of this function is to normalize the input angle x to the range [0, π], with
	// a separate flag to indicate if the result should be interpreted as negative. (Separating
	// out the sign is actually pretty convenient for the calling functions; sin applies the sign
	// after running the Taylor expansion, and cos just throws it away!)
	var unsignedX raw64
	var isNeg bool

	if !x.IsNeg() {
		unsignedX = raw64(x)
		isNeg = false
	} else {
		unsignedX = raw64(x.intMul(-1))
		isNeg = true
	}

	// Multiply the input by fix64TrigMultiplier
	hi, lo := mul64(unsignedX, unscaledRaw64(fix64TrigMultiplier))

	// Take the scaled up value, and compute the remainder when divided by 2π (at the same scale)
	_, rem := div64(hi, lo, raw64(fix64TwoPiScaled))

	// rem is now the input angle, modulo 2π, scaled up to fix64TrigScale. If the angle is greater
	// than π, subtract it from 2π to bring it into the range [0, π] and flip the sign flag.
	if ult64(raw64(fix64PiScaled), rem) {
		rem, _ = sub64(raw64(fix64TwoPiScaled), rem, 0)
		isNeg = !isNeg
	}

	return Fix64(rem), isNeg
}

// scaledSin returns sin(x), assuming x has already been normalized to [0, π] and scaled up
// by fix64TrigScale (i.e. has gone through clampAngle).
func (xScaled Fix64) innerSin64() (Fix64, error) {
	// Leverage the identity sin(x) = sin(π - x) to keep the input angle
	// in the range [0, π/2]. This is useful because the Taylor expansion for sin(x)
	// converges faster for smaller values of x
	if xScaled.Gt(fix64HalfPiScaled) {
		xScaled, _ = fix64PiScaled.Sub(xScaled)
	}

	if xScaled.Lte(fix64SinIotaScaled) {
		// If x is very small, we can just return x since sin(x) is linear for small x.
		return xScaled, nil
	}

	// sin(x) = x - x^3/3! + x^5/5! - x^7/7! + ...
	xSquared, _ := xScaled.FMD(xScaled, fix64TrigScale)

	sum := xScaled
	term := xScaled
	iter := int64(1)

	for {
		var err error

		term = term.intMul(-1) // Alternate the sign of the term

		term, err = term.FMD(xSquared, fix64TrigScale)

		if err == ErrUnderflow {
			break
		}

		term = term.intDiv((iter + 1) * (iter + 2))
		if term.IsZero() {
			break
		}

		iter += 2

		// Both sum and term are always close to zero, so we don't need to worry
		// about overflow
		sum, _ = sum.Add(term)
	}

	return sum, nil
}

func (x Fix64) Sin() (Fix64, error) {
	// Normalize the input angle to the range [0, π], with a flag indicating
	// if the result should be interpreted as negative.
	x_scaled, isNeg := clampAngle64(x)

	if x_scaled.IsZero() {
		return Fix64Zero, nil
	}

	res_scaled, err := x_scaled.innerSin64()
	if err != nil {
		return Fix64Zero, err
	}

	res, _ := res_scaled.Div(fix64TrigScale)

	// sin(-x) = -sin(x)
	if isNeg {
		res = res.intMul(-1)
	}

	return res, nil
}

// Cos returns the cosine of x (in radians).
func (x Fix64) Cos() (Fix64, error) {
	// Ignore the sign since cos(-x) = cos(x)
	xScaled, _ := clampAngle64(x)

	if xScaled.IsZero() {
		return Fix64One, nil // cos(0) = 1
	}

	// We use the following identities to compute cos(x):
	//     cos(x) = sin(π/2 - x)
	//     cos(x) = -sin(3π/2 − x)
	// If x is is less than or equal to π/2, we can use the first identity,
	// if x is greater than π/2, we use the second identity.
	// In both cases, we end up with a value in the range [0, π], to pass
	// to scaledSin().
	var yScaled Fix64
	var isNeg bool

	if xScaled.Lt(fix64HalfPiScaled) {
		// cos(x) = sin(π/2 - x)
		yScaled, _ = fix64HalfPiScaled.Sub(xScaled)
		isNeg = false
	} else {
		// cos(x) = -sin(3π/2 − x)
		yScaled, _ = fix64ThreeHalfPiScaled.Sub(xScaled)
		isNeg = true
	}

	resScaled, err := yScaled.innerSin64()
	if err != nil {
		return Fix64Zero, err
	}

	res, _ := resScaled.Div(fix64TrigScale)

	if isNeg {
		res = res.intMul(-1)
	}

	return res, nil
}
