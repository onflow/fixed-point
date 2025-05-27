# add64.py
# Generates Go test data for UFix64 and Fix64 addition (including overflow)

from decimal import Decimal, getcontext
from utils import to_ufix64, to_fix64, go_hex, FIX64_SCALE, MASK, parseInput64

getcontext().prec = 50

AddUFix64Tests = [
    # Simple cases
    ("1.0", "1.0"),
    ("1.0", "0.0"),
    ("0.0", "0.0"),
    ("0.0", "1.0"),
    ("1.0", "1e8"),
    ("1.0", "100000001.0"),

    # Random cases
    ("123.456", "789.012"),
    ("456.789", "123.456"),
    ("0.000123", "0.000456"),
    ("0.000789", "0.000321"),
    ("98765.4321", "12345.6789"),
    ("31415.9265", "27182.8182"),
    ("27182.8182", "31415.9265"),
    ("1.23456789", "0.98765432"),
    ("0.99999999", "0.00000001"),

    # Edge cases (upper limit)
    ("MaxUFix64 - 1.0", "1.0"),
    ("MaxUFix64 - 0.1", "0.1"),
    ("MaxUFix64 - 0.01", "0.01"),
    ("MaxUFix64 - 0.001", "0.001"),
    ("MaxUFix64 - 0.0001", "0.0001"),
    ("MaxUFix64 - 0.00001", "0.00001"),
    ("MaxUFix64 - 0.000001", "0.000001"),
    ("MaxUFix64 - 0.0000001", "0.0000001"),
    ("MaxUFix64 - 0.00000001", "0.00000001"),
    ("HalfMaxUFix64", "HalfMaxUFix64"),
    ("HalfMaxUFix64 + 0.00000001", "HalfMaxUFix64"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.00000001"),
]

AddUFix64OverflowTests = [
    ("MaxUFix64", "1.0"),
    ("MaxUFix64", "0.01"),
    ("MaxUFix64", "0.001"),
    ("MaxUFix64", "0.00001"),
    ("MaxUFix64", "0.0000001"),
    ("MaxUFix64", "0.00000001"),
    ("MaxUFix64", "MaxUFix64"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 1.0"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.1"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.01"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.001"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.0001"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.00001"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.000001"),
    ("HalfMaxUFix64 + 0.00000001", "HalfMaxUFix64 + 0.00000001"),
    ("HalfMaxUFix64 + 0.00000002", "HalfMaxUFix64"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.00000002"),
]

AddFix64Tests = [
    # Simple cases
    ("1.0", "1.0"),
    ("1.0", "0.0"),
    ("0.0", "0.0"),
    ("0.0", "1.0"),
    ("1.0", "2.0"),
    ("-1.0", "2.0"),
    ("1.0", "-2.0"),
    ("-1.0", "-2.0"),
    ("1.0", "1e8"),
    ("1.0", "100000001.0"),

    # Random cases
    ("1.0", "99999999.0"),
    ("123.456", "789.012"),
    ("-456.789", "123.456"),
    ("0.000123", "0.000456"),
    ("-0.000789", "0.000321"),
    ("98765.4321", "-12345.6789"),
    ("31415.9265", "27182.8182"),
    ("-27182.8182", "-31415.9265"),
    ("1.23456789", "-0.98765432"),
    ("0.99999999", "0.00000001"),
    ("-0.99999999", "-0.00000001"),

    # Edge cases (upper limit)
    ("MaxFix64 - 1.0", "1.0"),
    ("MaxFix64 - 0.1", "0.1"),
    ("MaxFix64 - 0.01", "0.01"),
    ("MaxFix64 - 0.001", "0.001"),
    ("MaxFix64 - 0.0001", "0.0001"),
    ("MaxFix64 - 0.00001", "0.00001"),
    ("MaxFix64 - 0.000001", "0.000001"),
    ("MaxFix64 - 0.0000001", "0.0000001"),
    ("MaxFix64 - 0.00000001", "0.00000001"),
    ("HalfMaxFix64", "HalfMaxFix64"),
    ("HalfMaxFix64 + 0.00000001", "HalfMaxFix64"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.00000001"),

    ("MaxFix64", "-1.0"),
    ("MaxFix64", "0.0"),
    ("MaxFix64", "-0.1"),
    ("MaxFix64", "-0.01"),
    ("MaxFix64", "-0.001"),
    ("MaxFix64", "-0.0001"),
    ("MaxFix64", "-0.00001"),
    ("MaxFix64", "-0.000001"),
    ("MaxFix64", "-0.0000001"),
    ("MaxFix64", "-0.00000001"),
    ("HalfMaxFix64", "HalfMaxFix64"),

    # Edge cases (lower limit)
    ("MinFix64 + 1.0", "-1.0"),
    ("MinFix64 + 0.1", "-0.1"),
    ("MinFix64 + 0.01", "-0.01"),
    ("MinFix64 + 0.001", "-0.001"),
    ("MinFix64 + 0.0001", "-0.0001"),
    ("MinFix64 + 0.00001", "-0.00001"),
    ("MinFix64 + 0.000001", "-0.000001"),
    ("MinFix64 + 0.0000001", "-0.0000001"),
    ("MinFix64 + 0.00000001", "-0.00000001"),

    ("MinFix64", "1.0"),
    ("MinFix64", "0.1"),
    ("MinFix64", "0.01"),
    ("MinFix64", "0.001"),
    ("MinFix64", "0.0001"),
    ("MinFix64", "0.00001"),
    ("MinFix64", "0.000001"),
    ("MinFix64", "0.0000001"),
    ("MinFix64", "0.00000001"),

    ("0", "MinFix64"),
    ("-0.1", "MinFix64 + 0.1"),
    ("-0.01", "MinFix64 + 0.01"),
    ("-0.001", "MinFix64 + 0.001"),
    ("-0.0001", "MinFix64 + 0.0001"),
    ("-0.00001", "MinFix64 + 0.00001"),
    ("-0.000001", "MinFix64 + 0.000001"),
    ("-0.0000001", "MinFix64 + 0.0000001"),
    ("-0.00000001", "MinFix64 + 0.00000001"),

    ("HalfMinFix64", "HalfMinFix64"),
    ("HalfMinFix64 + 0.00000001", "HalfMinFix64 - 0.00000001"),
]

AddFix64OverflowTests = [
    ("MaxFix64", "1.0"),
    ("MaxFix64", "0.1"),
    ("MaxFix64", "0.01"),
    ("MaxFix64", "0.001"),
    ("MaxFix64", "0.0001"),
    ("MaxFix64", "0.00001"),
    ("MaxFix64", "0.000001"),
    ("MaxFix64", "0.0000001"),
    ("MaxFix64", "0.00000001"),
    ("MaxFix64", "MaxFix64"),
    ("HalfMaxFix64", "HalfMaxFix64 + 1.0"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.1"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.01"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.001"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.0001"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.00001"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.000001"),
    ("HalfMaxFix64", "HalfMaxFix64 + 0.0000001"),
    ("HalfMaxFix64 + 0.00000001", "HalfMaxFix64 + 0.00000001"),
]

AddFix64NegOverflowTests = [
    ("MinFix64", "-1.0"),
    ("MinFix64", "-0.1"),
    ("MinFix64", "-0.01"),
    ("MinFix64", "-0.001"),
    ("MinFix64", "-0.0001"),
    ("MinFix64", "-0.00001"),
    ("MinFix64", "-0.000001"),
    ("MinFix64", "-0.0000001"),
    ("MinFix64", "-0.00000001"),
    ("MinFix64", "MinFix64"),
    ("HalfMinFix64", "HalfMinFix64 - 1.0"),
    ("HalfMinFix64", "HalfMinFix64 - 0.1"),
    ("HalfMinFix64", "HalfMinFix64 - 0.01"),
    ("HalfMinFix64", "HalfMinFix64 - 0.001"),
    ("HalfMinFix64", "HalfMinFix64 - 0.0001"),
    ("HalfMinFix64", "HalfMinFix64 - 0.00001"),
    ("HalfMinFix64", "HalfMinFix64 - 0.000001"),
    ("HalfMinFix64", "HalfMinFix64 - 0.0000001"),
]

def generate_add_ufix64_tests():
    lines = ["var AddUFix64Tests = []struct{ A, B, Expected uint64 }{"]
    for a_str, b_str in AddUFix64Tests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        c = a + b
        a_hex = go_hex(to_ufix64(a))
        b_hex = go_hex(to_ufix64(b))
        c_hex = go_hex(to_ufix64(c))
        comment = f"// {a_str} + {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_ufix64_overflow_tests():
    lines = ["var AddUFix64OverflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in AddUFix64OverflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex(to_ufix64(a))
        b_hex = go_hex(to_ufix64(b))
        comment = f"// {a_str} + {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_fix64_tests():
    lines = ["var AddFix64Tests = []struct{ A, B, Expected uint64 }{"]
    for a_str, b_str in AddFix64Tests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        c = a + b
        a_hex = go_hex(to_fix64(a))
        b_hex = go_hex(to_fix64(b))
        c_hex = go_hex(to_fix64(c))
        comment = f"// {a_str} + {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_fix64_overflow_tests():
    lines = ["var AddFix64OverflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in AddFix64OverflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex(to_fix64(a))
        b_hex = go_hex(to_fix64(b))
        comment = f"// {a_str} + {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_fix64_neg_overflow_tests():
    lines = ["var AddFix64NegOverflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in AddFix64NegOverflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex(to_fix64(a))
        b_hex = go_hex(to_fix64(b))
        comment = f"// {a_str} + {b_str} = neg overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def main():
    go_lines = [
        "// Code generated by testgen/add.py; DO NOT EDIT.",
        "package fixedPoint",
        "",
    ]
    go_lines.extend(generate_add_ufix64_tests())
    go_lines.append("")
    go_lines.extend(generate_add_ufix64_overflow_tests())
    go_lines.append("")
    go_lines.extend(generate_add_fix64_tests())
    go_lines.append("")
    go_lines.extend(generate_add_fix64_overflow_tests())
    go_lines.append("")
    go_lines.extend(generate_add_fix64_neg_overflow_tests())
    print('\n'.join(go_lines))

if __name__ == "__main__":
    main()
