package api

// String returns the underlying string value of `Id`.
func (id *Id) String() string {
	if id == nil {
		return ""
	}
	return string(*id)
}
