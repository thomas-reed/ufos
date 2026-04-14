package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/contacts"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (c *Client) printUFOList(list []api.UFO) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	// Print the Header
	fmt.Fprintln(w, "ID\tTYPE\tPREFIX\tNAME\tLINK")
	fmt.Fprintln(w, "--\t----\t------\t----\t----")

	// Decrypt the list
	for _, ufo := range list {
		metadataBytes, err := crypto.Decrypt(c.MasterKey, ufo.Metadata)
		if err != nil {
			return err
		}
		var metadata objects.ObjectMetadata
		if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("Error unmarshalling ufo metadata: %w", err)
		}

		filetype := "FILE"
		if metadata.SizeBytes < 0 {
			filetype = "DIR"
		}

		link := serverScheme + c.ActivePersona.BaseURL + api.RouteUFOs + "/" + ufo.ID

		// Write row to the tabwriter
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ufo.ID,
			filetype,
			metadata.Prefix,
			metadata.Name,
			link,
		)
		clear(metadataBytes)
		metadata = objects.ObjectMetadata{}
	}
	return w.Flush()
}

func (c *Client) printUFODetails(ufo api.UFOMetadataFromHeader) error {
	// Decrypt UFO metadata
	ufoMetaBytes, err := crypto.Decrypt(c.MasterKey, ufo.MetadataBlob)
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
	if len(accessList) == 0 {
		fmt.Println("- Private (No guests authorized)")
	} else {
		for _, contact := range accessList {
			fmt.Printf("- %s %s\n", contact.FirstName, contact.LastName)
		}
	}

	ufoMetadata = objects.ObjectMetadata{}
	return nil
}

func (c *Client) printOrbitList(orbit []api.Satellite) error {
	// Print the Header
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tID\tDOMAIN\tCOMPANY")
	fmt.Fprintln(w, "----\t--\t------\t-------")

	// Decrypt the list
	for _, sat := range orbit {
		metadataBytes, err := crypto.Decrypt(c.MasterKey, sat.Metadata)
		if err != nil {
			return err
		}
		var metadata contacts.ContactMetadata
		if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("Error unmarshalling satellite metadata: %w", err)
		}

		// Write row to the tabwriter
		fmt.Fprintf(w, "%s %s\t%s\t%s\t%s\n",
			metadata.FirstName,
			metadata.LastName,
			sat.PersonaID,
			metadata.Domain,
			metadata.Company,
		)
		clear(metadataBytes)
		metadata = contacts.ContactMetadata{}
	}
	return w.Flush()
}

func (c *Client) printSatelliteDetails(sat api.Satellite) error {
	// Decrypt UFO metadata
	satMetaBytes, err := crypto.Decrypt(c.MasterKey, sat.Metadata)
	if err != nil {
		return err
	}
	defer clear(satMetaBytes)
	var satMetadata contacts.ContactMetadata
	if err = json.Unmarshal(satMetaBytes, &satMetadata); err != nil {
		return fmt.Errorf("Error unmarshalling satellite metadata: %w", err)
	}
	fmt.Println("CONTACT DETAILS:")
	fmt.Printf("Name: %s %s\n", satMetadata.FirstName, satMetadata.LastName)
	fmt.Printf("Domain: %s\n", satMetadata.Domain)
	fmt.Printf("Notes:\n%s\n", satMetadata.Notes)
	fmt.Println("Phones:")
	if len(satMetadata.Phones) == 0 {
		fmt.Println("- No phone numbers saved.")
	} else {
		for _, phone := range satMetadata.Phones {
			fmt.Printf("- %s: %s\n", phone.PhoneType, phone.Number)
		}
	}
	fmt.Println("Addresses:")
	if len(satMetadata.Addresses) == 0 {
		fmt.Println("- No addresses saved.")
	} else {
		for _, address := range satMetadata.Addresses {
			fmt.Printf("--- %s\n", address.Label)
			fmt.Printf("- %s\n", address.Address1)
			if address.Address2 != "" {
				fmt.Printf("- %s\n", address.Address2)
			}
			fmt.Printf("- %s, %s  %s\n", address.City, address.Region, address.PostalCode)
			fmt.Printf("- %s\n", address.Country)
			fmt.Println("---")
		}
	}

	satMetadata = contacts.ContactMetadata{}
	return nil
}
