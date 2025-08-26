# Copyright Flow Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# utils.py
# Shared decimal and conversion helpers for testgen scripts
from decimal import Decimal, getcontext, InvalidOperation
import re

getcontext().prec = 100

FIX64_SCALE = 10**8
FIX128_SCALE = 10**24
MASK64 = 0xFFFFFFFFFFFFFFFF
MASK128 = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF

def to_ufix64(val):
    n = int((Decimal(val) * FIX64_SCALE).quantize(1, rounding='ROUND_HALF_UP'))
    if n < 0 or n > MASK64:
        raise ValueError(f"Value {val} out of UFix64 range")
    return n & MASK64

def to_fix64(val):
    n = int((Decimal(val) * FIX64_SCALE).quantize(1, rounding='ROUND_HALF_UP'))
    if n < -0x8000000000000000 or n > 0x7FFFFFFFFFFFFFFF:
        raise ValueError(f"Value {val} out of Fix64 range")
    return n & MASK64

def to_ufix128(val):
    n = int(Decimal(val) * FIX128_SCALE)
    if n < 0 or n > MASK128:
        raise ValueError(f"Value {val} out of UFix128 range")
    return n & MASK128

def to_fix128(val):
    n = int(Decimal(val) * FIX128_SCALE)
    if n < -0x80000000000000000000000000000000 or n > 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:
        raise ValueError(f"Value {val} out of Fix128 range")
    return n & MASK128

def go_hex64(val):
    return f"0x{val:016x}"

def go_hex128(val):
    val = int(val) & 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
    return f"raw128{{0x{(val >> 64):016x}, 0x{(val & 0xFFFFFFFFFFFFFFFF):016x}}}"

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

def parseInput128(s):
    s = s.strip()
    # Hex bit pattern: treat as exact uint64 bits
    if s.startswith("0x") or s.startswith("0X"):
        bits = int(s, 16)
        return Decimal(bits) / FIX128_SCALE
    # Symbolic special values with optional modifier
    specials = {
        "MaxUFix128": Decimal(2**128 - 1) / FIX128_SCALE,
        "MaxFix128":  Decimal(2**127 - 1) / FIX128_SCALE,
        "MinFix128":  Decimal(-(2**127)) / FIX128_SCALE,
        "HalfMaxUFix128": Decimal(2**127 - 1) / FIX128_SCALE,
        "HalfMaxFix128":  Decimal(2**126 - 1) / FIX128_SCALE,
        "HalfMinFix128":  Decimal(-(2**126)) / FIX128_SCALE,
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
        raise ValueError(f"Invalid input for parseInput128: {s}")
