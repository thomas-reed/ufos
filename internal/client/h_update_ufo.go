package client

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
	"golang.org/x/term"
)

func (c *Client) HandleUpdateUFO(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("update", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The id of the file you want to update")
	fs.StringVar(id, "i", "", "alias for --id")
	filename := fs.String("filename", "", "The new filename if you'd like to rename the file on the server")
	fs.StringVar(filename, "f", "", "alias for --filename")
	prefix := fs.String("prefix", "", "The 'folder' path on the server you want to move the file to, separated by '/'. Surround with quotes if the folder path contains any spaces")
	fs.StringVar(prefix, "p", "", "alias for --prefix")
	tagList := fs.String("tags", "", "The tags you wish to include for searching.  Surround multiple tags with quotes, separated by commas")
	fs.StringVar(tagList, "t", "", "alias for --tags")
	accessList := fs.String("access", "", "The Persona IDs of anyone you'd like to have downloadable access to the file.  Surround multiple IDs with quotes, separated by commas")
	fs.StringVar(accessList, "a", "", "alias for --access")

	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If name wasn't in Args, prompt
	if *name == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter desired persona name > ")
		if !scanner.Scan() {
			return fmt.Errorf("Input interrupted!")
		}
		n := scanner.Text()
		name = &n
	}

	// If id or file wasn't in Args, error out
	if *id == "" {
		return fmt.Errorf("Enter id of UFO you wish to update")
	}

	// If filename, prefix, tags, or access not in Args, error out
	if *filename == "" && *prefix == "" && *tagList == "" && *accessList == "" {
		return fmt.Errorf("Nothing entered to update - modify filename, prefix, tags, or access list")
	}

	// get master password to decrypt vault, find persona
	fmt.Printf("Enter master password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(password)
	err = c.GetPersonaFromVault(*name, password)
	if err != nil {
		return err
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	// Get UFO Metadata
	ufoUrl := c.ActivePersona.BaseURL + api.RouteUFOs + "/" + *id
	ufoRes, _, err := ufoSignedRequest[api.UFOMetadataFromHeader](c, http.MethodHead, ufoUrl, nil, nil)

	metadataBytes, err := base64.StdEncoding.DecodeString(string(ufoRes.MetadataBlob))
	if err != nil {
		return fmt.Errorf("Error decoding metadata: %w", err)
	}
	metadata, err := crypto.Decrypt(c.MasterKey, metadataBytes)
	if err != nil {
		return err
	}

	var ufoMeta objects.ObjectMetadata
	json.Unmarshal(metadata, &ufoMeta)

	searchSalt := crypto.DeriveSearchSalt(c.MasterKey, c.PersonaID)
	defer clear(searchSalt)

	// Populate the UFOMetadataRequest struct
	ufoReqData := api.UFOMetadataRequest{}

	tagsChanged := false
	// Filename change
	if *filename != "" {
		ufoMeta.Name = *filename
		hashedName := crypto.HashTag(searchSalt, strings.ToLower(*filename))
		ufoReqData.NameHash = &hashedName
		tagsChanged = true
	}

	// Prefix change
	if *prefix != "" {
		ufoMeta.Prefix = *prefix
		hashedPrefix := crypto.HashTag(searchSalt, strings.ToLower(*prefix))
		ufoReqData.PrefixHash = &hashedPrefix
		tagsChanged = true

		// If prefix changed, need to make sure the folders exist
		if err := c.CreatePrefixHierarchy(*prefix, searchSalt); err != nil {
			return fmt.Errorf("Error building prefix hierarchy: %w", err)
		}
	}

	// Tags change
	if *tagList != "" {
		ufoMeta.UserTags = strings.Split(*tagList, ",")
		tagsChanged = true
	}
	if tagsChanged {
		ufoMeta.SyncTags()
		hashedTags := make([]string, 0, len(ufoMeta.Tags))
		for i := range ufoMeta.Tags {
			hashedTags = append(
				hashedTags,
				crypto.HashTag(searchSalt, ufoMeta.Tags[i]),
			)
		}
		ufoReqData.TagHashes = hashedTags
	}

	// Get Orbit
	orbitUrl := c.ActivePersona.BaseURL + api.RouteOrbit
	orbit, _, err := ufoSignedRequest[[]api.OrbitItem](c, http.MethodGet, orbitUrl, nil, nil)

	// Make orbit into a map for faster searching for access list
	orbitMap := make(map[string]api.OrbitItem)
	for _, p := range orbit {
		orbitMap[p.PersonaID] = p
	}

	// Access List change
	if *accessList != "" {
		accessListMap := make(map[string][]byte)
		access := strings.Split(*accessList, ",")
		for i := range access {
			recipientID := strings.TrimSpace(access[i])
			if recipientID == "" {
				continue
			}
			if _, found := orbitMap[recipientID]; !found {
				fmt.Printf(
					"Skipping '%s' - persona is not in your orbit. Use 'ufos orbit add -u <Persona_ID>'",
					recipientID,
				)
				continue
			}
			// Provide way to allow the recipient to decrypt the file
			sharedSecret, err := crypto.GenerateSharedSecret(
				c.ActivePersona.PrivateExchangeKey,
				orbitMap[recipientID].ExchangeKey,
			)
			if err != nil {
				return err
			}

			// Get the original DEK
			dek, err := crypto.Decrypt(c.MasterKey, ufoMeta.OwnerWrappedKey)
			if err != nil {
				return err
			}
			defer clear(dek)

			// Derive the wrapping key for the recipient
			guestWrappingKey := crypto.DeriveWrappingKey(sharedSecret, recipientID)
			clear(sharedSecret)

			// Wrap the DEK
			wrappedDEK, err := ufoMeta.GrantAccess(recipientID, guestWrappingKey, dek)
			clear(guestWrappingKey)
			if err != nil {
				return err
			}

			// Add IV and hash to the wrappedDEK to package up decryption info
			envelope := make([]byte, 0, len(ufoMeta.IV)+len(ufoMeta.PlaintextHash)+len(wrappedDEK))
			envelope = append(envelope, ufoMeta.IV...)
			envelope = append(envelope, ufoMeta.PlaintextHash...)
			envelope = append(envelope, wrappedDEK...)

			accessListMap[recipientID] = envelope
		}
		defer func() {
			for _, envelope := range accessListMap {
				clear(envelope)
			}
			clear(accessListMap)
		}()
		ufoReqData.AccessList = accessListMap
	}

	// Encrypt the metadata
	metaBytes, err := json.Marshal(ufoMeta)
	if err != nil {
		return fmt.Errorf("Error marshalling metadata: %w", err)
	}
	metaBlob, err := crypto.Encrypt(c.MasterKey, metaBytes, crypto.CryptoSuiteV1)
	if err != nil {
		return err
	}
	ufoReqData.Metadata = metaBlob

	// Send the request to create the UFO database entry
	url := c.ActivePersona.BaseURL + api.RouteUFOs + "/" + *id
	res, _, err := ufoSignedRequest[api.UpdateUFOResponse](
		c,
		http.MethodPatch,
		url,
		ufoReqData,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error sending createUFO request: %w", err)
	}

	fmt.Printf("UFO %s updated successfully.\n", res.ID)

	return nil
}
