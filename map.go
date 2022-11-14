package tools

func Ok(value any, ok bool) any {
	if ok {
		return value
	}
	return nil
}
