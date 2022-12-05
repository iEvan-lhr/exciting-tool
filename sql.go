package tools

const sli = '_'

func (s *String) Save(model any) *String {
	s.appendAny("insert into ")
	s.marshalStruct(model)
	return s
}

func Query(model any) string {
	s := String{}
	s.queryStruct(model)
	return s.string()
}

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