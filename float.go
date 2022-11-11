package tools

import (
	"math"
)

var float32info = floatInfo{23, 8, -127}
var float64info = floatInfo{52, 11, -1023}
var optimize = true // set to false to force slow-path conversions for testing

func genericFtoa(dst *[]byte, val float64, fmt byte, prec, bitSize int) {
	var bits uint64
	var flt *floatInfo
	switch bitSize {
	case 32:
		bits = uint64(math.Float32bits(float32(val)))
		flt = &float32info
	case 64:
		bits = math.Float64bits(val)
		flt = &float64info
	default:
		panic("strconv: illegal AppendFloat/FormatFloat bitSize")
	}

	neg := bits>>(flt.expbits+flt.mantbits) != 0
	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1)
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 1<<flt.expbits - 1:
		// Inf, NaN
		switch {
		case mant != 0:
			*dst = append(*dst, []byte("NaN")...)
		case neg:
			*dst = append(*dst, []byte("-Inf")...)
		default:
			*dst = append(*dst, []byte("+Inf")...)
		}
		return
	case 0:
		// denormalized
		exp++

	default:
		// add implicit top bit
		mant |= uint64(1) << flt.mantbits
	}
	exp += flt.bias

	// Pick off easy binary, hex formats.
	if fmt == 'b' {
		fmtB(dst, neg, mant, exp, flt)
		return
	}
	if fmt == 'x' || fmt == 'X' {
		fmtX(dst, prec, fmt, neg, mant, exp, flt)
		return
	}

	if !optimize {
		bigFtoa(dst, prec, fmt, neg, mant, exp, flt)
		return
	}

	var digs decimalSlice
	ok := false
	// Negative precision means "only as much as needed to be exact."
	shortest := prec < 0
	if shortest {
		// Use Ryu algorithm.
		var buf [32]byte
		digs.d = buf[:]
		ryuFtoaShortest(&digs, mant, exp-int(flt.mantbits), flt)
		ok = true
		// Precision for shortest representation mode.
		switch fmt {
		case 'e', 'E':
			prec = max(digs.nd-1, 0)
		case 'f':
			prec = max(digs.nd-digs.dp, 0)
		case 'g', 'G':
			prec = digs.nd
		}
	} else if fmt != 'f' {
		// Fixed number of digits.
		digits := prec
		switch fmt {
		case 'e', 'E':
			digits++
		case 'g', 'G':
			if prec == 0 {
				prec = 1
			}
			digits = prec
		default:
			// Invalid mode.
			digits = 1
		}
		var buf [24]byte
		if bitSize == 32 && digits <= 9 {
			digs.d = buf[:]
			ryuFtoaFixed32(&digs, uint32(mant), exp-int(flt.mantbits), digits)
			ok = true
		} else if digits <= 18 {
			digs.d = buf[:]
			ryuFtoaFixed64(&digs, mant, exp-int(flt.mantbits), digits)
			ok = true
		}
	}
	if !ok {
		bigFtoa(dst, prec, fmt, neg, mant, exp, flt)
		return
	}
	formatDigits(dst, shortest, neg, digs, prec, fmt)
}

// bigFtoa uses multiprecision computations to format a float.
func bigFtoa(dst *[]byte, prec int, fmt byte, neg bool, mant uint64, exp int, flt *floatInfo) {
	d := new(decimal)
	d.Assign(mant)
	d.Shift(exp - int(flt.mantbits))
	var digs decimalSlice
	shortest := prec < 0
	if shortest {
		roundShortest(d, mant, exp, flt)
		digs = decimalSlice{d: d.d[:], nd: d.nd, dp: d.dp}
		// Precision for shortest representation mode.
		switch fmt {
		case 'e', 'E':
			prec = digs.nd - 1
		case 'f':
			prec = max(digs.nd-digs.dp, 0)
		case 'g', 'G':
			prec = digs.nd
		}
	} else {
		// Round appropriately.
		switch fmt {
		case 'e', 'E':
			d.Round(prec + 1)
		case 'f':
			d.Round(d.dp + prec)
		case 'g', 'G':
			if prec == 0 {
				prec = 1
			}
			d.Round(prec)
		}
		digs = decimalSlice{d: d.d[:], nd: d.nd, dp: d.dp}
	}
	formatDigits(dst, shortest, neg, digs, prec, fmt)
}

