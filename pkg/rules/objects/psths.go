package objects

func toPath(key any) string {
	switch x := key.(type) {
	case string:
		return x
	}
	return "unknown"
}
