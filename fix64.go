package fixedPoint

import (
	"math"
	"math/bits"
)

type UFix64 uint64
type Fix64 int64

const fix64Scale = 1e8
const fix64One Fix64 = fix64Scale

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

func unsignedMulHelper(a, b, scale uint64) (uint64, error) {
	hi, lo := bits.Mul64(uint64(a), uint64(b))

	if hi >= scale {
		return 0, ErrOverflow
	}

	if hi == 0 && lo != 0 && lo < scale {
		return 0, ErrUnderflow
	}

	quo, _ := bits.Div64(hi, lo, scale)

	return quo, nil
}

func signedMulHelper(a, b int64, scale uint64) (int64, error) {
	sign := int64(1)

	if a < 0 {
		sign = -sign
		a = -a
	}
	if b < 0 {
		sign = -sign
		b = -b
	}

	result, err := unsignedMulHelper(uint64(a), uint64(b), scale)
	if result > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return int64(result) * sign, err
}

func (a UFix64) Mul(b UFix64) (UFix64, error) {
	result, err := unsignedMulHelper(uint64(a), uint64(b), fix64Scale)
	return UFix64(result), err
}

func (a Fix64) Mul(b Fix64) (Fix64, error) {
	result, err := signedMulHelper(int64(a), int64(b), fix64Scale)
	return Fix64(result), err
}

func unsignedDivHelper(a, b, scale uint64) (uint64, error) {
	if a == 0 {
		return 0, nil
	}

	if b == 0 {
		return 0, ErrDivByZero
	}

	// The bits.Div64 accepts 128-bits for the dividend, and we can apply
	// the scale factor BEFORE we divide by using the bits.Mul64 function
	// (which returns a 128-bit result).
	//
	// We're starting with (a * scale) and (b * scale) and we want to end
	// up with (a / b) * scale. The concatended hi-lo values here are equivalent
	// to be equal to (a * scale * scale). When we divide by (b * scale) we'll
	// get our desired result.
	hi, lo := bits.Mul64(a, scale)

	// If the high part of the dividend is greater than the divisor, the
	// result won't fit into 64 bits.
	if hi >= b {
		return 0, ErrOverflow
	}

	quo, _ := bits.Div64(hi, lo, uint64(b))

	// We can't get here if a == 0 because we checked that first. So,
	// a quotient of 0 means the result is too small to represent, i.e. underflow.
	if quo == 0 {
		return 0, ErrUnderflow
	}

	return quo, nil
}

func signedDivHelper(a, b int64, scale uint64) (int64, error) {
	sign := int64(1)

	if a < 0 {
		sign = -sign
		a = -a
	}
	if b < 0 {
		sign = -sign
		b = -b
	}

	quo, err := unsignedDivHelper(uint64(a), uint64(b), scale)
	if err != nil {
		return 0, err
	}

	// TODO: Technically, we should accept a value that is one bigger if
	// the sign is negative, since the magnitude of the most negative int64 is one
	// bigger than the most positive int64.
	if quo > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return int64(quo) * sign, err
}

// DivUFix64 returns a / b, errors on division by zero and overflow.
func (a UFix64) Div(b UFix64) (UFix64, error) {
	quo, err := unsignedDivHelper(uint64(a), uint64(b), fix64Scale)
	return UFix64(quo), err
}

