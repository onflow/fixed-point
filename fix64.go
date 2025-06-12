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

func lnInnerLoop(x_extra Fix64, scale Fix64) Fix64 {
	// We will compute ln(x) using the approximation:
	// ln(x) = 2 * (z + z^3/3 + z^5/5 + z^7/7 + ...)
	// where z = (x - 1) / (x + 1)

	num, _ := x_extra.Sub(scale)
	den, _ := x_extra.Add(scale)
	z, err := num.FMD(scale, den)

	if err == ErrUnderflow {
		// If z is too small to represent, we just return 0
		return 0
	}

	// Precompute z^2 to avoid recomputing it in the loop.
	z2, err := z.FMD(z, scale)

	if err == ErrUnderflow {
		// If z^2 is too small, we just return 2*z, which is the first term
		// in the series.
		return z.intMul(2)
	}

	term := z
	sum := z
	iter := int64(1)

	// Keep interating until "term" and/or "next" rounds to zero
	for {
		term, err = term.FMD(z2, scale)

		if err == ErrUnderflow {
			break
		}

		// We can use basic arithmetic here since we are dividing by a
		// an integer constant
		next := term.intDiv(iter*2 + 1)

		if next.isZero() {
			break
		}

		sum, _ = sum.Add(next)
		iter += 1
	}

	return sum.intMul(2)
}

func (x UFix64) Ln() (Fix64, error) {
	if x == 0 {
		return 0, ErrDomain
	}

	// The Taylor expansion of ln(x) converges faster for values closer to 1.
	// If we scale the input to have exactly 37 leading zero bits, the input will
	// be a number in the range (0.67108864, 1.33584262). (0.67108864 = 1 << 36 / scale)
	// For every power of two removed (or added) by this shift, we add (or subtract)
	// a multiple of ln(2) at the end of the function.
	leadingZeros := bits.LeadingZeros64(uint64(x))
	k := 37 - leadingZeros

	computationBits := 16

	shift := k - computationBits

	if shift > 0 {
		x = x >> shift
	} else if shift < 0 {
		x = x << -shift
	}

	res_scaled := lnInnerLoop(Fix64(x), Fix64One<<computationBits)

	// Add/subtract as many ln(2)s as required to account for the scaling by 2^k we
	// did at the beginning.
	powerCorrection, _ := Fix64(ufix64_ln2Multiple).FMD(Fix64(k<<computationBits), Fix64(ufix64_ln2Factor))
	res_scaled, _ = res_scaled.Add(powerCorrection)

	res_scaled += 1 << (computationBits - 1) // Add a half to round to the nearest integer
	return res_scaled >> computationBits, nil
}

func (x Fix64) Exp() (UFix64, error) {
	var err error

	// If x is 0, return 1.
	if x == 0 {
		return UFix64One, nil
	}

	// We can quickly check to see if the input will overflow or underflow
	if x > maxLn64 {
		return 0, ErrOverflow
	} else if x < minLn64 {
		return 0, ErrUnderflow
	}

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

	var unsignedX uint64
	var isNeg bool

	if x >= 0 {
		unsignedX = uint64(x)
		isNeg = false
	} else {
		unsignedX = uint64(-x)
		isNeg = true
	}

	scaledHi, scaledLo := bits.Mul64(unsignedX, ufix64_ln2Factor)

	quo, rem := bits.Div64(scaledHi, scaledLo, uint64(ufix64_ln2Multiple))

	k := int64(quo)

	// We now have a value k indicating the number of times we need to scale the result after
	// our inner loop, and a remainder that is the input to our Taylor series expansion.
	// The remainder is in the range [0, ln(2)) (~0.69) but is still multiplied by ufix64_ln2Factor.
	// We can run the inner loop with this multiplier in place, by using ufix64_ln2Factor times
	// UFixScale as the scale for our arithmetic operations.

	seriesScale := UFix64One.intMul(int64(ufix64_ln2Factor))
	seriesInput := UFix64(rem)

	// We use the Taylor series to compute e^x. The series is:
	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...

	// Remember: Although we are using the UFix64 type here, the actual scale factor
	// of these values is seriesScale, provided we are careful to only use functions
	// that are scale agnostic (lke FMD and simple addition), this won't bite us.
	term := seriesScale // Starts as 1.0 in the series
	sum := seriesScale  // Starts as 1.0 in the series
	iter := int64(1)

	// This loop converges in under 20 iterations in testing
	for {
		// Multiply another power of x into the term, using FMD with seriesScale
		// as the divisor will keep the result in the correct scale.
		term, err = term.FMD(seriesInput, seriesScale)

		if err == ErrUnderflow {
			// If the term is too small to represent, we can just break out of the loop.
			break
		} else if err != nil {
			return 0, err
		}

		// Divide by the iteration number to account for the factorial.
		// We can use simple division here because we know that we are
		// dividing by an integer constant that can't overflow.
		term = term.intDiv(iter)

		// Break out of the loop when the term is too small to change the sum.
		if term == 0 {
			break
		}

		// Add the current term to the sum, we can use basic arithmetic
		// because we know we are adding converging terms that won't overflow.
		// (The value that comes out of this loop will always be <= 2, since we
		// quantized the input to be in the range [0, ln(2)).
		sum += term
		iter += 1
	}

	// What we have now is the value of e^x with two overlapping factors:
	//    - The scale factor, k, which is the power of two that we need to
	//      multiply the result by to get the final value. Note that k can
	// 		be negative!
	//    - The seriesScale, which is the scale factor we used for the Taylor
	//      series expansion to maintain precision.
	//
	// We can resolve both of these factors with a single FMD call and a shift.

	if !isNeg {
		// Our inner loop computed the result that needs to be multiplied by 2^k
		// and scaled down by seriesScale. We can resolve both of those with a single
		// FMD call.
		return sum.FMD(UFix64One<<k, seriesScale)
	} else {
		// We want the inverse of the the non-negative case. So, before
		// we wanted (sum * 2^k) / seriesScale, we want seriesScale / (sum * 2^k).
		// We can rearange this to be (seriesScale >> k) / sum.
		return (seriesScale >> k).FMD(UFix64One, sum)
	}
}

func (a UFix64) Pow(b Fix64) (UFix64, error) {
	if a == 0 {
		return 0, nil
	}
	if b == 0 {
		return UFix64One, nil
	}
	if b == Fix64One {
		return a, nil
	}

	lnA, err := UFix64(a).Ln()
	if err != nil {
		return 0, err
	}
	lnA_times_b, err := lnA.Mul(b)
	if err != nil {
		return 0, err
	}
	return lnA_times_b.Exp()
}
