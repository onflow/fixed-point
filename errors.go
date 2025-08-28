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

import "errors"

var (
	ErrOverflow    = errors.New("fixedPoint: overflow")
	ErrNegOverflow = errors.New("fixedPoint: negative overflow")
	ErrUnderflow   = errors.New("fixedPoint: underflow")
	ErrDivByZero   = errors.New("fixedPoint: division by zero")
	ErrDomain      = errors.New("fixedPoint: input out of domain")
)

func applySign(e error, sign int64) error {
	if e == ErrOverflow && sign < 0 {
		return ErrNegOverflow
	} else {
		return e
	}
}
