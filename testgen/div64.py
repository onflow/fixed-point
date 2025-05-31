# div64.py
# Generates Go test data for UFix64 and Fix64 division (including overflow/div by zero)

from decimal import Decimal, getcontext
from utils import to_ufix64, to_fix64, go_hex64, FIX64_SCALE, MASK64, parseInput64

getcontext().prec = 50

DivUFix64Tests = [
    # Simple cases
    ("1.0", "1.0"),
    ("1.0", "1e8"),
    ("10.0", "100000001.0"),
    ("1e8", "1e8"),
    ("1e8", "99999999.0"),
    ("1e8", "100000001.0"),
    ("5", "1"),
    ("5", "2"),
    ("5", "3"),
    ("5", "4"),
    ("5", "5"),
    ("5", "6"),
    ("5", "7"),
    ("5", "8"),
    ("5", "9"),
    ("5", "10"),

    # Random cases
    ("123.456", "789.012"),
    ("456.789", "123.456"),
    ("0.000123", "0.000456"),
    ("0.000789", "0.000321"),
    ("98765.4321", "12345.6789"),
    ("31415.9265", "27182.8182"),
    ("1.23456789", "0.98765432"),
    ("0.99999999", "0.00000001"),

    # The prime factors of UINT64_MAX are 3, 5, 17, 257, 641, 65537, and 6700417
    # The values on the right hand side are products of some of these factors
    ("MaxUFix64", "0x280fffffd7f"),
    ("MaxUFix64", "42007935"),
    ("MaxUFix64", "65535"),
    ("MaxUFix64", "255"),
    ("MaxUFix64", "4369"),
    ("MaxUFix64", "3"),
    ("MaxUFix64", "5"),
    ("MaxUFix64", "17"),
    ("MaxUFix64", "257"),
    ("MaxUFix64", "641"),
    ("MaxUFix64", "65537"),
    ("MaxUFix64", "6700417"),
    ("MaxUFix64", "494211"),

    # Near the square root of 2^64
    ("MaxUFix64", "429496.7296"),
    ("MaxUFix64", "429496.7295"),
    ("MaxUFix64", "429496.72959999"),

    # Starting big
    ("MaxUFix64", "1"),
    ("MaxUFix64", "10"),
    ("MaxUFix64", "100"),
    ("MaxUFix64", "1000"),
    ("MaxUFix64", "10000"),
    ("MaxUFix64", "100000"),
    ("MaxUFix64", "1000000"),
    ("MaxUFix64", "10000000"),
    ("MaxUFix64", "100000000"),
    ("MaxUFix64", "1000000000"),
    ("MaxUFix64", "10000000000"),
    ("MaxUFix64", "100000000000"),

    ("MaxUFix64 - 1", "1"),
    ("HalfMaxUFix64", "0.5"),

    ("MaxUFix64", "1.0"),
    ("MaxUFix64", "2.0"),
    ("MaxUFix64", "MaxUFix64"),
    ("HalfMaxUFix64", "2.0"),
    ("HalfMaxUFix64", "HalfMaxUFix64"),
    ("HalfMaxUFix64 + 0.00000001", "HalfMaxUFix64"),
    ("HalfMaxUFix64", "HalfMaxUFix64 + 0.00000001"),

    # Things that divide to the smallest UFix64
    ("0.00000001", "1"),
    ("0.0000001", "10"),
    ("0.000001", "100"),
    ("0.00001", "1000"),
    ("0.0001", "10000"),
    ("0.001", "100000"),
    ("0.01", "1000000"),
    ("0.1", "10000000"),
    ("1.0", "100000000"),
    ("0.00000005", "5"),
    ("0.00000002", "2"),

    # Same as above, but with a SLIGHTLY smaller divisor
    ("0.00000001", "0.99999999"),
    ("0.0000001", "9.99999999"),
    ("0.000001", "99.99999999"),
    ("0.00001", "999.99999999"),
    ("0.0001", "9999.99999999"),
    ("0.001", "99999.99999999"),
    ("0.01", "999999.99999999"),
    ("0.1", "9999999.99999999"),
    ("1.0", "99999999.99999999"),

    # Should divide to the largest UFix64
    ("184467440737.09551615", "1.0"),
    ("18446744073.70955161", "0.1"),
    ("1844674407.37095516", "0.01"),
    ("184467440.73709551", "0.001"),
    ("18446744.07370955", "0.0001"),
    ("1844674.40737095", "0.00001"),
    ("184467.44073709", "0.000001"),
    ("18446.74407370", "0.0000001"),
    ("1844.67440737", "0.00000001"),
]

