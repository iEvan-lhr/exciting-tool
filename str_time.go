package tools

import (
	"fmt"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

// strTime 用来格式化时间
func (s *String) strTime(t time.Time, layout ...string) {
	const bufSize = 64
	var b []byte
	var buf [bufSize]byte
	b = buf[:0]
	format := timeLayout
	if len(layout) > 0 && layout[0] != "" {
		format = layout[0]
	}
	s.appendAny(t.AppendFormat(b, format))
}

func (s *String) UpdateLayout(layout ...string) (t string, err error) {
	sourceLayout := "01-02-06"
	targetLayout := "2006-01-02"
	switch len(layout) {
	case 0:
	case 1:
		sourceLayout = layout[0]
	case 2:
		sourceLayout = layout[0]
		targetLayout = layout[1]
	default:
		return s.String(), fmt.Errorf("UpdateLayout accepts at most two layouts")
	}
	parsed, err := time.Parse(sourceLayout, s.String())
	if err != nil {
		return s.String(), err
	}
	return parsed.Format(targetLayout), nil
}
