package main

func IsNotNumber(val interface{}) bool {
	switch val.(type) {
	case float64, float32, int, int32, int64, uint, uint32, uint64:
		return false // It's some type of number
	default:
		return true // It's not a number
	}
}
