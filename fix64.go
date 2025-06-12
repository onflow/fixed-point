package fixedPoint

import (
	"math"
	"math/bits"
)

func (a UFix64) Add(b UFix64) (UFix64, error) {
	sum, carry := bits.Add64(uint64(a), uint64(b), 0)
	if carry != 0 {
		return 0, ErrOverflow
	}
	return UFix64(sum), nil
}

func (a Fix64) Add(b Fix64) (Fix64, error) {
	sum := a + b

	// Check for overflow by checking the sign bits of the operands and the result.
	if a > 0 && b > 0 && sum < 0 {
		return 0, ErrOverflow
	} else if a < 0 && b < 0 && sum >= 0 {
		return 0, ErrNegOverflow
	}

	return sum, nil
}

func (a UFix64) Sub(b UFix64) (UFix64, error) {
	diff, borrow := bits.Sub64(uint64(a), uint64(b), 0)
	if borrow != 0 {
		return 0, ErrNegOverflow
	}
	return UFix64(diff), nil
}

func (a Fix64) Sub(b Fix64) (Fix64, error) {
	diff := a - b

	// Overflow occurs when:
	// 1. Subtracting a positive from a non-positive results in a positive
	// 2. Subtracting a negative from a non-negative results in a negative
	// Subtracting two non-zero values with the same sign can't overflow in a signed int64
	if a >= 0 && b < 0 && diff < 0 {
		return 0, ErrOverflow
	} else if a < 0 && b >= 0 && diff >= 0 {
		return 0, ErrNegOverflow
	}

	return Fix64(diff), nil
}

func (a Fix64) Abs() Fix64 {
	if a < 0 {
		return -a
	}
	return a
}

func (a UFix64) Mul(b UFix64) (UFix64, error) {
	// It might seem strange to implement multiplication in terms of fused multiply-divide,
	// but it turns out that a simple mulitiplication fixed-point operation needs to both
	// multiply and divide anyway. (Multiply the inputs, and then divide by the scale factor.)

	// Additionally, the logic for handling rounding is REALLY not trivial, so
	// having that in one location is a big win. In the end, the only real cost
	// is the overhead of an extra function call, which is negligible.
	return a.FMD(b, UFix64One)
}

func (a Fix64) Mul(b Fix64) (Fix64, error) {
	// Same rationale as above for UFix64.Mul, but even more critical because handling the
	// signs correctly is ALSO not trivial.
	return a.FMD(b, Fix64One)
}

func (a UFix64) Div(b UFix64) (UFix64, error) {
	// Same rationale as above for UFix64.Mul
	return a.FMD(UFix64One, b)
}

func (a Fix64) Div(b Fix64) (Fix64, error) {
	// Same rationale as above...
	return a.FMD(Fix64One, b)
}

// returns a*b/c without intermediate rounding.
func (a UFix64) FMD(b, c UFix64) (UFix64, error) {
	// Must come before the check for a or b == 0 so we flag 0.0/0.0 as an error.
	if c == 0 {
		return 0, ErrDivByZero
	}

	if a == 0 || b == 0 {
		return 0, nil
	}

	hi, lo := bits.Mul64(uint64(a), uint64(b))

	// If the divsor isn't at least as big as the high part of the product,
	// the result won't fit into 64 bits.
	if uint64(c) <= hi {
		return 0, ErrOverflow
	}

	quo, rem := bits.Div64(hi, lo, uint64(c))

	// We want to round up if the remainder is at least half the divisor.
	// This isn't as simple as just checking if rem >= c/2, because dividing c by two
	// *loses precision*. A more accurate solution would be to multiply the
	// remainder by 2 and compare it to c, but that can overflow uint64 if the
	// remainder is large.
	//
	// However, we KNOW the remainder is less than c, and we know that c fits in 64 bits,
	// so if the remainder is so large that multiplying it by 2 would overflow,
	// then it must be at least half of c. So, we first check to see if it WOULD
	// overflow when doubled (in which case it is definitely larger than c/2),
	// and otherwise we can safely multiply it by 2 and compare it to c.
	if rem > (math.MaxUint64/2) || (rem*2) >= uint64(c) {
		// Make sure we don't "round up" to a value outside of the range of UFix64!
		if quo == 0xffffffffffffffff {
			return 0, ErrOverflow
		}

		quo++
	}

	// We can't get here if a == 0 or b == 0 because we checked that first. So,
	// a quotient of 0 means the result is too small to represent, i.e. underflow.
	if quo == 0 {
		return 0, ErrUnderflow
	}

	return UFix64(quo), nil
}

