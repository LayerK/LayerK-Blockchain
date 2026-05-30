package util

func ArrayToSet[T comparable](arr []T) map[T]struct{} {
	ret := make(map[T]struct{}, len(arr))
	for _, elem := range arr {
		ret[elem] = struct{}{}
	}
	return ret
}
