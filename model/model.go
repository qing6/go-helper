package model

type NumSpan [2]int64 // start, end + 1

func (span NumSpan) Split(d int64) []int64 {
	firstStart := span[0] / d
	lastStart := (span[1] - 1) / d
	result := make([]int64, lastStart - firstStart + 1)

	for i := 0; i < len(result); i++ {
		result[i] = (firstStart + int64(i)) * d
	}
	return result
}


















