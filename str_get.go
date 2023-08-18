package tools

import "bytes"

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

// Get 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 仅取出第一个匹配的结果
func (s *String) Get(model string) *String {
	if len(model) < 2 {
		return Make()
	}
	start := -1
	for i := 0; i < s.Len(); i++ {
		if s.buf[i] == model[0] {
			start = i
		} else if s.buf[i] == model[1] {
			return s.GetStrString(start+1, i)
		}
	}
	return Make()
}

// GetRune 中文支持 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 仅取出第一个匹配的结果
func (s *String) GetRune(model string) *String {
	if len(model) < 2 {
		return Make()
	}
	s.runes = bytes.Runes(s.buf)
	mRune := bytes.Runes([]byte(model))
	start := -1
	for i := 0; i < s.Len(); i++ {
		if s.runes[i] == mRune[0] {
			start = i
		} else if s.runes[i] == mRune[1] {
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
		} else if s.buf[i] == model[1] {
			res = append(res, s.GetStr(start+1, i))
			start = -1
		}
	}
	return res
}

// GetAllRune 此方法用于取出括号中的内容 支持输入字符model需要为2 下标0为左字符 1为右字符 取出所有匹配的结果
func (s *String) GetAllRune(model string) []string {
	if len(model) < 2 {
		return nil
	}
	var res []string
	s.runes = bytes.Runes(s.buf)
	mRune := bytes.Runes([]byte(model))
	start := -1
	for i := 0; i < s.Len(); i++ {
		if s.runes[i] == mRune[0] {
			start = i
		} else if s.runes[i] == mRune[1] {
			res = append(res, s.GetStrStringByRune(start+1, i).String())
			start = -1
		}
	}
	return res
}

// GetContent 此方法用于取出固定字符串中的内容,例如<a>mess</a>,注意 仅仅取出第一个匹配项，若要取出所有，请使用GetContentAll
// GetContent("<a>","</a>")
func (s *String) GetContent(label ...string) (content string) {
	temp := Make(s)
	for {
		if i, j := bytes.Index(temp.buf, []byte(label[0])), bytes.Index(temp.buf, []byte(label[1])); i != -1 && j != -1 {
			if i < j {
				content = s.GetStr(i+len(label[0]), j)
				return
			} else {
				temp.RemoveIndexStr(j + len(label[1]))
			}
		}
	}
}

// GetContentAll 此方法用于取出固定字符串中的内容,例如<a>mess</a>,注意 仅仅取出第一个匹配项，若要取出所有，请使用GetContentAll
// GetContentAll("<a>","</a>")
func (s *String) GetContentAll(label ...string) (content, other []string) {
	temp := Make(s)
	var tempOther string
	for {
		if i, j := bytes.Index(temp.buf, []byte(label[0])), bytes.Index(temp.buf, []byte(label[1])); i != -1 && j != -1 {
			if j < i {
				other[len(other)-1] = other[len(other)-1] + temp.GetStr(0, j+len(label[1]))
				temp.RemoveIndexStr(j + len(label[1]))
				tempOther = label[1]
			} else {
				content = append(content, temp.GetStr(i+len(label[0]), j))
				other = append(other, tempOther+temp.GetStr(0, i+len(label[0])))
				temp.RemoveIndexStr(j + len(label[1]))
				tempOther = label[1]
			}
		} else {
			other = append(other, tempOther+temp.string())
			break
		}
	}
	return
}
