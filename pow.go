package fixedPoint

// Exp returns e^x, or an error on overflow or underflow.
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

	// Switch to unsigned representation to simplify the logic and give
	// us a larger range of values to work with.
	var unsignedX UFix64
	var isNeg bool

	if !x.IsNeg() {
		unsignedX = UFix64(x)
		isNeg = false
	} else {
		unsignedX = UFix64(x.intMul(-1))
		isNeg = true
	}

	// Now that we've limited the input to minLn and maxLn, we can multiply
	// by ln2Multiplier without fear of overflow (ln2Multiplier was specifically
	// chosen not to overflow an unsigned fix when multiplied by values in that range).
	xScaled := unsignedX.intMul(fix64LnMultiplier)

	// We can do the opposite of the trick we did in ln(x). There, we scaled the input
	// by powers of 2, and then added/subtracted multiples of ln(2) at the end. Here
	// we can add/subtract multiples of ln(2) at the beginning, and then scale by
	// the appropriate power of 2 at the end.

	// The first step is to do a division of the input by ln(2); the quotient will
	// tell us how many shifts we need to do at the end to scale the result of our
	// inner loop to the correct value, and the remainder will be the input to our
	// Taylor series expansion.
	//
	// Of course, "divide by ln(2)" isn't actually as simple as it sounds because
	// we can't represent ln(2) exactly in fix64, we need to use an approximation.
	// So, in order to introduce as small of an error as possible, we scale the
	// input AND ln(2) by a factor that minimizes the error. See genFactors.py for
	// more details.

	quo, rem := div64(0, raw64(xScaled), raw64(fix64Ln2Scaled))

	k := uint64(quo)

	// We now have a value k indicating the number of times we need to scale the result after
	// our inner loop, and a remainder that is the input to our Taylor series expansion.
	// The remainder is in the range [0, ln(2)) (~0.69) using lnScale precision.
	seriesInput := UFix64(rem).intMul(13)
	seriesScale := ufix64LnScale.intMul(13)

	// We use the Taylor series to compute e^x. The series is:
	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...

	// Remember: Although we are using the UFix type here, the actual scale factor
	// of these values is lnScale. Provided we are careful to only use functions
	// that are scale agnostic (lke FMD and simple addition), this won't bite us.
	term := seriesScale // Starts as 1.0 in the series
	sum := term
	iter := uint64(1)
	var err error

	// This loop converges in under 20 iterations in testing
	for {
		// Multiply another power of x into the term, using FMD with seriesScale
		// as the divisor will keep the result in the correct scale.
		term, err = term.FMD(seriesInput, seriesScale)

		if err == ErrUnderflow {
			// If the term is too small to represent, we can just break out of the loop.
			break
		} else if err != nil {
			return UFix64Zero, err
		}

		// Divide by the iteration number to account for the factorial.
		term = term.intDiv(iter)

		// inDiv doesn't check for underflow, but a result of zero means
		// the term is too small to represent.
		if term.IsZero() {
			break
		}

		// Can't overflow
		sum, _ = sum.Add(term)
		iter += 1
	}

	var res UFix64

	// What we have now is the value of sum = e^x with three overlapping factors:
	//    - k, which is the power of two that we need to multiply the result by
	//      to get the final value.
	//    - lnScale, which is the scale factor we used for the Taylor series
	//      expansion to maintain precision.
	//    - isNeg, which is a flag indicating whether the input was negative. If it
	//      was, we need to return the multiplicative inverse of the sum.
	//
	// We resolve all of these factors with a FMD and bit shifting
	if !isNeg {
		// Our inner loop computed the result that needs to be multiplied by 2^k
		// and scaled down by seriesScale. We can resolve both of those with a single
		// FMD call.
		res, err = sum.FMD(UFix64One.shiftLeft(k), seriesScale)
	} else {
		// For negative input, we need the multiplicitive inverse of the result. The
		// positive result is (sum * 2^k) / lnScale, so the inverse is
		// (lnScale / (sum * 2^k)). We can rearange this to be (lnScale >> k) / sum.
		// Even better, we can save some precision by splitting the shift between
		// lnScale and FixOne and using FMD(lnScale, FixOne, sum) to compute the final result.
		// (Note that lnScale and FixOne each have the same number of trailing zero bits
		// because they both have the same large power of ten as a factor, and thus
		// the same large power of two as a factor.)
		oneShift := k / 2
		scaleShift := k - oneShift
		res, err = seriesScale.shiftRight(scaleShift).FMD(UFix64One.shiftRight(oneShift), sum)
	}

	if err != nil {
		return UFix64Zero, err
	}

	return res, nil
}

