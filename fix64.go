package fixedPoint

import (
	"math"
	"math/bits"
)

type UFix64 uint64
type Fix64 int64

const Fix64Scale = 1e8
const fix64One Fix64 = Fix64Scale

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

	quo, _ := bits.Div64(hi, lo, Fix64Scale)

	return UFix64(quo), nil
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
	if sign < 0 && prod == 0x8000000000000000 {
		return math.MinInt64, nil
	}

	if prod > math.MaxInt64 {
		if sign < 0 {
			return 0, ErrNegOverflow
		} else {
			return 0, ErrOverflow
		}
	}

	return Fix64(int64(prod) * sign), nil
}

// DivUFix64 returns a / b, errors on division by zero and overflow.
func (a UFix64) Div(b UFix64) (UFix64, error) {
	if b == 0 {
		// Must come before the check for a == 0 so we flag 0.0/0.0 as an error.
		return 0, ErrDivByZero
	}

	if a == 0 {
		return 0, nil
	}

	// The bits.Div64 accepts 128-bits for the dividend, and we can apply
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

	quo, _ := bits.Div64(hi, lo, uint64(b))

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

	if err != nil {
		if err == ErrOverflow && sign < 0 {
			return 0, ErrNegOverflow
		} else {
			return 0, err
		}
	}

	// Special case: if the result's sign should be negative and the quotient is 0x8000000000000000,
	// the result is valid and equal to math.MinInt64. (Note that 0x8000000000000000 is
	// LARGER than math.MaxInt64, so this requires special handling.)
	if sign < 0 && quo == 0x8000000000000000 {
		return math.MinInt64, nil
	}

	if quo > math.MaxInt64 {
		if sign < 0 {
			return 0, ErrNegOverflow
		} else {
			return 0, ErrOverflow
		}
	}
	return Fix64(int64(quo) * sign), nil
}

// returns a*b/c without intermediate rounding.
func (a UFix64) FMD(b, c UFix64) (UFix64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	if c == 0 {
		return 0, ErrDivByZero
	}

	hi, lo := bits.Mul64(uint64(a), uint64(b))

	// If the divsor isn't at least as big as the high part of the product,
	// the result won't fit into 64 bits.
	if uint64(c) <= hi {
		return 0, ErrOverflow
	}

	quo, _ := bits.Div64(hi, lo, uint64(c))

	if quo == 0 {
		return 0, ErrUnderflow
	}

	return UFix64(quo), nil
}

// returns a*b/c without intermediate rounding.
func (a Fix64) FMD(b, c Fix64) (Fix64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}

	if c == 0 {
		return 0, ErrDivByZero
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

	hi, lo := bits.Mul64(uint64(a), uint64(b))

	// If the divsor isn't at least as big as the high part of the product,
	// the result won't fit into 64 bits.
	if uint64(c) <= hi {
		return 0, ErrOverflow
	}

	quo, _ := bits.Div64(hi, lo, uint64(c))

	if quo == 0 {
		return 0, ErrUnderflow
	}

	return Fix64(int64(quo) * sign), nil
}

// func (x Fix64) Exp() (Fix64, error) {
// 	var err error

// 	// If x is 0, return 1.
// 	if x == 0 {
// 		return fix64One, nil
// 	}

// 	// TODO: These bounds could be tightened...
// 	// Values over e^26 are too large to represent in a Fix64, and values under
// 	// e^-18 are too small.
// 	if x >= 26*fix64One {
// 		return 0, ErrOverflow
// 	} else if x < -18*fix64One {
// 		return 0, ErrUnderflow
// 	}

// 	// If the input is negative, we could use the Taylor series below, but it
// 	// turns out that for negative numbers, the series diverges for some number
// 	// of iterations before starting to converge. Instead, we use the fact that
// 	// e^x is equal to 1/e^-x; positive exponents converge relatively quickly.
// 	inverted := false

// 	if x < 0 {
// 		inverted = true
// 		x = -x
// 	}

// 	// Use the higher precision fix64_12 type for the Taylor series approximation.
// 	x_12 := fixToTwelve(x)