DivUFix64OverflowTests = [
    # Overflow cases for UFix64 division
    ("MaxUFix64", "0.99999999"),
    ("MaxUFix64", "0.1"),
    ("MaxUFix64", "0.01"),
    ("MaxUFix64", "0.001"),
    ("MaxUFix64", "0.00001"),
    ("MaxUFix64", "0.0000001"),

    ("184467440737.09551615", "0.99999999"),
    ("18446744073.70955161", "0.09999999"),
    ("1844674407.37095516", "0.00999999"),
    ("184467440.73709551", "0.00099999"),
    ("18446744.07370955", "0.00009999"),
    ("1844674.40737095", "0.00000999"),
    ("184467.44073709", "0.00000099"),
    ("18446.74407370", "0.00000009"),

    ("HalfMaxUFix64", "0.00000001"),
    ("HalfMaxUFix64", "0.0000001"),
    ("HalfMaxUFix64", "0.000001"),
    ("HalfMaxUFix64", "0.0001"),
    ("HalfMaxUFix64", "0.001"),
    ("HalfMaxUFix64", "0.01"),
    ("HalfMaxUFix64", "0.1"),
]

DivUFix64UnderflowTests = [
    ("0.00000001", "2.0"),
    ("0.00000001", "10.0"),
    ("0.00000001", "100.0"),
    ("0.00000001", "1000.0"),
    ("0.00000001", "100000000.0"),
    ("1.0", "184467440737.09551615"),
    ("1.0", "100000000000.0"),
    ("1.0", "99999999999.0"),
    ("0.00000001", "184467440737.09551615"),
    ("0.00000001", "99999999999.0"),
]

DivUFix64DivByZeroTests = [
    ("1.0", "0.0"),
    ("0.0", "0.0"),
    ("MaxUFix64", "0.0"),
    ("HalfMaxUFix64", "0.0"),
]