// TODO: move elsewhere?
type floatInfo struct {
	mantbits uint
	expbits  uint
	bias     int
}

func formatDigits(dst *[]byte, shortest bool, neg bool, digs decimalSlice, prec int, fmt byte) {
	switch fmt {
	case 'e', 'E':
		fmtE(dst, neg, digs, prec, fmt)
		return
	case 'f':
		fmtF(dst, neg, digs, prec)
		return
	case 'g', 'G':
		// trailing fractional zeros in 'e' form will be trimmed.
		eprec := prec
		if eprec > digs.nd && digs.nd >= digs.dp {
			eprec = digs.nd
		}
		// %e is used if the exponent from the conversion
		// is less than -4 or greater than or equal to the precision.
		// if precision was the shortest possible, use precision 6 for this decision.
		if shortest {
			eprec = 6
		}
		exp := digs.dp - 1
		if exp < -4 || exp >= eprec {
			if prec > digs.nd {
				prec = digs.nd
			}
			fmtE(dst, neg, digs, prec-1, fmt+'e'-'g')
			return
		}
		if prec > digs.dp {
			prec = digs.nd
		}
		fmtF(dst, neg, digs, max(prec-digs.dp, 0))
		return
	}

	// unknown format
	*dst = append(*dst, '%', fmt)
}

