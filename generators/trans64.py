# trans64.py
# Test generator for transcendental functions (e.g., sqrt) for UFix64 and Fix64
# More functions can be added later.

from utils import *
from decTrig import *
import math

# Symbolic and numeric test cases for sqrt (UFix64 only)
SqrtUFix64Tests = [
    ("0"),
    ("1"),
    ("4"),
    ("16"),
    ("100"),
    ("10000"),
    ("MaxUFix64"),
    ("HalfMaxUFix64"),
    ("0.0001"),
    ("1e-8"),
    ("0.05"),
    ("0.5"),
    ("0.09"),
    ("0.0009"),
    ("2"),
    ("3.14159265"),  # Pi
    ("6.28318530"),  # 2*Pi
    ("1234567890.12345678"),  # Large number
]

LnUFix64Tests = [
    ("1"),
    ("2"),
    ("4"),
    ("10"),
    ("100"),
    ("10000"),
    ("2.71828183"),  # e
    ("7.38905610"),  # e^2
    ("0.5"),
    ("0.25"),
    ("0.125"),
    ("0.01"),
    ("3.14159265"), # Pi
    ("6.28318530"), # 2*Pi
    ("1234567890.12345678"),
    ("MaxUFix64"),
    ("HalfMaxUFix64"),
    ("0.0001"),
    ("1e-8"),
]

# Same input values for Sin and Cos
SinCosFix64Tests = [
    ("0"),
    ("0.1"),
    ("0.25"),
    ("0.5"),
    ("1"),
    ("2"),
    ("10"),
    ("123456.789"),
    ("-0.1"),
    ("-0.25"),
    ("-0.5"),
    ("-1"),
    ("-2"),
    ("-10"),
    ("-123456.789"),
    ("MaxFix64"),
    ("MaxFix64 - 1.0"),
    ("MaxFix64 - 1e-8"),
    ("HalfMaxFix64"),
    ("HalfMaxFix64 - 1.0"),
    ("HalfMaxFix64 - 1e-8"),
    ("HalfMaxFix64 + 1.0"),
    ("HalfMaxFix64 + 1e-8"),
    ("MinFix64"),
    ("MinFix64 + 1.0"),
    ("MinFix64 + 1e-8"),
    ("HalfMinFix64"), 
    ("HalfMinFix64 - 1.0"),
    ("HalfMinFix64 - 1e-8"),
    ("HalfMinFix64 + 1.0"),
    ("HalfMinFix64 + 1e-8"),
    ("0.0001"),
    ("1e-8"),

    # Values of common angles
    ("0.52359878"),   # pi/6
    ("0.78539816"),   # pi/4
    ("1.04719755"),   # pi/3
    ("1.57079633"),   # pi/2
    ("3.14159265"),   # pi
    ("4.71238898"),   # 3*pi/2
    ("6.28318531"),   # 2*pi
    ("2.35619449"),   # 3*pi/4
    ("-0.52359878"),  # -pi/6
    ("-1.57079633"),  # -pi/2
    ("-3.14159265"),  # -pi
    ("-2.35619449"),  # -3*pi/4

    # Same as above, plus 1e-8
    ("0.52359879"),   # pi/6
    ("0.78539817"),   # pi/4
    ("1.04719756"),   # pi/3
    ("1.57079634"),   # pi/2
    ("3.14159265"),   # pi
    ("4.71238899"),   # 3*pi/2
    ("6.28318532"),   # 2*pi
    ("2.35619450"),   # 3*pi/4
    ("-0.52359877"),  # -pi/6
    ("-1.57079632"),  # -pi/2
    ("-3.14159264"),  # -pi
    ("-2.35619448"),  # -3*pi/4

    # Same as above, minus 1e-8
    ("0.52359877"),   # pi/6
    ("0.78539815"),   # pi/4
    ("1.04719754"),   # pi/3
    ("1.57079632"),   # pi/2
    ("3.14159264"),   # pi
    ("4.71238897"),   # 3*pi/2
    ("6.28318530"),   # 2*pi
    ("2.35619448"),   # 3*pi/4
    ("-0.52359879"),  # -pi/6
    ("-1.57079634"),  # -pi/2
    ("-3.14159266"),  # -pi
    ("-2.35619450"),  # -3*pi/4
]

