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
	hi, lo := bits.Mul64(uint64(a), uint64(b))

	// If the high part of the product is greater than the scale factor, the
	// result won't fit into 64 bits after scaling back down.
	if hi >= Fix64Scale {
		return 0, ErrOverflow
	}

	// If the high part is 0, but the low part is non-zero but still less than the scale factor,
	// the result will round to zero when we scale it back down, we flag this as underflow.
	if hi == 0 && lo != 0 && lo < Fix64Scale {
		return 0, ErrUnderflow
	}

	quo, rem := bits.Div64(hi, lo, Fix64Scale)

	if rem >= (Fix64Scale / 2) {
		// If the remainder is at least half the scale factor, we round up.
		quo++
	}

	return UFix64(quo), nil
}

// Shared logic used by Mul, Div, and FMD that turns an unsigned result (including possible error)
// and a sign flag into a signed result or appropriate error (for a signed Fix64).
func resolveSign(val UFix64, err error, sign int64) (Fix64, error) {
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
	if sign < 0 && val == 0x8000000000000000 {
		return math.MinInt64, nil
	}

	if val > math.MaxInt64 {
		if sign < 0 {
			return 0, ErrNegOverflow
		} else {
			return 0, ErrOverflow
		}
	}

	return Fix64(int64(val) * sign), nil
}

func (a Fix64) Mul(b Fix64) (Fix64, error) {
	sign := int64(1)

	if a < 0 {
		sign = -sign
		a = -a
	}
	if b < 0 {
		sign = -sign
		b = -b
	}

	prod, err := UFix64(a).Mul(UFix64(b))

	return resolveSign(prod, err, sign)
}

// DivUFix64 returns a / b, errors on division by zero and overflow.
func (a UFix64) Div(b UFix64) (UFix64, error) {
	// Must come before the check for a == 0 so we flag 0.0/0.0 as an error.
	if b == 0 {
		return 0, ErrDivByZero
	}

	if a == 0 {
		return 0, nil
	}

	// The bits.Div64 accepts 128-bits for the dividend so we can apply
	// the scale factor BEFORE we divide by using the bits.Mul64 function
	// (which returns a 128-bit result).
	//
	// We're starting with (a * scale) and (b * scale) and we want to end
	// up with (a / b) * scale. The concatended hi-lo values here are equivalent
	// to be equal to (a * scale * scale). When we divide by (b * scale) we'll
	// get our desired result.
	hi, lo := bits.Mul64(uint64(a), Fix64Scale)

	// If the high part of the dividend is greater than the divisor, the
	// result won't fit into 64 bits.
	if hi >= uint64(b) {
		return 0, ErrOverflow
	}

	quo, rem := bits.Div64(hi, lo, uint64(b))

	// We want to round up if the remainder is at least half the divisor.
	// However, if c == 1, we get a false positive here (since 1/2 == 0),
	// so have to check for that specifically.
	if b != 1 && rem >= uint64(b)/2 {
		quo++
	}

	// We can't get here if a == 0 because we checked that first. So,
	// a quotient of 0 means the result is too small to represent, i.e. underflow.
	if quo == 0 {
		return 0, ErrUnderflow
	}

	return UFix64(quo), nil
}

func (a Fix64) Div(b Fix64) (Fix64, error) {
	sign := int64(1)

	if a < 0 {
		sign = -sign
		a = -a
	}
	if b < 0 {
		sign = -sign
		b = -b
	}

	quo, err := UFix64(a).Div(UFix64(b))

	return resolveSign(quo, err, sign)
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
	// However, if c == 1, we get a false positive here (since 1/2 == 0),
	// so have to check for that specifically.
	if c != 1 && rem >= uint64(c)/2 {
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

	res, err := UFix64(a).FMD(UFix64(b), UFix64(c))

	return resolveSign(res, err, sign)
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
	xHi, xLo := bits.Mul64(uint64(x), uint64(Fix64Scale))

	for {
		// This division can't fail: est is always a positive value somewhere between
		// x and 1, so it est will also be between x and 1.
		quo, rem := bits.Div64(xHi, xLo, est)

		if rem*2 >= uint64(est) {
			// If the remainder is at least half the divisor, we round up.
			quo++
		}

		// This swap originally started as an easy way not use unsigned arithmetic
		// to compare the two values, but empircal testing indicated that by always
		// approaching the true value "from below", the algorithm converges to the
		// correct value _including rounding_. When this loop used signed arithmetic
		// it would sometimes overshoot and end up with a value that was off by one.
		if est > quo {
			est, quo = quo, est
		}

		// We take the difference using basic arithmetic, since we know that quo
		// and est are close to each other and far away from zero, so the difference
		// will never overflow or underflow a signed int (although it can be negative).
		diff := quo - est

		diff = diff / 2

		// See if we've converged
		if diff == 0 {
			break
		}

		// Again, we can add the diff using simple signed arithmetic, since we know that
		// the diff value is small relative to est.
		est += diff
	}

	return UFix64(est), nil
}