DivFix64Tests = [
    # Simple cases
    ("1.0", "1.0"),
    ("1.0", "1e8"),
    ("1.1", "100000001.0"),
    ("10.0", "100000001.0"),
    ("1e8", "1e8"),
    ("1e8", "99999999.0"),
    ("1e8", "100000001.0"),
    ("5", "-1"),
    ("5", "-2"),
    ("5", "-3"),
    ("5", "-4"),
    ("5", "-5"),
    ("5", "-6"),
    ("5", "-7"),
    ("5", "-8"),
    ("5", "-9"),
    ("5", "-10"),
    ("1.0", "0.5"),
    ("1.0", "2.0"),
    ("2.0", "1.0"),
    ("1.0", "-2.0"),
    ("-2.0", "2.0"),
    ("-2.0", "-2.0"),
    ("1.0", "99999999.0"),

    # Random cases
    ("123.456", "789.012"),
    ("456.789", "123.456"),
    ("0.000123", "0.000456"),
    ("0.000789", "0.000321"),
    ("98765.4321", "12345.6789"),
    ("31415.9265", "27182.8182"),
    ("1.23456789", "0.98765432"),
    ("0.99999999", "0.00000001"),
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

    # Near the square root of 2^63
    ("MaxFix64", "303700.0499"),
    ("MaxFix64", "303700.0499"),
    ("MaxFix64", "303700.04999760"),
    ("MaxFix64", "-303700.0499"),
    ("MaxFix64", "-303700.0499"),
    ("MaxFix64", "-303700.04999760"),

    ("MinFix64", "303700.0499"),
    ("MinFix64", "303700.0499"),
    ("MinFix64", "303700.04999760"),
    ("MinFix64", "-303700.0499"),
    ("MinFix64", "-303700.0499"),
    ("MinFix64", "-303700.04999760"),

    # Starting big
    ("MaxFix64", "1"),
    ("MaxFix64", "10"),
    ("MaxFix64", "100"),
    ("MaxFix64", "1000"),
    ("MaxFix64", "10000"),
    ("MaxFix64", "100000"),
    ("MaxFix64", "1000000"),
    ("MaxFix64", "10000000"),
    ("MaxFix64", "100000000"),
    ("MaxFix64", "1000000000"),
    ("MaxFix64", "10000000000"),

    ("MaxFix64", "-1"),
    ("MaxFix64", "-10"),
    ("MaxFix64", "-100"),
    ("MaxFix64", "-1000"),
    ("MaxFix64", "-10000"),
    ("MaxFix64", "-100000"),
    ("MaxFix64", "-1000000"),
    ("MaxFix64", "-10000000"),
    ("MaxFix64", "-100000000"),
    ("MaxFix64", "-1000000000"),
    ("MaxFix64", "-10000000000"),

    ("MinFix64", "1"),
    ("MinFix64", "10"),
    ("MinFix64", "100"),
    ("MinFix64", "1000"),
    ("MinFix64", "10000"),
    ("MinFix64", "100000"),
    ("MinFix64", "1000000"),
    ("MinFix64", "10000000"),
    ("MinFix64", "100000000"),
    ("MinFix64", "1000000000"),
    ("MinFix64", "10000000000"),

    ("MinFix64 + 1e-8", "-1"),
    ("MinFix64", "-10"),
    ("MinFix64", "-100"),
    ("MinFix64", "-1000"),
    ("MinFix64", "-10000"),
    ("MinFix64", "-100000"),
    ("MinFix64", "-1000000"),
    ("MinFix64", "-10000000"),
    ("MinFix64", "-100000000"),
    ("MinFix64", "-1000000000"),
    ("MinFix64", "-10000000000"),

    ("MaxFix64", "1"),
    ("HalfMaxFix64", "0.5"),
    ("MaxFix64", "-1"),
    ("HalfMaxFix64", "-0.5"),
    ("MinFix64", "1"),
    ("HalfMinFix64 + 1e-8", "0.5"),
    ("MinFix64 + 1e-8", "-1"),
    ("HalfMinFix64 + 1e-8", "-0.5"),

    ("MaxFix64", "1.0"),
    ("MaxFix64", "2.0"),
    ("MaxFix64", "MaxFix64"),
    ("MinFix64", "1.0"),
    ("MinFix64", "2.0"),
    ("MinFix64", "MinFix64"),

    # Things that divide to the smallest Fix64 (in magnitude)
    ("0.00000001", "1"),
    ("0.0000001", "10"),
    ("0.000001", "100"),
    ("0.00001", "1000"),
    ("0.0001", "10000"),
    ("0.001", "100000"),
    ("0.01", "1000000"),
    ("0.1", "10000000"),
    ("1.0", "100000000"),
    ("0.00000005", "5"),
    ("0.00000002", "2"),

    ("0.00000001", "-1"),
    ("0.0000001", "-10"),
    ("0.000001", "-100"),
    ("0.00001", "-1000"),
    ("0.0001", "-10000"),
    ("0.001", "-100000"),
    ("0.01", "-1000000"),
    ("0.1", "-10000000"),
    ("1.0", "-100000000"),
    ("0.00000005", "-5"),
    ("0.00000002", "-2"),

    ("-0.00000001", "1"),
    ("-0.0000001", "10"),
    ("-0.000001", "100"),
    ("-0.00001", "1000"),
    ("-0.0001", "10000"),
    ("-0.001", "100000"),
    ("-0.01", "1000000"),
    ("-0.1", "10000000"),
    ("-1.0", "100000000"),
    ("-0.00000005", "5"),
    ("-0.00000002", "2"),

    ("-0.00000001", "-1"),
    ("-0.0000001", "-10"),
    ("-0.000001", "-100"),
    ("-0.00001", "-1000"),
    ("-0.0001", "-10000"),
    ("-0.001", "-100000"),
    ("-0.01", "-1000000"),
    ("-0.1", "-10000000"),
    ("-1.0", "-100000000"),
    ("-0.00000005", "-5"),
    ("-0.00000002", "-2"),

    # Same as above, but with a SLIGHTLY smaller divisor
    ("0.00000001", "0.99999999"),
    ("0.0000001", "9.99999999"),
    ("0.000001", "99.99999999"),
    ("0.00001", "999.99999999"),
    ("0.0001", "9999.99999999"),
    ("0.001", "99999.99999999"),
    ("0.01", "999999.99999999"),
    ("0.1", "9999999.99999999"),
    ("1.0", "99999999.99999999"),

    ("0.00000001", "-0.99999999"),
    ("0.0000001", "-9.99999999"),
    ("0.000001", "-99.99999999"),
    ("0.00001", "-999.99999999"),
    ("0.0001", "-9999.99999999"),
    ("0.001", "-99999.99999999"),
    ("0.01", "-999999.99999999"),
    ("0.1", "-9999999.99999999"),
    ("1.0", "-99999999.99999999"),

    ("-0.00000001", "0.99999999"),
    ("-0.0000001", "9.99999999"),
    ("-0.000001", "99.99999999"),
    ("-0.00001", "999.99999999"),
    ("-0.0001", "9999.99999999"),
    ("-0.001", "99999.99999999"),
    ("-0.01", "999999.99999999"),
    ("-0.1", "9999999.99999999"),
    ("-1.0", "99999999.99999999"),
    
    ("-0.00000001", "-0.99999999"),
    ("-0.0000001", "-9.99999999"),
    ("-0.000001", "-99.99999999"),
    ("-0.00001", "-999.99999999"),
    ("-0.0001", "-9999.99999999"),
    ("-0.001", "-99999.99999999"),
    ("-0.01", "-999999.99999999"),
    ("-0.1", "-9999999.99999999"),
    ("-1.0", "-99999999.99999999"),

    # Should divide to the largest Fix64
    ("92233720368.54775807", "1.0"),
    ("9223372036.85477580", "0.1"),
    ("922337203.68547758", "0.01"),
    ("92233720.36854775", "0.001"),
    ("9223372.03685477", "0.0001"),
    ("922337.20368547", "0.00001"),
    ("92233.72036854", "0.000001"),
    ("9223.37203685", "0.0000001"),
    ("922.33720368", "0.00000001"),

    ("92233720368.54775807", "-1.0"),
    ("9223372036.85477580", "-0.1"),
    ("922337203.68547758", "-0.01"),
    ("92233720.36854775", "-0.001"),
    ("9223372.03685477", "-0.0001"),
    ("922337.20368547", "-0.00001"),
    ("92233.72036854", "-0.000001"),
    ("9223.37203685", "-0.0000001"),
    ("922.33720368", "-0.00000001"),

    ("-92233720368.54775808", "1.0"),
    ("-9223372036.85477580", "0.1"),
    ("-922337203.68547758", "0.01"),
    ("-92233720.36854775", "0.001"),
    ("-9223372.03685477", "0.0001"),
    ("-922337.20368547", "0.00001"),
    ("-92233.72036854", "0.000001"),
    ("-9223.37203685", "0.0000001"),
    ("-922.33720368", "0.00000001"),

    ("-92233720368.54775807", "-1.0"),
    ("-9223372036.85477580", "-0.1"),
    ("-922337203.68547758", "-0.01"),
    ("-92233720.36854775", "-0.001"),
    ("-9223372.03685477", "-0.0001"),
    ("-922337.20368547", "-0.00001"),
    ("-92233.72036854", "-0.000001"),
    ("-9223.37203685", "-0.0000001"),
    ("-922.33720368", "-0.00000001"),

]

