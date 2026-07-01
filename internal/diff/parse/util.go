package parse

func objPath(parent, key string) string {
	if parent == "." {
		return "." + key
	}
	return parent + "." + key
}
