# add64.py
# Generates Go test data for UFix64 and Fix64 addition (including overflow)

from decimal import Decimal, getcontext, InvalidOperation, Overflow
#from utils import to_ufix64, to_fix64, go_hex64, FIX64_SCALE, MASK64, parseInput64
from data64 import *
import itertools
import sys
import inspect
import mpmath

getcontext().prec = 100
mpmath.mp.dps = 100

extraBits = 16  # The number of extra bits used in fix64_extra

decPi = Decimal(str(mpmath.pi))  # Pi to 100 decimal places

def hex64(val):
    """Convert a Decimal value to a hexadecimal string representation for a 64-bit type."""
    n = int((val * 10**8).quantize(1, rounding='ROUND_HALF_UP')) & 0xFFFFFFFFFFFFFFFF
    return f"0x{n:016x}"

def decSin(x: Decimal) -> Decimal:
    return Decimal(str(mpmath.sin(mpmath.mpf(str(x)))))

def decCos(x: Decimal) -> Decimal:
    return Decimal(str(mpmath.cos(mpmath.mpf(str(x)))))

def decTan(x: Decimal) -> Decimal:
    return Decimal(str(mpmath.tan(mpmath.mpf(str(x)))))

def decClamp(x: Decimal) -> Decimal:
    """ Normalize a Decimal value to the range of (-π, π)."""

    # start by reducing x to the range of [0, 2π)
    x = x % (decPi * 2)

    if x < 0:
        # if x is negative, add 2π to bring it into the range [0, 2π)
        x += decPi * 2

    # remove an extra 2π if x is greater than π
    if x > decPi:
        x -= decPi * 2

    # The clamp function in Go actually returns a fix64_extra value, so we multiply
    # by 2**12 to match that scale.
    x = x * Decimal(2**extraBits)

    return x

operations = {
    "Add": (lambda a, b: a + b, "{} + {} = {}"),
    "Sub": (lambda a, b: a - b, "{} - {} = {}"),
    "Mul": (lambda a, b: a * b, "{} * {} = {}"),
    "Div": (lambda a, b: a / b, "{} / {} = {}"),
    "FMD": (lambda a, b, c: a * b / c, "{} * {} / {} = {}"),
    "Sqrt": (lambda a: a.sqrt(), "sqrt({}) = {}"),
    "Ln": (lambda a: a.ln(), "ln({}) = {}"),
    "Exp": (lambda a: a.exp(), "exp({}) = {}"),
    "Clamp": (lambda a: decClamp(a), "clamp({}) = {}"),
    "Sin": (lambda a: decSin(a), "sin({}) = {}"),
    "Cos": (lambda a: decCos(a), "cos({}) = {}"),
    "Tan": (lambda a: decTan(a), "tan({}) = {}"),
}

types = {
    "UFix64":
        (Decimal(0),
         (Decimal(2**64) - 1) / Decimal("1e8"),
         Decimal("1e-8"),
         hex64),
    "Fix64":
        (Decimal(-2**63) / Decimal("1e8"),
         (Decimal(2**63) - 1) / Decimal('1e8'),
         Decimal("1e-8"),
         hex64),
}

def main():
    if len(sys.argv) != 3:
        print("Usage: add64.py <type> <operation>")
        sys.exit(1)

    type = sys.argv[1]
    operation = sys.argv[2]

    if type not in types:
        print(f"Invalid type: {type}. Must be one of {list(types.keys())}.")
        sys.exit(1)

    if operation not in operations:
        print(f"Invalid operation: {operation}. Must be one of {list(operations.keys())}.")
        sys.exit(1)

    typeInfo = types[type]
    operationInfo = operations[operation]
    operationFunc = operationInfo[0]
    operationFormat = operationInfo[1]

    # A bit of a hack: Ln() takes an unsigned argument, but returns a signed result. We want the
    # "type" value to stay as UFix to get the right input, but we want the output formatting and
    # validation to be for Fix.
    if operation == "Ln":
        if type[0] != 'U':
            print(f"Invalid operation {operation} for type {type}. Ln() only works with unsigned input.")
            sys.exit(1)

        typeInfo = types[type[1:]] # Trims off the first character, which should be U.

    minVal = typeInfo[0]
    maxVal = typeInfo[1]
    quanta = typeInfo[2]
    formatFunc = typeInfo[3]

    argCount = len(inspect.signature(operationFunc).parameters)

    match argCount:
        case 1:
            dataGen = itertools.chain(BaseData, ExtraData, BonusData)
        case 2:
            dataGen = itertools.chain(BaseData, ExtraData)
        case 3:
            dataGen = itertools.chain(BaseData)

    if type == "UFix64":
        generator = dataGen | generateUFix64Values | expandByIota | filterUFix64Values
    elif type == "Fix64":
        generator = dataGen | generateFix64Values | expandByIota | filterFix64Values
    
    for tuple in itertools.product(generator, repeat=argCount):
        descriptions, values = zip(*tuple)

        err = None

        try:
            result = operationFunc(*values)
            if result > 2**128:
                # The exp() operator can produce VERY large results, which can break
                # the quantization call below. We know that any value larger than 2**128
                # is an overflow in all of our types, so we can just skip the quantization step.
                err = "Overflow"
            elif not result.is_zero() and result.copy_sign(1) < (quanta / 2):
                err = "Underflow"
            else:
                result = result.quantize(quanta, rounding='ROUND_HALF_UP')
        except (ZeroDivisionError, InvalidOperation):
            err = "DivByZero"
        except Overflow:
            err = "Overflow"

        if operation == "Ln" and values[0] == 0:
            err = "DomainError"

        if err is None:
            if result > maxVal:
                err = "Overflow"
            elif result < minVal:
                err = "NegOverflow"

        if (operation == "Sin" or operation == "Cos") and err == "Underflow":
            # When sin or cost is called, they might produce values VERY, VERY close to 0
            # that would get tagged as underflow. However, for convenience, we want
            # to treat these as 0, so we just replace underflow errors with 0 results
            # for those two operations.
            result = Decimal(0)
            err = None
        
        if operation == "Exp" and not err and result.is_zero():
            # Technically, the exp() operation can only return positive results, so if
            # it did return zero, it must have been an underflow.
            err = "Underflow"

        # Wrap arguments in parentheses if they contain spaces (for readability)
        if argCount > 1:
            descriptions = map(lambda d: f'({d})' if ' ' in d else d, descriptions)

        if err is None:
            comment = operationFormat.format(*descriptions, result)
        else:
            comment = operationFormat.format(*descriptions, err)
            result = Decimal(0)

        hexValues = map(formatFunc, values)
        print(f'({', '.join(hexValues)}, {formatFunc(result)}, {err}, "{comment}")')

if __name__ == "__main__":
    main()