# Overflow and DivByZero test data for Fix64 division
DivFix64OverflowTests = [
    ("MinFix64", "-0.00000001"),
    ("MinFix64", "-0.0000001"),
    ("MinFix64", "-0.000001"),
    ("MinFix64", "-0.0001"),
    ("MinFix64", "-0.001"),
    ("MinFix64", "-0.01"),
    ("MinFix64", "-0.1"),
    ("MinFix64", "-0.99999999"),
    ("HalfMinFix64", "-0.00000001"),
    ("HalfMinFix64", "-0.0000001"),
    ("HalfMinFix64", "-0.000001"),
    ("HalfMinFix64", "-0.0001"),
    ("HalfMinFix64", "-0.001"),
    ("HalfMinFix64", "-0.01"),
    ("HalfMinFix64", "-0.1"),
    ("MaxFix64", "0.00000001"),
    ("MaxFix64", "0.0000001"),
    ("MaxFix64", "0.000001"),
    ("MaxFix64", "0.0001"),
    ("MaxFix64", "0.001"),
    ("MaxFix64", "0.01"),
    ("MaxFix64", "0.1"),
    ("HalfMaxFix64", "0.00000001"),
    ("HalfMaxFix64", "0.0000001"),
    ("HalfMaxFix64", "0.000001"),
    ("HalfMaxFix64", "0.0001"),
    ("HalfMaxFix64", "0.001"),
    ("HalfMaxFix64", "0.01"),
    ("HalfMaxFix64", "0.1"),
]

