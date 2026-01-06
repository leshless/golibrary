package xmaps

func Reverse[K comparable, V comparable](m map[K]V) map[V]K {
	res := make(map[V]K, len(m))
	for k, v := range m {
		res[v] = k
	}

	return res
}

func Keys[K comparable, V any](m map[K]V) []K {
	res := make([]K, 0, len(m))
	for k := range m {
		res = append(res, k)
	}

	return res
}

func ZipFunc[K comparable, V any, KV any](s []KV, zipFunc func(kv KV) (K, V)) map[K]V {
	res := make(map[K]V, len(s))
	for _, kv := range s {
		k, v := zipFunc(kv)
		res[k] = v
	}

	return res
}

func UnzipFunc[K comparable, V any, KV any](m map[K]V, unzipFunc func(K, V) KV) []KV {
	res := make([]KV, 0, len(m))
	for k, v := range m {
		kv := unzipFunc(k, v)
		res = append(res, kv)
	}

	return res
}
