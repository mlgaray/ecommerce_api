package models

// Image represents a unified image structure used across entities.
// Supports Category, Product, User, Shop via Exclusive Belongs-To pattern.
// Note: StorageRef is serialized for DB operations but NOT selected in read queries,
// so it won't be exposed to clients in API responses.
type Image struct {
	ID         int    `json:"id,omitempty"`
	URL        string `json:"url,omitempty"`
	StorageRef string `json:"storage_ref,omitempty"`
}
