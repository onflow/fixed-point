from decimal import Decimal, getcontext
import mpmath

# Internal helper to safely convert Decimal -> mpmath.mpf with precision sync
def _decimal_to_mpf(x: Decimal) -> mpmath.mpf:
    mpmath.mp.dps = getcontext().prec
    return mpmath.mpf(str(x))

# Internal helper to safely convert mpmath.mpf -> Decimal
def _mpf_to_decimal(x: mpmath.mpf) -> Decimal:
    return Decimal(str(x))

def decSin(x: Decimal) -> Decimal:
    return _mpf_to_decimal(mpmath.sin(_decimal_to_mpf(x))).quantize(Decimal('1e-8'), rounding='ROUND_HALF_UP')

def decCos(x: Decimal) -> Decimal:
    return _mpf_to_decimal(mpmath.cos(_decimal_to_mpf(x)))

def decTan(x: Decimal) -> Decimal:
    return _mpf_to_decimal(mpmath.tan(_decimal_to_mpf(x)))

def decAsin(x: Decimal) -> Decimal:
    return _mpf_to_decimal(mpmath.asin(_decimal_to_mpf(x)))

def decAcos(x: Decimal) -> Decimal:
    return _mpf_to_decimal(mpmath.acos(_decimal_to_mpf(x)))

def decAtan(x: Decimal) -> Decimal:
    return _mpf_to_decimal(mpmath.atan(_decimal_to_mpf(x)))