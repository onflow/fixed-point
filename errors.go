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

// PositiveOverflowError is reported when the value is positive and has a magnitude that is
// too large to be represented using the given bit length.
type PositiveOverflowError struct{}

var _ error = PositiveOverflowError{}

func (PositiveOverflowError) Error() string {
	return "overflow"
}

// NegativeOverflowError is reported when the value is negative and has a magnitude that is
// too large to be represented using the given bit length.
type NegativeOverflowError struct{}

var _ error = NegativeOverflowError{}

func (NegativeOverflowError) Error() string {
	return "negative overflow"
}

// UnderflowError is reported when the magnitude of the value is too small to be represented
// using the given bit length.
type UnderflowError struct{}

var _ error = UnderflowError{}

func (UnderflowError) Error() string {
	return "underflow"
}

type DivisionByZeroError struct{}

var _ error = DivisionByZeroError{}

func (DivisionByZeroError) Error() string {
	return "division by zero"
}

type OutOfDomainErrorError struct{}

var _ error = OutOfDomainErrorError{}

func (OutOfDomainErrorError) Error() string {
	return "input out of domain"
}

func applySign(e error, sign int64) error {
	if _, isUnderflowErr := e.(PositiveOverflowError); isUnderflowErr && sign < 0 {
		return NegativeOverflowError{}
	}

	return e
}
