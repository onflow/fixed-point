from decimal import *
import mpmath

getcontext().prec = 100
mpmath.mp.dps = 100

twoPi = Decimal(str(mpmath.pi * 2))  # 2π to 100 decimal places

minFactor = Decimal(0x7fffffffffffffff) / Decimal(5**24) / twoPi
minFactor = int(minFactor.quantize(Decimal('1'), rounding=ROUND_UP))

maxFactor = Decimal(0xffffffffffffffff) / Decimal(5**24) / twoPi
maxFactor = int(maxFactor.quantize(Decimal('1'), rounding=ROUND_DOWN))

bestFactor = None
bestError = Decimal(1)

for i in range(minFactor, maxFactor + 1):
    currentMultiplier = Decimal(i) * 5**24
    scaled2Pi = twoPi * currentMultiplier
    truncated2Pi = int(scaled2Pi.quantize(Decimal('1'), rounding=ROUND_HALF_UP))

    estTwoPi = Decimal(truncated2Pi) / currentMultiplier
    err = abs(estTwoPi - twoPi)

    if err < bestError:
        bestError = err
        bestFactor = i

print(f"Best factor: {bestFactor}, Error: {bestError:0.3g}")


