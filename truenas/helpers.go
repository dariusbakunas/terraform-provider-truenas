package truenas

func flattenInt64List(list []int64) []interface{} {
	result := make([]interface{}, 0, len(list))
	for _, num := range list {
		result = append(result, num)
	}
	return result
}