DivFix64NegOverflowTests = [
    ("MinFix64", "0.00000001"),
    ("MinFix64", "0.0000001"),
    ("MinFix64", "0.000001"),
    ("MinFix64", "0.0001"),
    ("MinFix64", "0.001"),
    ("MinFix64", "0.01"),
    ("MinFix64", "0.1"),
    ("MinFix64", "0.99999999"),
    ("HalfMinFix64", "0.00000001"),
    ("HalfMinFix64", "0.0000001"),
    ("HalfMinFix64", "0.000001"),
    ("HalfMinFix64", "0.0001"),
    ("HalfMinFix64", "0.001"),
    ("HalfMinFix64", "0.01"),
    ("HalfMinFix64", "0.1"),
    ("MaxFix64", "-0.00000001"),
    ("MaxFix64", "-0.0000001"),
    ("MaxFix64", "-0.000001"),
    ("MaxFix64", "-0.0001"),
    ("MaxFix64", "-0.001"),
    ("MaxFix64", "-0.01"),
    ("MaxFix64", "-0.1"),
    ("HalfMaxFix64", "-0.00000001"),
    ("HalfMaxFix64", "-0.0000001"),
    ("HalfMaxFix64", "-0.000001"),
    ("HalfMaxFix64", "-0.0001"),
    ("HalfMaxFix64", "-0.001"),
    ("HalfMaxFix64", "-0.01"),
    ("HalfMaxFix64", "-0.1"),
]


DivFix64UnderflowTests = [
    # Underflow cases for Fix64 division (results too small to represent)
    ("0.00000001", "1.0000001"),
    ("0.00000001", "1.1"),
    ("0.00000001", "2.0"),
    ("0.00000001", "10.0"),
    ("0.00000001", "100.0"),
    ("0.00000001", "1000.0"),
    ("0.00000001", "100000000.0"),
    ("1.0", "92233720368.54775807"),
    ("1.0", "90000000000.0"),
    ("1.0", "80000000000.0"),
    ("0.00000001", "92233720368.54775807"),
    ("0.00000001", "80000000000.0"),
]

