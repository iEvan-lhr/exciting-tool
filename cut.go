package tools

func cutStructMessage(sm string) string {
	sms := Make(sm)
	split := sms.Split(".")
	sms.coverWrite(split[len(split)-1])
	return sms.Split(" ")[0]
}
