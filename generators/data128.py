# data64.py
# Defines test data for UFix64 and Fix64 types

from decimal import Decimal
import re
from pipe import Pipe

from data64 import BaseData64, ExtraData64, BonusData64

BaseData128 = BaseData64 + [
    # NOTE: We include all of the BaseData64 values here as well, including those values
    # specified as "MAX" and "MIN", so we just need values that are specific for 128-bit types.

    # Very small values in the range of UFix128
    ("", "3e-24"),
    ("", "6e-24"),
    ("", "9e-24"),

    # sqrt(MaxUFix128) and nearby values
    ("sqrt(MaxUFix128)", "18446744.073709551616"),
    ("sqrt(MaxUFix128) - epsilon", "18446744.073709551616999999999998"),
    ("sqrt(MaxUFix128) + epsilon", "18446744.073709551616000000000002"),

    # sqrt(HalfMaxUFix128) nearby values
    ("sqrt(HalfMaxUFix128)", "13043817.825332782212349571806253"),
    ("sqrt(HalfMaxUFix128) - epsilon", "13043817.825332782212349571806251"),
    ("sqrt(HalfMaxUFix128) + epsilon", "13043817.825332782212349571806255"),
]

ExtraData128 = ExtraData64 + [
    # MaxUFix128 divided by powers of ten
    ("MaxUFix128 / 10", "34028236692093.846346337460743176821146"),
    ("MaxUFix128 / 100", "3402823669209.384634633746074317682115"),
    ("MaxUFix128 / 1000", "340282366920.938463463374607431768211"),
    ("MaxUFix128 / 10000", "34028236692.093846346337460743176821"),
    ("MaxUFix128 / 100000", "3402823669.209384634633746074317682"),
    ("MaxUFix128 / 1000000", "340282366.920938463463374607431768"),
    ("MaxUFix128 / 10000000", "34028236.692093846346337460743177"),
    ("MaxUFix128 / 100000000", "3402823.669209384634633746074318"),
    ("MaxUFix128 / 1000000000", "340282.366920938463463374607432"),
    ("MaxUFix128 / 10000000000", "34028.236692093846346337460743"),
    ("MaxUFix128 / 100000000000", "3402.823669209384634633746074"),
    ("MaxUFix128 / 1000000000000", "340.282366920938463463374607"),
    ("MaxUFix128 / 10000000000000", "34.028236692093846346337460"),
    ("MaxUFix128 / 100000000000000", "3.402823669209384634633746"),

    # MaxFix64 divided by powers of ten
    ("MaxFix128 / 1", "170141183460469.231731687303715884105727"),
    ("MaxFix128 / 10", "17014118346046.923173168730371588410573"),
    ("MaxFix128 / 100", "1701411834604.692317316873037158841057"),
    ("MaxFix128 / 1000", "170141183460.469231731687303715884106"),
    ("MaxFix128 / 10000", "17014118346.046923173168730371588411"),
    ("MaxFix128 / 100000", "1701411834.604692317316873037158841"),
    ("MaxFix128 / 1000000", "170141183.460469231731687303715884"),
    ("MaxFix128 / 10000000", "17014118.346046923173168730371588"),
    ("MaxFix128 / 100000000", "1701411.834604692317316873037159"),
    ("MaxFix128 / 1000000000", "170141.183460469231731687303716"),
    ("MaxFix128 / 10000000000", "17014.118346046923173168730372"),
    ("MaxFix128 / 100000000000", "1701.411834604692317316873037"),
    ("MaxFix128 / 1000000000000", "170.141183460469231731687304"),
    ("MaxFix128 / 10000000000000", "17.014118346046923173168730"),
    ("MaxFix128 / 100000000000000", "1.701411834604692317316873"),
    
    # Powers of ten beyond the range of UFix64
    ("", "1e-23"),
    ("", "1e-22"),
    ("", "1e-21"),
    ("", "1e-20"),
    ("", "1e-15"),
    ("", "1e12"),
    ("", "1e13"),
    ("", "1e14"),
    ("", "1e15"),

    # Powers of 2 beyond the range of UFix64
    ("2^-16", "0.0000152587890625"),
    ("2^-15", "0.000030517578125"),
    ("2^-14", "0.00006103515625"),
    ("2^-13", "0.0001220703125"),
    ("2^-12", "0.000244140625"),
    ("2^-11", "0.00048828125"),
    ("2^-10", "0.0009765625"),
    ("2^-9", "0.001953125"),
    ("2^37", "137438953472"),
    ("2^38", "274877906944"),
    ("2^39", "549755813888"),
    ("2^40", "1099511627776"),
    ("2^41", "2199023255552"),
    ("2^42", "4398046511104"),
    ("2^43", "8796093022208"),
    ("2^44", "17592186044416"),
    ("2^45", "35184372088832"),
    ("2^46", "70368744177664"),
    ("2^47", "140737488355328"),
    ("2^48", "281474976710656"),

    # Trigonometric values at higher precision
    ("pi/6", "0.523598775598298873077107"),
    ("pi/4", "0.785398163397448309615661"),
    ("pi/3", "1.047197551196597746154214"),
    ("pi/2", "1.570796326794896619231322"),
    ("pi", "3.141592653589793238462643"),
    ("3*pi/2", "4.712388980384689857693965"),
    ("2*pi", "6.283185307179586476925287"),
    ("3*pi/4", "2.356194490192344928846983"),
    ("sqrt(2)", "1.414213562373095048801689"),
    ("sqrt(2)/2", "0.707106781186547524400844"),

    # Logarithmic values at higher precision
    ("ln(2)", "0.693147180559945309417232"),
    ("ln(10)", "2.302585092994045684017991"),
    ("e", "2.718281828459045235360287"),
    ("e^2", "7.389056098930650227230427"),
]

