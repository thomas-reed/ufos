package client

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/contacts"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
)

func buildContact() (contacts.ContactMetadata, error) {
	fmt.Println("----- Contact Builder -----")
	fmt.Println("Enter your new contact's metadata below. Hit enter with no value to skip to next field.")
	var contact contacts.ContactMetadata
	var err error

	contact.PersonaID, err = getInput("persona ID")
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	d, err := getInput("domain")
	if err != nil {
		return contacts.ContactMetadata{}, err
	}
	if strings.HasPrefix(d, "http") {
		scheme, domain, ok := strings.Cut(d, "://")
		if !ok {
			d = scheme // assume the domain actually starts with http
		} else {
			d = domain
		}
	}
	if strings.HasSuffix(d, "/") {
		d = d[:len(d)-1]
	}
	contact.Domain = d

	contact.FirstName, err = getInput("first name")
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	contact.LastName, err = getInput("last name")
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	contact.Company, err = getInput("company")
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	for {
		n, err := getInput("phone number")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		if n == "" {
			break
		}
		t, err := getInput("phone type (mobile, home, main, etc.)")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		contact.Phones = append(contact.Phones, contacts.Phone{
			PhoneType: t,
			Number: n,
		})
	}

	for {
		a1, err := getInput("street address (line 1)")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		if a1 == "" {
			break
		}
		a2, err := getInput("street address (line 2)")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		c, err := getInput("city/town")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		s, err := getInput("state/region")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		z, err := getInput("zip/postal code")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		n, err := getInput("country")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		contact.Addresses = append(contact.Addresses, contacts.Address{
			Address1: a1,
			Address2: a2,
			City: c,
			StateRegion: s,
			Zip: z,
			Country: n,
		})
	}
	contact.Notes, err = getInput("notes")
	if err != nil {
		return contacts.ContactMetadata{}, err
	}
	return contact, nil
}

func getInput(inputStr string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter %s > ", inputStr)
	if !scanner.Scan() {
		return "", fmt.Errorf("Input interrupted!")
	}
	inputTxt := scanner.Text()
	return inputTxt, nil
}

func formatPrefix(prefix *string) {
	if *prefix == "" {
		*prefix = "/"
		return
	}
	*prefix = strings.ToLower(*prefix)
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

func (c *Client) getRecursiveIDs(folder api.UFOMetadataFromHeader, searchSalt []byte) ([]string, error) {
	// Get UFO Metadata to get the plaintext Prefix
	metadataBytes, err := base64.StdEncoding.DecodeString(string(folder.MetadataBlob))
	if err != nil {
		return nil, fmt.Errorf("Error decoding metadata: %w", err)
	}
	metadata, err := crypto.Decrypt(c.MasterKey, metadataBytes)
	if err != nil {
		return nil, err
	}

	var ufoMeta objects.ObjectMetadata
	if err = json.Unmarshal(metadata, &ufoMeta); err != nil {
		return nil, fmt.Errorf("Error unmarshalling metadata: %w", err)
	}
	
	// Build the hashedPath for this folder's children
	folderPrefix := ufoMeta.Prefix + ufoMeta.Name
	formatPrefix(&folderPrefix)
	hashedPrefix := crypto.HashTag(searchSalt, folderPrefix)
	
	// List children
	queryValue := url.Values{}
	queryValue.Add("prefix", hashedPrefix)

	// Send the request to list UFOs
	url := c.ActivePersona.BaseURL + api.RouteUFOs + "?" + queryValue.Encode()
	children, _, err := ufoSignedRequest[[]api.UFO](
		c,
		http.MethodGet,
		url,
		nil,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Error getting UFO list for %s: %w", folderPrefix, err)
	}
	
	ids := []string{}
	
	for _, child := range children {
		ids = append(ids, child.ID)
		if child.SizeBytes < 0 {
			// It's a folder, Recurse
			childMeta := api.UFOMetadataFromHeader{
				MetadataBlob: child.Metadata,
			}
			subIDs, err := c.getRecursiveIDs(childMeta, searchSalt)
			if err != nil { return nil, err }
			ids = append(ids, subIDs...)
		}
	}
	
	return ids, nil
}
