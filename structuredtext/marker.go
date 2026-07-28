package structuredtext

import (
	"errors"
	"strings"
)

const defaultMaxMarkerBytes = 64 << 10

var (
	ErrEmptyMarkerDelimiter = errors.New("marker delimiter cannot be empty")
	ErrMarkerTooLarge       = errors.New("marker exceeds configured size limit")
	ErrTokenizerClosed      = errors.New("marker tokenizer is closed")
)

type TokenKind uint8

const (
	TextToken TokenKind = iota
	MarkerToken
)

type Token struct {
	Kind  TokenKind
	Value string
}

// MarkerTokenizer recognizes markers that may span arbitrary input chunks.
// It is intended for one streaming consumer.
type MarkerTokenizer struct {
	start    string
	end      string
	maxBytes int
	buffer   string
	closed   bool
	failed   error
}

func NewMarkerTokenizer(start, end string) (*MarkerTokenizer, error) {
	return NewMarkerTokenizerLimit(start, end, defaultMaxMarkerBytes)
}

func NewMarkerTokenizerLimit(start, end string, maxMarkerBytes int) (*MarkerTokenizer, error) {
	if start == "" || end == "" {
		return nil, ErrEmptyMarkerDelimiter
	}
	if maxMarkerBytes <= 0 {
		maxMarkerBytes = defaultMaxMarkerBytes
	}
	return &MarkerTokenizer{
		start:    start,
		end:      end,
		maxBytes: maxMarkerBytes,
	}, nil
}

// Push accepts one chunk and returns all complete text and marker tokens.
func (t *MarkerTokenizer) Push(chunk string) ([]Token, error) {
	if t == nil {
		return nil, errors.New("marker tokenizer cannot be nil")
	}
	if t.closed {
		return nil, ErrTokenizerClosed
	}
	if t.failed != nil {
		return nil, t.failed
	}
	t.buffer += chunk
	tokens, err := t.scan(false)
	if err != nil {
		t.failed = err
	}
	return tokens, err
}

// Flush emits remaining incomplete input as text and closes the tokenizer.
func (t *MarkerTokenizer) Flush() ([]Token, error) {
	if t == nil {
		return nil, errors.New("marker tokenizer cannot be nil")
	}
	if t.closed {
		return nil, ErrTokenizerClosed
	}
	t.closed = true
	if t.failed != nil {
		return nil, t.failed
	}
	return t.scan(true)
}

func (t *MarkerTokenizer) scan(final bool) ([]Token, error) {
	var tokens []Token
	for t.buffer != "" {
		startIndex := strings.Index(t.buffer, t.start)
		if startIndex < 0 {
			if final {
				tokens = appendToken(tokens, TextToken, t.buffer)
				t.buffer = ""
				return tokens, nil
			}
			retain := partialDelimiterSuffix(t.buffer, t.start)
			emitEnd := len(t.buffer) - retain
			tokens = appendToken(tokens, TextToken, t.buffer[:emitEnd])
			t.buffer = t.buffer[emitEnd:]
			return tokens, nil
		}
		if startIndex > 0 {
			tokens = appendToken(tokens, TextToken, t.buffer[:startIndex])
			t.buffer = t.buffer[startIndex:]
		}

		contentStart := len(t.start)
		endOffset := strings.Index(t.buffer[contentStart:], t.end)
		if endOffset < 0 {
			if final {
				tokens = appendToken(tokens, TextToken, t.buffer)
				t.buffer = ""
				return tokens, nil
			}
			if len(t.buffer) > t.maxBytes {
				return tokens, ErrMarkerTooLarge
			}
			return tokens, nil
		}
		contentEnd := contentStart + endOffset
		markerEnd := contentEnd + len(t.end)
		if markerEnd > t.maxBytes {
			return tokens, ErrMarkerTooLarge
		}
		tokens = appendToken(tokens, MarkerToken, t.buffer[contentStart:contentEnd])
		t.buffer = t.buffer[markerEnd:]
	}
	return tokens, nil
}

func partialDelimiterSuffix(value, delimiter string) int {
	maximum := min(len(value), len(delimiter)-1)
	for length := maximum; length > 0; length-- {
		if strings.HasSuffix(value, delimiter[:length]) {
			return length
		}
	}
	return 0
}

func appendToken(tokens []Token, kind TokenKind, value string) []Token {
	if value == "" {
		return tokens
	}
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == kind {
		tokens[len(tokens)-1].Value += value
		return tokens
	}
	return append(tokens, Token{Kind: kind, Value: value})
}
