package error

import "fmt"

type UtilErr uint32

func (h UtilErr) Error() string {
	return GetErrMsg(uint32(h))
}
func (h UtilErr) ErrorCode() uint32 {
	return uint32(h)
}

var errMap = make(map[uint32]string)

func RegisterError(code uint32, msg string) {
	if _, dup := errMap[code]; dup {
		// panic(fmt.Sprintf("error: dumplicate code %d", code))
		fmt.Println(fmt.Sprintf("warn: dumplicate code %d msg:%s", code, msg))
	}
	errMap[code] = msg
}

func GetErrMsg(code uint32) string {
	if v, ok := errMap[code]; ok {
		return v
	}
	return "unknown error"
}