// 	// We scale the input down somewhat so we can fit the *result* inside our
// 	// internal fix64_12 type. For each power of 2 that we shift the input down,
// 	// we will compensate at the end by squaring.
// 	scaleFactor := 0

// 	// NOTE: Because the highest value that can get this far is 26 (see above)
// 	//       this loop will run at most twice. As above, this bound could be
// 	//       tightened.
// 	for x_12 > 8*fix64_12One {
// 		x_12 = x_12 >> 1
// 		scaleFactor += 1
// 	}

// 	term := fix64_12One
// 	sum := fix64_12One

// 	// Use the Taylor series to compute e^x. The series is:
// 	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...
// 	// This loop tends to converge in 20-40 iterations, but we hardcap to 50
// 	// to avoid runaway.
// 	for i := int64(1); i < 50; i++ {
// 		// Multiply in another power of x to the term.
// 		term, err = mulFix64_12(term, x_12)
// 		if err != nil {
// 			return 0, err
// 		}

// 		// Divide by the iteration number to account for the factorial.
// 		// We can use simple division here because we know that we are
// 		// dividing by an integer constant that can't overflow.
// 		term = fix64_12(int64(term) / i)

// 		// Break out of the loop when the term is too small to change the sum.
// 		// TODO: We should probably break out of this loop when term < 1000
// 		// (which is the smallest fix64_12 value representable in Fix64)
// 		if term == 0 {
// 			break
// 		}

// 		// Add the current term to the sum, we can use basic arithmetic
// 		// because we know we are adding converging terms that can't overflow.
// 		sum += term
// 	}

// 	// Convert the result back to the lower precision Fix64 type.
// 	result := twelveToFix(sum)

// 	// Unwind our scaling factor by squaring the result as necessary.
// 	for scaleFactor > 0 {
// 		result, err = result.Mul(result)

// 		if err != nil {
// 			return 0, err
// 		}
// 		scaleFactor -= 1
// 	}

// 	// If we inverted the result on the way in, return 1/result.
// 	if inverted {
// 		result, err = fix64One.Div(result)
// 		if err != nil {
// 			return 0, err
// 		}
// 	}

// 	return result, nil
// }

// func (a Fix64) Pow(b Fix64) (Fix64, error) {
// 	if a == 0 {
// 		return 0, nil
// 	}
// 	if b == fix64One {
// 		return a, nil
// 	}

// 	lnA, err := UFix64(a).Ln()
// 	if err != nil {
// 		return 0, err
// 	}
// 	lnA_times_b, err := lnA.Mul(b)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return lnA_times_b.Exp()
// }

// const fix64PI = Fix64(314159265)

// func clampAngle(x Fix64) Fix64 {
// 	x = x % (fix64PI * 2)

// 	if x > fix64PI {
// 		x -= fix64PI * 2
// 	} else if x <= -fix64PI {
// 		x += fix64PI * 2
// 	}

// 	return x
// }

// func (x Fix64) Sin() (Fix64, error) {
// 	// Remove multiples of 2*pi from the input, and reduce the input to the range
// 	// (-pi, pi].
// 	x = clampAngle(x)

// 	if x == 0 {
// 		return 0, nil
// 	}

// 	// We will compute sin(x) using the approximation:
// 	// sin(x) = x - x^3/3! + x^5/5! - x^7/7! + ...

// 	// If x is small, we can just return x since sin(x) is approximately x for small x.
// 	// In fact, any value less than 391486/100000000 (0.00391486) will round to 0 in the
// 	// space of a Fix64 when we do the first step below (x^3/3!).
// 	if int64(x.Abs()) < 391486 {
// 		return x, nil
// 	}

// 	x_12 := fixToTwelve(x)

// 	// Can't fail; can't overflow (x is between -pi and pi), can't underflow (we checked for small values above)
// 	x_squared, _ := mulFix64_12(x_12, x_12)

// 	sum := x_12
// 	term := x_12
// 	iter := int64(1)

// 	for {
// 		var err error

// 		term = -term

// 		term, err = mulFix64_12(term, x_squared)
// 		if err == ErrUnderflow {
// 			break
// 		} else if err != nil {
// 			return 0, err
// 		}

