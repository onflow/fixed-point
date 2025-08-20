# constgen.py
# Generates Go constant definitions for Fix64 and Fix128 types.
# Uses Decimal and mpmath for high-precision computation.

from decimal import *
import mpmath as mp

# Note: We use both Decimal and mpmath for high-precision calculations, even though there is a TON
# of overlap between them. We prefer Decimal for most calculations because it works in base 10,
# while mpmath works in base 2. However, Decimal is lacking trigonometric and analytic functions
# (i.e. Chebyshev), so we use mpmath for those. The "cleanest" way to convert values between the two
# libraries is to go throught their string representations. Hardly efficient, but we don't really
# care about performance here, just correctness.

# Set the precision for Decimal and mpmath to 100 decimal places, probably overkill... but why not?
getcontext().prec = 100
mp.mp.dps = 100

# Fixed-point scales. Note that a significant amount of the Go code assumes these specific values
# and doesn't look at the constant definition which enables certain optimizations. So, while a lot of
# code will automatically adapt to updates in these constants, a bunch of other code will not. So,
# changing these values is NOT trivial.
Fix64Scale = Decimal('1e8')
Fix128Scale = Decimal('1e24')

# Smallest representable values for each type
fix64Epsilon = Decimal('1e-8')
fix128Epsilon = Decimal('1e-24')

# Bounds for each type
UFix64Max = Decimal(0xffffffffffffffff) / Fix64Scale
Fix64Max = Decimal(0x7fffffffffffffff) / Fix64Scale
Fix64Min = Decimal(-0x8000000000000000) / Fix64Scale

UFix128Max = Decimal(0xffffffffffffffffffffffffffffffff) / Fix128Scale
Fix128Max =  Decimal(0x7fffffffffffffffffffffffffffffff) / Fix128Scale
Fix128Min = Decimal(-0x80000000000000000000000000000000) / Fix128Scale

# Base transcendental constants
pi = Decimal(str(mp.pi)) # Pi to 100 decimal places!
ln2 = Decimal(2).ln() # Natural logarithm of 2

# In order to avoid some complicated and expensive math in clampAngle(), we need a "magic" scale
# factor for 2π. The details of this are explained better in the fix192.go file, but for the
# purposes of this file, we just need a value that meets the following criteria:
# 1. It is a multiple of 2π that fits into an unsigned 64-bit integer.
# 2. It is at least as large as the first word of any valid input to clampAngle(). Since
#    clampAngle() uses the absolute value of a signed input, the first word will always be positive
#    and at most 2^63 (0x8000000000000000).
# 3. It shoud be a multiple of 5^24. Again, more details in fix192.go, but using this factor allows
#    us to replace a potentially expensive division operation with a simple bit shift.
# 4. The value divided by its scale factor should be as close to 2π as possible, to minimize errors.
#
# Since there are only a few valid scale factors that meet the first three criteria, we can look at
# them all using a brute-force search. The code below does just that, and then picks out the scale
# factor with the smallest error.
twoPi = Decimal(str(mp.pi * 2))

# Later in this file, we will output 2π as a fix192 value. However, one of the operations we do in
# clampAngle() needs _even more precision_ than that, so we include an additional 64 bits of
# precision as a value called "twoPiResidual". We can compute that by multiplying 2π by the normal
# scale factor of fix192 (10^24 * 2^64) and then multiplying that by an additional 2^64 and taking
# the bottom 64 bits of the result as an integer.
twoPiResidual = int((twoPi * Decimal(10**24) * Decimal(2**128)).quantize(Decimal(1), rounding=ROUND_HALF_UP)) & 0xffffffffffffffff

# The smallest factor that results in a value larger than 2^63
minFactor = Decimal(0x8000000000000000) / Decimal(5**24) / twoPi
minFactor = int(minFactor.quantize(Decimal('1'), rounding=ROUND_UP))

# The largest factor that results in a value that still fits in 64-bits
maxFactor = Decimal(0xffffffffffffffff) / Decimal(5**24) / twoPi
maxFactor = int(maxFactor.quantize(Decimal('1'), rounding=ROUND_DOWN))

# Initial values that are updated in the loop below
bestFactor = None
bestError = Decimal(1)

# We now loop through each factor in the valid range to find the one with the smallest error
for i in range(minFactor, maxFactor + 1):
    currentMultiplier = Decimal(i) * 5**24
    scaled2Pi = twoPi * currentMultiplier
    # Always round down, the logic in clampAngle() assumes that this value of 2π is a slight
    # underestimate
    truncated2Pi = scaled2Pi.quantize(Decimal('1'), rounding=ROUND_DOWN)

    estTwoPi = truncated2Pi / currentMultiplier
    err = twoPi - estTwoPi

    if err < 0:
        # Should never happen
        raise ValueError(f"Error in twoPi calculation is negative: {err}")

    if err < bestError:
        bestError = err
        bestFactor = i

