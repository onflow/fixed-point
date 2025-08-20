# add64.py
# Generates Go test data for UFix64 and Fix64 addition (including overflow)

from decimal import Decimal, getcontext, InvalidOperation, Overflow
from data64 import *
from data128 import *
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
    intValue = int((val * 10**8).quantize(1, rounding='ROUND_HALF_UP')) & 0xFFFFFFFFFFFFFFFF
    return f"0x{intValue:016x}"

def hex128(val):
    """Convert a Decimal value to a hexadecimal string representation for a 128-bit type."""
    intValue = int((val * 10**24).quantize(1, rounding='ROUND_HALF_UP')) & 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
    return f"0x{((intValue >> 64) & 0xffffffffffffffff):016x}, 0x{(intValue & 0xffffffffffffffff):016x}"

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

    # The clamp function in Go actually returns a scaled value, so we multiply
    # to match that scale.
    x = x * 21264757054

    return x

operations = {
    "Add": (lambda a, b: a + b, "{} + {} = {}"),
    "Sub": (lambda a, b: a - b, "{} - {} = {}"),
    "Mul": (lambda a, b: a * b, "{} * {} = {}"),
    "Div": (lambda a, b: a / b, "{} / {} = {}"),
    "FMD": (lambda a, b, c: a * b / c, "{} * {} / {} = {}"),
    "Mod": (lambda a, b: a % b, "{} % {} = {}"),
    "Sqrt": (lambda a: a.sqrt(), "sqrt({}) = {}"),
    "Ln": (lambda a: a.ln(), "ln({}) = {}"),
    "Exp": (lambda a: a.exp(), "exp({}) = {}"),
    "Pow": (lambda a, b: a ** b, "{} ** {} = {}"),
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
    "UFix128":
        (Decimal(0),
         (Decimal(2**128) - 1) / Decimal("1e24"),
         Decimal("1e-24"),
         hex128),
    "Fix128":
        (Decimal(-2**127) / Decimal("1e24"),
         (Decimal(2**127) - 1) / Decimal('1e24'),
         Decimal("1e-24"),
         hex128),
}

def main():
    if len(sys.argv) != 3:
        print("Usage: add64.py <type> <operation>")
        sys.exit(1)

    outputType = sys.argv[1]
    operation = sys.argv[2]

    if outputType not in types:
        print(f"Invalid output type: {outputType}. Must be one of {list(types.keys())}.")
        sys.exit(1)

    if operation not in operations:
        print(f"Invalid operation: {operation}. Must be one of {list(operations.keys())}.")
        sys.exit(1)

    outputTypeInfo = types[outputType]
    operationInfo = operations[operation]
    operationFunc = operationInfo[0]
    operationFormat = operationInfo[1]

    minVal = outputTypeInfo[0]
    maxVal = outputTypeInfo[1]
    quanta = outputTypeInfo[2]
    formatFunc = outputTypeInfo[3]

    argCount = len(inspect.signature(operationFunc).parameters)

    # By default, we assume that the inputs are the same type as the output. This will be true
    # for most operations.
    argTypes = [outputType] * argCount

    match operation:
        case "Ln":
            # Ln goes unsigned -> signed
            if outputType[0] == 'U':
                exit("Ln operation requires a signed output type (Fix64 or Fix128).")
            
            argTypes[0] = "U" + outputType  # set the argument type to be unsigned
        case "Exp":
            # Exp goes signed -> unsigned
            if outputType[0] != 'U':
                exit("Exp operation requires an unsigned output type (UFix64 or UFix128).")

            argTypes[0] = outputType[1:]  # set the argument type to be signed
        case "Pow":
            # Pow goes (unsigned, signed) -> unsigned
            if outputType[0] != 'U':
                exit("Pow operation requires an unsigned output type (UFix64 or UFix128).")

            argTypes = [outputType, outputType[1:]]  # first argument unsigned, second is signed
        case _:
            pass  # No change needed for other operations

    bitLength = "64" if outputType[-1] == '4' else "128"

    baseData = globals()[f"BaseData{bitLength}"]
    extraData = globals()[f"ExtraData{bitLength}"]
    bonusData = globals()[f"BonusData{bitLength}"]

    match argCount:
        case 1:
            dataList = baseData + extraData + bonusData
        case 2:
            dataList = baseData + extraData
        case 3:
            dataList = baseData

    argGenerators = []

    for argType in argTypes:
        if argType == "UFix64":
            argGenerators.append(dataList | generateUFix64Values | expandByIota64 | filterUFix64Values)
        elif argType == "Fix64":
            argGenerators.append(dataList | generateFix64Values | expandByIota64 | filterFix64Values)
        elif argType == "UFix128":
            argGenerators.append(dataList | generateUFix128Values | expandByIota128 | filterUFix128Values)
        elif argType == "Fix128":
            argGenerators.append(dataList | generateFix128Values | expandByIota128 | filterFix128Values)
        else:
            raise ValueError(f"Unknown argument type: {argType}")

    for tuple in itertools.product(*argGenerators):
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
                beforeRounding = result
                result = result.quantize(quanta, rounding='ROUND_DOWN')
                if result.is_zero() and not beforeRounding.is_zero():
                    err = "Underflow"

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
            # When sin or cos is called, they might produce values VERY, VERY close to 0
            # that would get tagged as underflow. However, for convenience, we want
            # to treat these as 0, so we just replace underflow errors with 0 results
            # for those two operations.
            result = Decimal(0)
            err = None
        
        # if (operation == "Tan"):
        #     # Producing accurate results for tan() close to ±π/2 is difficult, so we
        #     # treat any result that is outside the range of -50M to 50M as an overflow.
        #     if result > Decimal('5e7'):
        #         err = "Overflow"
        #     elif result < Decimal('-5e7'):
        #         err = "NegOverflow"
        
        if (operation == "Exp" or operation == "Pow") and not err and result.is_zero():
            # Technically, the exp() and pow() operations can only return positive, non-zero
            # results, so if it did return zero, it must have been an underflow.
            err = "Underflow"

        # if (operation == "Exp" or operation == "Pow") and not err and result > Decimal('1e14'):
            # For exp() and pow(), we round results that are larger than 100 trillion to "just"
            # 23 decimal places, which is a million times better than any floating-point library
            # can do, and way more trouble than it's worth to get that two digits for such large
            # numbers...
            # result = result.quantize(Decimal('1e-23'), rounding='ROUND_HALF_UP')

        if operation == "Pow" and values[0] == 0:
            # The Decimal library treats 0^x differently that we want to, so we override
            # some if its behavior here
            if values[1] < 0:
                err = "DivByZero"
                result = Decimal(0)
            elif values[1] == 0:
                err = None
                result = Decimal(1)
            else:
                err = None
                result = Decimal(0)

        # Wrap arguments in parentheses if they contain spaces (for readability)
        if argCount > 1:
            descriptions = map(lambda d: f'({d})' if ' ' in d else d, descriptions)

        if err is None:
            comment = operationFormat.format(*descriptions, result)
        else:
            comment = operationFormat.format(*descriptions, err)
            result = Decimal(0)

        hexValues = map(formatFunc, values)
        print(f'({", ".join(hexValues)}, {formatFunc(result)}, {err}, "{comment}")')

if __name__ == "__main__":
    main()
