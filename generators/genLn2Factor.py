from decimal import *
import argparse
import re
import sys
import time

getcontext().prec = 100

UFix64Max = Decimal(0xffffffffffffffff) / Decimal(1e8)  # UFix64 max value
Fix64Max = Decimal(0x7fffffffffffffff) / Decimal(1e8)  # Fix64 max value

decLn2 = Decimal(2).ln()
smallestK = Decimal(1e-8).ln() / decLn2

def genLn2Factor64(maxFactor):
    """Generate a factor for ln(2) that minimizes error in the space of 64-bits."""

    # About 250 billion... :sweat_smile:
    minError = Decimal(1)
    bestFactor = None

    start_timer()

    # Start with the highest possible factor and work downwards
    for factor in range(int(maxFactor), 1, -1):
        scaledEstimate = (decLn2 * Decimal(factor)).quantize(Decimal('1e-8'), rounding=ROUND_HALF_UP)
        est = scaledEstimate / Decimal(factor)
        err = abs(est - decLn2)
        if err < minError:
            minError = err
            bestFactor = factor
            print(f"Best factor: {bestFactor}, Error: {minError:.03g}")
        
        if check_timer():
            print("Computation timed out.")
            break

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

    print("Calculating best factor of ln(2) for UFix64")
    maxFactor64 = (UFix64Max / 2).quantize(1, rounding=ROUND_DOWN)
    genLn2Factor64(maxFactor64)

if __name__ == "__main__":
    main()