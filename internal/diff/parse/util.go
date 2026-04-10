package parse

// objPath builds a dot-notation child path for an object key.
// When parent is "." (the root), avoids the double-dot "..key" form.
func objPath(parent, key string) string {
	if parent == "." {
		return "." + key
	}
	return parent + "." + key
}
