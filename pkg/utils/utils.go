package utils

func CalcTotalPage(total, pageSize int64) int64 {
	if pageSize <= 0 {
		return 0
	}
	if total == 0 {
		return 0
	}
	if total%pageSize == 0 {
		return total / pageSize
	}
	return total/pageSize + 1
}