func (a Fix64) Div(b Fix64) (Fix64, error) {
	quo, err := signedDivHelper(int64(a), int64(b), fix64Scale)
	return Fix64(quo), err
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

// An internal signed fixed-point type for numbers close to 1.
// Its denominator is 1e12, giving it an extra 4 decimal places of precision.
// This increases the precision of our transcendental calculations.
type fix64_12 int64

const fix64_12Scale = 1e12
const fix64_12One fix64_12 = fix64_12Scale

func addFix64_12(a, b fix64_12) (fix64_12, error) {
	// Adding two Fix64_12 numbers is equivalent to adding two Fix64 numbers
	sum, e := Fix64(a).Add(Fix64(b))
	return fix64_12(sum), e
}

func mulFix64_12(a, b fix64_12) (fix64_12, error) {
	product, e := signedMulHelper(int64(a), int64(b), fix64_12Scale)
	return fix64_12(product), e
}

func divFix64_12(a, b fix64_12) (fix64_12, error) {
	quo, err := signedDivHelper(int64(a), int64(b), fix64_12Scale)
	return fix64_12(quo), err
}

func fixToTwelve(x Fix64) fix64_12 {
	return fix64_12(x * (fix64_12Scale / fix64Scale))
}

func twelveToFix(x fix64_12) Fix64 {
	return Fix64(x / (fix64_12Scale / fix64Scale))
}

func (x UFix64) Sqrt() (UFix64, error) {
	if x == 0 {
		return 0, nil
	}

	n := bits.Len64(uint64(x))

	// We start our estimate with a number that has a bit length
	// halfway between the original number and the fixed-point representation
	// of 1. This will be of the same order of magnitude as the square root.
	n = (n + 27) / 2
	est := Fix64(1 << n)

	// Use Newton's method to approximate the square root.
	for {
		quo, err := x.Div(UFix64(est))
		if err != nil {
			return 0, err
		}

		diff, err := Fix64(quo).Sub(est)
		if err != nil {
			return 0, err
		}

		// When error is too small to matter, we can stop.
		if uint64(diff.Abs()) < 3 {
			break
		}

		est, err = est.Add(diff / 2)
		if err != nil {
			return 0, err
		}
	}

	return UFix64(est), nil
}

func (x UFix64) Ln() (Fix64, error) {
	if x <= 0 {
		return 0, ErrDomain
	}

	// The Taylor expansion of ln(x) converges faster for values closer to 1.
	// If we scale the input to have exactly 37 leading zero bits, the input will
	// be a number in the range (0.67108864, 1.33584262). (0.67108864 = 1 << 36 / scale)
	// For every power of two removed (or added) by this shift, we add (or subtract)
	// a multiple of ln(2) at the end of the function.
	leadingZeros := bits.LeadingZeros64(uint64(x))
	k := 37 - leadingZeros

	if k > 0 {
		x = x >> k
	} else if k < 0 {
		x = x << -k
	}

	// We use the higher precision fix64_12 type for the Taylor series approximation.
	x_12 := fixToTwelve(Fix64(x))

	// We will compute ln(x) using the approximation:
	// ln(x) = 2 * (z + z^3/3 + z^5/5 + z^7/7 + ...)
	// where z = (x - 1) / (x + 1)
	z, err := divFix64_12(x_12-fix64_12One, x_12+fix64_12One)
	if err != nil {
		return 0, err
	}

	z2, err := mulFix64_12(z, z)
	if err != nil {
		return 0, err
	}
	term := z
	sum := z
	iter := int64(1)

	// Keep interating until "term" and/or "next" rounds to zero (at the precision
	// of a Fix64).
	for {
		term, err = mulFix64_12(term, z2)

		if err == ErrUnderflow {
			break
		} else if err != nil {
			return 0, err
		}

		// We can use basic arithmetic here since we are dividing by a
		// an integer constant
		next := fix64_12(int64(term) / (iter*2 + 1))

		if next == 0 {
			break
		}

		// We can use basic arithmetic here since we know we are adding
		// small terms that can't overflow (positive or negative)
		sum += next
		iter += 1
	}

	// Multiply by 2 to account for the global 2x at the begining of the Tayler
	// expansion, and then add/subtract as many ln(2)s as required to account
	// for the scaling by 2^k we did at the beginning.
	ln2_fix64_12 := fix64_12(693147180559) // ln(2) * 1e12
	sum = sum*2 + fix64_12(k)*ln2_fix64_12

	// Convert the result back to the lower precision Fix64 type.
	return twelveToFix(sum), nil
}

func (x Fix64) Exp() (Fix64, error) {
	var err error

	// If x is 0, return 1.
	if x == 0 {
		return fix64One, nil
	}

	// TODO: These bounds could be tightened...
	// Values over e^26 are too large to represent in a Fix64, and values under
	// e^-18 are too small.
	if x >= 26*fix64One {
		return 0, ErrOverflow
	} else if x < -18*fix64One {
		return 0, ErrUnderflow
	}

	// If the input is negative, we could use the Taylor series below, but it
	// turns out that for negative numbers, the series diverges for some number
	// of iterations before starting to converge. Instead, we use the fact that
	// e^x is equal to 1/e^-x; positive exponents converge relatively quickly.
	inverted := false

	if x < 0 {
		inverted = true
		x = -x
	}

	// Use the higher precision fix64_12 type for the Taylor series approximation.
	x_12 := fixToTwelve(x)

	// We scale the input down somewhat so we can fit the *result* inside our
	// internal fix64_12 type. For each power of 2 that we shift the input down,
	// we will compensate at the end by squaring.
	scaleFactor := 0

	// NOTE: Because the highest value that can get this far is 26 (see above)
	//       this loop will run at most twice. As above, this bound could be
	//       tightened.
	for x_12 > 8*fix64_12One {
		x_12 = x_12 >> 1
		scaleFactor += 1
	}

	term := fix64_12One
	sum := fix64_12One

	// Use the Taylor series to compute e^x. The series is:
	// e^x = 1 + x + x^2/2! + x^3/3! + x^4/4! + ...
	// This loop tends to converge in 20-40 iterations, but we hardcap to 50
	// to avoid runaway.
	for i := int64(1); i < 50; i++ {
		// Multiply in another power of x to the term.
		term, err = mulFix64_12(term, x_12)
		if err != nil {
			return 0, err
		}

		// Divide by the iteration number to account for the factorial.
		// We can use simple division here because we know that we are
		// dividing by an integer constant that can't overflow.
		term = fix64_12(int64(term) / i)

		// Break out of the loop when the term is too small to change the sum.
		// TODO: We should probably break out of this loop when term < 1000
		// (which is the smallest fix64_12 value representable in Fix64)
		if term == 0 {
			break
		}

		// Add the current term to the sum, we can use basic arithmetic
		// because we know we are adding converging terms that can't overflow.
		sum += term
	}

	// Convert the result back to the lower precision Fix64 type.
	result := twelveToFix(sum)

	// Unwind our scaling factor by squaring the result as necessary.
	for scaleFactor > 0 {
		result, err = result.Mul(result)

		if err != nil {
			return 0, err
		}
		scaleFactor -= 1
	}

	// If we inverted the result on the way in, return 1/result.
	if inverted {
		result, err = fix64One.Div(result)
		if err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (a Fix64) Pow(b Fix64) (Fix64, error) {
	if a == 0 {
		return 0, nil
	}
	if b == fix64One {
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

const fix64PI = Fix64(314159265)

func clampAngle(x Fix64) Fix64 {
	x = x % (fix64PI * 2)

	if x > fix64PI {
		x -= fix64PI * 2
	} else if x <= -fix64PI {
		x += fix64PI * 2
	}

	return x
}

func (x Fix64) Sin() (Fix64, error) {
	// Remove multiples of 2*pi from the input, and reduce the input to the range
	// (-pi, pi].
	x = clampAngle(x)

	if x == 0 {
		return 0, nil
	}

	// We will compute sin(x) using the approximation:
	// sin(x) = x - x^3/3! + x^5/5! - x^7/7! + ...

	// If x is small, we can just return x since sin(x) is approximately x for small x.
	// In fact, any value less than 391486/100000000 (0.00391486) will round to 0 in the
	// space of a Fix64 when we do the first step below (x^3/3!).
	if int64(x.Abs()) < 391486 {
		return x, nil
	}

	x_12 := fixToTwelve(x)

	// Can't fail; can't overflow (x is between -pi and pi), can't underflow (we checked for small values above)
	x_squared, _ := mulFix64_12(x_12, x_12)

	sum := x_12
	term := x_12
	iter := int64(1)

	for {
		var err error

		term = -term

		term, err = mulFix64_12(term, x_squared)
		if err == ErrUnderflow {
			break
		} else if err != nil {
			return 0, err
		}

		term, err = divFix64_12(term, fix64_12((iter+1)*(iter+2)*fix64_12Scale))
		if err == ErrUnderflow {
			break
		} else if err != nil {
			return 0, err
		}

		iter += 2

		// Both sum and term are always close to zero, so we don't need to worry
		// about overflow and can use the simple + operator.
		sum = sum + term
	}

	return twelveToFix(sum), nil
}

func addHalfPi(x Fix64) Fix64 {
	if x > 0 {
		// We subtract 3/2*pi from x when x is positive to prevent overflow.
		x = x - (3 * fix64PI / 2)
	} else {
		x = x + (fix64PI / 2)
	}

	return x
}

func (x Fix64) Cos() (Fix64, error) {
	if x == 0 {
		return fix64Scale, nil
	} else {
		// cos(x) = sin(x + π/2)
		x = addHalfPi(x)
		return x.Sin()
	}
}

func (x Fix64) Tan() (Fix64, error) {
	sinX, err := x.Sin()
	if err != nil {
		return 0, err
	}
	cosX, err := x.Cos()
	if err != nil {
		return 0, err
	}

	// TODO: This would be more accurate in the last bit or so if we had internal
	// versions of Sin and Cos that returned fix64_12...
	return sinX.Div(cosX)
}

func (x Fix64) Atan() (Fix64, error) {
	if x == 0 {
		return 0, nil
	}

	var err error

	neg := false

	if x < 0 {
		neg = true
		x = -x
	}

	inverse := false

	if x > fix64Scale {
		inverse = true
		x, err = Fix64(fix64Scale).Div(x)
		if err != nil {
			return 0, err
		}
	}

	// We now have x in the range [0, 1).

	// We will compute atan(x) using the approximation:
	// atan(x) = x - x^3/3 + x^5/5 - x^7/7 + ...
	x_squared, err := x.Mul(x)
	if err != nil {
		return 0, err
	}

	sum := x
	term := x
	div := int64(1)

	for {
		var err error

		term = -term

		term, err = term.Mul(x_squared)
		if err == ErrUnderflow {
			break
		} else if err != nil {
			return 0, err
		}

		div += 2

		term, err = term.Div(Fix64(div * fix64Scale))
		if err == ErrUnderflow {
			break
		} else if err != nil {
			return 0, err
		}

		sum, _ = sum.Add(term)
	}

	if inverse {
		sum, err = Fix64(fix64PI / 2).Sub(sum)
		if err != nil {
			return 0, err
		}
	}

	if neg {
		sum = -sum
	}

	return sum, nil
}

func (x Fix64) Asin() (Fix64, error) {
	if x < -fix64Scale || x > fix64Scale {
		return 0, ErrDomain
	}

	x_squared, err := x.Mul(x)
	if err != nil {
		return 0, err
	}

	one_minus_x_squared, err := UFix64(fix64Scale).Sub(UFix64(x_squared))
	if err != nil {
		return 0, err
	}

	sqrt_one_minus_x_squared, err := one_minus_x_squared.Sqrt()
	if err != nil {
		return 0, err
	}
	div, err := x.Div(Fix64(sqrt_one_minus_x_squared))
	if err != nil {
		return 0, err
	}

	atanX, err := div.Atan()

	if err != nil {
		return 0, err
	}

	return atanX, nil
}

func (x Fix64) Acos() (Fix64, error) {
	if x < -fix64Scale || x > fix64Scale {
		return 0, ErrDomain
	}

	asinX, err := x.Asin()
	if err != nil {
		return 0, err
	}

	return fix64PI/2 - asinX, nil
}
