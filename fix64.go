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

// Neg returns the additive inverse of a (i.e. -a), or a negative overflow error
func (a Fix64) Neg() (Fix64, error) {
	if a == Fix64Min {
		// Special case: negating the minimum value will overflow.
		return Fix64Zero, ErrNegOverflow
	}

	return Fix64(neg64(raw64(a))), nil
}

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
	// Subtracting two, non-zero values with the same sign can't overflow in a signed int64
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
		// Neg of a raw type equal to "min value" (0x80000...) is a no-op!
		// And, the correct UNSIGNED value of the absolute value of min value is 0x80000...
		return UFix64(neg64(raw64(a))), -1
	}

	return UFix64(a), 1
}

// ApplySign converts a UFix64 to a Fix64, applying the sign specified by the input.
func (a UFix64) ApplySign(sign int64) (Fix64, error) {
	if sign == 1 {
		if a.Gt(UFix64(Fix64Max)) {
			return Fix64Zero, ErrOverflow
		}
		return Fix64(a), nil
	} else {
		// Special case: if the result's sign should be negative and the converted
		// value is the minimum representable value, we can just return the minimum
		// value. We need to do this because the comparison against FixMax will fail
		// below, even thought would be a valid result.
		if isEqual64(raw64(a), raw64(Fix64Min)) {
			return Fix64Min, nil
		}
		if a.Gt(UFix64(Fix64Max)) {
			return Fix64Zero, ErrNegOverflow
		}

		return Fix64(neg64(raw64(a))), nil
	}
}

// Mul returns the product of a and b, or an error on overflow or underflow.
func (a UFix64) Mul(b UFix64) (UFix64, error) {
	// It might seem strange to implement multiplication in terms of fused multiply-divide,
	// but it turns out that a simple fixed-point multiplication needs to both
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

	return res.ApplySign(sign)
}

// Sqrt returns the square root of x using Newton-Rhaphson. Note that this
// method returns an error result for consistency with other methods,
// but can't actually ever fail...
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
			_, estLo := mul64(est, est)
			estError, _ := sub64(xLo, estLo, 0)

			_, quoLo := mul64(quo, quo)
			quoError, _ := sub64(quoLo, xLo, 0)

			if ult64(quoError, estError) {
				// If quo has a lower error, use that instead of est.
				est = quo
			}
			break
		} else if isNegIota64(diff) {
			// Same logic as above, except diff is negative, so quo is smaller
			_, estLo := mul64(est, est)
			estError, _ := sub64(estLo, xLo, 0)

			_, quoLo := mul64(quo, quo)
			quoError, _ := sub64(xLo, quoLo, 0)

			if ult64(quoError, estError) {
				// If the estimate is further away, we can just use quo.
				est = quo
			}
			// If quo has a lower error, use that instead of est.
			break
		}

		diff = sshiftRight64(diff, 1)

		est, _ = add64(est, diff, 0)
	}

	return UFix64(est), nil
}

// A utility function to prescale inputs to ln() and pow()
func (x UFix64) prescale() (UFix64, uint64) {
	// For very small values, converting to fix192 will lose precision. To avoid this,
	// we first check to see if the input has more leading zero bits than the
	// fixed-point representation of 1, and if so, we scale it up by shifting
	// it left by one less than the difference.
	prescale := uint64(0)
	zeros := leadingZeroBits64(raw64(x))

	if zeros > Fix64OneLeadingZeros {
		// We need to shift left by enough so that we have one more leading zero bit
		// than the fixed point representation of 1.
		prescale = zeros - Fix64OneLeadingZeros - 1

		x = x.shiftLeft(prescale)
	}

	return x, prescale
}

func (x UFix64) Ln() (Fix64, error) {
	// Prescale to avoid precision loss when converting to fix192
	x, prescale := x.prescale()

	// TODO: x192.ln() provides a ton of precision that we don't need, it
	// would be ideal if we could pass an error limit to it so it could
	// stop early when we don't need the full precision.
	res192, err := x.toFix192_old().ln(prescale)

	if err != nil {
		return Fix64Zero, err
	}

	return res192.toFix64()
}

// Exp(x) returns e^x, or an error on overflow or underflow. Note that although the
// input is a Fix64, the output is a UFix64, since e^x is always positive.
func (x Fix64) Exp() (UFix64, error) {
	// If x is 0, return 1.
	if x.IsZero() {
		return UFix64One, nil
	}

	// We can quickly check to see if the input will overflow or underflow
	if x.Gt(maxLn64) {
		return UFix64Zero, ErrOverflow
	} else if x.Lt(minLn64) {
		return UFix64Zero, ErrUnderflow
	}

	// Use the fix192 implementation of Exp
	res192, err := x.toFix192_old().exp()

	if err != nil {
		return UFix64Zero, err
	}

	return res192.toUFix64()
}