# Capture the two values we need to output to the constants file
twoPiFactor = Decimal(bestFactor)
twoPiMultiple = (twoPi * twoPiFactor * 5**24).quantize(Decimal('1'), rounding=ROUND_HALF_UP)

# Sanitity check: Make sure that our "magic" calculation works as expected for the largest possible
# 192-bit fixed value.
fix192Max = Decimal(2**191 - 1) / Decimal(10**24) / Decimal(2**64)
correctQuotient = fix192Max // twoPi
magicQuotient = (2**191 - 1) // twoPiMultiple * twoPiFactor // 2**88

if correctQuotient != magicQuotient:
    raise ValueError(f"Magic quotient does not compute the correct quotient for Fix128Max // 2π")

# Largest input to exp() that doesn't overflow
maxLn64 = UFix64Max.ln().quantize(fix64Epsilon, rounding='ROUND_DOWN')
maxLn128 = UFix128Max.ln().quantize(fix128Epsilon, rounding='ROUND_DOWN')

# Smallest input to exp() that doesn't underflow
minLn64 = (fix64Epsilon / 2).ln().quantize(fix64Epsilon, rounding='ROUND_DOWN')
minLn128 = (fix128Epsilon / 2).ln().quantize(fix128Epsilon, rounding='ROUND_DOWN')

# For sin(), exp(), and ln() we use Chebyshev polynomial approximations. This function generates the
# coefficients for a Chebyshev polynomial approximation of a given function over a specified range
# (using mpmath.chebyfit). It also runs the polynomial calculation at both ends of the range,
# ensuring that each coefficient and each intermediate result fits within the range of a fix192.
#
# We do a couple of tricks here for efficiency:
# 1. We wrap the function with scaling so that the the Chebyshev fit function scales to the factor
#    we use for fix192 values (10^24 * 2^64).
# 2. We scale each coefficient by a the closest power of 2 to that scale factor (2^144). This allows
#    our multiplication function to use a bit shift instead of having to divide by 10^24.
#
# See chebyPoly() and chebyMul() in fix192.go for more details.
def chebyFitWithOverflowCheck(func, bounds, degree):
    """ Compute a Chebyshev polynomial fit for a function, checking for overflow at the end of the range. """

    # Scale factors to account for the denominator for the fix192 type.
    scaleIn = mp.mpf(10**24) * mp.mpf(2**64)
    scaleOut = mp.mpf(10**24) * mp.mpf(2**64)

    # Scale factor used in our chebyMul() method, close to 10^24 * 2^64, to keep the coefficients
    # close to the value of one in fix192, but a power of two so that we don't have a division in the
    # inner loop.
    mulScale = mp.mpf(2**145)

    # Scale up the bounds by the input factor
    scaledBounds = [x * scaleIn for x in bounds]

    # Wrap the target function to scale both input and output
    wrappedFunc = lambda x: func(x/scaleIn) * scaleOut

    # Call the mpmath Chebyshev fit function
    (coeffs, err) = mp.chebyfit(wrappedFunc, scaledBounds, degree, error=True)

    # If the error is larger than 1, that means we don't match the precision of fix192, we don't
    # want that!
    if err > 1:
        raise ValueError(f"Chebyshev fit error {err} exceeds maximum allowed error for degree {degree}")

    # Create a list to represent the degree that each coefficient corresponds to.
    degrees = list(range(len(coeffs)))
    degrees.reverse()

    # Scale each coefficient by the multiplication scale factor taken to the same power of its
    # degree.
    coeffs = [x * (mulScale ** i) for i, x in zip(degrees, coeffs)]

    # We now have all of the coefficients, but we need to check to make sure that neither they, nor
    # the intermediate computations, overflow fix192. If the range of the input function is
    # monotonic, checking the computation at the endpoints can provide strong assurances that the
    # polynomial evaluation will not overflow in the middle of the range either.

    # Accumulators for the polynomial evaluation at the minimum and maximum of the range.
    minAccum, maxAccum = 0, 0

    # The end points to evaluate the polynomial at.
    minX = scaledBounds[0]
    maxX = scaledBounds[1]

    # The largest and smallest values that we can represent in 192 bits (signed).
    upperBound = mp.mpf(2**191) - 1
    lowerBound = mp.mpf(-2**191)

    for i, coeff in enumerate(coeffs):
        # Check if the coefficient is fits in 192 bits
        if coeff > upperBound or coeff < lowerBound:
            raise ValueError(f"Coefficient {coeff} for x^{len(coeffs) - i - 1} outside of Fix128 range")
        
        # Compute a temporary product of the accumulator and the x value, using the scale factor
        # that the chebyMul() function will use. Check to ensure the result won't overflow fix192.
        prod = minAccum * minX / mulScale
        if prod > upperBound or prod < lowerBound:
            raise ValueError(f"Overflow in Chebyshev polynomial evaluation at x={minX} for degree {degree}")
        
        # Add the next coefficient and check for overflow.
        minAccum = prod + coeff
        if minAccum > upperBound or minAccum < lowerBound:
            raise ValueError(f"Overflow in Chebyshev polynomial evaluation at x={minX} for degree {degree}")

        # Repeat the same for the maximum value in the range
        prod = maxAccum * maxX / mulScale
        if prod > upperBound or prod < lowerBound:
            raise ValueError(f"Overflow in Chebyshev polynomial evaluation at x={maxX} for degree {degree}")
        
        maxAccum = prod + coeff
        if maxAccum > upperBound or maxAccum < lowerBound:
            raise ValueError(f"Overflow in Chebyshev polynomial evaluation at x={maxX} for degree {degree}")


    # As a sanity check, let's compute the actual result at each end of the range when put through
    # the input function (scaled up by the output scale factor) to see if the result we computed
    # above is within our expected error bounds.
    realResultMin = func(bounds[0]) * scaleOut

    if abs(realResultMin - minAccum) > 1:
        raise ValueError(f"Chebyshev polynomial evaluation at x={bounds[0]} for degree {degree} does not meet error bound: ±{abs(realResultMin - minAccum)}")


    realResultMax = func(bounds[1]) * scaleOut

    if abs(realResultMax - maxAccum) > 1:
        raise ValueError(f"Chebyshev polynomial evaluation at x={bounds[1]} for degree {degree} does not meet error bound: ±{abs(realResultMax - maxAccum)}")

    return coeffs


