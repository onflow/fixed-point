# Archived Code

This file contains various code snippets that didn't make it into the final version of the code, but which might still be promising for future iterations. Rather than "hide them" in a git branch, I'm just going to drop them here. Maybe someone else can do something useful with them!

### Fix192 Inverse

A relatively fast geometric inverse function. Was used to compute tan(x) when x is close to π/2.

The number of iterations of the main loop (set to 8) was never properly calibrated. So its possible it's larger than it needs to be, or maybe too small for some values. Additionally, there might be a way to get a better initial estimate. However, it is quite fast and worked quite well.

```go
func (x fix192) inverse() (fix192, error) {

	if isNeg64(x.Hi) || x.isZero() {
		// If the input is negative or zero, we can't compute the inverse.
		return fix192{}, ErrDomain
	}

	// NOTE: Returns 128 if x == 0
	zeros := leadingZeroBits192(x)

	if zeros >= 96 {
		// It turns out that any value with more than 96 leading zero bits is so small that it's inverse
		// would overflow the 192-bit fixed-point representation.
		return fix192{}, ErrOverflow
	}

	// We use the following recursive formula to compute the inverse:
	// rₙ₊₁ = rₙ (2 - x·rₙ)

	// We start our estimate by taking the input and shifting it "past" the representation of one
	// by the same number of bits as the original distance. This should get us within a factor of 2
	// of the inverse we are looking for.
	fix192Two := fix192{0x00000000000001a784, 0x379d99db42000000, 0x0000000000000000}
	est := x

	if zeros > Fix128OneLeadingZeros {
		// The input has more leading zeros that one. Shift left by twice the difference.
		est = est.shiftLeft(uint64(zeros-Fix128OneLeadingZeros) * 2)
	} else if zeros < Fix128OneLeadingZeros {
		// The input has fewer leading zeros than one. Shift right by twice the difference.
		est = est.ushiftRight(uint64(Fix128OneLeadingZeros-zeros) * 2)
	}

	for i := 0; i < 8; i++ {
		prod, _ := est.umul(x)
		est, _ = est.umul(fix192Two.sub(prod))
	}

	return est, nil
}
```