// 		term, err = divFix64_12(term, fix64_12((iter+1)*(iter+2)*fix64_12Scale))
// 		if err == ErrUnderflow {
// 			break
// 		} else if err != nil {
// 			return 0, err
// 		}

// 		iter += 2

// 		// Both sum and term are always close to zero, so we don't need to worry
// 		// about overflow and can use the simple + operator.
// 		sum = sum + term
// 	}

// 	return twelveToFix(sum), nil
// }

// func addHalfPi(x Fix64) Fix64 {
// 	if x > 0 {
// 		// We subtract 3/2*pi from x when x is positive to prevent overflow.
// 		x = x - (3 * fix64PI / 2)
// 	} else {
// 		x = x + (fix64PI / 2)
// 	}

// 	return x
// }

// func (x Fix64) Cos() (Fix64, error) {
// 	if x == 0 {
// 		return fix64One, nil
// 	} else {
// 		// cos(x) = sin(x + π/2)
// 		x = addHalfPi(x)
// 		return x.Sin()
// 	}
// }

// func (x Fix64) Tan() (Fix64, error) {
// 	sinX, err := x.Sin()
// 	if err != nil {
// 		return 0, err
// 	}
// 	cosX, err := x.Cos()
// 	if err != nil {
// 		return 0, err
// 	}

// 	// TODO: This would be more accurate in the last bit or so if we had internal
// 	// versions of Sin and Cos that returned fix64_12...
// 	return sinX.Div(cosX)
// }

// func (x Fix64) Atan() (Fix64, error) {
// 	if x == 0 {
// 		return 0, nil
// 	}

// 	var err error

// 	neg := false

// 	if x < 0 {
// 		neg = true
// 		x = -x
// 	}

// 	inverse := false

// 	if x > fix64One {
// 		inverse = true
// 		x, err = Fix64(fix64One).Div(x)
// 		if err != nil {
// 			return 0, err
// 		}
// 	}

// 	// We now have x in the range [0, 1).

// 	// We will compute atan(x) using the approximation:
// 	// atan(x) = x - x^3/3 + x^5/5 - x^7/7 + ...
// 	x_squared, err := x.Mul(x)
// 	if err != nil {
// 		return 0, err
// 	}

// 	sum := x
// 	term := x
// 	div := int64(1)

// 	for {
// 		var err error

// 		term = -term

// 		term, err = term.Mul(x_squared)
// 		if err == ErrUnderflow {
// 			break
// 		} else if err != nil {
// 			return 0, err
// 		}

// 		div += 2

// 		// We can use basic arithmetic here since we know that we are dividing
// 		// by an integer constant that can't overflow.
// 		term = Fix64(int64(term) / div)

// 		// We've converged!
// 		if term == 0 {
// 			break
// 		}

// 		sum, _ = sum.Add(term)
// 	}

// 	if inverse {
// 		sum, err = Fix64(fix64PI / 2).Sub(sum)
// 		if err != nil {
// 			return 0, err
// 		}
// 	}

// 	if neg {
// 		sum = -sum
// 	}

// 	return sum, nil
// }

// func (x Fix64) Asin() (Fix64, error) {
// 	if x < -fix64One || x > fix64One {
// 		return 0, ErrDomain
// 	}

// 	x_squared, err := x.Mul(x)
// 	if err != nil {
// 		return 0, err
// 	}

// 	one_minus_x_squared, err := UFix64(fix64One).Sub(UFix64(x_squared))
// 	if err != nil {
// 		return 0, err
// 	}

// 	sqrt_one_minus_x_squared, err := one_minus_x_squared.Sqrt()
// 	if err != nil {
// 		return 0, err
// 	}
// 	div, err := x.Div(Fix64(sqrt_one_minus_x_squared))
// 	if err != nil {
// 		return 0, err
// 	}

// 	atanX, err := div.Atan()

// 	if err != nil {
// 		return 0, err
// 	}

// 	return atanX, nil
// }

// func (x Fix64) Acos() (Fix64, error) {
// 	if x < -fix64One || x > fix64One {
// 		return 0, ErrDomain
// 	}

// 	asinX, err := x.Asin()
// 	if err != nil {
// 		return 0, err
// 	}

// 	return fix64PI/2 - asinX, nil
// }
