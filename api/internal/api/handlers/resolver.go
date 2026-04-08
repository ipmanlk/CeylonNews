package handlers

// SourceResolver provides source name lookup by ID
type SourceResolver interface {
	GetSourceNameByID(id string) (string, bool)
}