# Additional inputs used for single-argument methods
BonusData128 = BonusData64 + [
    # Odd multiples of pi/2, at higher precision
    ("3/2*pi", "4.712388980384689857693965"),
    ("5/2*pi", "7.853981633974483096156608"),
    ("7/2*pi", "10.995574287564276334619252"),
    ("9/2*pi", "14.137166941154069573081895"),
    ("11/2*pi", "17.278759594743862811544539"),
    ("13/2*pi", "20.420352248333656050007182"),
    ("15/2*pi", "23.561944901923449288469825"),
    ("17/2*pi", "26.703537555513242526932469"),
    ("19/2*pi", "29.845130209103035765395112"),

    # VERY large multiples of pi/2, the largest possible in the space of Fix128
    ("108315241484939/2*pi", "170141183460444.384801716215411905272453"),
    ("108315241484941/2*pi", "170141183460447.526394369805205143735096"),
    ("108315241484943/2*pi", "170141183460450.667987023394998382197740"),
    ("108315241484945/2*pi", "170141183460453.809579676984791620660383"),
    ("108315241484947/2*pi", "170141183460456.951172330574584859123027"),
    ("108315241484949/2*pi", "170141183460460.092764984164378097585670"),
    ("108315241484951/2*pi", "170141183460463.234357637754171336048313"),
    ("108315241484953/2*pi", "170141183460466.375950291343964574510957"),

    # Values that are close to the boundaries of UFix128 for exp()
    ("", "32.5"),
    ("", "32.6"),
    ("", "32.7"),
    ("", "32.8"),
    ("", "32.9"),
    ("", "33.0"),
    ("", "33.1"),
    ("", "33.2"),
    ("", "33.3"),
    ("", "33.4"),
    ("", "33.5"),
    ("ln(UFix64Max)", "33.460796879815903188973917"),

    ("", "54.5"),
    ("", "54.6"),
    ("", "54.7"),
    ("", "54.8"),
    ("", "54.9"),
    ("", "55.0"),
    ("", "55.1"),
    ("", "55.2"),
    ("", "55.3"),
    ("ln(UFix64Iota)", "55.262042231857096416431795")
]

# Generates a sequence of UFix64 test values based on the provided raw values.
# Each test value is a tuple of (string representation, Decimal value).
@Pipe
def generateUFix128Values(raw_values):
    for (rawD, rawV) in raw_values:
        if len(rawD) == 0:
            rawD = rawV

        if rawV[0].isdigit():
            val = Decimal(rawV)
            string = str(rawD)

            if val.quantize(Decimal("1e-24"), rounding='ROUND_HALF_UP') != val:
                raise ValueError(f"Bad raw value: {rawV}")

            yield (string, val)
            continue

        if rawV == "Max":
            val = Decimal(2**128 - 1) / Decimal(10**24)
            string = "MaxUFix128"
            yield (string, val)
            continue

        if rawV == "HalfMax":
            val = Decimal(2**127 - 1) / Decimal(10**24)
            string = "HalfMaxUFix128"
            yield (string, val)
            continue

        m = re.match(r"^([MH][a-zA-Z0-9]+) ([+-]) (\S+)$", rawV)
        if m:
            key = m.group(1)
            op = m.group(2)
            mod = Decimal(m.group(3))
            if key == "Max":
                val = Decimal(2**128 - 1) / Decimal(10**24)
            elif key == "HalfMax":
                val = Decimal(2**127 - 1) / Decimal(10**24)
            else:
                raise ValueError(f"Unknown symbolic value: {key}")
            if op == "+":
                val += mod
            elif op == "-":
                val -= mod
            else:
                raise ValueError(f"Unknown operator: {op}")
            string = f"{key}UFix128 {op} {mod}"
        else:
            raise ValueError(f"Invalid raw value format: {rawV}")

        yield (string, val)