// returns a*b/c without intermediate rounding.
func (a Fix64) FMD(b, c Fix64) (Fix64, error) {
	// Must come before the check for a or b == 0 so we flag 0.0/0.0 as an error.
	if c == 0 {
		return 0, ErrDivByZero
	}

	if a == 0 || b == 0 {
		return 0, nil
	}

	// Determine the sign of the result based on the signs of a, b, and c.
	sign := int64(1)

	if a < 0 {
		sign = -sign
		a = -a
	}
	if b < 0 {
		sign = -sign
		b = -b
	}
	if c < 0 {
		sign = -sign
		c = -c
	}

	// Compute the result using unsigned arithmetic.
	res, err := UFix64(a).FMD(UFix64(b), UFix64(c))

	if err != nil {
		if err == ErrOverflow && sign < 0 {
			return 0, ErrNegOverflow
		} else {
			return 0, err
		}
	}

	// Special case: if the result's sign should be negative and the product is 0x8000000000000000,
	// the result is valid and equal to math.MinInt64. (Note that 0x8000000000000000 is
	// LARGER than math.MaxInt64, so this requires special handling.)
	if sign < 0 && res == 0x8000000000000000 {
		return math.MinInt64, nil
	}

	if res > math.MaxInt64 {
		if sign < 0 {
			return 0, ErrNegOverflow
		} else {
			return 0, ErrOverflow
		}
	}

	return Fix64(int64(res) * sign), nil
}

func (x UFix64) Sqrt() (UFix64, error) {
	if x == 0 {
		return 0, nil
	}

	// Count the number of set bits in x, this is a cheap way of estimating
	// the order of magnitude of the input.
	n := bits.Len64(uint64(x))

	// We start our estimate with a number that has a bit length
	// halfway between the original number and the fixed-point representation
	// of 1. This will be of the same order of magnitude as the square root, allowing
	// our Newton-Raphson loop below to converge quickly.
	n = (n + 27) / 2
	est := uint64(1 << n)

	// The inner loop here will frequently divide the input by the current estimate,
	// so instead of using the Fix64.Div method, we expand the numerator once outside
	// the loop, and then directly call bits.Div64 in the loop.
	xHi, xLo := bits.Mul64(uint64(x), Fix64Scale)

	for {
		// This division can't fail: est is always a positive value somewhere between
		// x and 1, so it est will also be between x and 1.
		quo, rem := bits.Div64(xHi, xLo, est)

		if rem*2 >= uint64(est) {
			// If the remainder is at least half the divisor, we round up.
			quo++
		}

		// We take the difference using basic arithmetic, since we know that quo
		// and est are close to each other and far away from zero, so the difference
		// will never overflow or underflow a signed int (although it can be negative).
		diff := int64(quo) - int64(est)

		// If the difference is zero, we've converged cleanly.
		if diff == 0 {
			break
		}

		// If the difference is 1 or -1, we know that the correct answer is either
		// quo or est, but we can't be sure which one is closer! The easiest way to
		// be sure is to just square the two values and see which one is closer to
		// the original input.
		if diff == 1 {
			// Diff is positive, so quo is larger than est, and quo^2 will be larger than x

			// NOTE: This math could overflow with both the multiplication AND the subtraction,
			// but since we know the error will fit into a uint64, it'll all work out in the wash!
			estError := xLo - (est * est)
			quoError := (quo * quo) - xLo

			if estError > quoError {
				// If the estimate is further away, we can just use quo.
				est = quo
			}
			break
		} else if diff == -1 {
			// Diff is negative, so quo is smaller than est, and quo^2 will be smaller than x
			// Otherwise, this is the same logic as above.
			estError := (est * est) - xLo
			quoError := xLo - (quo * quo)

			if estError > quoError {
				// If the estimate is further away, we can just use quo.
				est = quo
			}
			break
		}

		diff = diff / 2

		// Again, we can add the diff using simple signed arithmetic, since we know that
		// the diff value is small relative to est.
		est = uint64(int64(est) + diff)
	}

	return UFix64(est), nil
}