// roundShortest rounds d (= mant * 2^exp) to the shortest number of digits
// that will let the original floating point value be precisely reconstructed.
func roundShortest(d *decimal, mant uint64, exp int, flt *floatInfo) {
	// If mantissa is zero, the number is zero; stop now.
	if mant == 0 {
		d.nd = 0
		return
	}

	// Compute upper and lower such that any decimal number
	// between upper and lower (possibly inclusive)
	// will round to the original floating point number.

	// We may see at once that the number is already shortest.
	//
	// Suppose d is not denormal, so that 2^exp <= d < 10^dp.
	// The closest shorter number is at least 10^(dp-nd) away.
	// The lower/upper bounds computed below are at distance
	// at most 2^(exp-mantbits).
	//
	// So the number is already shortest if 10^(dp-nd) > 2^(exp-mantbits),
	// or equivalently log2(10)*(dp-nd) > exp-mantbits.
	// It is true if 332/100*(dp-nd) >= exp-mantbits (log2(10) > 3.32).
	minexp := flt.bias + 1 // minimum possible exponent
	if exp > minexp && 332*(d.dp-d.nd) >= 100*(exp-int(flt.mantbits)) {
		// The number is already shortest.
		return
	}

	// d = mant << (exp - mantbits)
	// Next highest floating point number is mant+1 << exp-mantbits.
	// Our upper bound is halfway between, mant*2+1 << exp-mantbits-1.
	upper := new(decimal)
	upper.Assign(mant*2 + 1)
	upper.Shift(exp - int(flt.mantbits) - 1)

	// d = mant << (exp - mantbits)
	// Next lowest floating point number is mant-1 << exp-mantbits,
	// unless mant-1 drops the significant bit and exp is not the minimum exp,
	// in which case the next lowest is mant*2-1 << exp-mantbits-1.
	// Either way, call it mantlo << explo-mantbits.
	// Our lower bound is halfway between, mantlo*2+1 << explo-mantbits-1.
	var mantlo uint64
	var explo int
	if mant > 1<<flt.mantbits || exp == minexp {
		mantlo = mant - 1
		explo = exp
	} else {
		mantlo = mant*2 - 1
		explo = exp - 1
	}
	lower := new(decimal)
	lower.Assign(mantlo*2 + 1)
	lower.Shift(explo - int(flt.mantbits) - 1)

	// The upper and lower bounds are possible outputs only if
	// the original mantissa is even, so that IEEE round-to-even
	// would round to the original mantissa and not the neighbors.
	inclusive := mant%2 == 0

	// As we walk the digits we want to know whether rounding up would fall
	// within the upper bound. This is tracked by upperdelta:
	//
	// If upperdelta == 0, the digits of d and upper are the same so far.
	//
	// If upperdelta == 1, we saw a difference of 1 between d and upper on a
	// previous digit and subsequently only 9s for d and 0s for upper.
	// (Thus rounding up may fall outside the bound, if it is exclusive.)
	//
	// If upperdelta == 2, then the difference is greater than 1
	// and we know that rounding up falls within the bound.
	var upperdelta uint8

	// Now we can figure out the minimum number of digits required.
	// Walk along until d has distinguished itself from upper and lower.
	for ui := 0; ; ui++ {
		// lower, d, and upper may have the decimal points at different
		// places. In this case upper is the longest, so we iterate from
		// ui==0 and start li and mi at (possibly) -1.
		mi := ui - upper.dp + d.dp
		if mi >= d.nd {
			break
		}
		li := ui - upper.dp + lower.dp
		l := byte('0') // lower digit
		if li >= 0 && li < lower.nd {
			l = lower.d[li]
		}
		m := byte('0') // middle digit
		if mi >= 0 {
			m = d.d[mi]
		}
		u := byte('0') // upper digit
		if ui < upper.nd {
			u = upper.d[ui]
		}

		// Okay to round down (truncate) if lower has a different digit
		// or if lower is inclusive and is exactly the result of rounding
		// down (i.e., and we have reached the final digit of lower).
		okdown := l != m || inclusive && li+1 == lower.nd

		switch {
		case upperdelta == 0 && m+1 < u:
			// Example:
			// m = 12345xxx
			// u = 12347xxx
			upperdelta = 2
		case upperdelta == 0 && m != u:
			// Example:
			// m = 12345xxx
			// u = 12346xxx
			upperdelta = 1
		case upperdelta == 1 && (m != '9' || u != '0'):
			// Example:
			// m = 1234598x
			// u = 1234600x
			upperdelta = 2
		}
		// Okay to round up if upper has a different digit and either upper
		// is inclusive or upper is bigger than the result of rounding up.
		okup := upperdelta > 0 && (inclusive || upperdelta > 1 || ui+1 < upper.nd)

		// If it's okay to do either, then round to the nearest one.
		// If it's okay to do only one, do it.
		switch {
		case okdown && okup:
			d.Round(mi + 1)
			return
		case okdown:
			d.RoundDown(mi + 1)
			return
		case okup:
			d.RoundUp(mi + 1)
			return
		}
	}
}

type decimalSlice struct {
	d      []byte
	nd, dp int
	neg    bool
}

// %e: -d.ddddde±dd
func fmtE(dst *[]byte, neg bool, d decimalSlice, prec int, fmt byte) {
	// sign
	if neg {
		*dst = append(*dst, '-')
	}

	// first digit
	ch := byte('0')
	if d.nd != 0 {
		ch = d.d[0]
	}
	*dst = append(*dst, ch)

	// .moredigits
	if prec > 0 {
		*dst = append(*dst, '.')
		i := 1
		m := min(d.nd, prec+1)
		if i < m {
			*dst = append(*dst, d.d[i:m]...)
			i = m
		}
		for ; i <= prec; i++ {
			*dst = append(*dst, '0')
		}
	}

	// e±
	*dst = append(*dst, fmt)
	exp := d.dp - 1
	if d.nd == 0 { // special case: 0 has exponent 0
		exp = 0
	}
	if exp < 0 {
		ch = '-'
		exp = -exp
	} else {
		ch = '+'
	}
	*dst = append(*dst, ch)

	// dd or ddd
	switch {
	case exp < 10:
		*dst = append(*dst, '0', byte(exp)+'0')
	case exp < 100:
		*dst = append(*dst, byte(exp/10)+'0', byte(exp%10)+'0')
	default:
		*dst = append(*dst, byte(exp/100)+'0', byte(exp/10)%10+'0', byte(exp%10)+'0')
	}

}

