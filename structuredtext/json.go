// Package structuredtext extracts structured values and markers from text
// produced by humans or language models.
package structuredtext

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrJSONNotFound      = errors.New("no JSON object or array found")
	ErrInvalidRepairJSON = errors.New("JSON repair returned invalid JSON")
)

// JSONRepairFunc repairs one JSON candidate.
type JSONRepairFunc func(string) (string, error)

// StripFence removes one wrapping Markdown code fence and its optional
// language name. Other surrounding text is left unchanged.
func StripFence(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 6 {
		return trimmed
	}
	var fence string
	switch {
	case strings.HasPrefix(trimmed, "```"):
		fence = "```"
	case strings.HasPrefix(trimmed, "~~~"):
		fence = "~~~"
	default:
		return trimmed
	}
	lineEnd := strings.IndexByte(trimmed, '\n')
	if lineEnd < 0 {
		return trimmed
	}
	closingStart := strings.LastIndex(trimmed, "\n"+fence)
	if closingStart < lineEnd || strings.TrimSpace(trimmed[closingStart+1:]) != fence {
		return trimmed
	}
	if closingStart == lineEnd {
		return ""
	}
	return strings.TrimSpace(trimmed[lineEnd+1 : closingStart])
}

// JSONCandidates returns balanced object and array candidates. Brackets
// inside JSON strings are ignored. Candidates may still require JSON repair.
func JSONCandidates(value string) []string {
	value = StripFence(value)
	var candidates []string
	for cursor := 0; cursor < len(value); {
		offset := strings.IndexAny(value[cursor:], "{[")
		if offset < 0 {
			break
		}
		start := cursor + offset
		candidate, ok := balancedJSON(value[start:])
		if !ok {
			cursor = start + 1
			continue
		}
		candidates = append(candidates, candidate)
		cursor = start + len(candidate)
	}
	return candidates
}

// ExtractJSON returns the first valid JSON object or array.
func ExtractJSON(value string) (string, bool) {
	stripped := StripFence(value)
	if startsWithContainer(stripped) && json.Valid([]byte(stripped)) {
		return stripped, true
	}
	for _, candidate := range JSONCandidates(stripped) {
		if json.Valid([]byte(candidate)) {
			return candidate, true
		}
	}
	return "", false
}

// ExtractJSONWithRepair returns valid JSON, invoking repair only when direct
// extraction fails. The repaired value is always validated before returning.
func ExtractJSONWithRepair(value string, repair JSONRepairFunc) (string, error) {
	if candidate, ok := ExtractJSON(value); ok {
		return candidate, nil
	}
	if repair == nil {
		return "", ErrJSONNotFound
	}

	stripped := StripFence(value)
	inputs := append([]string(nil), JSONCandidates(stripped)...)
	if strings.TrimSpace(stripped) != "" && !containsString(inputs, stripped) {
		inputs = append(inputs, stripped)
	}
	if len(inputs) == 0 {
		return "", ErrJSONNotFound
	}

	var lastErr error
	for _, input := range inputs {
		repaired, err := repair(input)
		if err != nil {
			lastErr = err
			continue
		}
		repaired = StripFence(repaired)
		if startsWithContainer(repaired) && json.Valid([]byte(repaired)) {
			return repaired, nil
		}
		lastErr = ErrInvalidRepairJSON
	}
	if lastErr == nil {
		return "", ErrJSONNotFound
	}
	return "", errors.Join(ErrJSONNotFound, fmt.Errorf("repair failed: %w", lastErr))
}

func balancedJSON(value string) (string, bool) {
	if value == "" || (value[0] != '{' && value[0] != '[') {
		return "", false
	}
	stack := make([]byte, 0, 8)
	inString := false
	escaped := false
	for index := 0; index < len(value); index++ {
		current := value[index]
		if inString {
			switch {
			case escaped:
				escaped = false
			case current == '\\':
				escaped = true
			case current == '"':
				inString = false
			}
			continue
		}
		switch current {
		case '"':
			inString = true
		case '{':
			stack = append(stack, '}')
		case '[':
			stack = append(stack, ']')
		case '}', ']':
			if len(stack) == 0 || stack[len(stack)-1] != current {
				return "", false
			}
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return value[:index+1], true
			}
		}
	}
	return "", false
}

func startsWithContainer(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[")
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
