// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tools

import (
	"errors"
)

// ErrRange indicates that a value is out of range for the target type.
var ErrRange = errors.New("value out of range")

// ErrSyntax indicates that a value does not have the right syntax for the target type.
var ErrSyntax = errors.New("invalid syntax")

// A NumError records a failed conversion.
type NumError struct {
	Func *String // the failing function (ParseBool, ParseInt, ParseUint, ParseFloat, ParseComplex)
	Num  *String // the input
	Err  error   // the reason the conversion failed (e.g. ErrRange, ErrSyntax, etc.)
}

func (e *NumError) Error() string {
	Quote(e.Num)
	return "strconv." + e.Func.string() + ": " + "parsing " + e.Num.string() + ": " + e.Err.Error()
}

func (e *NumError) Unwrap() error { return e.Err }

func syntaxError(fn, str *String) *NumError {
	return &NumError{fn, str, ErrSyntax}
}

func rangeError(fn, str *String) *NumError {
	return &NumError{fn, str, ErrRange}
}

func baseError(fn, str *String, base int) *NumError {
	errorf := Strings("invalid base ")
	errorf.AppendAny(base)
	return &NumError{fn, str, errors.New(errorf.string())}
}

func bitSizeError(fn, str *String, bitSize int) *NumError {
	errorf := Strings("invalid bit size ")
	errorf.AppendAny(bitSize)
	return &NumError{fn, str, errors.New(errorf.string())}
}

const intSize = 32 << (^uint(0) >> 63)

// IntSize is the size in bits of an int or uint value.
const IntSize = intSize

const maxUint64 = 1<<64 - 1

// ParseUint is like ParseInt but for unsigned numbers.
//
// A sign prefix is not permitted.
func ParseUint(s []byte, base int, bitSize int) (uint64, error) {
	const fnParseUint = "ParseUint"

	if s == nil || len(s) == 0 {
		return 0, syntaxError(Strings(fnParseUint), BytesString(s))
	}

	base0 := base == 0

	s0 := s
	switch {
	case 2 <= base && base <= 36:
		// valid base; nothing to do

	case base == 0:
		// Look for octal, hex prefix.
		base = 10
		if s[0] == '0' {
			switch {
			case len(s) >= 3 && lower(s[1]) == 'b':
				base = 2
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'o':
				base = 8
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'x':
				base = 16
				s = s[2:]
			default:
				base = 8
				s = s[1:]
			}
		}

	default:
		return 0, baseError(Strings(fnParseUint), BytesString(s0), base)
	}

	if bitSize == 0 {
		bitSize = IntSize
	} else if bitSize < 0 || bitSize > 64 {
		return 0, bitSizeError(Strings(fnParseUint), BytesString(s0), bitSize)
	}

	// Cutoff is the smallest number such that cutoff*base > maxUint64.
	// Use compile-time constants for common cases.
	var cutoff uint64
	switch base {
	case 10:
		cutoff = maxUint64/10 + 1
	case 16:
		cutoff = maxUint64/16 + 1
	default:
		cutoff = maxUint64/uint64(base) + 1
	}

	maxVal := uint64(1)<<uint(bitSize) - 1

	underscores := false
	var n uint64
	for _, c := range s {
		var d byte
		switch {
		case c == '_' && base0:
			underscores = true
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= lower(c) && lower(c) <= 'z':
			d = lower(c) - 'a' + 10
		default:
			return 0, syntaxError(Strings(fnParseUint), BytesString(s0))
		}

		if d >= byte(base) {
			return 0, syntaxError(Strings(fnParseUint), BytesString(s0))
		}

		if n >= cutoff {
			// n*base overflows
			return maxVal, rangeError(Strings(fnParseUint), BytesString(s0))
		}
		n *= uint64(base)

		n1 := n + uint64(d)
		if n1 < n || n1 > maxVal {
			// n+d overflows
			return maxVal, rangeError(Strings(fnParseUint), BytesString(s0))
		}
		n = n1
	}

	if underscores && !underscoreOK(s0) {
		return 0, syntaxError(Strings(fnParseUint), BytesString(s0))
	}

	return n, nil
}