// Tan returns the tangent of x (in radians).
func (x Fix64) Tan() (Fix64, error) {
	// tan(x) = sin(x) / cos(x)
	// We can't use the Sin() and Cos() methods directly because they don't provide
	// enough precision once we divide the results.

	// Normalize the input angle to the range [0, π]
	xScaled, isNeg := clampAngle64(x)

	if xScaled == 0 {
		return 0, nil
	}

	// We compute y the same way we did in the cos() function above.
	var yScaled Fix64

	if xScaled < fix64HalfPiScaled {
		// cos(x) = sin(π/2 - x)
		yScaled, _ = fix64HalfPiScaled.Sub(xScaled)
	} else {
		// cos(x) = -sin(3π/2 − x)
		yScaled, _ = fix64ThreeHalfPiScaled.Sub(xScaled)
		isNeg = !isNeg
	}

	sinX, err := xScaled.innerSin64()
	if err != nil {
		return 0, err
	}

	cosX, err := yScaled.innerSin64()
	if err != nil {
		return 0, err
	}

	res, err := sinX.FMD(Fix64One, cosX)

	if err != nil {
		return 0, err
	}

	if res > Fix64(5e15) {
		if isNeg {
			// If the result is too large and negative, we return a negative overflow error
			return 0, ErrNegOverflow
		} else {
			// If the result is too large and positive, we return a positive overflow error
			return 0, ErrOverflow
		}
	}

	if isNeg {
		res = res.intMul(-1)
	}

	return res, nil
}

func (a UFix64) Pow(b Fix64) (UFix64, error) {
	// The order of these guards is important!

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

	lnA, err := a.Ln()

	if err != nil {
		return UFix64Zero, err
	}

	prod, err := lnA.Mul(b)

	if err == ErrUnderflow {
		// If the product is too small to represent, the result will effectively be 1
		return UFix64One, nil
	} else if err == ErrOverflow {
		// The multiplication can easily overflow (due to fix64_ln2Multiplier)
		// if ln(a)•b is larger than 27. However, if that is true, then the exp()
		// call below would also overflow, so returning an overflow error here is
		// appropriate.
		return UFix64Zero, ErrOverflow
	} else if err == ErrNegOverflow {
		// If the product overflows in the negative direction, the exponential
		// would underflow.
		return UFix64Zero, ErrUnderflow
	}

	return prod.Exp()
}

// func (a UFix64) powNearOne(b Fix64) (UFix64, error) {
// 	// We can use a different expansion to handle powers of numbers close to 1.
// 	// We use the Taylor/Maclaurin series:
// 	//     ln(1 + e) ≈ e - (e^2)/2 + (e^3)/3 - ...
// 	// We combine this with:
// 	//     a^b = exp(b * ln(a))
// 	//
// 	// Instead of computing ln(a) by itself, we compute b * ln(1 + e) as:
// 	//     b * ln(1 + e) ≈ b•e - b•(e^2)/2 + b•(e^3)/3 - ...
// 	seriesScale := Fix64One.intMul(int64(fix64LnMultiplier))
// 	e := (Fix64(a) - Fix64One).intMul(int64(fix64LnMultiplier))

// 	term, err := b.Mul(e)

// 	if err == ErrUnderflow {
// 		// If the product is too small to represent, the result will effectivly be 1
// 		return UFix64One, nil
// 	} else if err == ErrOverflow {
// 		// The multiplication can easily overflow (due to fix64_ln2Multiplier)
// 		// if ln(a)•b is larger than 27. However, if that is true, then the exp()
// 		// call below would also overflow, so returning an overflow error here is
// 		// appropriate.
// 		return 0, ErrOverflow
// 	} else if err == ErrNegOverflow {
// 		// If the product overflows in the negative direction, the exponential
// 		// would underflow.
// 		return 0, ErrUnderflow
// 	} else if err != nil {
// 		// I don't think there are any other errors that can occur here, but... :shrug:
// 		return 0, err
// 	}

