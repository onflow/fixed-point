# add64.py
# Generates Go test data for UFix128 and Fix128 addition (including overflow)

from decimal import Decimal, getcontext
from utils import *

AddUFix128Tests = [
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
    ("0.999999999999999999999999", "1e-24"),

    # Edge cases (upper limit)
    ("MaxUFix128 - 1.0", "1.0"),
    ("MaxUFix128 - 0.1", "0.1"),
    ("MaxUFix128 - 0.01", "0.01"),
    ("MaxUFix128 - 0.001", "0.001"),
    ("MaxUFix128 - 0.0001", "0.0001"),
    ("MaxUFix128 - 0.00001", "0.00001"),
    ("MaxUFix128 - 0.000001", "0.000001"),
    ("MaxUFix128 - 0.0000001", "0.0000001"),
    ("MaxUFix128 - 0.00000001", "0.00000001"),
    ("HalfMaxUFix128", "HalfMaxUFix128"),
    ("HalfMaxUFix128 + 1e-24", "HalfMaxUFix128"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 1e-24"),
]

AddUFix128OverflowTests = [
    ("MaxUFix128", "1.0"),
    ("MaxUFix128", "0.01"),
    ("MaxUFix128", "0.001"),
    ("MaxUFix128", "0.00001"),
    ("MaxUFix128", "0.0000001"),
    ("MaxUFix128", "0.00000001"),
    ("MaxUFix128", "1e-24"),
    ("MaxUFix128", "MaxUFix128"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 1.0"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 0.1"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 0.01"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 0.001"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 0.0001"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 0.00001"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 0.000001"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 1e-23"),
    ("HalfMaxUFix128 + 1e-24", "HalfMaxUFix128 + 1e-24"),
    ("HalfMaxUFix128 + 2e-24", "HalfMaxUFix128"),
    ("HalfMaxUFix128", "HalfMaxUFix128 + 2e-24"),
]

AddFix128Tests = [
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
    ("MaxFix128 - 1.0", "1.0"),
    ("MaxFix128 - 0.1", "0.1"),
    ("MaxFix128 - 0.01", "0.01"),
    ("MaxFix128 - 0.001", "0.001"),
    ("MaxFix128 - 0.0001", "0.0001"),
    ("MaxFix128 - 0.00001", "0.00001"),
    ("MaxFix128 - 0.000001", "0.000001"),
    ("MaxFix128 - 0.0000001", "0.0000001"),
    ("MaxFix128 - 0.00000001", "0.00000001"),
    ("MaxFix128 - 1e-24", "1e-24"),
    ("HalfMaxFix128", "HalfMaxFix128"),
    ("HalfMaxFix128 + 1e-24", "HalfMaxFix128"),
    ("HalfMaxFix128", "HalfMaxFix128 + 1e-24"),

    ("MaxFix128", "-1.0"),
    ("MaxFix128", "0.0"),
    ("MaxFix128", "-0.1"),
    ("MaxFix128", "-0.01"),
    ("MaxFix128", "-0.001"),
    ("MaxFix128", "-0.0001"),
    ("MaxFix128", "-0.00001"),
    ("MaxFix128", "-0.000001"),
    ("MaxFix128", "-0.0000001"),
    ("MaxFix128", "-0.00000001"),
    ("HalfMaxFix128", "HalfMaxFix128"),

    # Edge cases (lower limit)
    ("MinFix128 + 1.0", "-1.0"),
    ("MinFix128 + 0.1", "-0.1"),
    ("MinFix128 + 0.01", "-0.01"),
    ("MinFix128 + 0.001", "-0.001"),
    ("MinFix128 + 0.0001", "-0.0001"),
    ("MinFix128 + 0.00001", "-0.00001"),
    ("MinFix128 + 0.000001", "-0.000001"),
    ("MinFix128 + 0.0000001", "-0.0000001"),
    ("MinFix128 + 0.00000001", "-0.00000001"),
    ("MinFix128 + 1e-24", "-1e-24"),

    ("MinFix128", "1.0"),
    ("MinFix128", "0.1"),
    ("MinFix128", "0.01"),
    ("MinFix128", "0.001"),
    ("MinFix128", "0.0001"),
    ("MinFix128", "0.00001"),
    ("MinFix128", "0.000001"),
    ("MinFix128", "0.0000001"),
    ("MinFix128", "0.00000001"),
    ("MinFix128", "1e-24"),

    ("0", "MinFix128"),
    ("-0.1", "MinFix128 + 0.1"),
    ("-0.01", "MinFix128 + 0.01"),
    ("-0.001", "MinFix128 + 0.001"),
    ("-0.0001", "MinFix128 + 0.0001"),
    ("-0.00001", "MinFix128 + 0.00001"),
    ("-0.000001", "MinFix128 + 0.000001"),
    ("-0.0000001", "MinFix128 + 0.0000001"),
    ("-0.00000001", "MinFix128 + 0.00000001"),
    ("-1e-24", "MinFix128 + 1e-24"),

    ("HalfMinFix128", "HalfMinFix128"),
    ("HalfMinFix128 + 1e-24", "HalfMinFix128 - 1e-24"),
]

AddFix128OverflowTests = [
    ("MaxFix128", "1.0"),
    ("MaxFix128", "0.1"),
    ("MaxFix128", "0.01"),
    ("MaxFix128", "0.001"),
    ("MaxFix128", "0.0001"),
    ("MaxFix128", "0.00001"),
    ("MaxFix128", "0.000001"),
    ("MaxFix128", "0.0000001"),
    ("MaxFix128", "0.00000001"),
    ("MaxFix128", "1e-24"),
    ("MaxFix128", "MaxFix128"),
    ("HalfMaxFix128", "HalfMaxFix128 + 1.0"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.1"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.01"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.0001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.00001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.000001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.0000001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 1e-23"),
    ("HalfMaxFix128 + 1e-24", "HalfMaxFix128 + 1e-24"),
]

AddFix128NegOverflowTests = [
    ("MinFix128", "-1.0"),
    ("MinFix128", "-0.1"),
    ("MinFix128", "-0.01"),
    ("MinFix128", "-0.001"),
    ("MinFix128", "-0.0001"),
    ("MinFix128", "-0.00001"),
    ("MinFix128", "-0.000001"),
    ("MinFix128", "-0.0000001"),
    ("MinFix128", "-0.00000001"),
    ("MinFix128", "-1e-24"),
    ("MinFix128", "MinFix128"),
    ("HalfMinFix128", "HalfMinFix128 - 1.0"),
    ("HalfMinFix128", "HalfMinFix128 - 0.1"),
    ("HalfMinFix128", "HalfMinFix128 - 0.01"),
    ("HalfMinFix128", "HalfMinFix128 - 0.001"),
    ("HalfMinFix128", "HalfMinFix128 - 0.0001"),
    ("HalfMinFix128", "HalfMinFix128 - 0.00001"),
    ("HalfMinFix128", "HalfMinFix128 - 0.000001"),
    ("HalfMinFix128", "HalfMinFix128 - 0.0000001"),
    ("HalfMinFix128", "HalfMinFix128 - 1e-24"),
]

def generate_add_ufix128_tests():
    lines = ["var AddUFix128Tests = []struct{ A, B, Expected raw128 }{"]
    for a_str, b_str in AddUFix128Tests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        c = a + b
        a_hex = go_hex128(to_ufix128(a))
        b_hex = go_hex128(to_ufix128(b))
        c_hex = go_hex128(to_ufix128(c))
        comment = f"// {a_str} + {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_ufix128_overflow_tests():
    lines = ["var AddUFix128OverflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in AddUFix128OverflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_ufix128(a))
        b_hex = go_hex128(to_ufix128(b))
        comment = f"// {a_str} + {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_fix128_tests():
    lines = ["var AddFix128Tests = []struct{ A, B, Expected raw128 }{"]
    for a_str, b_str in AddFix128Tests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        c = a + b
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        c_hex = go_hex128(to_fix128(c))
        comment = f"// {a_str} + {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_fix128_overflow_tests():
    lines = ["var AddFix128OverflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in AddFix128OverflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        comment = f"// {a_str} + {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_add_fix128_neg_overflow_tests():
    lines = ["var AddFix128NegOverflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in AddFix128NegOverflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        comment = f"// {a_str} + {b_str} = neg overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def main():
    getcontext().prec = 100

    go_lines = [
        "// Code generated by testgen/add128.py; DO NOT EDIT.",
        "package fixedPoint",
        "",
    ]
    go_lines.extend(generate_add_ufix128_tests())
    go_lines.append("")
    go_lines.extend(generate_add_ufix128_overflow_tests())
    go_lines.append("")
    go_lines.extend(generate_add_fix128_tests())
    go_lines.append("")
    go_lines.extend(generate_add_fix128_overflow_tests())
    go_lines.append("")
    go_lines.extend(generate_add_fix128_neg_overflow_tests())
    print('\n'.join(go_lines))

if __name__ == "__main__":
    main()
