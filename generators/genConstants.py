# constgen.py
# Generates Go constant definitions for Fix64 and Fix128 types.
# Uses Decimal and mpmath for high-precision computation.

from decimal import *
import mpmath as mp

# Note: We use both Decimal and mpmath for high-precision calculations, even though there
# is a TON of overlap between them. We prefer Decimal for most calculations because it
# works in base 10, while mpmath works in base 2. However, Decimal is lacking trigonometric
# and analytic functions (i.e. Chebyshev), so we use mpmath for those. The "cleanest"
# way to convert between them is to convert the value from one library into a string, and
# then parse it with the other library. Hardly efficient, but we don't really care about
# performance here, just correctness.

# Set the precision for Decimal and mpmath to 100 decimal places.
getcontext().prec = 100
mp.mp.dps = 100

# Fixed-point scales
Fix64Scale = Decimal('1e8')
Fix128Scale = Decimal('1e24')
fix64Epsilon = Decimal('1e-8')  # Smallest representable difference in Fix64
fix128Epsilon = Decimal('1e-24')  # Smallest representable difference in Fix128

UFix64Max = Decimal(0xffffffffffffffff) / Fix64Scale
Fix64Max = Decimal(0x7fffffffffffffff) / Fix64Scale
Fix64Min = Decimal(-0x8000000000000000) / Fix64Scale

UFix128Max = Decimal(0xffffffffffffffffffffffffffffffff) / Fix128Scale
Fix128Max =  Decimal(0x7fffffffffffffffffffffffffffffff) / Fix128Scale
Fix128Min = Decimal(-0x80000000000000000000000000000000) / Fix128Scale

# Base constants
pi = Decimal(str(mp.pi)) # Pi to 50 decimal places!
ln2 = Decimal(2).ln() # Natural logarithm of 2


# Here we calculate the computation base for the trig functions for Fix64.
# 
# The constant below comes from running genFactors.py. It finds a large multiple
# for 2π that minimizes the error after the 8th decimal place, while still ensuring
# that trig values that have been scaled by this factor can still fit in a Fix64.
# 
# This allows us to have a "constant" for 2π that is accurate to nearly 30 decimal
# places, even though we can fit it into to a 64-bit value. (Which is almost magic
# given that 64-bits can only represent 19 decimal places accurately!)
fix64TrigMultiplier = Decimal(21264757054)
fix64PiScaled = pi * fix64TrigMultiplier
fix64TwoPiScaled = fix64PiScaled * 2
fix64HalfPiScaled = fix64PiScaled / 2
fix64ThreeHalfPiScaled = fix64PiScaled * 3 / 2

# When appoximating sin(x), the first non-linear term would be (x^3)/6. We use this fact
# to find a reasonable limit on using a simple linear approximation for sin(x).
# The following calculation finds the largest value of x for which the (x^3)/6
# is too small to affect the result (when scaled by fix64TrigMultiplier).
fix64_sinIota = (Decimal(3) / Decimal(Fix64Scale * fix64TrigMultiplier)) ** (Decimal(1) / Decimal(3))
fix64SinIotaScaled = fix64_sinIota * fix64TrigMultiplier

# A factor for calculating ln() an exp() in Fix64, comes from genFactors.py
fix64LnMultiplier = Decimal(3095485757)
fix64Ln2Scaled = ln2 * fix64LnMultiplier

 # Largest input to exp() that doesn't overflow
maxLn64 = UFix64Max.ln().quantize(fix64Epsilon, rounding='ROUND_DOWN')
 # Smallest input to exp() that doesn't underflow
minLn64 = (fix64Epsilon / 2).ln().quantize(fix64Epsilon, rounding='ROUND_DOWN')


# Similar values to the above, but for Fix128
# Best factor: 40384726982017, Error: 3.87e-51
# Best factor: 39291483001871, Error: 4.32e-51
# Best factor: 36941826758740, Error: 1.19e-51
# Best factor: 33498926535463, Error: 2.04e-51
# Computation timed out, last checked factor: 32390106065117
fix128TrigMultiplier = Decimal(36941826758740)
fix128PiScaled = pi * fix128TrigMultiplier
fix128TwoPiScaled = fix128PiScaled * 2
fix128HalfPiScaled = fix128PiScaled / 2
fix128ThreeHalfPiScaled = fix128PiScaled * 3 / 2