// 	sum := term
// 	iter := int64(2)
// 	for {
// 		term = term.intMul(-1)
// 		term, err = term.FMD(e, seriesScale)

// 		if err == ErrUnderflow {
// 			break
// 		}

// 		sum += term.intDiv(iter)
// 	}

// 	exp, k, isNeg, err := sum.scaledExp()

// 	if err != nil {
// 		return 0, err
// 	}

// 	return unscaleExp(exp, k, isNeg)
// }

// func (a UFix64) Pow(b Fix64) (UFix64, error) {
// 	// The order of these guards is important!

// 	// We accept 0^0 as 1.
// 	if b == 0 {
// 		return UFix64One, nil
// 	}

// 	if a == 0 {
// 		if b < 0 {
// 			// 0^negative is undefined, so we return an error.
// 			return 0, ErrDivByZero // 0^negative is undefined
// 		} else {
// 			// 0^positive is 0.
// 			return 0, nil
// 		}
// 	}

// 	if a == UFix64One {
// 		// 1^b is always 1, so we can return it directly.
// 		return UFix64One, nil
// 	}

// 	// a^1 is just a, so we can return it directly.
// 	if b == Fix64One {
// 		return a, nil
// 	}

// 	if a > (Fix64Scale-1000) && a < (Fix64Scale+1000) {
// 		// If a is close to 1, we can efficiently use a direct Taylor series expansion to compute the result.
// 		return a.powNearOne(b)
// 	}

// 	// Compute ln(a), but keep the result scaled up for greater precision.
// 	scaledLnA, err := a.scaledLn()

// 	if err != nil {
// 		return 0, err
// 	}

// 	// Mulitply the scaled up value by b.
// 	scaledProduct, err := scaledLnA.Mul(b)

// 	if err == ErrUnderflow {
// 		// If the product is too small to represent, the result will effectivly be 1
// 		return UFix64One, nil
// 	} else if err == ErrOverflow {
// 		// The multiplication can easily overflow (due to fix64_ln2Multiplier)
// 		// if ln(a)•b is larger than 27. However, if that is true, then the exp()
// 		// call below would also overflow, so returning an overflow error here is
// 		// appropriate.
// 		return 0, ErrOverflow
// 	} else if err == ErrNegOverflow {
// 		// If the product overflows in the negative direction, the exponential
// 		// would underflow.
// 		return 0, ErrUnderflow
// 	} else if err != nil {
// 		// I don't think there are any other errors that can occur here, but... :shrug:
// 		return 0, err
// 	}

// 	if scaledProduct > maxLn64.intMul(int64(fix64LnMultiplier)) {
// 		return 0, ErrOverflow
// 	} else if scaledProduct < minLn64.intMul(int64(fix64LnMultiplier)) {
// 		return 0, ErrUnderflow
// 	}

// 	exp, k, isNeg, err := scaledProduct.scaledExp()

// 	if err != nil {
// 		return 0, err
// 	}

// 	return unscaleExp(exp, k, isNeg)
// }

// func (a UFix64) intPow(b uint64) (UFix128, error) {

// 	remainingB := b

// 	accum := UFix128One
// 	term := a.ToUFix128()

// 	if remainingB&0x01 == 1 {
// 		accum = term
// 	}

// 	remainingB >>= 1

// 	for remainingB != 0 {
// 		var err error

// 		// Square the term
// 		term, err = term.Mul(term)

// 		if err != nil {
// 			return UFix128Zero, err
// 		}

// 		if remainingB&0x01 == 1 {
// 			// If the current bit is set, multiply the accumulator by the term
// 			accum, err = accum.Mul(term)

// 			if err != nil {
// 				return UFix128Zero, err
// 			}
// 		}
// 		remainingB >>= 1
// 	}

// 	return accum, nil
// }

// func (a UFix64) fracPow(b Fix64) (UFix128, error) {
// 	scaledLn, err := a.scaledLn()

// 	if err != nil {
// 		return UFix128Zero, err
// 	}

// 	// Both scaledLN and b are scaled up by 10**8, by multiplying them together,
// 	// we get a value that is scaled up by 10**16. We then divide out the
// 	// fix64_ln2Multiplier factor and do our Taylor series expansion
// 	// at a precision of 10**20.
// 	prodHi, prodLo := mul64(raw64(scaledLn), raw64(b*1e3))
// 	fractionalProd, _ := div64(prodHi, prodLo, raw64(fix64_ln2Multiplier))