# Generate Chebyshev coefficients for sin() (also used for cos()). A 30-degree polynomial gets us
# to fix192 precision (found with trial and error)
sinCoeffs = chebyFitWithOverflowCheck(mp.sin, [0, mp.pi/2], 30)

# When computing exp(), we only need the Chebyshev polynomial for the fractional part of the
# exponent; we compute the integer part using a lookup table and then multiply the two together. A
# 28-degree polynomial is sufficient for the fractional part (found with trial and error).
expCoeffs = chebyFitWithOverflowCheck(mp.exp, [0, 1], 28)

# Computing a polynomial for ln() is much more complicated than for sin() or exp(). First, the
# range: We can add a multiple of ln(2) to the result, and scale the input by a power of 2, so that
# we only need to compute ln() in a range [x, x*2] for some value x. We choose x = 2^79 / 10^24,
# which is largest power of 2 that is less than 1 in fix192. We can easily scale any input into the
# range [x, x*2] by doing a binary shift of that input so that it has the same nummber of leading
# zero bits as x. (Both shifts and counting leading zero bits are cheap operations.)
#
# Even for this relatively small range, we would need a polynomial of quite a large degree (~64
# terms!) to achieve our desired precision. Not only is this quite expensive, but each additional
# term provides additional compounding error. To that end, we break the range into 16 sub-ranges,
# each of which is [x * scaleFactor^i, x * scaleFactor^(i+1)], where scaleFactor is slightly larger
# than 2^(1/16). We then compute a Chebyshev polynomial for each sub-range.
#
# It might seem easier to break this range into equal sub-ranges linearly, but we either then
# have different number of coefficients in each range, or a different error bound in each range. By
# making the sub-ranges logarithmic, we can use the same number of coefficients in each range.
lnLowerBound = mp.mpf(2**79) / mp.mpf(10**24)
scaleFactor = 1.0443 # Slightly more than 2 ** (1/16)

# An array to accumulate the coefficients for each sub-range. This array will have 16 elements by
# the end.
lnCoeffs = []

# An array that represents the input bounds for each sub-range, used for a binary search to find the
# appropriate sub-range for any given input. This array will have 17 elements by the end.
lnBounds = []

