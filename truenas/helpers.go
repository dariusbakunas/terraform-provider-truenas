package truenas

func flattenInt64List(list []int64) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, num := range list {
		result = append(result, num)
	}
	return result
}

func flattenInt32List(list []int32) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, num := range list {
		result = append(result, num)
	}
	return result
}

func getStringPtr(s string) *string {
	val := s
	return &val
}

func getInt64Ptr(i int64) *int64 {
	val := i
	return &val
}

func getInt32Ptr(i int32) *int32 {
	val := i
	return &val
}

func getBoolPtr(b bool) *bool {
	val := b
	return &val
}