# Generates a sequence of Fix64 test values based on the provided raw values.
# Each test value is a tuple of (string representation, Decimal value).
@Pipe
def generateFix128Values(raw_values):
    for (rawD, rawV) in raw_values:
        if len(rawD) == 0:
            rawD = rawV

        if rawV == "0":
            val = Decimal("0")
            string = "0"
            yield (string, val)
            continue

        if rawV[0].isdigit():
            val = Decimal(rawV)
            string = str(rawD)

            yield (string, val)
            yield ("-" + string, -val)
            continue

        if rawV == "Max":
            val = Decimal(2**127 - 1) / Decimal(10**24)
            yield ("MaxFix128", val)
            val = -Decimal(2**127) / Decimal(10**24)
            yield ("MinFix128", val)
            continue

        if rawV == "HalfMax":
            val = Decimal(2**126 - 1) / Decimal(10**24)
            yield ("HalfMaxFix128", val)
            val = -Decimal(2**126) / Decimal(10**24)
            yield ("HalfMinFix128", val)
            continue

        m = re.match(r"^([MH][a-zA-Z0-9]+) ([+-]) (\S+)$", rawV)
        if m:
            key = m.group(1)
            op = m.group(2)
            mod = Decimal(m.group(3))
            if key == "Max":
                baseVal = Decimal(2**127 - 1) / Decimal(10**24)
            elif key == "HalfMax":
                baseVal = Decimal(2**127 - 1) / Decimal(10**24)
            else:
                raise ValueError(f"Unknown symbolic value: {key}")
            
            # First output the positive value
            if op == "+":
                val = baseVal + mod
            elif op == "-":
                val = baseVal - mod
            else:
                raise ValueError(f"Unknown operator: {op}")
            string = f"{key}Fix128 {op} {mod}"
            yield (string, val)

            # Now output the corresponding negative value
            key = key.replace("Max", "Min")
            if op == "+":
                op = "-"
            else:
                op = "+"

            # Negative min values are one larger than positive max values
            baseVal += Decimal('1e-24')

            if op == "+":
                val = baseVal + mod
            elif op == "-":
                val = baseVal - mod
            else:
                raise ValueError(f"Unknown operator: {op}")
            string = f"{key}Fix128 {op} {mod}"
            yield (string, val)
        else:
            raise ValueError(f"Invalid raw value format: {rawV}")

@Pipe
def expandByIota128(values):
    for val in values:
        yield val

        # Generate values ± 1e-8 around the original value
        newVal = val[1] + Decimal("1e-24")
        newStr = val[0] + " + 1e-24"

        if len(str(newVal)) < len(newStr):
            newStr = str(newVal).lower()

        yield (newStr, newVal)

        newVal = val[1] - Decimal("1e-24")
        newStr = val[0] + " - 1e-24"

        if len(str(newVal)) < len(newStr):
            newStr = str(newVal).lower()

        yield (newStr, newVal)

def isUFix128Value(val):
    """Check if the value is a valid UFix64 value."""
    if isinstance(val, Decimal):
        testVal = val.quantize(Decimal("1e-24"), rounding='ROUND_HALF_UP')
        return 0 <= testVal <= Decimal("340282366920938.463463374607431768211455") and \
                (testVal == 0 or testVal >= Decimal("1e-24"))
    return False

@Pipe
def filterUFix128Values(values):
    return (x for x in values if isUFix128Value(x[1]))

def isFix128Value(val):
    """Check if the value is a valid Fix64 value."""
    if isinstance(val, Decimal):
        testVal = val.quantize(Decimal("1e-24"), rounding='ROUND_HALF_UP')
        return Decimal("-170141183460469.231731687303715884105728") <= testVal <= Decimal("170141183460469.231731687303715884105727") and \
                (testVal == 0 or testVal.copy_sign(1) >= Decimal("1e-24"))
    return False

@Pipe
def filterFix128Values(values):
    return [x for x in values if isFix128Value(x[1])]
