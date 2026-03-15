package objects

import (
	"regexp"
	"strings"
)

type UFOStatus string

const (
	StatusPending   UFOStatus = "pending"   // Row created, waiting for file
	StatusUploading UFOStatus = "uploading" // Server is currently io.Copy-ing
	StatusActive    UFOStatus = "active"    // File is safe on disk
	StatusFailed    UFOStatus = "failed"    // Upload was interrupted or error occurred
)

type ObjectMetadata struct {
	Name        string        `json:"name"`         // "original_filename.ext"
	ContentType string        `json:"content_type"` // "image/jpeg"
	Prefix      string        `json:"prefix"`       // "Path" to file
	AccessList  []AccessEntry `json:"access_list"`  // Wrapped keys for sharing
	Tags        []string      `json:"tags"`         // Human-readable tags
	SizeBytes   uint64        `json:"size_bytes"`   // Filesize in bytes
}

type AccessEntry struct {
	RecipientID string `json:"id"`
	WrappedKey  []byte `json:"key"` // The DEK encrypted for this recipient
}

func GrantAccess(m *ObjectMetadata, recipientID string) {
	// TODO: create wrapped key
	var wrappedKey []byte

	m.AccessList = append(
		m.AccessList,
		AccessEntry{
			RecipientID: recipientID,
			WrappedKey:  wrappedKey,
		},
	)
}

func RevokeAccess(m *ObjectMetadata, recipientID string) {
	for i := range m.AccessList {
		if m.AccessList[i].RecipientID == recipientID {
			// Remove the entry from the slice
			m.AccessList = append(m.AccessList[:i], m.AccessList[i+1:]...)
			return
		}
	}
}

func (m *ObjectMetadata) AddTags(tags ...string) {
	tagMap := make(map[string]struct{})
	// add existing tags to the map
	for _, tag := range m.Tags {
		tagMap[tag] = struct{}{}
	}
	// add new tags
	for _, tag := range tags {
		tagMap[strings.ToLower(tag)] = struct{}{}
	}
	// add original name, prefix, and content type for exact matches
	tagMap[strings.ToLower(m.Name)] = struct{}{}
	tagMap[strings.ToLower(m.Prefix)] = struct{}{}
	tagMap[strings.ToLower(m.ContentType)] = struct{}{}
	// add sanitized individual words from name and prefix
	for _, word := range strings.Fields(cleanString(m.Name)) {
		tagMap[strings.ToLower(word)] = struct{}{}
	}
	for _, word := range strings.Fields(cleanString(m.Prefix)) {
		tagMap[strings.ToLower(word)] = struct{}{}
	}
	// convert the map back to []string and save it
	var newTags []string
	for tag := range tagMap {
		newTags = append(newTags, tag)
	}
	m.Tags = newTags
}

func cleanString(s string) string {
	regex := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	return regex.ReplaceAllString(s, " ")
}
