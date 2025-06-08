# add64.py
# Generates Go test data for UFix64 and Fix64 addition (including overflow)

from decimal import Decimal, getcontext
from utils import to_ufix64, to_fix64, go_hex64, FIX64_SCALE, MASK64, parseInput64

getcontext().prec = 50

UFix64TestValues = [
    # Simple cases
    "0",
    "1",
    "5",

    # Common repeating decimals
    "0.111111111",  # 1/9
    "0.333333333",  # 1/3
    "0.666666666",  # 2/3
    "0.142857142",  # 1/7
    "0.285714285",  # 2/7

    # The smallest non-zero values
    # The code that consumes this list also generages values that
    # are ± 1e-8, adding these two cases (on top of zero) results in
    # the 10th "smallest" values.
    "3e-8",
    "6e-8",
    "9e-8",

    # Random cases
    "123.45678901",
    "456.78901234",
    "0.00012345",
    "0.00045678",
    "98765.4321",
    "31415.9265",
    "27182.8182",
    "1234567890.12345678",

    # Powers of ten
    # "1e-8", # generated above
    # "1e-7", # generated above
    "1e-6",
    "1e-5",
    "1e-4",
    "1e-3",
    "1e-2",
    "1e-1",
    "1e1",
    "1e2",
    "1e3",
    "1e4",
    "1e5",
    "1e6",
    "1e7",
    "1e8",
    "1e9",
    "1e10",
    "1e11",

    # Powers of 2
    "0.00390625",
    "0.0078125",
    "0.015625",
    "0.03125",
    "0.0625",
    "0.125",
    "0.25",
    "0.5",
    "2",
    "4",
    "8",
    "16",
    "32",
    "64",
    "128",
    "256",
    "512",
    "1024",
    "1048576",  # 2^20
    "1073741824",  # 2^30
    "137438953472", # 2^37

    # The prime factors of UINT64_MAX are 3, 5, 17, 257, 641, 65537, and 6700417
    # The values below are different subsets of those numbers multipled together to
    # create values for which some pairs should multiply to exactly UFix64Max.
    "3",
    "15",
    "4391.25228929",
    "27530.74036095",
    "65535",
    "6700417",
    "2814792.71743489",
    "42007935",
    "12297829382.47303441",
    "61489146912.36517205",

    # sqrt(MaxUFix64) and nearby values
    "429496.7296",
    "429496.72959998",
    "429496.72960002",

    # sqrt(HalfMaxUFix64) and sqrt(MaxFix64) and nearby values
    "303700.04999760",
    "303700.04999758",
    "303700.04999762",

    # MaxUFix64 divided by powers of ten
    "184467440737.09551615",
    "18446744073.70955161",
    "1844674407.37095516",
    "184467440.73709552",
    "18446744.07370955",
    "1844674.40737096",
    "184467.44073710",
    "18446.74407370",
    "1844.67440737",
    "184.46744074",
    "18.44674407",
    "1.84467441",

    # MaxFix64 divided by powers of ten
    "92233720368.54775807",
    "9223372036.85477581",
    "922337203.68547758",
    "92233720.36854776",
    "9223372.03685478",
    "922337.20368548",
    "92233.72036855",
    "9223.37203685",
    "922.33720369",
    "92.23372037",
    "9.22337204",

    # Trigonometric values
    "0.52359878",   # pi/6
    "0.78539816",   # pi/4
    "1.04719755",   # pi/3
    "1.57079633",   # pi/2
    "3.14159265",   # pi
    "4.71238898",   # 3*pi/2
    "6.28318531",   # 2*pi
    "2.35619449",   # 3*pi/4
    "1.41421356",   # sqrt(2)
    "0.70710678",   # sqrt(2) / 2

    # Logarithmic values
    "0.69314718",   # ln(2)
    "2.302585092",  # ln(10)
    "2.71828183",   # e
    "7.38905610",   # e^2

    # Near the limits
    "MaxUFix64",
    "MaxUFix64 - 1",
    "HalfMaxUFix64",
    "HalfMaxUFix64 + 1",
    "HalfMaxUFix64 - 1",
]

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
        a_hex = go_hex64(to_ufix64(a))
        b_hex = go_hex64(to_ufix64(b))
        c_hex = go_hex64(to_ufix64(c))
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
        a_hex = go_hex64(to_ufix64(a))
        b_hex = go_hex64(to_ufix64(b))
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
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        c_hex = go_hex64(to_fix64(c))
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
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
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
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        comment = f"// {a_str} + {b_str} = neg overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def main():
    go_lines = [
        "// Code generated by testgen/add64.py; DO NOT EDIT.",
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