// %f: -ddddddd.ddddd
func fmtF(dst *[]byte, neg bool, d decimalSlice, prec int) {
	// sign
	if neg {
		*dst = append(*dst, '-')
	}

	// integer, padded with zeros as needed.
	if d.dp > 0 {
		m := min(d.nd, d.dp)
		*dst = append(*dst, d.d[:m]...)
		for ; m < d.dp; m++ {
			*dst = append(*dst, '0')
		}
	} else {
		*dst = append(*dst, '0')
	}

	// fraction
	if prec > 0 {
		*dst = append(*dst, '.')
		for i := 0; i < prec; i++ {
			ch := byte('0')
			if j := d.dp + i; 0 <= j && j < d.nd {
				ch = d.d[j]
			}
			*dst = append(*dst, ch)
		}
	}

}

// %b: -ddddddddp±ddd
func fmtB(dst *[]byte, neg bool, mant uint64, exp int, flt *floatInfo) {
	// sign
	if neg {
		*dst = append(*dst, '-')
	}

	// mantissa
	formatBits(dst, mant, 10, false)

	// p
	*dst = append(*dst, 'p')

	// ±exponent
	exp -= int(flt.mantbits)
	if exp >= 0 {
		*dst = append(*dst, '+')
	}
	formatBits(dst, uint64(exp), 10, exp < 0)
}

// %x: -0x1.yyyyyyyyp±ddd or -0x0p+0. (y is hex digit, d is decimal digit)
func fmtX(dst *[]byte, prec int, fmt byte, neg bool, mant uint64, exp int, flt *floatInfo) {
	if mant == 0 {
		exp = 0
	}

	// Shift digits so leading 1 (if any) is at bit 1<<60.
	mant <<= 60 - flt.mantbits
	for mant != 0 && mant&(1<<60) == 0 {
		mant <<= 1
		exp--
	}

	// Round if requested.
	if prec >= 0 && prec < 15 {
		shift := uint(prec * 4)
		extra := (mant << shift) & (1<<60 - 1)
		mant >>= 60 - shift
		if extra|(mant&1) > 1<<59 {
			mant++
		}
		mant <<= 60 - shift
		if mant&(1<<61) != 0 {
			// Wrapped around.
			mant >>= 1
			exp++
		}
	}

	hex := lowerhex
	if fmt == 'X' {
		hex = upperhex
	}

	// sign, 0x, leading digit
	if neg {
		*dst = append(*dst, '-')
	}
	*dst = append(*dst, '0', fmt, '0'+byte((mant>>60)&1))

	// .fraction
	mant <<= 4 // remove leading 0 or 1
	if prec < 0 && mant != 0 {
		*dst = append(*dst, '.')
		for mant != 0 {
			*dst = append(*dst, hex[(mant>>60)&15])
			mant <<= 4
		}
	} else if prec > 0 {
		*dst = append(*dst, '.')
		for i := 0; i < prec; i++ {
			*dst = append(*dst, hex[(mant>>60)&15])
			mant <<= 4
		}
	}

	// p±
	ch := byte('P')
	if fmt == lower(fmt) {
		ch = 'p'
	}
	*dst = append(*dst, ch)
	if exp < 0 {
		ch = '-'
		exp = -exp
	} else {
		ch = '+'
	}
	*dst = append(*dst, ch)

	// dd or ddd or dddd
	switch {
	case exp < 100:
		*dst = append(*dst, byte(exp/10)+'0', byte(exp%10)+'0')
	case exp < 1000:
		*dst = append(*dst, byte(exp/100)+'0', byte((exp/10)%10)+'0', byte(exp%10)+'0')
	default:
		*dst = append(*dst, byte(exp/1000)+'0', byte(exp/100)%10+'0', byte((exp/10)%10)+'0', byte(exp%10)+'0')
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const (
	lowerhex = "0123456789abcdef"
	upperhex = "0123456789ABCDEF"
)

func lower(c byte) byte {
	return c | ('x' - 'X')
}