func (a UFix64) Pow(b Fix64) (UFix64, error) {
	// We accept 0^0 as 1.
	if b.IsZero() {
		return UFix64One, nil
	}

	if a.IsZero() {
		if b.IsNeg() {
			// 0^negative is undefined, so we return an error.
			return UFix64Zero, ErrDivByZero // 0^negative is undefined
		} else {
			// 0^positive is 0.
			return UFix64Zero, nil
		}
	}

	if a.Eq(UFix64One) {
		// 1^b is always 1, so we can return it directly.
		return UFix64One, nil
	}

	// a^1 is just a, so we can return it directly.
	if b.Eq(Fix64One) {
		return a, nil
	}

	// Prescale the base to avoid precision loss when converting to fix192
	a, prescale := a.prescale()

	a192 := a.toFix192_old()
	b192 := b.toFix192_old()

	res192, err := a192.pow(b192, prescale)

	if err != nil {
		return UFix64Zero, err
	}

	return res192.toUFix64()
}

func trigResult64(res192 fix192_old, err error) (Fix64, error) {
	if err != nil {
		return Fix64Zero, err
	}

	res, err := res192.toFix64()

	if err == ErrUnderflow {
		// For trig underflows, we just return 0.
		return Fix64Zero, nil
	} else if err != nil {
		return Fix64Zero, err
	}

	return res, nil
}

func (x Fix64) Sin() (Fix64, error) {
	return trigResult64(x.toFix192_old().sin())
}

func (x Fix64) Cos() (Fix64, error) {
	return trigResult64(x.toFix192_old().cos())
}

func (x Fix64) Tan() (Fix64, error) {
	// Unlike with sin and cos, we want tan() to return an underflow error
	// TODO: Do we really? :laughing:
	res, err := x.toFix192_old().tan()

	if err != nil {
		return Fix64Zero, err
	}

	return res.toFix64()
}

// Cos returns the cosine of x (in radians).
// func (x Fix64) Cos() (Fix64, error) {
// 	// Ignore the sign since cos(-x) = cos(x)
// 	xScaled, _ := clampAngle64(x)

// 	if xScaled.IsZero() {
// 		return Fix64One, nil // cos(0) = 1
// 	}

// 	// We use the following identities to compute cos(x):
// 	//     cos(x) = sin(π/2 - x)
// 	//     cos(x) = -sin(3π/2 − x)
// 	// If x is is less than or equal to π/2, we can use the first identity,
// 	// if x is greater than π/2, we use the second identity.
// 	// In both cases, we end up with a value in the range [0, π], to pass
// 	// to scaledSin().
// 	var yScaled UFix64
// 	var sign int64

// 	if xScaled.Lt(ufix64HalfPiScaled) {
// 		// cos(x) = sin(π/2 - x)
// 		yScaled, _ = ufix64HalfPiScaled.Sub(xScaled)
// 		sign = 1
// 	} else {
// 		// cos(x) = -sin(3π/2 − x)
// 		yScaled, _ = ufix64ThreeHalfPiScaled.Sub(xScaled)
// 		sign = -1
// 	}

// 	resScaled, err := yScaled.innerSin64()
// 	if err != nil {
// 		return Fix64Zero, err
// 	}

// 	res, _ := resScaled.Div(fix64TrigScale)

// 	res = res.intMul(sign)

// 	return res, nil
// }

// Tan returns the tangent of x (in radians).
// func (x Fix64) Tan() (Fix64, error) {
// tan(x) = sin(x) / cos(x)
// We can't use the Sin() and Cos() methods directly because they don't provide
// enough precision once we divide the results.

// Normalize the input angle to the range [0, π]
// xScaled, sign := clampAngle64(x)

// if xScaled.IsZero() {
// 	return Fix64Zero, nil
// }
// // We compute y the same way we did in the cos() function above.
// var yScaled UFix64

// if xScaled.Lt(ufix64HalfPiScaled) {
// 	// cos(x) = sin(π/2 - x)
// 	yScaled, _ = ufix64HalfPiScaled.Sub(xScaled)
// } else {
// 	// cos(x) = -sin(3π/2 − x)
// 	yScaled, _ = ufix64ThreeHalfPiScaled.Sub(xScaled)
// 	sign *= -1
// }

// sinX, err := xScaled.innerSin64()
// if err != nil {
// 	return Fix64Zero, err
// }

// cosX, err := yScaled.innerSin64()
// if err != nil {
// 	return Fix64Zero, err
// }

// res, err := sinX.FMD(Fix64One, cosX)

// if err != nil {
// 	return Fix64Zero, err
// }

// // if res > Fix64(5e15) {
// // 	if sign < 0 {
// // 		// If the result is too large and negative, we return a negative overflow error
// // 		return 0, ErrNegOverflow
// // 	} else {
// // 		// If the result is too large and positive, we return a positive overflow error
// // 		return 0, ErrOverflow
// // 	}
// // }

// res = res.intMul(sign)

// return res, nil
// }
