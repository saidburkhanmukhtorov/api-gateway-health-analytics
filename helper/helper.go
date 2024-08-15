package helper

import "strconv"

func StringToInt(num string) int32 {
	res, err := strconv.Atoi(num)
	if err != nil {
		return 0
	}
	return int32(res)
}
