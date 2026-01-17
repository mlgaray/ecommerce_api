package models

type Timezone struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	UTCOffset  string `json:"utc_offset,omitempty"`
}