TanFix64Tests = [
    ("0"),
    ("0.1"),
    ("0.25"),
    ("0.5"),
    ("1"),
    ("2"),
    ("10"),
    ("123456.789"),
    ("-0.1"),
    ("-0.25"),
    ("-0.5"),
    ("-1"),
    ("-2"),
    ("-10"),
    ("-123456.789"),
    ("MaxFix64"),
    ("MaxFix64 - 1.0"),
    ("MaxFix64 - 1e-8"),
    ("HalfMaxFix64"),
    ("HalfMaxFix64 - 1.0"),
    ("HalfMaxFix64 - 1e-8"),
    ("HalfMaxFix64 + 1.0"),
    ("HalfMaxFix64 + 1e-8"),
    ("MinFix64"),
    ("MinFix64 + 1.0"),
    ("MinFix64 + 1e-8"),
    ("HalfMinFix64"), 
    ("HalfMinFix64 - 1.0"),
    ("HalfMinFix64 - 1e-8"),
    ("HalfMinFix64 + 1.0"),
    ("HalfMinFix64 + 1e-8"),
    ("0.0001"),
    ("1e-8"),

    # Values of common angles
    ("0.52359878"),   # pi/6
    ("0.78539816"),   # pi/4
    ("1.04719755"),   # pi/3
    ("1.57079633"),   # pi/2
    ("3.14159265"),   # pi
    ("4.71238898"),   # 3*pi/2
    ("6.28318531"),   # 2*pi
    ("2.35619449"),   # 3*pi/4
    ("-0.52359878"),  # -pi/6
    ("-1.57079633"),  # -pi/2
    ("-3.14159265"),  # -pi
    ("-2.35619449"),  # -3*pi/4

    # Same as above, plus 1e-8
    ("0.52359879"),   # pi/6
    ("0.78539817"),   # pi/4
    ("1.04719756"),   # pi/3
    ("1.57079634"),   # pi/2
    ("3.14159265"),   # pi
    ("4.71238899"),   # 3*pi/2
    ("6.28318532"),   # 2*pi
    ("2.35619450"),   # 3*pi/4
    ("-0.52359877"),  # -pi/6
    ("-1.57079632"),  # -pi/2
    ("-3.14159264"),  # -pi
    ("-2.35619448"),  # -3*pi/4

    # Same as above, minus 1e-8
    ("0.52359877"),   # pi/6
    ("0.78539815"),   # pi/4
    ("1.04719754"),   # pi/3
    ("1.57079632"),   # pi/2
    ("3.14159264"),   # pi
    ("4.71238897"),   # 3*pi/2
    ("6.28318530"),   # 2*pi
    ("2.35619448"),   # 3*pi/4
    ("-0.52359879"),  # -pi/6
    ("-1.57079634"),  # -pi/2
    ("-3.14159266"),  # -pi
    ("-2.35619450"),  # -3*pi/4
]

TanFix64OverflowTests = [
    ("1.57079632"),   # pi/2
    ("4.71238898"),   # 3*pi/2
    ("-1.57079632"),  # -pi/2
    ("-4.71238898"),  # -3*pi/2
    ("7.85398163"),    # pi/2 + 2*pi
    ("11.78097245"),   # 3*pi/2 + 2*pi
    ("4.71238898"),    # -pi/2 + 2*pi (which is 3*pi/2)
    ("1.57079632"),    # -3*pi/2 + 2*pi (which is pi/2)
    ("-4.71238898"),   # pi/2 - 2*pi (which is -3*pi/2)
    ("-2.35619449"),   # 3*pi/2 - 2*pi (which is -pi/4)
    ("-7.85398163"),   # -pi/2 - 2*pi
    ("-11.78097245"),  # -3*pi/2 - 2*pi
]

# Same input values for Asin and Acos
AsinAcosFix64Tests = [
    ("0"),
    ("0.1"),
    ("0.25"),
    ("0.5"),
    ("1"),
    ("-0.1"),
    ("-0.25"),
    ("-0.5"),
    ("-1"),
    ("0.99999999"),  # Close to 1
    ("-0.99999999"), # Close to -1
    ("0.70710678"),  # sin(pi/4)
    ("-0.70710678"), # sin(-pi/4)
    ("0.86602540"),  # sin(pi/3)
    ("-0.86602540"), # sin(-pi/3)
    ("0.5"),         # sin(pi/6)
    ("-0.5"),        # sin(-pi/6)
    ("0.25881905"),  # sin(pi/12)
    ("-0.25881905"), # sin(-pi/12)
    ("0.38268343"),  # sin(pi/8)
    ("-0.38268343"), # sin(-pi/8)
    ("1e-8"),
    ("-1e-8"),
]

AsinAcosFix64RangeTests = [
    ("1.00000001"),
    ("-1.00000001"),
    ("2.0"),
    ("-2.0"),
    ("MaxFix64"),
    ("MinFix64"),
    ("HalfMaxFix64"),
    ("HalfMinFix64"),
]


