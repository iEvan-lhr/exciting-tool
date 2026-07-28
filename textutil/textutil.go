// Package textutil provides small, bounds-safe text helpers.
package textutil

import (
	"errors"
	"strings"
)

var (
	ErrEmptyDelimiter = errors.New("delimiter cannot be empty")
	ErrInvalidRange   = errors.New("rune range is out of bounds")
)

// Between returns the first value between start and end.
func Between(value, start, end string) (string, bool) {
	if start == "" || end == "" {
		return "", false
	}
	startIndex := strings.Index(value, start)
	if startIndex < 0 {
		return "", false
	}
	contentStart := startIndex + len(start)
	endOffset := strings.Index(value[contentStart:], end)
	if endOffset < 0 {
		return "", false
	}
	return value[contentStart : contentStart+endOffset], true
}

// AllBetween returns non-overlapping values between start and end.
func AllBetween(value, start, end string) []string {
	if start == "" || end == "" {
		return nil
	}
	var result []string
	for cursor := 0; cursor < len(value); {
		startOffset := strings.Index(value[cursor:], start)
		if startOffset < 0 {
			break
		}
		contentStart := cursor + startOffset + len(start)
		endOffset := strings.Index(value[contentStart:], end)
		if endOffset < 0 {
			break
		}
		contentEnd := contentStart + endOffset
		result = append(result, value[contentStart:contentEnd])
		cursor = contentEnd + len(end)
	}
	return result
}

// SliceRunes returns the half-open rune range [start, end).
func SliceRunes(value string, start, end int) (string, error) {
	runes := []rune(value)
	if start < 0 || end < start || end > len(runes) {
		return "", ErrInvalidRange
	}
	return string(runes[start:end]), nil
}

// TruncateRunes limits value to at most limit runes and appends suffix when
// truncation occurs.
func TruncateRunes(value string, limit int, suffix string) (string, error) {
	if limit < 0 {
		return "", ErrInvalidRange
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value, nil
	}
	return string(runes[:limit]) + suffix, nil
}