fix128_sinIota = (Decimal(3) / Decimal(Fix128Scale * fix128TrigMultiplier)) ** (Decimal(1) / Decimal(3))
fix128SinIotaScaled = fix128_sinIota * fix128TrigMultiplier

# Best factor: 2606255174222, Error: 2.06e-33
fix128LnMultiplier = Decimal(2606255174222)
fix128Ln2Scaled = ln2 * fix128LnMultiplier

maxLn128 = UFix128Max.ln().quantize(fix128Epsilon, rounding='ROUND_DOWN')
minLn128 = (fix128Epsilon / 2).ln().quantize(fix128Epsilon, rounding='ROUND_DOWN')

# We use a Chebyshev polynomial to approximate sin(x), working at the precision
# of fix192 (~1e-38). In order to keep the polynomial power terms less than
# one, we reduce the input to values less than 1.
#
# Fortunately for us, there are utilities in mpmath that can compute a Chebyshev
# polynomial for us. Rather than picking a specific degree, we target a specific
# error bound (1 / 2^128, the limit of fix192 precision).
#
# NOTE: The code that computes the coefficients assumes that all of the
# Chebyshev coefficients for sin(x) in the range [0, 1] are less than
# 1 EXCEPT for the linear term, which it assumes is close to 1 (maybe a little over,
# maybe a little under). In practice, having looked at the coefficients
# for a number of options, this assumption seems to hold true. (Very true,
# actually, with higher degree terms being much smaller than 1...)

# chebyDegree = 10 # Start with a degree of 10
maxError = Decimal('0.5') / Decimal(2**128)

# When approximating sin(x) near zero (using a Taylor series or Chebyshev), the first
# non-linear term is (x^3)/6 (or very, very close). We use this fact to find a
# reasonable lower for using a simple linear approximation for sin(x) that we
# can use for small values.
#
# The following calculation finds the largest value of x for which an (x^3)/6
# term is still too small to affect the result at are target error bound. For the
# values below this, we can just treat sin(x) = x, saving a lot of computation.
#
# NOTE: err = (i^3)/6 --> i = (6 * err) ** (1/3).
sinIota = (Decimal(6) * maxError) ** (Decimal(1) / Decimal(3))

# while True:
#     # Keep trying higher and higher degrees until we find one that fits our error bound.
#     # This lets us potentially just fiddle with different error bounds instead of hard
#     # coding a specific degree.
#     chebyDegree += 1

#     # Since we are going to use a linear approximation for sin(x) for values
#     # less than fix64_extra_sinIota, we can start the approximation range there.
#     (sinCoeffs, error) = mp.chebyfit(mp.sin, [mp.mpf(str(sinIota)), 1], chebyDegree, error=True)
#     if error < mp.mpf(str(maxError)):
#         break

# We now have a list of coefficients for the Chebyshev polynomial for sin(x).
# in the range [sinIota, pi/4], with error less than maxError.
sinCoeffs = mp.chebyfit(mp.sin, [0, 1], 25)
# sinCoeffs = reversed(sinCoeffs)  # Reverse so that the index matches the power of x
tanCoeffs = mp.chebyfit(mp.tan, [0, 1/8], 25)
# tanCoeffs = reversed(tanCoeffs)  # Reverse so that the index matches the power of x
expCoeffs = mp.chebyfit(mp.exp, [0, 1], 25)

def printChebyCoeff(coeffs):
    for i, coeff in enumerate(coeffs):
        decCoeff = Decimal(str(coeff)) # * (2**20)
        intValue = int((decCoeff * 2**128).to_integral_value(rounding=ROUND_HALF_UP))

        intString = f"0x{(intValue >> 128 & 0xffffffffffffffff):016x}"
        fracString= f"raw128{{0x{(intValue >> 64 & 0xffffffffffffffff):016x}, 0x{(intValue & 0xffffffffffffffff):016x}}}"
        hexString = f"{{i: {intString}, f: {fracString}}}"

        print(f"    fix192{hexString}, // x^{i}")


