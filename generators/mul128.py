# mul128.py - Generates Go test data for UFix128 and Fix128 multiplication (including overflow/underflow)

from decimal import Decimal, getcontext
from utils import *

MulUFix128Tests = [
    # Simple cases
    ("1.0", "1.0"),
    ("1.0", "0.0"),
    ("0.0", "0.0"),
    ("1.0", "1e8"),
    ("1.0", "100000001.0"),
    ("1.0", "100000001.00000001"),
    ("3.0", "6700417.0"),
    ("3.0", "5.0"),
    ("3.0", "17.0"),
    ("3.0", "257.0"),
    ("3.0", "641.0"),
    ("3.0", "65537.0"),

    # Random cases
    ("123.456", "789.012"),
    ("456.789", "123.456"),
    ("0.000123", "0.000456"),
    ("0.000789", "0.000321"),
    ("98765.4321", "12345.6789"),
    ("31415.9265", "27182.8182"),
    ("1.23456789", "0.98765432"),
    ("0.99999999", "0.00000001"),
    ("1e-9", "MaxUFix128"),

    # Slightly less than the sqrt of UFix128Max
    ("18446744.073709551615", "18446744.073709551615"),
    ("18446744.073709551615999999999999", "18446744.073709551615999999999999"),

    ("MaxUFix128", "1.0"),
    ("MaxUFix128", "0.1"),
    ("MaxUFix128", "0.01"),
    ("MaxUFix128", "0.001"),
    ("MaxUFix128", "0.0001"),
    ("MaxUFix128", "0.00001"),
    ("MaxUFix128", "0.000001"),
    ("MaxUFix128", "0.0000001"),
    ("MaxUFix128", "0.00000001"),
    ("MaxUFix128", "1e-24"),
    ("MaxUFix128", "0.0"),
    ("MaxUFix128 - 1.0", "1.0"),
    ("MaxUFix128 - 1e-24", "1e-24"),
    ("HalfMaxUFix128", "2.0"),

    # Multiply to the minimum UFix128 value
    ("1e-24", "1e0"),
    ("1e-23", "1e-1"),
    ("1e-22", "1e-2"),
    ("1e-21", "1e-3"),
    ("1e-20", "1e-4"),
    ("1e-15", "1e-9"),
    ("5e-24", "0.2"),
    ("2e-24", "0.5"),
]

MulUFix128OverflowTests = [
    ("MaxUFix128", "1.1"),
    ("MaxUFix128", "1.01"),
    ("MaxUFix128", "1.001"),
    ("MaxUFix128", "1.00001"),
    ("MaxUFix128", "1.0000001"),
    ("MaxUFix128", "1.000000000000000000000001"),
    ("MaxUFix128", "MaxUFix128"),

    # sqrt(MaxUFix128)
    ("18446744.073709551616", "18446744.073709551616"),

    ("340282366920938.463463374607431768211455", "1.000000000000000000000001"),
    ("34028236692093.846346337460743176821146", "10.000000000000000000000001"),
    ("3402823669209.384634633746074317682115", "100.000000000000000000000001"),
    ("340282366920.938463463374607431768212", "1000.000000000000000000000001"),
    ("34028236692.093846346337460743176822", "10000.000000000000000000000001"),
    ("3402823669.209384634633746074317683", "100000.000000000000000000000001"),
    ("340282366.920938463463374607431769", "1000000.000000000000000000000001"),
    ("34028236.692093846346337460743177", "10000000.000000000000000000000001"),
    ("3402823.669209384634633746074318", "100000000.000000000000000000000001"),
    ("340282.366920938463463374607432", "1000000000.000000000000000000000001"),
    ("34028.236692093846346337460744", "10000000000.000000000000000000000001"),
    ("3402.823669209384634633746075", "100000000000.000000000000000000000001"),
    ("340.282366920938463463374608", "1000000000000.000000000000000000000001"),
    ("34.028236692093846346337461", "10000000000000.000000000000000000000001"),
    ("3.402823669209384634633747", "100000000000000.000000000000000000000001"),

    ("HalfMaxUFix128", "2.000000000000000000000001"),
    ("HalfMaxUFix128", "2.00000001"),
    ("HalfMaxUFix128", "2.0000001"),
    ("HalfMaxUFix128", "2.000001"),
    ("HalfMaxUFix128", "2.0001"),
    ("HalfMaxUFix128", "2.001"),
    ("HalfMaxUFix128", "2.01"),
    ("HalfMaxUFix128", "2.1"),
]

