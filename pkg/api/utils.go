package api

// NewClusterID creates a new id from a string
func NewClusterID(id string) ClusterId {
	return ClusterId{
		Id: Id(id),
	}
}

// NewTenantID creates a new id from a string
func NewTenantID(id string) TenantId {
	return TenantId{
		Id: Id(id),
	}
}
