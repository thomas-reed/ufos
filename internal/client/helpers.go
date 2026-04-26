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

func buildContact(existing contacts.ContactMetadata) (contacts.ContactMetadata, error) {
	fmt.Println("----- Contact Builder -----")
	if existing.PersonaID == "" {
		fmt.Println("Enter your new contact's metadata below. Hit enter skip to optional fields. Hit ctrl-c to cancel.")
	} else {
		fmt.Println("Update your contact's metadata below. Hit enter to keep the current value or skip optional empty fields. Hit ctrl-c to cancel.")
	}
	var contact contacts.ContactMetadata
	var err error

	contact.PersonaID, err = getInputWithDefault("persona ID", existing.PersonaID, true)
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	d, err := getInputWithDefault("domain", existing.Domain, true)
	if err != nil {
		return contacts.ContactMetadata{}, err
	}
	formatDomain(&d)
	contact.Domain = d

	contact.FirstName, err = getInputWithDefault("first name", existing.FirstName, false)
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	contact.LastName, err = getInputWithDefault("last name", existing.LastName, false)
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	contact.Company, err = getInputWithDefault("company", existing.Company, false)
	if err != nil {
		return contacts.ContactMetadata{}, err
	}

	// Go through any existing phones first
	if len(existing.Phones) > 0 {
		for i := range existing.Phones {
			n, err := getInputWithDefault("phone number", existing.Phones[i].Number, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			t, err := getInputWithDefault("phone type (mobile, home, main, etc.)", existing.Phones[i].PhoneType, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			contact.Phones = append(contact.Phones, contacts.Phone{
				PhoneType: t,
				Number:    n,
			})
		}
	}
	// Then loop for any new phone numbers
	for {
		n, err := getInput("new phone number", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		if n == "" {
			break
		}
		t, err := getInput("phone type (mobile, home, main, etc.)", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		contact.Phones = append(contact.Phones, contacts.Phone{
			PhoneType: t,
			Number:    n,
		})
	}

	// Go through any existing addresses first
	if len(existing.Addresses) != 0 {
		for i := range existing.Addresses {
			a1, err := getInputWithDefault("street address (line 1)", existing.Addresses[i].Address1, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			a2, err := getInputWithDefault("street address (line 2)", existing.Addresses[i].Address2, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			c, err := getInputWithDefault("city/town", existing.Addresses[i].City, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			r, err := getInputWithDefault("state/province/region", existing.Addresses[i].Region, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			p, err := getInputWithDefault("ZIP/postal code", existing.Addresses[i].PostalCode, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			n, err := getInputWithDefault("country", existing.Addresses[i].Country, false)
			if err != nil {
				return contacts.ContactMetadata{}, err
			}
			contact.Addresses = append(contact.Addresses, contacts.Address{
				Address1:   a1,
				Address2:   a2,
				City:       c,
				Region:     r,
				PostalCode: p,
				Country:    n,
			})
		}
	}
	// Then loop for any new addresses
	for {
		a1, err := getInput("street address (line 1)", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		if a1 == "" {
			break
		}
		a2, err := getInput("street address (line 2)", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		c, err := getInput("city/town", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		r, err := getInput("state/province/region", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		p, err := getInput("ZIP/postal code", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		n, err := getInput("country", false)
		if err != nil {
			return contacts.ContactMetadata{}, err
		}
		contact.Addresses = append(contact.Addresses, contacts.Address{
			Address1:   a1,
			Address2:   a2,
			City:       c,
			Region:     r,
			PostalCode: p,
			Country:    n,
		})
	}
	contact.Notes, err = getInputWithDefault("notes", existing.Notes, false)
	if err != nil {
		return contacts.ContactMetadata{}, err
	}
	return contact, nil
}

func getInput(prompt string, required bool) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if required {
			fmt.Printf("Enter %s (required): ", prompt)
		} else {
			fmt.Printf("Enter %s: ", prompt)
		}

		if !scanner.Scan() {
			return "", fmt.Errorf("Input interrupted!")
		}
		input := scanner.Text()
		if required && input == "" {
			continue
		}
		return input, nil
	}
}

func getInputWithDefault(prompt, defaultValue string, required bool) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if required {
			fmt.Printf("Enter %s (required) [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("Enter %s [%s]: ", prompt, defaultValue)
		}

		if !scanner.Scan() {
			return "", fmt.Errorf("input interrupted")
		}

		input := scanner.Text()
		if input == "" {
			if required && defaultValue == "" {
				continue
			}
			return defaultValue, nil
		}
		return input, nil
	}
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
}

func formatDomain(d *string) {
	*d = strings.ToLower(*d)
	if strings.HasPrefix(*d, "http") {
		scheme, domain, ok := strings.Cut(*d, "://")
		if !ok {
			*d = scheme // Domain must start with http?
		} else {
			*d = domain
		}
	}
	*d = strings.TrimSuffix(*d, "/")
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
	url := serverScheme + c.ActivePersona.BaseURL + api.RouteUFOs + "?" + queryValue.Encode()
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
			if err != nil {
				return nil, err
			}
			ids = append(ids, subIDs...)
		}
	}

	return ids, nil
}
