package vugu

type buildCacheKey any

func makeBuildCacheKey(v any) buildCacheKey {
	return v
}
