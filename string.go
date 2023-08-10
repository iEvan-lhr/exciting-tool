package tools

import (
	"bytes"
	"errors"
	"reflect"
	"time"
	"unicode/utf8"
	"unsafe"
)

const (
	TRUE  = "true"
	FALSE = "false"
)

type String struct {
	addr  *String
	runes []rune
	buf   []byte
}

// Strings 根据字符串来构建一个String
func Strings(str string) *String {
	s := String{}
	ReturnValue(s.writeString(str))
	s.runes = bytes.Runes(s.buf)
	return &s
}

// Make 根据指定类型来构建一个String
func Make(value ...any) *String {
	s := &String{}
	s.Append(value...)
	return s
}

func BytesString(b []byte) *String {
	s := String{}
	ReturnValue(s.Write(b))
	s.runes = bytes.Runes(s.buf)
	return &s
}

// ToString 字符串转型输出
func (s *String) String() string {
	return s.string()
}

// Runes 返回中文支持的字符
func (s *String) Runes() []rune {
	return s.runes
}

// Bytes 返回中文支持的字符
func (s *String) Bytes() []byte {
	return s.buf
}

// Check 检查是否相等
func (s *String) Check(str any) bool {
	switch str.(type) {
	case *String:
		if inTF(len(s.buf), str.(*String).Len()) {
			for i, v := range str.(*String).buf {
				if s.buf[i] != v {
					return false
				}
			}
			return true
		}
	case string:
		if inTF(len(s.buf), len(str.(string))) {
			for i, v := range []byte(str.(string)) {
				if s.buf[i] != v {
					return false
				}
			}
			return true
		}
	case []byte:
		if inTF(len(s.buf), len(str.([]byte))) {
			for i, v := range str.([]byte) {
				if s.buf[i] != v {
					return false
				}
			}
			return true
		}
	case []rune:
		if inTF(len(s.runes), len(str.([]rune))) {
			for i, v := range str.([]rune) {
				if s.runes[i] != v {
					return false
				}
			}
			return true
		}
	}
	return false

}

// appendAny  拼接字符串
func (s *String) appendAny(join any) int {
	switch join.(type) {
	case *String:
		return ReturnValue(s.Write(join.(*String).buf)).(int)
	case string:
		return ReturnValue(s.writeString(join.(string))).(int)
	case []byte:
		return ReturnValue(s.Write(join.([]byte))).(int)
	case byte:
		ReturnValue(s.WriteByte(join.(byte)))
		return 1
	case int:
		return appendInt(join.(int), &s.buf)
	case int8:
		return appendInt(int(join.(int8)), &s.buf)
	case int16:
		return appendInt(int(join.(int16)), &s.buf)
	case int32:
		return appendInt(int(join.(int32)), &s.buf)
	case int64:
		return appendInt(int(join.(int64)), &s.buf)
	case uint:
		return appendUint64(uint64(join.(uint)), &s.buf)
	case uint16:
		return appendUint64(uint64(join.(uint16)), &s.buf)
	case uint32:
		return appendUint64(uint64(join.(uint32)), &s.buf)
	case uint64:
		return appendUint64(join.(uint64), &s.buf)
	case float32:
		l1 := s.Len()
		genericFtoa(&s.buf, float64(join.(float32)), 'f', 2, 32)
		return s.Len() - l1
	case float64:
		l1 := s.Len()
		genericFtoa(&s.buf, join.(float64), 'f', 2, 32)
		return s.Len() - l1
	case bool:
		if join.(bool) {
			return ReturnValue(s.writeString(TRUE)).(int)
		} else {
			return ReturnValue(s.writeString(FALSE)).(int)
		}
	case time.Time:
		s.strTime(join.(time.Time))
	default:
		value := reflect.ValueOf(join)
		switch value.Kind() {
		case reflect.Struct, reflect.Pointer:
			if value.MethodByName("String").IsValid() {
				return ReturnValue(s.writeString(value.MethodByName("String").Call(nil)[0].String())).(int)
			} else {
				s.Marshal(join)
			}
		case reflect.Slice:
			ReturnValue(s.WriteByte('['))
			for i := 0; i < value.Len(); i++ {
				s.appendAny(value.Index(i).Interface())
				ReturnValue(s.WriteByte(','))
			}
			if value.Len() > 0 {
				s.RemoveLastStr(1)
			}
			ReturnValue(s.WriteByte(']'))
		}
	}
	return -1
}

func (s *String) coverWrite(key any) *String {
	s.buf = nil
	s.appendAny(key)
	return s
}

// Append  拼接字符串后返回String
func (s *String) Append(join ...any) *String {
	for i := range join {
		s.appendAny(join[i])
	}
	return s
}

// AppendLens  拼接字符串后返回String
func (s *String) AppendLens(join any) int {
	return s.appendAny(join)
}

// Index 返回数据中含有字串的下标 没有返回-1
func (s *String) Index(str any) int {
	switch str.(type) {
	case *String:
		return bytes.Index(s.buf, str.(*String).buf)
	case string:
		return bytes.Index(s.buf, []byte(str.(string)))
	case []byte:
		return bytes.Index(s.buf, str.([]byte))
	case rune:
		return s.indexByRune(str.(rune))
	}
	return -1
}

// Split 按照string片段来分割字符串 返回[]string
func (s *String) Split(str string) []string {
	var order []string
	for _, v := range bytes.Split(s.buf, []byte(str)) {
		order = append(order, string(v))
	}
	return order
}

// SplitString 按照*String来分割字符串 返回[]*String
func (s *String) SplitString(str String) []*String {
	byt := bytes.Split(s.buf, str.buf)
	var order []*String
	for i := range byt {
		order = append(order, &String{buf: byt[i]})
	}
	return order
}

