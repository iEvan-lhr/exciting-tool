package tools

import "bytes"

// humpName 格式化驼峰命名
func humpName(buf string) (ans []byte) {
	if len(buf) > 0 {
		for i := range buf {
			if buf[i] < 97 {
				if i == 0 {
					ans = append(ans, buf[0]+32)
				} else {
					ans = append(ans, sli)
					ans = append(ans, buf[i]+32)
				}
			} else {
				ans = append(ans, buf[i])
			}
		}
	}
	return
}

func righteousCharacter(s *String) *String {
	var runes []rune
	for _, v := range bytes.Runes(s.buf) {
		if v == '\'' || v == '`' {
			runes = append(runes, append([]rune(`\`), v)...)
		} else {
			runes = append(runes, v)
		}
	}
	s.runes = runes
	s.buf = runesToBytes(s.Runes())
	return s
}
