package tools

import "math/bits"

const nSmalls = 100

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

const host32bit = ^uint(0)>>32 == 0

const fastSmalls = true // enable fast path for small integers

const digits = "0123456789abcdefghijklmnopqrstuvwxyz"

// small returns the string for an i with 0 <= i < nSmalls.
func small(i int) []byte {
	if i < 10 {
		return []byte(digits[i : i+1])
	}
	return []byte(smallsString[i*2 : i*2+2])
}

// formatInt returns the string representation of i in the given base,
// for 2 <= base <= 36. The result uses the lower-case letters 'a' to 'z'
// for digit values >= 10.
func formatInt(i int64, base int, byt *[]byte) int {
	if fastSmalls && 0 <= i && i < nSmalls && base == 10 {
		if i < 10 {
			*byt = append(*byt, digits[i : i+1][0])
			return 1
		} else {
			*byt = append(*byt, []byte(smallsString[i*2:i*2+2])...)
			return 2
		}

	}
	return formatBits(byt, uint64(i), base, i < 0)
}

// appendBytes is quick append int to String
func appendBytes(i int, byt *[]byte) int {
	return formatInt(int64(i), 10, byt)
}

func formatBits(dst *[]byte, u uint64, base int, neg bool) int {
	if base < 2 || base > len(digits) {
		panic("strconv: illegal AppendInt/FormatInt base")
	}
	// 2 <= base && base <= len(digits)

	var a [64 + 1]byte // +1 for sign of 64bit value in base 2
	i := len(a)

	if neg {
		u = -u
	}

	// convert bits
	// We use uint values where we can because those will
	// fit into a single register even on a 32bit machine.
	if base == 10 {
		// common case: use constants for / because
		// the compiler can optimize it into a multiply+shift

		if host32bit {
			// convert the lower digits using 32bit operations
			for u >= 1e9 {
				// Avoid using r = a%b in addition to q = a/b
				// since 64bit division and modulo operations
				// are calculated by runtime functions on 32bit machines.
				q := u / 1e9
				us := uint(u - q*1e9) // u % 1e9 fits into a uint
				for j := 4; j > 0; j-- {
					is := us % 100 * 2
					us /= 100
					i -= 2
					a[i+1] = smallsString[is+1]
					a[i+0] = smallsString[is+0]
				}

				// us < 10, since it contains the last digit
				// from the initial 9-digit us.
				i--
				a[i] = smallsString[us*2+1]

				u = q
			}
			// u < 1e9
		}

		// u guaranteed to fit into a uint
		us := uint(u)
		for us >= 100 {
			is := us % 100 * 2
			us /= 100
			i -= 2
			a[i+1] = smallsString[is+1]
			a[i+0] = smallsString[is+0]
		}

		// us < 100
		is := us * 2
		i--
		a[i] = smallsString[is+1]
		if us >= 10 {
			i--
			a[i] = smallsString[is]
		}

	} else if isPowerOfTwo(base) {
		// Use shifts and masks instead of / and %.
		// Base is a power of 2 and 2 <= base <= len(digits) where len(digits) is 36.
		// The largest power of 2 below or equal to 36 is 32, which is 1 << 5;
		// i.e., the largest possible shift count is 5. By &-ind that value with
		// the constant 7 we tell the compiler that the shift count is always
		// less than 8 which is smaller than any register width. This allows
		// the compiler to generate better code for the shift operation.
		shift := uint(bits.TrailingZeros(uint(base))) & 7
		b := uint64(base)
		m := uint(base) - 1 // == 1<<shift - 1
		for u >= b {
			i--
			a[i] = digits[uint(u)&m]
			u >>= shift
		}
		// u < base
		i--
		a[i] = digits[uint(u)]
	} else {
		// general case
		b := uint64(base)
		for u >= b {
			i--
			// Avoid using r = a%b in addition to q = a/b
			// since 64bit division and modulo operations
			// are calculated by runtime functions on 32bit machines.
			q := u / b
			a[i] = digits[uint(u-q*b)]
			u = q
		}
		// u < base
		i--
		a[i] = digits[uint(u)]
	}

	// add sign, if any
	if neg {
		i--
		a[i] = '-'
	}
	*dst = append(*dst, a[i:]...)
	return len(a[i:])
}

func isPowerOfTwo(x int) bool {
	return x&(x-1) == 0
}
