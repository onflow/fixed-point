# Archived Code

This file contains various code snippets that didn't make it into the final version of the code, but which might still be promising for future iterations. Rather than "hide them" in a git branch, I'm just going to drop them here. Maybe someone else can do something useful with them!

## Chebyshev Approximation

There is an alternative to Taylor expansion that is frequently used in math libraries (at least, according to ChatGPT!) that uses polynomial approximation. The mpmath library for Python can compute the coeffficients for these polygons for any smooth function and in theory, we could use this for our sin() implementation, and possibly exp() and ln().

In theory, this should improve performance, because you will end up computing fewer terms of the polynomial than you will do iterations on the Taylor series, and the computational cost of each term will be comparable, or even slighly cheaper for the polynomial. It would also have the advantage that we could use the same inner loop for multiple functions, just passing in a different array of coeffficents for each computation.

In practice, this ended up creating problems. For sin(), if I used a polynomial optimized for the range up to π/2, I needed about 12 terms for Fix64. This meant that the highest x^n term could be as large as π/2^12. This is around 200, but the scale factor we used for π (fix64_twoPiScale) could only represent values up to 17. We could use a smaller scale factor, but smaller scale factors ended up with much larger error terms and introduced non-trivial errors into clampAngle(). The fix for this (which is included in the code below) is to limit the Chebyshev to values below π/4. By tightening the bounds, we end up needing fewer coefficients for the same precision AND we know that x is <1 so each power of x is smaller than the last and we have no worries about overflow. Unfortunately, the identity we use for mapping values above π/4 to values under π/4 is sin(x) = 2•sin(x/2)•cos(x/2). cos(x/2) can be approximated with sin(x/4), so we can just call sin() twice with smaller values. However, this doubles the cost of the function call for those values, which more than obliterates the computational savings of using the Chebyshev polynomial!

The upshot was that using Chebyshve ended with a larger average runtime, even though it could be made to be equally precise and had a (slightly) shorter runtime for values under π/4.

### Computing the Coefficients

This code was in constget.py to compute the Chebyshev coefficients. Picking the right maximum error value was tricky. You might think that just using fix64Epsilon would be sufficent, but in my testing, I found that sample inputs that required very high precision to get exact answers. For example, cos(0.0001) is equal to 0.9999999950000000041. So, an error less than 1e-17 can be enough to get the 8th decimal place wrong. 

```python
# We compute the Chebyshev polynomial for sin(x) using the utilites in mpmath.
# Since we are going to use a linear approximation for sin(x) for values
# less than fix64_extra_sinIota, we can start the approximation range there.
chebyDegree = 4

while True:
    chebyDegree += 1
    (sinCoeff64, error) = mp.chebyfit(mp.sin, [mp.mpf(str(fix64_sinIota)), mp.pi/4], chebyDegree, error=True)
    if error < mp.mpf(str(fix64Epsilon * fix64Epsilon)):
        break

# We now have a list of coefficients for the Chebyshev polynomial for sin(x) in Fix64,
# in the range [0, pi/4], with error less than fix64Epsilon.
sinCoeff64 = reversed(sinCoeff64)  # Reverse the order so that the index matches the degree

print("// Chebyshev coefficients for sin(x) at fix64_twoPiScale")
print("var sinChebyCoeff64 = []Fix64{")
for i, coeff in enumerate(sinCoeff64):
    print(f"    Fix64(0x{int(coeff * Fix64Scale * fix64_twoPiMultiplier) & 0xffffffffffffffff:016x}), // Coefficient {i}")
print("}")
```

### Computing Sin()

Other than expanding the identity `sin(x) = 2•sin(x/2)•cos(x/2)`, this function is a pretty simple application of the Chebyshev coefficients to the input. This code DOES work, and is very accurate, it's just MUCH slower than the Taylor expansion for values over π/4, and only slightly faster for values below that.

```go
// Returns sin(x), assumes that the input has already been normalized to the range [0, π]
// and scaled up by fix64_TwoPiScale.
func (x_scaled Fix64) scaledSinTest() (Fix64, error) {
	// Leverage the identity sin(x) = sin(π - x) to keep the input angle
	// in the range [0, π/2].
	if x_scaled > fix64_halfPiTerm {
		x_scaled, _ = fix64_onePiTerm.Sub(x_scaled)
	}
	// We can use the following identity to reduce the input angle to [0, π/4]:
	//     sin(x) = 2•sin(x/2)•cos(x/2)
	// At the same time, cos(y) = 1-2•sin²(y/2), so we can further expand this to:
	//     sin(x) = 2•sin(x/2)•(1 - 2•sin²(x/4))
	if x_scaled > fix64_halfPiTerm/2 {
		// recursively call sin(x/2) and sin(x/4), this should only recurse once since we
		// already reduced the input to the range [0, π/2] at the beginning of this function.
		sin_half, _ := x_scaled.intDiv(2).scaledSinTest()
		sin_quarter, _ := x_scaled.intDiv(4).scaledSinTest()
		sin_quarter_squared, _ := sin_quarter.FMD(sin_quarter, fix64_twoPiScale)
		cosTerm, _ := fix64_twoPiScale.Sub(sin_quarter_squared.intMul(2)) // cos(x/2) = 1 - 2•sin²(x/4)
		res, _ := sin_half.intMul(2).FMD(cosTerm, fix64_twoPiScale)       // sin(x) = 2•sin(x/2)•cos(x/2)
		return res, nil
	}

	if x_scaled.Lte(fix64_sinIotaTerm) {
		// If x is very small, we can just return x since sin(x) is linear for small x.
		return x_scaled, nil
	}

	// Start with the constant term of the chebyshev polynomial
	sum := sinChebyCoeff64[0]

	// manually compute the linear term of the series outside the loop to
	// avoid an extra multiplication in the loop.
	var err error
	pow := x_scaled
	term, _ := pow.FMD(sinChebyCoeff64[1], fix64_twoPiScale)
	sum, _ = sum.Add(term)

	for i := 2; i < len(sinChebyCoeff64); i++ {
		pow, err = pow.FMD(x_scaled, fix64_twoPiScale)
		if err == ErrUnderflow {
			break
		} else if err != nil {
			return 0, err
		}

		term, _ = pow.FMD(sinChebyCoeff64[i], fix64_twoPiScale)
		sum, _ = sum.Add(term)
	}

	return sum, nil
}
```