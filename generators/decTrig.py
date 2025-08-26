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