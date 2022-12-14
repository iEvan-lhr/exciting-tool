package tools

func Itoa(i interface{}) *String {
	s := String{}
	if s.appendAny(i) != -1 {
		return &s
	}
	s.Append('0')
	return &s
}