// 	term := fractionalProd
// 	sum := fractionalProd
// 	iter := uint64(2)

// 	for {
// 		var r raw64
// 		// Multiply another power of x into the term
// 		hi, lo := mul64(raw64(term), raw64(fractionalProd))

// 		// Because we just multiplied two numbers together that are each scaled
// 		// up by 2^20, the 128-bit result, (hi, lo) is scaled up by 2^40. We can't
// 		// actually divide out the extra 2^20 with bits.Div64, because the denominator
// 		// is bigger than 2^64! Fortunately, we know that 2^20 = 2^20 * 5^20, so we can
// 		// shift by 20 bits, and the divide by 5^20, which is 95367431640625. Even better,
// 		// because we are using this smaller denominator, we can include the iter division
// 		// as part of the same operation. (iter is always a positive integer, and this loop
// 		// converges LONG before iter gets large enough to overflow.)

// 		lo = hi<<(64-19) | (lo >> 19)
// 		hi = hi >> 19
// 		denom := raw64(19073486328125 * iter)
// 		term, r = div64(hi, lo, denom)

// 		if ushouldRound64(r, denom) {
// 			term, _ = add64(term, rawZero, 1)
// 		}

// 		// Break out of the loop when the term is too small to change the sum.
// 		if term == 0 {
// 			break
// 		}

// 		sum, _ = add64(sum, term, 0)
// 		iter += 1
// 	}

// 	// We have e^prod - 1 now, scaled up by 10^16. We can very easily turn this
// 	// into a UFix128 by multiplying by 10^8, to get the 10^24 scaling factor used
// 	// for UFix128.
// 	hi, lo := mul64(sum, 1e5)
// 	res := UFix128{Hi: uint64(hi), Lo: uint64(lo)}

// 	// Now we can add in the 1.0 we ignored earlier!
// 	return res.Add(UFix128One)
// }

// func (a UFix64) PowTest(b Fix64) (UFix64, error) {
// 	// The order of these guards is important!

// 	// We accept 0^0 as 1.
// 	if b == 0 {
// 		return UFix64One, nil
// 	}

// 	if a == 0 {
// 		if b < 0 {
// 			// 0^negative is undefined, so we return an error.
// 			return 0, ErrDivByZero // 0^negative is undefined
// 		} else {
// 			// 0^positive is 0.
// 			return 0, nil
// 		}
// 	}

// 	if a == UFix64One {
// 		// 1^b is always 1, so we can return it directly.
// 		return UFix64One, nil
// 	}

// 	// a^1 is just a, so we can return it directly.
// 	if b == Fix64One {
// 		return a, nil
// 	}

// 	if a > (Fix64Scale-1000) && a < (Fix64Scale+1000) {
// 		// If a is close to 1, we can efficiently use a direct Taylor series expansion to compute the result.
// 		return a.powNearOne(b)
// 	}

// 	// We'll handle negative values later...
// 	if b < 0 {
// 		return a.Pow(b)
// 	}

// 	var err error

// 	b += Fix64One / 2

// 	b_int := uint64(b / Fix64One)
// 	b_frac := b % Fix64One

// 	intResult, err := a.intPow(b_int)

// 	if err != nil {
// 		return 0, err
// 	}

// 	b_frac -= Fix64One / 2
// 	fracNeg := false

// 	if b_frac < 0 {
// 		fracNeg = true
// 		b_frac = -b_frac
// 	}

// 	fracResult, err := a.fracPow(b_frac)

// 	if err != nil {
// 		return 0, err
// 	}

// 	var combinedResult UFix128

// 	if fracNeg {
// 		combinedResult, err = intResult.Div(fracResult)
// 	} else {
// 		combinedResult, err = intResult.Mul(fracResult)
// 	}

// 	if err != nil {
// 		return 0, err
// 	}

// 	return combinedResult.ToUFix64()

// 	// prod, _ := scaledLn.Mul(b_frac)

// 	// exp, k, isNeg, err := prod.scaledExp()

// 	// if err != nil {
// 	//     return 0, err
// 	// }

