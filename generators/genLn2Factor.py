from decimal import *
import argparse
import re
import sys
import time
import mpmath
from numba import njit

getcontext().prec = 100
mpmath.mp.dps = 100

UFix64Max = Decimal(0xffffffffffffffff) / Decimal(1e8)  # UFix64 max value
Fix64Max = Decimal(0x7fffffffffffffff) / Decimal(1e8)  # Fix64 max value

UFix128Max = Decimal(0xffffffffffffffffffffffffffffffff) / Decimal(1e24)  # UFix64 max value
Fix128Max = Decimal(0x7fffffffffffffffffffffffffffffff) / Decimal(1e24)  # Fix64 max value

def genFactor(maxFactor, baseValue, quanta):
    """Generate a factor for some input value that minimizes error for a particular quanta."""

    minError = Decimal(1)
    bestFactor = None

    start_timer()

    # Start with the highest possible factor and work downwards
    for factor in range(int(maxFactor), 1, -1):
        scaledEstimate = (baseValue * Decimal(factor)).quantize(quanta, rounding=ROUND_HALF_UP)
        est = scaledEstimate / Decimal(factor)
        err = abs(est - baseValue)
        if err < minError:
            minError = err
            bestFactor = factor
            print(f"Best factor: {bestFactor}, Error: {minError:0.3g}")
        
        if check_timer():
            print(f"Computation timed out, last checked factor: {factor}")
            break

@njit("uint64(uint64, uint64, uint64, uint64)", cache=True)
def checkChunk(startValue, incr, count, modulus):
    smallest = startValue
    smallestIndex = 0
    curValue = startValue
    for i in range(count):
        if curValue < smallest:
            smallest = curValue
            smallestIndex = i
        curValue = (curValue + incr) % modulus
    return smallestIndex

def genFactorJit(maxFactor, baseValue: Decimal, quanta):
    """Generate a factor for some input value that minimizes error for a particular quanta using numpy."""

    minError = Decimal(1)
    bestFactor = None

    start_timer()

    chunkSize = 100000
    intScale = 10**19
    incrementInt = int((baseValue - baseValue.quantize(quanta, rounding=ROUND_DOWN)) / quanta * intScale)

    # Start with the highest possible factor and work downwards
    for factor in range(int(maxFactor-chunkSize), 1, -chunkSize):
        # Compute the multiple of the input that we will start with for this chunk 
        chunkBase = baseValue * factor

        # Take the residual of the chunk base when quantized to the nearest quanta
        # You can think of this as the "error" in the quantization, or the "remainder"
        # when dividing by the quanta.
        chunkResidual = chunkBase - chunkBase.quantize(quanta, rounding=ROUND_DOWN)

        # Turn that remainder into an integer so we can do the primary loop using integer arithmetic.
        chunkResidualInt = int(chunkResidual / quanta * intScale)

        bestFactorInChunk = checkChunk(chunkResidualInt, incrementInt, chunkSize, intScale) + factor
        # print(f"Best factor in chunk {factor} to {factor + chunkSize}: {bestFactorInChunk}")

        scaledEstimate = (baseValue * Decimal(bestFactorInChunk)).quantize(quanta, rounding=ROUND_HALF_UP)
        est = scaledEstimate / Decimal(bestFactorInChunk)
        err = abs(est - baseValue)
        # print(f"Best factor in chunk: {bestFactorInChunk}, Error: {err:0.3g}")
        if err < minError:
            minError = err
            bestFactor = bestFactorInChunk
            print(f"Best factor: {bestFactorInChunk}, Error: {minError:0.3g}")
        
        if check_timer():
            print(f"Computation timed out, last checked factor: {factor}")
            return
    
    print(f"Checked all factors {factor}-{maxFactor}.")

deadline = 30.0  # default in seconds

_timer_start = None

def start_timer():
    global _timer_start
    _timer_start = time.time()

def check_timer():
    """Returns True if more than 'deadline' seconds have passed since start_timer() was called, otherwise False."""
    if _timer_start is None:
        return True # If timer was never started, stop immediately to avoid infinite loops
    return (time.time() - _timer_start) > deadline

def parse_duration(s):
    match = re.fullmatch(r'(\d+(?:\.\d+)?)([smh]?)', s.strip())
    if not match:
        raise ValueError(f"Invalid duration: {s}")
    value, unit = match.groups()
    value = float(value)
    if unit == 's' or unit == '':
        return value
    elif unit == 'm':
        return value * 60
    elif unit == 'h':
        return value * 3600
    else:
        raise ValueError(f"Unknown unit: {unit}")

def main():
    global deadline
    parser = argparse.ArgumentParser(description="Generate ln(2) factor for 64-bit fixed point.")
    parser.add_argument('deadline', nargs='?', default='30s', help="Time limit (e.g. 10s, 2m, 1.5h). Default: 30s")
    args = parser.parse_args()
    try:
        deadline = parse_duration(args.deadline)
    except ValueError as e:
        print(e, file=sys.stderr)
        sys.exit(1)
    print(f"Deadline set to {deadline} seconds.")

    print("Calculating best factor of 2π for UFix64")
    maxFactor = (Fix64Max / 7).quantize(1, rounding=ROUND_DOWN)
    genFactorJit(maxFactor, Decimal(str(mpmath.pi)) * 2, Decimal('1e-8'))

    # print("Calculating best factor of ln(2) for UFix64")
    # maxFactor = (Fix64Max / Decimal('25.95')).quantize(1, rounding=ROUND_DOWN)
    # genFactorJit(maxFactor, Decimal(2).ln(), Decimal('1e-8'))

    print("Calculating best factor of 2π for UFix64")
    maxFactor = (Fix128Max / 7).quantize(1, rounding=ROUND_DOWN)
    genFactorJit(maxFactor, Decimal(str(mpmath.pi)) * 2, Decimal('1e-24'))

    # print("Calculating best factor of ln(2) for UFix64")
    # maxFactor = (Fix64Max / Decimal('25.95')).quantize(1, rounding=ROUND_DOWN)
    # genFactorJit(maxFactor, Decimal(2).ln(), Decimal('1e-8'))

if __name__ == "__main__":
    main()