DivFix64DivByZeroTests = [
    ("1.0", "0.0"),
    ("0.0", "0.0"),
    ("MaxFix64", "0.0"),
    ("MinFix64", "0.0"),
    ("HalfMaxFix64", "0.0"),
    ("HalfMinFix64", "0.0"),
]


def generate_div_ufix64_tests():
    lines = ["var DivUFix64Tests = []struct{ A, B, Expected uint64 }{"]
    for a_str, b_str in DivUFix64Tests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        try:
            c = a / b
        except Exception:
            c = Decimal('0')
        a_hex = go_hex64(to_ufix64(a))
        b_hex = go_hex64(to_ufix64(b))
        c_hex = go_hex64(to_ufix64(c))
        comment = f"// {a_str} / {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_ufix64_overflow_tests():
    lines = ["var DivUFix64OverflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivUFix64OverflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_ufix64(a))
        b_hex = go_hex64(to_ufix64(b))
        comment = f"// {a_str} / {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_ufix64_underflow_tests():
    lines = ["var DivUFix64UnderflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivUFix64UnderflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_ufix64(a))
        b_hex = go_hex64(to_ufix64(b))
        comment = f"// {a_str} / {b_str} = underflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_ufix64_divbyzero_tests():
    lines = ["var DivUFix64DivByZeroTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivUFix64DivByZeroTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_ufix64(a))
        b_hex = go_hex64(to_ufix64(b))
        comment = f"// {a_str} / {b_str} = div by zero"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_fix64_tests():
    lines = ["var DivFix64Tests = []struct{ A, B, Expected uint64 }{"]
    for a_str, b_str in DivFix64Tests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        try:
            c = a / b
        except Exception:
            c = Decimal('0')
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        c_hex = go_hex64(to_fix64(c))
        comment = f"// {a_str} / {b_str} = {c}"
        pad = " " * (60 - len(f"    {{{a_hex}, {b_hex}, {c_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}, {c_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_fix64_overflow_tests():
    lines = ["var DivFix64OverflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivFix64OverflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        comment = f"// {a_str} / {b_str} = overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_fix64_neg_overflow_tests():
    lines = ["var DivFix64NegOverflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivFix64NegOverflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        comment = f"// {a_str} / {b_str} = negative overflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_fix64_underflow_tests():
    lines = ["var DivFix64UnderflowTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivFix64UnderflowTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        comment = f"// {a_str} / {b_str} = underflow"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def generate_div_fix64_divbyzero_tests():
    lines = ["var DivFix64DivByZeroTests = []struct{ A, B uint64 }{"]
    for a_str, b_str in DivFix64DivByZeroTests:
        a = parseInput64(a_str)
        b = parseInput64(b_str)
        a_hex = go_hex64(to_fix64(a))
        b_hex = go_hex64(to_fix64(b))
        comment = f"// {a_str} / {b_str} = div by zero"
        pad = " " * (40 - len(f"    {{{a_hex}, {b_hex}}},"))
        lines.append(f"    {{{a_hex}, {b_hex}}},{pad}{comment}")
    lines.append("}")
    return lines

def main():
    go_lines = [
        "// Code generated by testgen/div.py; DO NOT EDIT.",
        "package fixedPoint",
        "",
    ]
    go_lines.extend(generate_div_ufix64_tests())
    go_lines.extend(generate_div_ufix64_overflow_tests())
    go_lines.extend(generate_div_ufix64_underflow_tests())
    go_lines.extend(generate_div_ufix64_divbyzero_tests())
    go_lines.extend(generate_div_fix64_tests())
    go_lines.extend(generate_div_fix64_overflow_tests())
    go_lines.extend(generate_div_fix64_neg_overflow_tests())
    go_lines.extend(generate_div_fix64_underflow_tests())
    go_lines.extend(generate_div_fix64_divbyzero_tests())
    print('\n'.join(go_lines))

if __name__ == "__main__":
    main()
