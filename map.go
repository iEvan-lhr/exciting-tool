package tools

import (
	"sync"
)

func Ok(value interface{}, ok bool) interface{} {
	if ok {
		return value
	}
	return nil
}

type Spider struct {
	Values interface{}
	Key    []byte
	Next   [255]*Spider
	sync   sync.Mutex
	reader *String
	len    int
}

func (s *Spider) Add(key interface{}, value interface{}) {
	spider := s.getWriteSpider(s.reader.coverWrite(key).buf)
	//log.Println(spider)
	spider.add(value)
}

func (s *Spider) add(value interface{}) {
	s.Values = value
}

func (s *Spider) Get(key interface{}) interface{} {
	return s.getReadSpider(s.reader.coverWrite(key).buf).Values
}

func MakeSpider(key interface{}, value interface{}) *Spider {
	spider := &Spider{}
	spider.reader = &String{}
	spider.Key = spider.reader.Append(key).buf
	spider.getWriteSpider(spider.Key).add(value)
	return spider
}

func (s *Spider) getWriteSpider(key []byte) *Spider {
	temp := s
	for i := 0; i < len(key); i++ {
		if i == len(key)-1 {
			if temp.Next[key[i]] != nil {
				//位移至新地址
				trans := temp.Next[key[i]]
				temp.Next[key[i]] = &Spider{Key: key}
				s.getWriteSpider(trans.Key).add(trans.Values)
				s.addLen()
				return temp.Next[key[i]]
			} else {
				s.addLen()
				temp.Next[key[i]] = &Spider{Key: key}
				return temp.Next[key[i]]
			}
		}
		if temp.Next[key[i]] == nil {
			temp.Next[key[i]] = &Spider{Key: key}
			s.addLen()
			return temp.Next[key[i]]
		} else if checkBytes(temp.Next[key[i]].Key, key) {
			return temp.Next[key[i]]
		} else {
			temp = temp.Next[key[i]]
		}
	}
	return nil
}

func (s *Spider) getReadSpider(key []byte) *Spider {
	temp := s
	for i := 0; i < len(key); i++ {
		if temp.Next[key[i]] == nil {
			return &Spider{}
		} else if checkBytes(temp.Next[key[i]].Key, key) {
			return temp.Next[key[i]]
		} else {
			temp = temp.Next[key[i]]
		}
	}
	return nil
}

func (s *Spider) addLen() {
	s.sync.Lock()
	s.len++
	s.sync.Unlock()
}

func (s *Spider) Len() int {
	return s.len
}