# Compute the 16 polynomials, each with 22 coefficients, and store them in the lnCoeffs array. At
# the same time, we build up the lower bounds for each sub-range (which is also the upper bound for
# for the first 15 sub-ranges).
for i in range(16):
    lnBounds.append(lnLowerBound * (scaleFactor ** i))
    lnCoeffs.append(chebyFitWithOverflowCheck(mp.ln, [lnLowerBound * (scaleFactor ** i), lnLowerBound * (scaleFactor ** (i+1))], 22))

# Add the upper bound for the last sub-range
lnBounds.append(lnLowerBound * (scaleFactor ** 16))

# Convert the bounds to Decimal for our output functions
lnBounds = [Decimal(str(b)) for b in lnBounds]

# A function to print the Chebyshev coefficients in a format suitable for Go code.
def printChebyCoeff(coeffs):
    for i, coeff in enumerate(coeffs):
        decCoeff = Decimal(str(coeff))

        intValue = int(decCoeff.to_integral_value(rounding=ROUND_HALF_UP))
        hexString = hexString192(intValue)
        print(f"    fix192{hexString}, // x^{len(coeffs) - i - 1}")

def hexString192(intValue):
    """ Convert a 192-bit integer to a Go hex string representation. """
    hiString = f"0x{(intValue >> 128 & 0xffffffffffffffff):016x}"
    midString = f"0x{(intValue >> 64 & 0xffffffffffffffff):016x}"
    loString = f"0x{(intValue >> 0 & 0xffffffffffffffff):016x}"
    return f"{{Hi: {hiString}, Mid: {midString}, Lo: {loString}}}"

# Conversts a value to a Go constant definition.
def go_const(name, value, typ, rounding=ROUND_HALF_UP):
    match typ:
        case 'int64' | 'uint64' | 'raw64':
            scaledValue = Decimal(value)
            bitLength = 64

        case 'Fix64' | 'UFix64':
            scaledValue = Decimal(value) * Fix64Scale
            bitLength = 64

        case 'Fix128' | 'UFix128':
            scaledValue = Decimal(value) * Fix128Scale
            bitLength = 128
        
        case 'raw128':
            scaledValue = Decimal(value)
            bitLength = 128
        
        case 'fix192':
            scaledValue = Decimal(value) * Decimal(10**24) * Decimal(2**64)
            bitLength = 192
        
        case _:
            raise ValueError(f"Unknown type: {typ}")

    intValue = int(scaledValue.to_integral_value(rounding=rounding))

    match bitLength:
        case 64:
            if intValue >= 2**64:
                raise ValueError(f"Value {value} for {name} exceeds 64-bit range: {intValue}")

            decl = 'const'
            hexString = f"(0x{(intValue & 0xffffffffffffffff):016x})"

        case 128:
            if intValue >= 2**128:
                raise ValueError(f"Value {value} for {name} exceeds 128-bit range: {intValue}")

            decl = 'var'

            hiString = f"0x{(intValue >> 64 & 0xffffffffffffffff):016x}"
            loString = f"0x{(intValue >> 0 & 0xffffffffffffffff):016x}"
            hexString = f"{{Hi: {hiString}, Lo: {loString}}}"
        
        case 192:
            if intValue >= 2**192:
                raise ValueError(f"Value {value} for {name} exceeds 192-bit range: {intValue}")
            
            decl = 'var'

            hexString = hexString192(intValue)


    return f"{decl} {name} = {typ}{hexString}"