// FirstUpper 首字母大写
func (s *String) FirstUpper() {
	if s.buf[0] > 90 {
		s.buf[0] = s.buf[0] - 32
	}
	s.runes = bytes.Runes(s.buf)
}

// FirstLower 首字母小写
func (s *String) FirstLower() {
	if s.buf[0] < 97 {
		s.buf[0] = s.buf[0] + 32
	}
	s.runes = bytes.Runes(s.buf)
}

// FirstUpperBackString 首字母大写后返回string
func (s *String) FirstUpperBackString() string {
	s.FirstUpper()
	return s.string()
}

// FirstLowerBackString 首字母小写后返回string
func (s *String) FirstLowerBackString() string {
	s.FirstLower()
	return s.string()
}

func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func (s *String) copyCheck() {
	if s.addr == nil {
		s.addr = (*String)(noescape(unsafe.Pointer(s)))
	} else if s.addr != s {
		panic("strings: illegal use of non-zero String copied by value")
	}
}

func (s *String) string() string {
	return *(*string)(unsafe.Pointer(&s.buf))
}

// Len 返回字符串长度
func (s *String) Len() int { return len(s.buf) }

// LenByRune 返回含有中文的字符串长度
func (s *String) LenByRune() int { return len(bytes.Runes(s.buf)) }

func (s *String) cap() int { return cap(s.buf) }

func (s *String) reset() {
	s.addr = nil
	s.buf = nil
}

func (s *String) grow(n int) {
	buf := make([]byte, len(s.buf), 2*cap(s.buf)+n)
	copy(buf, s.buf)
	s.buf = buf
}

// Grow  扩充大小
func (s *String) Grow(n int) {
	s.copyCheck()
	if n < 0 {
		panic("strings.String.Grow: negative count")
	}
	if cap(s.buf)-len(s.buf) < n {
		s.grow(n)
	}
}

// WriteByte 写入[]Byte的数据
func (s *String) Write(p []byte) (int, error) {
	s.copyCheck()
	s.buf = append(s.buf, p...)
	return len(p), nil
}

// WriteByte 写入Byte字符格式的数据
func (s *String) WriteByte(c byte) error {
	s.copyCheck()
	s.buf = append(s.buf, c)
	return nil
}

// WriteRune 写入Rune字符格式的数据
func (s *String) WriteRune(r rune) (int, error) {
	s.copyCheck()
	if r < utf8.RuneSelf {
		s.buf = append(s.buf, byte(r))
		return 1, nil
	}
	l := len(s.buf)
	if cap(s.buf)-l < utf8.UTFMax {
		s.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(s.buf[l:l+utf8.UTFMax], r)
	s.buf = s.buf[:l+n]
	return n, nil
}

func (s *String) writeString(str string) (int, error) {
	s.copyCheck()
	s.buf = append(s.buf, str...)
	return len(str), nil
}

// RemoveLastStr 从尾部移除固定长度的字符
func (s *String) RemoveLastStr(lens int) {
	if lens > s.Len() {
		LogError(errors.New("RemoveLens>stringLens Please Check"))
		return
	}
	s.buf = s.buf[:s.Len()-lens]
	s.runes = bytes.Runes(s.buf)
}

// ReplaceLastStr 从尾部移除固定长度的字符
func (s *String) ReplaceLastStr(lens int, str any) {
	s.buf = s.buf[:s.Len()-lens]
	s.appendAny(str)
	s.runes = bytes.Runes(s.buf)
}

// RemoveLastStrByRune 从尾部移除固定长度的字符 并且支持中文字符的移除
func (s *String) RemoveLastStrByRune(lens int) {
	runes := bytes.Runes(s.buf)
	if lens > len(runes) {
		LogError(errors.New("RemoveLens>stringLens Please Check"))
		return
	}
	s.buf = runesToBytes(runes[:len(runes)-lens])
}

// GetByte 获取字符串的单个字符值
func (s *String) GetByte(index int) byte {
	return s.buf[index]
}

// GetStr 获取字符串的某个片段 返回String
func (s *String) GetStr(index, end int) string {
	return string(s.buf[index:end])
}

// GetStrString 获取字符串的某个片段 返回String结构
func (s *String) GetStrString(index, end int) *String {
	return &String{buf: s.buf[index:end]}
}

// GetStrStringByRune 获取字符串的某个片段 返回String结构
func (s *String) GetStrStringByRune(index, end int) *String {
	return &String{buf: runesToBytes(s.runes[index:end])}
}

// RemoveIndexStr 移除头部固定长度的字符
func (s *String) RemoveIndexStr(lens int) {
	if lens > s.Len() {
		LogError(errors.New("RemoveLens>stringLens Please Check"))
		return
	}
	s.buf = s.buf[lens:]
	s.runes = bytes.Runes(s.buf)
}

// RemoveIndexRune 移除头部固定长度的字符（中文支持）
func (s *String) RemoveIndexRune(lens int) {
	if lens > len(s.runes) {
		LogError(errors.New("RemoveLens>stringLens Please Check"))
		return
	}
	s.runes = s.runes[lens:]
	s.buf = runesToBytes(s.runes)
}

// CheckIsNull 检查字符串是否为空 只包含' '与'\t'与'\n'都会被视为不合法的值
func (s *String) CheckIsNull() bool {
	for _, b := range s.buf {
		if b != 32 && b != 9 && b != 10 {
			return false
		}
	}
	return true
}

func (s *String) indexByRune(r rune) int {
	if s.runes == nil || len(s.runes) == 0 {
		s.runes = bytes.Runes(s.buf)
	}
	for i := range s.runes {
		if s.runes[i] == r {
			return i
		}
	}
	return -1
}
