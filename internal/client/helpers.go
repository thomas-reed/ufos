package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/contacts"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
)

func formatPrefix(prefix *string) {
	if *prefix == "" {
		*prefix = "/"
		return
	}
	if !strings.HasPrefix(*prefix, "/") {
		*prefix = "/" + *prefix
	}
	if !strings.HasSuffix(*prefix, "/") {
		*prefix = *prefix + "/"
	}
	return
}

func (c *Client) printUFOList(list []api.UFO) error {
	// (Params: output, minwidth, tabwidth, padding, padchar, flags)
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	// Print the Header
	fmt.Fprintln(w, "ID\tTYPE\tPATH\tNAME")
	fmt.Fprintln(w, "--\t----\t----\t----")
	for _, ufo := range list {
		metadataBytes, err := crypto.Decrypt(c.MasterKey, ufo.Metadata)
		if err != nil {
			return err
		}
		var metadata objects.ObjectMetadata
		if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("Error unmarshalling metadata: %w", err)
		}

		filetype := "FILE"
		if metadata.SizeBytes < 0 {
			filetype = "DIR"
		}

		// Write row to the tabwriter
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			ufo.ID,
			filetype,
			metadata.Prefix,
			metadata.Name,
		)
		clear(metadataBytes)
		metadata = objects.ObjectMetadata{}
	}
	return w.Flush()
}

func (c *Client) printUFODetails(item api.UFO) error {
	// Decrypt UFO metadata
	ufoMetaBytes, err := crypto.Decrypt(c.MasterKey, item.Metadata)
	if err != nil {
		return err
	}
	defer clear(ufoMetaBytes)
	var ufoMetadata objects.ObjectMetadata
	if err = json.Unmarshal(ufoMetaBytes, &ufoMetadata); err != nil {
		return fmt.Errorf("Error unmarshalling ufo metadata: %w", err)
	}

	// Get Orbit and construct access list map
	orbitUrl := c.ActivePersona.BaseURL + api.RouteOrbit
	orbit, _, err := ufoSignedRequest[[]api.Satellite](c, http.MethodGet, orbitUrl, nil, nil)
	orbitMap := make(map[string]contacts.ContactMetadata)
	for _, sat := range orbit {
		// Decrypt satellite metadata
		satMetaBytes, err := crypto.Decrypt(c.MasterKey, sat.Metadata)
		if err != nil {
			return err
		}

		var satMetadata contacts.ContactMetadata
		if err = json.Unmarshal(satMetaBytes, &satMetadata); err != nil {
			return fmt.Errorf("Error unmarshalling satellite metadata: %w", err)
		}

		orbitMap[sat.PersonaID] = satMetadata
		clear(satMetaBytes)
		satMetadata = contacts.ContactMetadata{}
	}

	// Loop through access list to make list of contacts
	accessList := make([]contacts.ContactMetadata, 0, len(ufoMetadata.AccessList))
	for _, entry := range ufoMetadata.AccessList {
		if contact, ok := orbitMap[entry.RecipientID]; ok {
			accessList = append(accessList, contact)
		}
	}

	filetype := "FILE"
	sizeStr := fmt.Sprintf("%d B", ufoMetadata.SizeBytes)
	if ufoMetadata.SizeBytes < 0 {
		filetype = "DIR"
		sizeStr = "-" // Folders don't have a binary size
	}

	fmt.Println("UFO DETAILS:")
	fmt.Printf("Name: %s\n", ufoMetadata.Name)
	fmt.Printf("Prefix: %s\n", ufoMetadata.Prefix)
	fmt.Printf("Type: %s\n", filetype)
	fmt.Printf("Size: %s\n", sizeStr)
	fmt.Println("Tags:")
	fmt.Println(ufoMetadata.UserTags)
	fmt.Println("Access List:")
	for _, contact := range accessList {
		fmt.Printf("%s %s\n", contact.FirstName, contact.LastName)
	}

	ufoMetadata = objects.ObjectMetadata{}
	return nil
}
