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

// Exported fixed-point types
type UFix64 raw64
type Fix64 raw64

type UFix128 raw128
type Fix128 raw128

// Internal types
type raw64 uint64
type raw128 struct {
	Hi raw64
	Lo raw64
}

func NewFix128(hi, lo uint64) Fix128 {
	return Fix128{
		Hi: raw64(hi),
		Lo: raw64(lo),
	}
}

func NewUFix128(hi, lo uint64) UFix128 {
	return UFix128{
		Hi: raw64(hi),
		Lo: raw64(lo),
	}
}
