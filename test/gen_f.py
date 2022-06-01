#!/usr/bin/env python3

import sys

if len(sys.argv) != 3:
	print("Usage: ./gen_f.py F_NAME F_NUM_BYTES")
	sys.exit(1)

f = open(sys.argv[1], "wb+")

barr = bytearray([n for n in range(0, 256)])
len_barr = len(barr)
nrem = int(sys.argv[2])

while nrem > 0:
	n = min(nrem, len_barr)
	f.write(barr[0:n])
	nrem -= n


f.close()