# Output Go code
def go_const(name, value, typ):
    match typ:
        case 'int64' | 'uint64':
            scaledValue = value
            bitLength = 64

        case 'Fix64' | 'UFix64':
            scaledValue = value * Fix64Scale
            bitLength = 64

        case 'Fix128' | 'UFix128':
            scaledValue = value * Fix128Scale
            bitLength = 128
        
        case 'raw128':
            scaledValue = value
            bitLength = 128
        
        case 'fix192':
            scaledValue = value * Decimal(2**128)
            bitLength = 192
        
        case _:
            raise ValueError(f"Unknown type: {typ}")

    intValue = int(scaledValue.to_integral_value(rounding=ROUND_HALF_UP))

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
            hexString = f"{{0x{(intValue >> 64 & 0xffffffffffffffff):016x}, 0x{(intValue & 0xffffffffffffffff):016x}}}"
        
        case 192:
            if intValue >= 2**192:
                raise ValueError(f"Value {value} for {name} exceeds 192-bit range: {intValue}")
            
            decl = 'var'

            intString = f"0x{(intValue >> 128 & 0xffffffffffffffff):016x}"
            fracString= f"raw128{{0x{(intValue >> 64 & 0xffffffffffffffff):016x}, 0x{(intValue & 0xffffffffffffffff):016x}}}"
            hexString = f"{{i: {intString}, f: {fracString}}}"


    return f"{decl} {name} = {typ}{hexString}"

def go_hex128(val):
    val = int(val) & 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
    return f"raw128{{0x{(val >> 64):016x}, 0x{(val & 0xFFFFFFFFFFFFFFFF):016x}}}"

