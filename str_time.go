package tools

import "time"

const timeLayout = "2006-01-02 15:04:05"

// strTime 用来格式化时间
func (s *String) strTime(t time.Time, layout ...string) {
	const bufSize = 64
	var b []byte
	var buf [bufSize]byte
	b = buf[:0]
	s.appendAny(t.AppendFormat(b, timeLayout))
}

func (s *String) UpdateLayout(layout ...string) (t string, err error) {
	defer func() {
		if e := recover(); e != nil {
			t = s.string()
			err = e.(error)
		}
	}()
	switch len(layout) {
	case 0:
		ti := ReturnValue(time.Parse("01-02-06", s.string())).(time.Time)
		t = ti.Format("2006-01-02")
	case 1:
		ti := ReturnValue(time.Parse(layout[0], s.string())).(time.Time)
		t = ti.Format("2006-01-02")
	case 2:
		ti := ReturnValue(time.Parse(layout[0], s.string())).(time.Time)
		t = ti.Format(layout[1])
	default:
		panic("unknown Insert")
	}
	return
}
