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

const scaleFactor64To128 = raw64(Fix128Scale / Fix64Scale)

// Converts a UFix64 to a UFix128, can't fail since UFix128 has a larger range than UFix64.
func (a UFix64) ToUFix128() UFix128 {
	hi, lo := mul64(raw64(a), scaleFactor64To128)

	return UFix128{Hi: hi, Lo: lo}
}

// Converts a Fix64 to a Fix128, can't fail since Fix128 has a larger range than Fix64.
func (a Fix64) ToFix128() Fix128 {
	unsignedX, sign := a.Abs()

	unsignedRes := unsignedX.ToUFix128()

	res, _ := unsignedRes.ApplySign(sign)

	return res
}

// Converts a UFix128 to a UFix64, returns an error if the value can't be represented in UFix64,
// including overflow and underflow cases.
func (x UFix128) ToUFix64(round RoundingMode) (UFix64, error) {
	// Return zero immediately when possible.
	if x.IsZero() {
		return UFix64Zero, nil
	}

	// Fix128 has a larger range than UFix64, so we check to see that this
	// value will fit in UFix64 after division
	if !ult64(x.Hi, scaleFactor64To128) {
		return UFix64Zero, OverflowError{}
	}

	quo, rem := div64(x.Hi, x.Lo, scaleFactor64To128)

	if ushouldRound64(quo, rem, scaleFactor64To128, round) {
		var carry uint64
		quo, carry = add64(quo, raw64Zero, 1)

		// If there's a carry, the rounding overflowed.
		if carry != 0 {
			return UFix64Zero, OverflowError{}
		}
	} else {
		// If the quotient is zero, the result is an overflow. (We had a fast return
		// at the top of the file to check for the case where the input was zero to
		// begin with.)
		if isZero64(quo) {
			return UFix64Zero, UnderflowError{}
		}
	}

	return UFix64(quo), nil
}

// Converts a Fix128 to a Fix64, returns an error if the value can't be represented in Fix64,
// including overflow, negative overflow, and underflow cases.
func (x Fix128) ToFix64(round RoundingMode) (Fix64, error) {
	unsignedX, sign := x.Abs()

	res, err := unsignedX.ToUFix64(round)

	if err != nil {
		return Fix64Zero, err
	}

	return res.ApplySign(sign)
}
