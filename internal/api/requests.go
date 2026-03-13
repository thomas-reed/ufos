package api

// for use with CreateUFO, UpdateUFO
type UFOMetadataRequest struct {
	PrefixHash string `json:"prefix_hash"`
	SizeBytes  int64  `json:"size_bytes"`
	Metadata   []byte `json:"metadata"`
}
