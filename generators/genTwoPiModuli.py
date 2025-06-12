from decimal import *
import mpmath

getcontext().prec = 50
mpmath.mp.dps = 50

decPi = Decimal(str(mpmath.pi))  # Pi to 50 decimal places

# We scale the residual by a large multiplier so we can do the primary loop using integer arithmetic.
decimalMultiplier = Decimal(1e20)

def printSmallestErrors(scaleFactor, maxIterations=10000000):
    scaleFactor = Decimal(scaleFactor)

    scaled2Pi = decPi * 2 * scaleFactor
    truncated2Pi = int(scaled2Pi)

    # Turn that all into an integer so we can do the primary loop using integer arithmetic.
    baseTerm = int((scaled2Pi - truncated2Pi) * decimalMultiplier)

    currentTerm = 0
    smallestErr = baseTerm + 1

    for multiplier in range(1, maxIterations):
        # print(f"Multiplier: {multiplier}, Current Error: {currentErr}")
        currentTerm = (currentTerm + baseTerm) % decimalMultiplier
        currentErr = min(currentTerm, decimalMultiplier - currentTerm)
        if currentErr < smallestErr:
            smallestErr = currentErr
            print(f"New smallest err: {Decimal(smallestErr) / decimalMultiplier / scaleFactor} (multiplier: {multiplier})")

print("Smallest errors for Fix64:")
printSmallestErrors(10**8) #, maxIterations=92233720368)

print("\nSmallest errors for fix64_extra:")
printSmallestErrors(Decimal(10**8 * 2**12))  # This is the scale factor used in fix64_extra
