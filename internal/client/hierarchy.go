package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (c *Client) CreatePrefixHierarchy(prefix string, searchSalt []byte) error {
	// Need to make sure the folders exist for a given prefix
	segments := objects.GetPrefixSegments(prefix)
	for _, s := range segments {
		// Build metadata for the folder
		// Get folder name and hash it
		folders := strings.Split(s, "/")
		folderName := folders[len(folders)-1]
		hashedFolderName := crypto.HashTag(searchSalt, strings.ToLower(folderName))
		// Get prefix and hash it
		folderPrefix := strings.Split(s, folderName)[0]
		hashedFolderPrefix := crypto.HashTag(searchSalt, strings.ToLower(folderPrefix))

		// Create the struct
		folderMeta := objects.ObjectMetadata{
			Name:      folderName,
			Prefix:    folderPrefix,
			SizeBytes: -1,
		}

		// Make the Tags and tag hashes
		folderMeta.SyncTags()
		hashedFolderTags := make([]string, 0, len(folderMeta.Tags))
		for i := range folderMeta.Tags {
			hashedFolderTags = append(
				hashedFolderTags,
				crypto.HashTag(searchSalt, folderMeta.Tags[i]),
			)
		}

		// Encrypt the metadata
		folderMetaBytes, err := json.Marshal(folderMeta)
		if err != nil {
			return fmt.Errorf("Error marshalling metadata: %w", err)
		}
		folderMetaBlob, err := crypto.Encrypt(c.MasterKey, folderMetaBytes, crypto.CryptoSuiteV1)

		// Populate the folder UFOMetadataRequest struct
		folderReqData := api.UFOMetadataRequest{
			NameHash:   &hashedFolderName,
			PrefixHash: &hashedFolderPrefix,
			SizeBytes:  &folderMeta.SizeBytes,
			Metadata:   folderMetaBlob,
			TagHashes:  hashedFolderTags,
		}

		// Send the request to create the folder UFO database entry
		folderUrl := c.ActivePersona.BaseURL + api.RouteUFOs
		folderRes, status, err := ufoSignedRequest[api.CreateUFOResponse](
			c,
			http.MethodPost,
			folderUrl,
			folderReqData,
			nil,
		)
		if err != nil {
			return fmt.Errorf("Error sending createUFO request: %w", err)
		}
		switch status {
		case http.StatusCreated:
			fmt.Printf("'%s' folder created successfully - UFO %s.\n", folderMeta.Name, folderRes.ID)
		case http.StatusOK:
			fmt.Printf("'%s' folder exists - UFO %s.\n", folderMeta.Name, folderRes.ID)
		default:
			fmt.Printf("Unexpected status code (%s)\n", status)
		}
	}
	return nil
}
