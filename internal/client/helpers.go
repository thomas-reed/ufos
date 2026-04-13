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
		r, err := getInput("state/province/region")
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		p, err := getInput("ZIP/postal code")
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
			Region: r,
			PostalCode: p,
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

func (c *Client) getRecursiveIDs(folder api.UFOMetadataFromHeader, searchSalt []byte) ([]string, error) {
	// Get UFO Metadata to get the plaintext Prefix
	metadataBytes, err := base64.StdEncoding.DecodeString(string(folder.MetadataBlob))
	if err != nil {
		return nil, fmt.Errorf("Error decoding ufo metadata: %w", err)
	}
	metadata, err := crypto.Decrypt(c.MasterKey, metadataBytes)
	if err != nil {
		return nil, err
	}

	var ufoMeta objects.ObjectMetadata
	if err = json.Unmarshal(metadata, &ufoMeta); err != nil {
		return nil, fmt.Errorf("Error unmarshalling ufo metadata: %w", err)
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