def generate_sqrt_ufix64_tests():
    lines = ["var SqrtUFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in SqrtUFix64Tests:
        a = parseInput64(a_str)
        expected = a.sqrt()
        a_hex = go_hex64(to_ufix64(a))
        expected_hex = go_hex64(to_ufix64(expected))
        data = f"    {{{a_hex}, {expected_hex}}},"
        comment = f"// sqrt({a_str}) = {float(expected)}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")
    return lines

def generate_ln_ufix64_tests():
    lines = ["var LnUFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in LnUFix64Tests:
        a = parseInput64(a_str)
        # ln(x) is undefined for x <= 0, skip or handle as needed
        if a <= 0:
            continue
        expected = a.ln()
        a_hex = go_hex64(to_ufix64(a))
        expected_hex = go_hex64(to_fix64(expected))  # ln can be negative, so use Fix64 encoding
        data = f"    {{{a_hex}, {expected_hex}}},"
        comment = f"// ln({a_str}) = {float(expected):.8f}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")
    return lines

def generate_sin_fix64_tests():
    lines = ["var SinFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in SinCosFix64Tests:
        a = parseInput64(a_str)
        sin_val = decSin(a)

        sin_hex = go_hex64(to_fix64(sin_val))

        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}, {sin_hex}}},"
        comment = f"// sin({a_str}) = {float(sin_val):.8f}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")
    return lines

def generate_cos_fix64_tests():
    lines = ["var CosFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in SinCosFix64Tests:
        a = parseInput64(a_str)
        cos_val = decCos(a)

        cos_hex = go_hex64(to_fix64(cos_val))

        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}, {cos_hex}}},"
        comment = f"// cos({a_str}) = {float(cos_val):.8f}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")
    return lines

def generate_tan_fix64_tests():
    lines = ["var TanFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in TanFix64Tests:
        a = parseInput64(a_str)
        tan_val = decTan(a)

        tan_hex = go_hex64(to_fix64(tan_val))

        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}, {tan_hex}}},"
        comment = f"// tan({a_str}) = {float(tan_val):.8f}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")
    # Handle overflow cases for tan
    lines.append("var TanFix64OverflowTests = []struct{ A uint64 }{")
    for a_str in TanFix64OverflowTests:
        a = parseInput64(a_str)
        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}}},"
        comment = f"// tan({a_str}) = overflow"
        pad = " " * (40 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")

    return lines

def generate_asin_fix64_tests():
    lines = ["var AsinFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in AsinAcosFix64Tests:
        a = parseInput64(a_str)
        if abs(a) > 1:
            continue  # asin is only defined for -1 <= x <= 1
        asin_val = decAsin(a)

        asin_hex = go_hex64(to_fix64(asin_val))

        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}, {asin_hex}}},"
        comment = f"// asin({a_str}) = {float(asin_val):.8f}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")

    lines.append("var AsinFix64RangeTests = []struct{ A uint64 }{")
    for a_str in AsinAcosFix64RangeTests:
        a = parseInput64(a_str)
        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}}},"
        comment = f"// asin({a_str}) = range error"
        pad = " " * (40 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")

    return lines

def generate_acos_fix64_tests():
    lines = ["var AcosFix64Tests = []struct{ A, Expected uint64 }{"]
    for a_str in AsinAcosFix64Tests:
        a = parseInput64(a_str)
        if abs(a) > 1:
            continue  # acos is only defined for -1 <= x <= 1
        acos_val = decAcos(a)

        acos_hex = go_hex64(to_fix64(acos_val))

        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}, {acos_hex}}},"
        comment = f"// acos({a_str}) = {float(acos_val):.8f}"
        pad = " " * (50 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")

    lines.append("var AcosFix64RangeTests = []struct{ A uint64 }{")
    for a_str in AsinAcosFix64RangeTests:
        a = parseInput64(a_str)
        a_hex = go_hex64(to_fix64(a))
        data = f"    {{{a_hex}}},"
        comment = f"// acos({a_str}) = range error"
        pad = " " * (40 - len(data))
        lines.append(f"{data}{pad}{comment}")
    lines.append("}")
    lines.append("")

    return lines

if __name__ == "__main__":
    go_lines = [
        "// Code generated by testgen/trans64.py; DO NOT EDIT.",
        "package fixedPoint",
        "",
    ]
    go_lines.extend(generate_sqrt_ufix64_tests())
    go_lines.extend(generate_ln_ufix64_tests())
    go_lines.extend(generate_sin_fix64_tests())
    go_lines.extend(generate_cos_fix64_tests())
    go_lines.extend(generate_tan_fix64_tests())
    go_lines.extend(generate_asin_fix64_tests())
    go_lines.extend(generate_acos_fix64_tests())
    print('\n'.join(go_lines))