def main():
    print("// Code generated by testgen/constgen.py; DO NOT EDIT.")
    print("package fixedPoint")
    print()
    print("// Exported scale constants")
    print(f"const Fix64Scale = {Fix64Scale}")
    print("const UFix64Zero = UFix64(0)")
    print("const Fix64Zero = Fix64(0)")
    print("const UFix64One = UFix64(1 * Fix64Scale) // 1 in fix64")
    print("const Fix64One = Fix64(1 * Fix64Scale) // 1 in fix64")
    print("const UFix64Iota = UFix64(1)")
    print("const Fix64Iota = Fix64(1)")
    print(f"const Fix64OneLeadingZeros = {64 - int(Fix64Scale).bit_length()} // Number of leading zero bits for Fix64One")    
    print("const UFix64Max = UFix64(0xffffffffffffffff) // Max value for UFix64")
    print("const Fix64Max = Fix64(0x7fffffffffffffff) // Max value for Fix64")
    print("const Fix64Min = Fix64(0x8000000000000000) // Min value for Fix64")
    print()
    print(f"const Fix128Scale = {Fix128Scale} // NOTE: Bigger than uint64! Mostly here as documentation...")
    print(go_const('UFix128Zero', Decimal(0), 'UFix128'))
    print(go_const('Fix128Zero', Decimal(0), 'Fix128'))
    print(go_const('UFix128One', Decimal(1), 'UFix128'))
    print(go_const('Fix128One', Decimal(1), 'Fix128'))
    print("var UFix128Iota = UFix128{0, 1}")
    print("var Fix128Iota = Fix128{0, 1}")
    print(f"const Fix128OneLeadingZeros = {128 - int(Fix128Scale).bit_length()} // Number of leading zero bits for Fix128One")    
    print(go_const('UFix128Max', UFix128Max, 'UFix128'))
    print(go_const('Fix128Max', Fix128Max, 'Fix128'))
    print(go_const('Fix128Min', Fix128Min, 'Fix128'))
    print()
    print("// Fix64 transcendental constants (see constgen.py for more information)")
    print(go_const('Fix64Pi', pi, 'Fix64'))
    print(go_const('fix64TrigMultiplier', fix64TrigMultiplier, 'uint64'))
    print(go_const('fix64TrigScale', fix64TrigMultiplier, 'Fix64'))
    print(go_const('ufix64PiScaled', fix64PiScaled, 'UFix64'))
    print(go_const('ufix64TwoPiScaled', fix64TwoPiScaled, 'UFix64'))
    print(go_const('ufix64HalfPiScaled', fix64HalfPiScaled, 'UFix64'))
    print(go_const('ufix64ThreeHalfPiScaled', fix64ThreeHalfPiScaled, 'UFix64'))
    print(go_const('fix64SinIotaScaled', fix64SinIotaScaled, 'Fix64'))
    print()
    print(go_const('fix64LnMultiplier', fix64LnMultiplier, 'uint64'))
    print(go_const('fix64LnScale', fix64LnMultiplier, 'Fix64'))
    print(go_const('ufix64LnScale', fix64LnMultiplier, 'UFix64'))
    print(go_const('fix64Ln2Scaled', fix64Ln2Scaled, 'Fix64'))
    print()
    print("// Valid logarithm bounds for Fix64")
    print(go_const('maxLn64', maxLn64, 'Fix64'))
    print(go_const('minLn64', minLn64, 'Fix64'))
    print()
    print("// Fix128 transcendental constants (see constgen.py for more information)")
    print(go_const('Fix128Pi', pi, 'Fix128'))
    print(go_const('fix128TrigMultiplier', fix128TrigMultiplier, 'uint64'))
    print(go_const('fix128TrigScale', fix128TrigMultiplier, 'Fix128'))
    print(go_const('ufix128PiScaled', fix128PiScaled, 'UFix128'))
    print(go_const('ufix128TwoPiScaled', fix128TwoPiScaled, 'UFix128'))
    print(go_const('ufix128HalfPiScaled', fix128HalfPiScaled, 'UFix128'))
    print(go_const('ufix128ThreeHalfPiScaled', fix128ThreeHalfPiScaled, 'UFix128'))
    print(go_const('fix128SinIotaScaled', fix128SinIotaScaled, 'Fix128'))
    print()
    print(go_const('fix128LnMultiplier', fix128LnMultiplier, 'uint64'))
    print(go_const('fix128LnScale', fix128LnMultiplier, 'Fix128'))
    print(go_const('ufix128LnScale', fix128LnMultiplier, 'UFix128'))
    print(go_const('fix128Ln2Scaled', fix128Ln2Scaled, 'Fix128'))
    print()
    print(go_const('fix192Pi', pi, 'fix192'))
    print(go_const('fix192HalfPi', pi/2, 'fix192'))
    print(go_const('fix192ThreeHalfPi', pi*3/2, 'fix192'))
    print()
    print("// Valid logarithm bounds for Fix128")
    print(go_const('maxLn128', maxLn128, 'Fix128'))
    print(go_const('minLn128', minLn128, 'Fix128'))
    print()
    print("// The value of e^x for all integer values of x between minLn128 and maxLn128")
    print("// expressed as fix192 values.")
    print("var expIntPowers = []fix192{")
    for intPower in range(int(minLn128) - 1, int(maxLn128) + 1):
        expValue = Decimal(intPower).exp()
        intPart = int(expValue) # Must truncate
        fracPart = int(((expValue - intPart) * 2**128).quantize(1, rounding=ROUND_HALF_UP))
        print(f"    fix192{{i: {intPart}, f: {go_hex128(fracPart)}}}, // e^{intPower}")
    print("}")
    print("const smallestExpIntPower = ", int(minLn128) - 1)
    print()
    # print("// Used for Chebyshev coefficients")
    # print("type coeff struct { isNeg bool; value raw128 }")
    # print()
    print("// Chebyshev coefficients for sin(x) in the range [0, 1]")
    print("var sinChebyCoeffs = []fix192{")
    printChebyCoeff(sinCoeffs)
    print("}")
    print()
    print("// Chebyshev coefficients for tan(x) in the range [0, 1/8]")
    print("var tanChebyCoeffs = []fix192{")
    printChebyCoeff(tanCoeffs)
    print("}")
    print()
    print("// Chebyshev coefficients for exp(x) in the range [0, 1]")
    print("var expChebyCoeffs = []fix192{")
    printChebyCoeff(expCoeffs)
    print("}")
    print()

if __name__ == "__main__":
    main()