// ParseInt interprets a string s in the given base (0, 2 to 36) and
// bit size (0 to 64) and returns the corresponding value i.
//
// The string may begin with a leading sign: "+" or "-".
//
// If the base argument is 0, the true base is implied by the string's
// prefix following the sign (if present): 2 for "0b", 8 for "0" or "0o",
// 16 for "0x", and 10 otherwise. Also, for argument base 0 only,
// underscore characters are permitted as defined by the Go syntax for
// integer literals.
//
// The bitSize argument specifies the integer type
// that the result must fit into. Bit sizes 0, 8, 16, 32, and 64
// correspond to int, int8, int16, int32, and int64.
// If bitSize is below 0 or above 64, an error is returned.
//
// The errors that ParseInt returns have concrete type *NumError
// and include err.Num = s. If s is empty or contains invalid
// digits, err.Err = ErrSyntax and the returned value is 0;
// if the value corresponding to s cannot be represented by a
// signed integer of the given size, err.Err = ErrRange and the
// returned value is the maximum magnitude integer of the
// appropriate bitSize and sign.
func ParseInt(s []byte, base int, bitSize int) (i int64, err error) {
	const fnParseInt = "ParseInt"

	if s == nil || len(s) == 0 {
		return 0, syntaxError(Strings(fnParseInt), BytesString(s))
	}

	// Pick off leading sign.
	s0 := s
	neg := false
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		neg = true
		s = s[1:]
	}

	// Convert unsigned and check range.
	var un uint64
	un, err = ParseUint(s, base, bitSize)
	if err != nil && err.(*NumError).Err != ErrRange {
		err.(*NumError).Func = Strings(fnParseInt)
		err.(*NumError).Num = BytesString(s0)
		return 0, err
	}

	if bitSize == 0 {
		bitSize = IntSize
	}

	cutoff := uint64(1 << uint(bitSize-1))
	if !neg && un >= cutoff {
		return int64(cutoff - 1), rangeError(Strings(fnParseInt), BytesString(s0))
	}
	if neg && un > cutoff {
		return -int64(cutoff), rangeError(Strings(fnParseInt), BytesString(s0))
	}
	n := int64(un)
	if neg {
		n = -n
	}
	return n, nil
}

// Atoi 方法在实际使用中的效率近似接近strconv包中的方法，但如果是对同一个string对象进行多次转换 推荐使用strconv包中的方法
func (s *String) Atoi() (int, error) {
	const fnAtoi = "Atoi"
	sLen := s.Len()
	if intSize == 32 && (0 < sLen && sLen < 10) ||
		intSize == 64 && (0 < sLen && sLen < 19) {
		// Fast path for small integers that fit int type.
		s0 := s
		if s.buf[0] == '-' || s.buf[0] == '+' {
			s.RemoveIndexStr(1)
			if s.Len() < 1 {
				return 0, &NumError{Strings(fnAtoi), s0, ErrSyntax}
			}
		}

		n := 0
		for _, ch := range s.buf {
			ch -= '0'
			if ch > 9 {
				return 0, &NumError{Strings(fnAtoi), s0, ErrSyntax}
			}
			n = n*10 + int(ch)
		}
		if s0.buf[0] == '-' {
			n = -n
		}
		return n, nil
	}

	// Slow path for invalid, big, or underscored integers.
	i64, err := ParseInt(s.buf, 10, 0)
	if nerr, ok := err.(*NumError); ok {
		nerr.Func = Strings(fnAtoi)
	}
	return int(i64), err
}

// underscoreOK reports whether the underscores in s are allowed.
// Checking them in this one function lets all the parsers skip over them simply.
// Underscore must appear only between digits or between a base prefix and a digit.
func underscoreOK(s []byte) bool {
	// saw tracks the last character (class) we saw:
	// ^ for beginning of number,
	// 0 for a digit or base prefix,
	// _ for an underscore,
	// ! for none of the above.
	saw := '^'
	i := 0

	// Optional sign.
	if len(s) >= 1 && (s[0] == '-' || s[0] == '+') {
		s = s[1:]
	}

	// Optional base prefix.
	hex := false
	if len(s) >= 2 && s[0] == '0' && (lower(s[1]) == 'b' || lower(s[1]) == 'o' || lower(s[1]) == 'x') {
		i = 2
		saw = '0' // base prefix counts as a digit for "underscore as digit separator"
		hex = lower(s[1]) == 'x'
	}

	// Number proper.
	for ; i < len(s); i++ {
		// Digits are always okay.
		if '0' <= s[i] && s[i] <= '9' || hex && 'a' <= lower(s[i]) && lower(s[i]) <= 'f' {
			saw = '0'
			continue
		}
		// Underscore must follow digit.
		if s[i] == '_' {
			if saw != '0' {
				return false
			}
			saw = '_'
			continue
		}
		// Underscore must also be followed by digit.
		if saw == '_' {
			return false
		}
		// Saw non-digit, non-underscore.
		saw = '!'
	}
	return saw != '_'
}