def main():
    print("// Code generated by generators/genConstants.py; DO NOT EDIT.")
    print("package fixedPoint")
    print()
    print("// Basic constants for Fix64 and UFix64")
    print(f"const Fix64Scale = {Fix64Scale}")
    print("const UFix64Zero = UFix64(0)")
    print("const Fix64Zero = Fix64(0)")
    print("const UFix64One = UFix64(1 * Fix64Scale) // 1 in fix64")
    print("const Fix64One = Fix64(1 * Fix64Scale) // 1 in fix64")
    print(f"const Fix64OneLeadingZeros = {64 - int(Fix64Scale).bit_length()} // Number of leading zero bits for Fix64One")    
    print("const UFix64Max = UFix64(0xffffffffffffffff) // Max value for UFix64")
    print("const Fix64Max = Fix64(0x7fffffffffffffff) // Max value for Fix64")
    print("const Fix64Min = Fix64(0x8000000000000000) // Min value for Fix64")
    print()
    print("// Basic constants for Fix128 and UFix128")
    print(f"const Fix128Scale = {Fix128Scale} // NOTE: Bigger than uint64! Mostly here as documentation...")
    print(go_const('UFix128Zero', Decimal(0), 'UFix128'))
    print(go_const('Fix128Zero', Decimal(0), 'Fix128'))
    print(go_const('UFix128One', Decimal(1), 'UFix128'))
    print(go_const('Fix128One', Decimal(1), 'Fix128'))
    print(f"const Fix128OneLeadingZeros = {128 - int(Fix128Scale).bit_length()} // Number of leading zero bits for Fix128One")    
    print(go_const('UFix128Max', UFix128Max, 'UFix128'))
    print(go_const('Fix128Max', Fix128Max, 'Fix128'))
    print(go_const('Fix128Min', Fix128Min, 'Fix128'))
    print()
    print("// Transcendental constants")
    print(go_const('Fix64Pi', pi, 'Fix64'))
    print(go_const('Fix64TwoPi', pi * 2, 'Fix64'))
    print(go_const('Fix64HalfPi', pi / 2, 'Fix64'))
    print(go_const('Fix128Pi', pi, 'Fix128'))
    print(go_const('Fix128TwoPi', pi * 2, 'Fix128'))
    print(go_const('Fix128HalfPi', pi / 2, 'Fix128'))
    print()
    print("// Internal constants for Fix64 and Fix128")
    print(go_const('maxLn64', maxLn64, 'Fix64'))
    print(go_const('minLn64', minLn64, 'Fix64'))
    print(go_const('maxLn128', maxLn128, 'Fix128'))
    print(go_const('minLn128', minLn128, 'Fix128'))
    print()
    print("// Internal constants for fix192")
    print(go_const('fix192Zero', Decimal(0), 'fix192'))
    print(go_const('fix192One', Decimal(1), 'fix192'))
    print(go_const('fix192Pi', pi, 'fix192'))
    print(go_const('fix192TwoPi', pi * 2, 'fix192'))
    print(go_const('fix192HalfPi', pi / 2, 'fix192'))
    print(go_const('fix192Ln2', ln2, 'fix192'))
    print(go_const('fiveToThe24', 5**24, 'raw64'))
    print()
    print("// Extra constants for clampAngle(), see fix192.go for details")
    # NOTE: We must use a value for 2π in clampAngle that rounds down to ensure that we can always
    # add the error term (computed using twoPiResidual). See fix192.go for more details.
    print(go_const('clampAngleTwoPi', twoPi, 'fix192', rounding=ROUND_DOWN))
    print(go_const('clampAngleTwoPiMultiple', twoPiMultiple, 'raw64'))
    print(go_const('clampAngleTwoPiFactor', twoPiFactor, 'raw64'))
    print(go_const('clampAngleTwoPiResidual', twoPiResidual, 'raw64'))
    print()
    print("// The value of e^x for all integer values of x between minLn128 and maxLn128")
    print("// expressed as fix192 values.")
    print("var expIntPowers = []fix192{")
    for intPower in range(int(minLn128) - 1, int(maxLn128) + 1):
        expValue = Decimal(intPower).exp()
        intValue = int((expValue * Decimal(10**24) * Decimal(2**64)).to_integral_value(rounding=ROUND_HALF_UP))
        hexString = hexString192(intValue)

        print(f"    fix192{hexString}, // e^{intPower} = {expValue:.30f}")
    print("}")
    print("const smallestExpIntPower = ", int(minLn128) - 1)
    print()
    print("// Chebyshev coefficients for sin(x) in the range [0, π/2]")
    print("var sinChebyCoeffs = []fix192{")
    printChebyCoeff(sinCoeffs)
    print("}")
    print()
    print("// Chebyshev coefficients for exp(x) in the range [0, 1]")
    print("var expChebyCoeffs = []fix192{")
    printChebyCoeff(expCoeffs)
    print("}")
    print()
    print("// Ranges for ln(x) polynomial coefficients")
    print("var lnBounds = []fix192{")
    for bound in lnBounds:
        intValue = int((bound * Decimal(10**24) * Decimal(2**64)).to_integral_value(rounding=ROUND_HALF_UP))
        hexString = hexString192(intValue)
        print(f"    fix192{hexString}, // {bound:.3f}")
    print("}")
    print()
    print(f"// Chebyshev coefficients for ln(x) in the range [{lnBounds[0]:.3f}, {lnBounds[-1]:.3f}]")
    print(f"var lnChebyCoeffs = [{len(lnCoeffs)}][]fix192{{")
    for i in range(len(lnCoeffs)):
        print(f"    // Coefficients for ln(x) in the range [{lnBounds[i]:.3f}, {lnBounds[i+1]:.3f}]")
        print("    {")
        printChebyCoeff(lnCoeffs[i])
        print("    },")
    print("}")
    print()

if __name__ == "__main__":
    main()
