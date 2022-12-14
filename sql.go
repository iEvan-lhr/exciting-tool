package tools

const sli = '_'

// Save 根据参数组装insert语句
func Save(model interface{}) (result []*String) {
	result = marshalStruct(model)
	return
}

// SaveTable 根据参数组装insert语句
func SaveTable(model interface{}, tableName string) (result []*String) {
	result = marshalStructTable(model, tableName)
	return
}

// Query 根据参数组装sql查询语句
func Query(model interface{}) string {
	s := String{}
	s.queryStruct(model)
	return s.string()
}

func Update(model interface{}) {

}

func Check(model interface{}) string {
	s := String{}
	s.checkStruct(model)
	return s.string()
}
