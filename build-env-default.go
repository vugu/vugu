package vugu

type buildCacheKey interface{}

func makeBuildCacheKey(v interface{}) buildCacheKey {
	return v
}
