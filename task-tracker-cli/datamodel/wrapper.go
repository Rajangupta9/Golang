package datamodel

type Wrapper[RecordType any] struct {
	CurrentIncrement int          `json:"current_increment"`
	Records          []RecordType `json:"records"`
}
