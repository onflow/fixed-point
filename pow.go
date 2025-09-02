/*
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fixedPoint

// func (x Fix128) SinTest() (Fix128, error) {
// 	// Normalize the input angle to the range [0, π], with a flag indicating
// 	// if the result should be interpreted as negative.
// 	unsignedX, sign := x.Abs()
// 	x192 := unsignedX.toFix192()
// 	clampedX, sign := x192.clampAngle(sign)

// 	resFrac := clampedX.chebySin()

// 	res, _ := resFrac.toUFix128()

// 	return res.ApplySign(sign)
// }

// func (x Fix128) Exp() (UFix128, error) {
// 	// If x is 0, return 1.
// 	if x.IsZero() {
// 		return UFix128One, nil
// 	}

// 	// We can quickly check to see if the input will overflow or underflow
// 	if x.Gt(maxLn128) {
// 		return UFix128Zero, PositiveOverflowError{}
// 	} else if x.Lt(minLn128) {
// 		return UFix128Zero, UnderflowError{}
// 	}

// 	res192, err := x.toFix192().exp()

// 	if err != nil {
// 		return UFix128Zero, err
// 	}

// 	res, err := res192.toUFix128()

// 	if err != nil {
// 		return UFix128Zero, err
// 	}

// 	var UFix128OneTrillion = UFix128{0x00c097ce7bc90715, 0xb34b9f1000000000}

// 	if res.Gt(UFix128OneTrillion) {
// 		// 2^64 % 100 = 16, so we multiply the modulus of the high part by 16
// 		lastDigits := ((res.Hi%100)*16 + (res.Lo % 100)) % 100
// 		var roundedRes raw128
// 		if lastDigits >= 50 {
// 			// If the last digits are 50 or greater, we round up.
// 			roundedRes, _ = add128(raw128(res), raw128{0, 100 - lastDigits}, 0)
// 		} else {
// 			// Round down
// 			roundedRes, _ = sub128(raw128(res), raw128{0, lastDigits}, 0)
// 		}

// 		res = UFix128(roundedRes)
// 	}

// 	return res, nil
// }

// func (a UFix64) Pow(b Fix64) (UFix64, error) {
// 	// Use the 128-bit exp() function to compute the result
// 	a128 := a.ToUFix128()
// 	b128 := b.ToFix128()

// 	res128, err := a128.Pow(b128)

// 	if err != nil {
// 		return UFix64Zero, err
// 	}

// 	return res128.ToUFix64()
// }

// // Exp returns e^x, or an error on overflow or underflow.
// func (x Fix128) Exp() (UFix128, error) {
// 	// If x is 0, return 1.
// 	if x.IsZero() {
// 		return UFix128One, nil
// 	}

// 	// We can quickly check to see if the input will overflow or underflow
// 	if x.Gt(maxLn128) {
// 		return UFix128Zero, PositiveOverflowError{}
// 	} else if x.Lt(minLn128) {
// 		return UFix128Zero, UnderflowError{}
// 	}

// 	// Switch to unsigned representation to simplify the logic and give
// 	// us a larger range of values to work with.
// 	var unsignedX UFix128
// 	var isNeg bool

// 	if !x.IsNeg() {
// 		unsignedX = UFix128(x)
// 		isNeg = false
// 	} else {
// 		unsignedX = UFix128(x.intMul(-1))
// 		isNeg = true
// 	}

// 	// Now that we've limited the input to minLn and maxLn, we can multiply
// 	// by ln2Multiplier without fear of overflow (ln2Multiplier was specifically
// 	// chosen not to overflow an unsigned fix when multiplied by values in that range).
// 	xScaled := unsignedX.intMul(fix128LnMultiplier)

// 	// We can do the opposite of the trick we did in ln(x). There, we scaled the input
// 	// by powers of 2, and then added/subtracted multiples of ln(2) at the end. Here
// 	// we can add/subtract multiples of ln(2) at the beginning, and then scale by
// 	// the appropriate power of 2 at the end.

// 	// The first step is to do a division of the input by ln(2); the quotient will
// 	// tell us how many shifts we need to do at the end to scale the result of our
// 	// inner loop to the correct value, and the remainder will be the input to our
// 	// Taylor series expansion.
// 	//
// 	// Of course, "divide by ln(2)" isn't actually as simple as it sounds because
// 	// we can't represent ln(2) exactly in fix64, we need to use an approximation.
// 	// So, in order to introduce as small of an error as possible, we scale the
// 	// input AND ln(2) by a factor that minimizes the error. See genFactors.py for
// 	// more details.

// 	quo, rem := div128(raw128Zero, raw128(xScaled), raw128(fix128Ln2Scaled))

// 	k := uint64(quo.Lo)

// 	// We now have a value k indicating the number of times we need to scale the result after
// 	// our inner loop, and a remainder that is the input to our Taylor series expansion.
// 	// The remainder is in the range [0, ln(2)) (~0.69) using lnScale precision.
// 	seriesInput := UFix128(rem)
// 	seriesScale := ufix128LnScale

// 	// We use the Taylor series to compute e^x. The series is:
// 	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...

// 	// Remember: Although we are using the UFix type here, the actual scale factor
// 	// of these values is lnScale. Provided we are careful to only use functions
// 	// that are scale agnostic (lke FMD and simple addition), this won't bite us.
// 	term := seriesScale // Starts as 1.0 in the series
// 	sum := term
// 	iter := uint64(1)
// 	var err error

// 	// This loop converges in under 20 iterations in testing
// 	for {
// 		// Multiply another power of x into the term, using FMD with seriesScale
// 		// as the divisor will keep the result in the correct scale.
// 		term, err = term.FMD(seriesInput, seriesScale)

// 		if _, ok := err.(UnderflowError); ok {
// 			// If the term is too small to represent, we can just break out of the loop.
// 			break
// 		} else if err != nil {
// 			return UFix128Zero, err
// 		}

// 		// Divide by the iteration number to account for the factorial.
// 		term = term.intDiv(iter)

// 		// inDiv doesn't check for underflow, but a result of zero means
// 		// the term is too small to represent.
// 		if term.IsZero() {
// 			break
// 		}

// 		// Can't overflow
// 		sum, _ = sum.Add(term)
// 		iter += 1
// 	}

// 	var res UFix128

// 	// What we have now is the value of sum = e^x with three overlapping factors:
// 	//    - k, which is the power of two that we need to multiply the result by
// 	//      to get the final value.
// 	//    - lnScale, which is the scale factor we used for the Taylor series
// 	//      expansion to maintain precision.
// 	//    - isNeg, which is a flag indicating whether the input was negative. If it
// 	//      was, we need to return the multiplicative inverse of the sum.
// 	//
// 	// We resolve all of these factors with a FMD and bit shifting
// 	if !isNeg {
// 		// Our inner loop computed the result that needs to be multiplied by 2^k
// 		// and scaled down by seriesScale. We can resolve both of those with a single
// 		// FMD call.
// 		res, err = sum.FMD(UFix128One.shiftLeft(k), seriesScale)
// 	} else {
// 		// For negative input, we need the multiplicitive inverse of the result. The
// 		// positive result is (sum * 2^k) / lnScale, so the inverse is
// 		// (lnScale / (sum * 2^k)). We can rearange this to be (lnScale >> k) / sum.
// 		// Even better, we can save some precision by splitting the shift between
// 		// lnScale and FixOne and using FMD(lnScale, FixOne, sum) to compute the final result.
// 		// (Note that lnScale and FixOne each have the same number of trailing zero bits
// 		// because they both have the same large power of ten as a factor, and thus
// 		// the same large power of two as a factor.)
// 		oneShift := k / 2
// 		scaleShift := k - oneShift
// 		res, err = seriesScale.shiftRight(scaleShift).FMD(UFix128One.shiftRight(oneShift), sum)
// 	}

// 	if err != nil {
// 		return UFix128Zero, err
// 	}

// 	return res, nil
// }

// func (a UFix128) Pow(b Fix128) (UFix128, error) {
// 	// The order of these guards is important!

// 	// We accept 0^0 as 1.
// 	if b.IsZero() {
// 		return UFix128One, nil
// 	}

// 	if a.IsZero() {
// 		if b.IsNeg() {
// 			// 0^negative is undefined, so we return an error.
// 			return UFix128Zero, DivisionByZeroError{} // 0^negative is undefined
// 		} else {
// 			// 0^positive is 0.
// 			return UFix128Zero, nil
// 		}
// 	}

// 	if a.Eq(UFix128One) {
// 		// 1^b is always 1, so we can return it directly.
// 		return UFix128One, nil
// 	}

// 	// a^1 is just a, so we can return it directly.
// 	if b.Eq(Fix128One) {
// 		return a, nil
// 	}

// 	lnA, err := a.Ln()

// 	if err != nil {
// 		return UFix128Zero, err
// 	}

// 	prod, err := lnA.Mul(b)

// 	if err == ErrUnderflow {
// 		// If the product is too small to represent, the result will effectively be 1
// 		return UFix128One, nil
// 	} else if err == ErrOverflow {
// 		// The multiplication can easily overflow (due to fix64_ln2Multiplier)
// 		// if ln(a)•b is larger than 27. However, if that is true, then the exp()
// 		// call below would also overflow, so returning an overflow error here is
// 		// appropriate.
// 		return UFix128Zero, PositiveOverflowError{}
// 	} else if err == ErrNegOverflow {
// 		// If the product overflows in the negative direction, the exponential
// 		// would underflow.
// 		return UFix128Zero, UnderflowError{}
// 	}

// 	return prod.Exp()
// }