// 	// // What we want to do here is unscale exp, and then multiply by intResult.
// 	// // However, if we unscale exp in the space of a UFix64, we will lose enough
// 	// // precision to cause problems if intResult is large enough. So, we replicate
// 	// // the logic from unscaleExp() here, but we do it in the space of a UFix128
// 	// seriesScale := UFix128One.intMul(int64(ufix64_ln2Multiplier))
// 	// exp128 := exp.ToUFix128()
// 	// var fracResult UFix128

// 	// if !isNeg {
// 	//     // Our inner loop computed the result that needs to be multiplied by 2^k
// 	//     // and scaled down by seriesScale. We can resolve both of those with a single
// 	//     // FMD call.
// 	//     fracResult, err = exp128.FMD(UFix128One.shiftLeft(k), seriesScale)
// 	// } else {
// 	//     // We want the inverse of the the non-negative case. So, before
// 	//     // we wanted (sum * 2^k) / seriesScale, we want seriesScale / (sum * 2^k).
// 	//     // We can rearange this to be (seriesScale >> k) / sum.
// 	//     fracResult, err = seriesScale.shiftRight(k).FMD(UFix128One, exp128)
// 	// }

// 	// if err != nil {
// 	//     return 0, err
// 	// }

// 	// // HACK TIME: This should be an alternative to the code above, instead of just overwriting
// 	// // fracResult
// 	// if prod > 0 && prod < Fix64One.intMul(int64(ufix64_ln2Multiplier)/10) {
// 	//     // We know that prod is a fractional value < 0.1. We use an more precise expansion
// 	//     // for this very small value. We start with the same Taylor series as above:
// 	//     // e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...
// 	//     // However, we make a couple of key changes:
// 	//     // 1. We ignore the first term in the expansion (the 1). We can account for that later.
// 	//     // 2. We scale the input to use a scale multiple of 10^20. This is actually _larger_ than
// 	//     //    2^64, but since all of the terms in our expansion below will be less than 0.1,
// 	//     //    each term can still be resresented in 2^64.

// 	//     // Scale up by 10^12 so that the total result is scaled up by 10^20.
// 	//     prodHi, prodLo := bits.Mul64(uint64(scaledLn), uint64(b_frac*1e4))
// 	//     // Take out the ufix64_ln2Multiplier factor which we don't need anymore.
// 	//     fractionalProd, _ := bits.Div64(prodHi, prodLo, ufix64_ln2Multiplier)

// 	//     term := fractionalProd
// 	//     sum := fractionalProd
// 	//     iter := uint64(2)

// 	//     for {
// 	//         var r uint64
// 	//         // Multiply another power of x into the term
// 	//         hi, lo := bits.Mul64(term, fractionalProd)

// 	//         // Because we just multiplied two numbers together that are each scaled
// 	//         // up by 2^20, the 128-bit result, (hi, lo) is scaled up by 2^40. We can't
// 	//         // actually divide out the extra 2^20 with bits.Div64, because the denominator
// 	//         // is bigger than 2^64! Fortunately, we know that 2^20 = 2^20 * 5^20, so we can
// 	//         // shift by 20 bits, and the divide by 5^20, which is 95367431640625. Even better,
// 	//         // because we are using this smaller denominator, we can include the iter division
// 	//         // as part of the same operation. (iter is always a positive integer, and this loop
// 	//         // converges LONG before iter gets large enough to overflow.)

// 	//         lo = hi<<(64-20) | (lo >> 20)
// 	//         hi = hi >> 20
// 	//         denom := uint64(95367431640625 * iter)
// 	//         term, r = bits.Div64(hi, lo, denom)

// 	//         if r*2 >= denom {
// 	//             term++
// 	//         }

// 	//         // Break out of the loop when the term is too small to change the sum.
// 	//         if term == 0 {
// 	//             break
// 	//         }

// 	//         // Add the current term to the sum, we can use basic arithmetic
// 	//         sum += term
// 	//         iter += 1
// 	//     }

// 	//     // We have e^prod - 1 now, scaled up by 10^20. We can very easily turn this
// 	//     // into a UFix128 by multiplying by 10^4, to get the 10^24 scaling factor used
// 	//     // for UFix128.
// 	//     hi, lo := bits.Mul64(sum, 10000)
// 	//     fracResult = UFix128{Hi: hi, Lo: lo}

// 	//     // Now we can add in the 1.0 we ignored earlier!
// 	//     fracResult, _ = fracResult.Add(UFix128One)
// 	// }

// }