MulUFix128UnderflowTests = [
    ("1e-24", "1e-24"),
    ("1e-23", "1e-23"),
    ("1e-20", "1e-20"),
    ("1e-15", "1e-15"),
    ("9e-13", "9e-13"),

    ("1e-24", "1e-1"),
    ("1e-23", "1e-2"),
    ("1e-22", "1e-3"),
    ("1e-21", "1e-4"),
    ("1e-20", "1e-5"),
    ("1e-15", "1e-10"),
    ("5e-24", "0.02"),
    ("2e-24", "0.05"),

    ("0.999999999999999999999999", "1e-24"),
    ("0.099999999999999999999999", "1e-23"),
    ("0.009999999999999999999999", "1e-22"),

    ("5e-24", "0.199999999999999999999999"),
    ("2e-24", "0.499999999999999999999999"),
]

MulFix128Tests = [
    # Simple cases
    ("1.0", "1.0"),
    ("-1.0", "-1.0"),
    ("1.0", "-1.0"),
    ("-1.0", "1.0"),
    ("1.0", "0.0"),
    ("0.0", "0.0"),
    ("1.0", "1e8"),
    ("1.0", "100000001.0"),
    ("1.0", "100000000000000.000000000000000000000001"),
    ("3.0", "-6700417.0"),
    ("3.0", "5.0"),
    ("-3.0", "17.0"),
    ("3.0", "257.0"),
    ("-3.0", "641.0"),
    ("-3.0", "-65537.0"),

    # Slightly less than the sqrt of Fix128Max
    
    ("13043817.8253327817147311798", "13043817.8253327817147311798"),
    ("13043817.825332781714731179800509", "13043817.825332781714731179800509"),
    ("13043817.8253327817147311798", "-13043817.8253327817147311798"),
    ("13043817.825332781714731179800509", "-13043817.825332781714731179800509"),
    ("-13043817.8253327817147311798", "13043817.8253327817147311798"),
    ("-13043817.825332781714731179800509", "13043817.825332781714731179800509"),
    ("-13043817.8253327817147311798", "-13043817.8253327817147311798"),
    ("-13043817.825332781714731179800509", "-13043817.825332781714731179800509"),


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
    ("MaxFix128", "0.0"),
    ("MaxFix128 - 1.0", "1.0"),
    ("HalfMaxFix128", "2.0"),

    ("MaxFix128", "-1.0"),
    ("MaxFix128", "-0.1"),
    ("MaxFix128", "-0.01"),
    ("MaxFix128", "-0.001"),
    ("MaxFix128", "-0.0001"),
    ("MaxFix128", "-0.00001"),
    ("MaxFix128", "-0.000001"),
    ("MaxFix128", "-0.0000001"),
    ("MaxFix128", "-0.00000001"),
    ("MaxFix128", "-1e-24"),
    ("MaxFix128", "0.0"),
    ("MaxFix128 - 1.0", "-1.0"),
    ("HalfMaxFix128", "-2.0"),

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
    ("MinFix128", "0.0"),
    ("MinFix128 + 1.0", "1.0"),
    ("HalfMinFix128", "2.0"),

    ("MinFix128 + 1e-24", "-1.0"),
    ("MinFix128", "-0.1"),
    ("MinFix128", "-0.01"),
    ("MinFix128", "-0.001"),
    ("MinFix128", "-0.0001"),
    ("MinFix128", "-0.00001"),
    ("MinFix128", "-0.000001"),
    ("MinFix128", "-0.0000001"),
    ("MinFix128", "-0.00000001"),
    ("MinFix128", "-1e-24"),
    ("MinFix128", "0.0"),
    ("MinFix128 + 1.0", "-1.0"),
    ("HalfMinFix128 + 1e-24", "-2.0"),


    # Multiply to the smallest magnitude Fix128 value
    ("1e-24", "1e0"),
    ("1e-23", "1e-1"),
    ("1e-22", "1e-2"),
    ("1e-21", "1e-3"),
    ("1e-20", "1e-4"),
    ("1e-15", "1e-9"),
    ("5e-24", "0.2"),
    ("2e-24", "0.5"),

    ("1e-24", "-1e0"),
    ("1e-23", "-1e-1"),
    ("1e-22", "-1e-2"),
    ("1e-21", "-1e-3"),
    ("1e-20", "-1e-4"),
    ("1e-15", "-1e-9"),
    ("5e-24", "-0.2"),
    ("2e-24", "-0.5"),

    ("-1e-24", "1e0"),
    ("-1e-23", "1e-1"),
    ("-1e-22", "1e-2"),
    ("-1e-21", "1e-3"),
    ("-1e-20", "1e-4"),
    ("-1e-15", "1e-9"),
    ("-5e-24", "0.2"),
    ("-2e-24", "0.5"),

    ("-1e-24", "-1e0"),
    ("-1e-23", "-1e-1"),
    ("-1e-22", "-1e-2"),
    ("-1e-21", "-1e-3"),
    ("-1e-20", "-1e-4"),
    ("-1e-15", "-1e-9"),
    ("-5e-24", "-0.2"),
    ("-2e-24", "-0.5"),

]

