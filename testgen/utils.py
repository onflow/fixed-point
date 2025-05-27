# utils.py
# Shared decimal and conversion helpers for testgen scripts
from decimal import Decimal, getcontext, InvalidOperation
import re

getcontext().prec = 50

FIX64_SCALE = 10**8
MASK = 0xFFFFFFFFFFFFFFFF

def to_ufix64(val):
    n = int(Decimal(val) * FIX64_SCALE)
    if n < 0 or n > MASK:
        raise ValueError(f"Value {val} out of UFix64 range")
    return n & MASK

def to_fix64(val):
    n = int(Decimal(val) * FIX64_SCALE)
    return n & MASK

def go_hex(val):
    return f"0x{val:016X}"

def parseInput64(s):
    s = s.strip()
    # Hex bit pattern: treat as exact uint64 bits
    if s.startswith("0x") or s.startswith("0X"):
        bits = int(s, 16)
        return Decimal(bits) / FIX64_SCALE
    # Symbolic special values with optional modifier
    specials = {
        "MaxUFix64": Decimal("184467440737.09551615"),
        "MaxFix64":  Decimal("92233720368.54775807"),
        "MinFix64":  Decimal("-92233720368.54775808"),
        "HalfMaxUFix64": Decimal("92233720368.54775807"),
        "HalfMaxFix64":  Decimal("46116860184.27387903"),
        "HalfMinFix64":  Decimal("-46116860184.27387904"),
    }
    # Regex for symbolic value with optional modifier
    import re
    m = re.match(r"^([MH][a-zA-Z0-9]+)\s*([+-])?\s*(\S+)?$", s)
    if m:
        key = m.group(1)
        op = m.group(2)
        mod = m.group(3)
        # Accept exact match or case-insensitive
        base = None
        if key in specials:
            base = specials[key]
        else:
            for k in specials:
                if k.lower() == key.lower():
                    base = specials[k]
                    break
        if base is None:
            raise ValueError(f"Unknown symbolic value: {key}")
        if op and mod:
            try:
                mod_val = Decimal(mod)
            except Exception:
                raise ValueError(f"Invalid modifier for symbolic value: {mod}")
            if op == '+':
                return base + mod_val
            else:
                return base - mod_val
        return base
    # Default: parse as decimal
    try:
        return Decimal(s)
    except InvalidOperation:
        raise ValueError(f"Invalid input for parseInput64: {s}")
