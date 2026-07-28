package tools

import (
	"bytes"
	"errors"
)

var ErrStringRange = errors.New("string index or range is out of bounds")

// ByteAt returns a byte without panicking on an invalid index.
func (s *String) ByteAt(index int) (byte, bool) {
	if s == nil || index < 0 || index >= len(s.buf) {
		return 0, false
	}
	return s.buf[index], true
}

// Slice returns a byte-indexed half-open range.
func (s *String) Slice(index, end int) (string, error) {
	if s == nil || index < 0 || end < index || end > len(s.buf) {
		return "", ErrStringRange
	}
	return string(s.buf[index:end]), nil
}

// RuneAt returns a Unicode code point without panicking.
func (s *String) RuneAt(index int) (rune, bool) {
	if s == nil {
		return 0, false
	}
	runes := bytes.Runes(s.buf)
	if index < 0 || index >= len(runes) {
		return 0, false
	}
	return runes[index], true
}

// SliceRunes returns a rune-indexed half-open range.
func (s *String) SliceRunes(index, end int) (string, error) {
	if s == nil {
		return "", ErrStringRange
	}
	runes := bytes.Runes(s.buf)
	if index < 0 || end < index || end > len(runes) {
		return "", ErrStringRange
	}
	return string(runes[index:end]), nil
}

// GetByte 获取字符串的单个字符值
// Deprecated: Use ByteAt for bounds-safe access.
func (s *String) GetByte(index int) byte {
	value, ok := s.ByteAt(index)
	if !ok {
		panic(ErrStringRange)
	}
	return value
}

// GetStr 获取字符串的某个片段 返回String
// Deprecated: Use Slice for bounds-safe access.
func (s *String) GetStr(index, end int) string {
	value, err := s.Slice(index, end)
	if err != nil {
		panic(err)
	}
	return value
}

// GetStrString 获取字符串的某个片段 返回String结构
func (s *String) GetStrString(index, end int) *String {
	return BytesString(s.buf[index:end])
}

// GetStrStringByRune 获取字符串的某个片段 返回String结构
func (s *String) GetStrStringByRune(index, end int) *String {
	runes := bytes.Runes(s.buf)
	return BytesString(runesToBytes(runes[index:end]))
}

// Get 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 仅取出第一个匹配的结果
func (s *String) Get(model string) *String {
	if len(model) < 2 {
		return Make()
	}
	start := -1
	for i := 0; i < s.Len(); i++ {
		if s.buf[i] == model[0] {
			start = i
		} else if s.buf[i] == model[1] && start >= 0 {
			return s.GetStrString(start+1, i)
		}
	}
	return Make()
}

// GetRune 中文支持 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 仅取出第一个匹配的结果
func (s *String) GetRune(model string) *String {
	mRune := bytes.Runes([]byte(model))
	if len(mRune) < 2 {
		return Make()
	}
	runes := bytes.Runes(s.buf)
	start := -1
	for i := 0; i < len(runes); i++ {
		if runes[i] == mRune[0] {
			start = i
		} else if runes[i] == mRune[1] && start >= 0 {
			return s.GetStrStringByRune(start+1, i)
		}
	}
	return Make()
}

// GetAll 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 取出所有匹配的结果
func (s *String) GetAll(model string) []string {
	if len(model) < 2 {
		return nil
	}
	var res []string
	start := -1
	for i := 0; i < s.Len(); i++ {
		if s.buf[i] == model[0] {
			start = i
		} else if s.buf[i] == model[1] && start >= 0 {
			res = append(res, s.GetStr(start+1, i))
			start = -1
		}
	}
	return res
}

// GetAllRune 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 取出所有匹配的结果
func (s *String) GetAllRune(model string) []string {
	mRune := bytes.Runes([]byte(model))
	if len(mRune) < 2 {
		return nil
	}
	var res []string
	runes := bytes.Runes(s.buf)
	start := -1
	for i := 0; i < len(runes); i++ {
		if runes[i] == mRune[0] {
			start = i
		} else if runes[i] == mRune[1] && start >= 0 {
			res = append(res, s.GetStrStringByRune(start+1, i).String())
			start = -1
		}
	}
	return res
}

// GetContent 此方法用于取出固定字符串中的内容,例如<a>mess</a>,注意 仅仅取出第一个匹配项，若要取出所有，请使用GetContentAll
// GetContent("<a>","</a>")
func (s *String) GetContent(label ...string) (content string) {
	if len(label) < 2 || label[0] == "" || label[1] == "" {
		return ""
	}
	start := bytes.Index(s.buf, []byte(label[0]))
	if start < 0 {
		return ""
	}
	contentStart := start + len(label[0])
	end := bytes.Index(s.buf[contentStart:], []byte(label[1]))
	if end < 0 {
		return ""
	}
	return s.GetStr(contentStart, contentStart+end)
}

// GetContentAll 此方法用于取出固定字符串中的内容,例如<a>mess</a>,注意 仅仅取出第一个匹配项，若要取出所有，请使用GetContentAll
// GetContentAll("<a>","</a>")
func (s *String) GetContentAll(label ...string) (content, other []string, steps map[int]struct {
	Model int
	Index int
}) {
	if len(label) < 2 || label[0] == "" || label[1] == "" {
		return nil, []string{s.String()}, map[int]struct {
			Model int
			Index int
		}{0: {Model: 1, Index: 0}}
	}
	steps = make(map[int]struct {
		Model int
		Index int
	})
	remaining := s.Bytes()
	pendingClose := ""
	ste := 0
	for {
		start := bytes.Index(remaining, []byte(label[0]))
		if start < 0 {
			other = append(other, pendingClose+string(remaining))
			steps[ste] = struct {
				Model int
				Index int
			}{Model: 1, Index: len(other) - 1}
			break
		}
		contentStart := start + len(label[0])
		endOffset := bytes.Index(remaining[contentStart:], []byte(label[1]))
		if endOffset < 0 {
			other = append(other, pendingClose+string(remaining))
			steps[ste] = struct {
				Model int
				Index int
			}{Model: 1, Index: len(other) - 1}
			break
		}

		other = append(other, pendingClose+string(remaining[:contentStart]))
		steps[ste] = struct {
			Model int
			Index int
		}{Model: 1, Index: len(other) - 1}
		ste++

		end := contentStart + endOffset
		content = append(content, string(remaining[contentStart:end]))
		steps[ste] = struct {
			Model int
			Index int
		}{Model: 0, Index: len(content) - 1}
		ste++

		remaining = remaining[end+len(label[1]):]
		pendingClose = label[1]
	}
	return
}