MulFix128OverflowTests = [
    ("MaxFix128", "1.1"),
    ("MaxFix128", "1.01"),
    ("MaxFix128", "1.001"),
    ("MaxFix128", "1.00001"),
    ("MaxFix128", "1.0000001"),
    ("MaxFix128", "1.000000000000000000000001"),
    ("MaxFix128", "MaxFix128"),
    ("HalfMaxFix128", "HalfMaxFix128 + 1.0"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.1"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.01"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.0001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.00001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 0.000001"),
    ("HalfMaxFix128", "HalfMaxFix128 + 1e-23"),
    ("HalfMaxFix128 + 0.00000001", "HalfMaxFix128 + 0.00000001"),
    ("HalfMinFix128", "HalfMinFix128 + 1.0"),
    ("HalfMinFix128", "HalfMinFix128 + 0.1"),
    ("HalfMinFix128", "HalfMinFix128 + 0.01"),
    ("HalfMinFix128", "HalfMinFix128 + 0.001"),
    ("HalfMinFix128", "HalfMinFix128 + 0.0001"),
    ("HalfMinFix128", "HalfMinFix128 + 0.00001"),
    ("HalfMinFix128", "HalfMinFix128 + 0.000001"),
    ("HalfMinFix128", "HalfMinFix128 + 0.0000001"),
    ("HalfMinFix128", "HalfMinFix128 + 1e-24"),
]

MulFix128UnderflowTests = [
    ("1e-24", "1e-24"),
    ("1e-23", "1e-23"),
    ("1e-20", "1e-20"),
    ("1e-15", "1e-15"),
    ("9e-13", "9e-13"),

    ("1e-24", "-1e-24"),
    ("1e-23", "-1e-23"),
    ("1e-20", "-1e-20"),
    ("1e-15", "-1e-15"),
    ("9e-13", "-9e-13"),

    ("-1e-24", "1e-24"),
    ("-1e-23", "1e-23"),
    ("-1e-20", "1e-20"),
    ("-1e-15", "1e-15"),
    ("-9e-13", "9e-13"),

    ("-1e-24", "-1e-24"),
    ("-1e-23", "-1e-23"),
    ("-1e-20", "-1e-20"),
    ("-1e-15", "-1e-15"),
    ("-9e-13", "-9e-13"),

    ("1e-24", "1e-1"),
    ("1e-23", "1e-2"),
    ("1e-22", "1e-3"),
    ("1e-21", "1e-4"),
    ("1e-20", "1e-5"),
    ("1e-15", "1e-10"),
    ("5e-24", "0.02"),
    ("2e-24", "0.05"),

    ("1e-24", "-1e-1"),
    ("1e-23", "-1e-2"),
    ("1e-22", "-1e-3"),
    ("1e-21", "-1e-4"),
    ("1e-20", "-1e-5"),
    ("1e-15", "-1e-10"),
    ("5e-24", "-0.02"),
    ("2e-24", "-0.05"),

    ("-1e-24", "1e-1"),
    ("-1e-23", "1e-2"),
    ("-1e-22", "1e-3"),
    ("-1e-21", "1e-4"),
    ("-1e-20", "1e-5"),
    ("-1e-15", "1e-10"),
    ("-5e-24", "0.02"),
    ("-2e-24", "0.05"),

    ("-1e-24", "-1e-1"),
    ("-1e-23", "-1e-2"),
    ("-1e-22", "-1e-3"),
    ("-1e-21", "-1e-4"),
    ("-1e-20", "-1e-5"),
    ("-1e-15", "-1e-10"),
    ("-5e-24", "-0.02"),
    ("-2e-24", "-0.05"),


    ("0.999999999999999999999999", "1e-24"),
    ("0.099999999999999999999999", "1e-23"),
    ("0.009999999999999999999999", "1e-22"),
    ("5e-24", "0.199999999999999999999999"),
    ("2e-24", "0.499999999999999999999999"),

    ("0.999999999999999999999999", "-1e-24"),
    ("0.099999999999999999999999", "-1e-23"),
    ("0.009999999999999999999999", "-1e-22"),
    ("5e-24", "-0.199999999999999999999999"),
    ("2e-24", "-0.499999999999999999999999"),

    ("-0.999999999999999999999999", "1e-24"),
    ("-0.099999999999999999999999", "1e-23"),
    ("-0.009999999999999999999999", "1e-22"),
    ("-5e-24", "0.199999999999999999999999"),
    ("-2e-24", "0.499999999999999999999999"),

    ("-0.999999999999999999999999", "-1e-24"),
    ("-0.099999999999999999999999", "-1e-23"),
    ("-0.009999999999999999999999", "-1e-22"),
    ("-5e-24", "-0.199999999999999999999999"),
    ("-2e-24", "-0.499999999999999999999999"),
]


MulFix128NegOverflowTests = [
    ("MinFix128", "1.000000000000000000000001"),
    ("MinFix128", "1.00000001"),
    ("MinFix128", "1.0000001"),
    ("MinFix128", "1.1"),
    ("MinFix128", "2.0"),
    ("MinFix128", "10"),
    ("MinFix128", "100"),
    ("MinFix128", "1000"),
    ("MinFix128", "10000"),
    ("MinFix128", "100000"),
    ("MinFix128", "1000000"),
    ("MinFix128", "10000000"),
    ("MinFix128", "100000000"),
    ("MinFix128", "1000000000"),
    ("MinFix128", "MaxFix128"),
    ("HalfMaxFix128", "HalfMinFix128 + 1.0"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.1"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.01"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.001"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.0001"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.00001"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.000001"),
    ("HalfMaxFix128", "HalfMinFix128 + 0.0000001"),
    ("HalfMaxFix128", "HalfMinFix128 + 1e-24"),
]

def generate_mul_ufix128_tests():
    lines = ["var MulUFix128Tests = []struct{ A, B, Expected raw128 }{"]
    for a_str, b_str in MulUFix128Tests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        c = a * b
        a_hex = go_hex128(to_ufix128(a))
        b_hex = go_hex128(to_ufix128(b))
        c_hex = go_hex128(to_ufix128(c))
        comment = f"// {a_str} * {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_mul_ufix128_overflow_tests():
    lines = ["var MulUFix128OverflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in MulUFix128OverflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_ufix128(a))
        b_hex = go_hex128(to_ufix128(b))
        comment = f"// {a_str} * {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_mul_ufix128_underflow_tests():
    lines = ["var MulUFix128UnderflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in MulUFix128UnderflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_ufix128(a))
        b_hex = go_hex128(to_ufix128(b))
        comment = f"// {a_str} * {b_str} = underflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_mul_fix128_tests():
    lines = ["var MulFix128Tests = []struct{ A, B, Expected raw128 }{"]
    for a_str, b_str in MulFix128Tests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        c = a * b
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        c_hex = go_hex128(to_fix128(c))
        comment = f"// {a_str} * {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_mul_fix128_overflow_tests():
    lines = ["var MulFix128OverflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in MulFix128OverflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        comment = f"// {a_str} * {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_mul_fix128_underflow_tests():
    lines = ["var MulFix128UnderflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in MulFix128UnderflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        comment = f"// {a_str} * {b_str} = underflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_mul_fix128_neg_overflow_tests():
    lines = ["var MulFix128NegOverflowTests = []struct{ A, B raw128 }{"]
    for a_str, b_str in MulFix128NegOverflowTests:
        a = parseInput128(a_str)
        b = parseInput128(b_str)
        a_hex = go_hex128(to_fix128(a))
        b_hex = go_hex128(to_fix128(b))
        comment = f"// {a_str} * {b_str} = neg overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def main():
    getcontext().prec = 100  # Set precision for Decimal operations

    go_lines = [
        "// Code generated by testgen/mul.py; DO NOT EDIT.",
        "package fixedPoint",
        "",
    ]
    go_lines.extend(generate_mul_ufix128_tests())
    go_lines.extend(generate_mul_ufix128_overflow_tests())
    go_lines.extend(generate_mul_ufix128_underflow_tests())
    go_lines.extend(generate_mul_fix128_tests())
    go_lines.extend(generate_mul_fix128_overflow_tests())
    go_lines.extend(generate_mul_fix128_underflow_tests())
    go_lines.extend(generate_mul_fix128_neg_overflow_tests())
    print('\n'.join(go_lines))

if __name__ == "__main__":
    main()
