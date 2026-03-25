package client

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
	"golang.org/x/term"
)

func (c *Client) HandleUploadUFO(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	filePath := fs.String("file", "", "The local path (relative or absolute) to the file you want to upload")
	fs.StringVar(filePath, "f", "", "alias for --file")
	prefix := fs.String("prefix", "", "The 'folder' path on the server you want to upload the file to, separated by '/'. Surround with quotes if the folder path contains any spaces")
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

	// If file wasn't in Args, prompt, and open the file
	if *filePath == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter desired personfile name > ")
		if !scanner.Scan() {
			return fmt.Errorf("Input interrupted!")
		}
		fp := scanner.Text()
		filePath = &fp
	}

	// If prefix wasn't in Args, default to root '/'
	if *prefix == "" {
		p := "/"
		prefix = &p
	}

	// get master password to decrypt vault, find persona
	fmt.Printf("Enter master password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(pw)
	err = c.GetPersonaFromVault(*name, pw)
	if err != nil {
		return err
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	// Generate data encryption key
	dek, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("Error generating encryption key", err)
	}
	defer clear(dek)

	// Verify/Open the file on disk
	file, err := os.Open(*filePath)
	if err != nil {
		return fmt.Errorf("Cant open '%s': %w", *filePath, err)
	}
	defer file.Close()

	// Build Object metadata
	// File name, size
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Content Type
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	// Fallback if generic
	if contentType == "application/octet-stream" {
		ext := filepath.Ext(*filePath)
		if m := mime.TypeByExtension(ext); m != "" {
			contentType = m
		}
	}
	file.Seek(0, io.SeekStart)

	// Owner data
	wrappingKey := crypto.DeriveWrappingKey(c.MasterKey, c.PersonaID)
	defer clear(wrappingKey)
	wrappedDEK, err := crypto.Encrypt(wrappingKey, dek, crypto.CryptoSuiteV1)
	if err != nil {
		return fmt.Errorf("Error encrypting DEK: %w", err)
	}

	ufoMeta := objects.ObjectMetadata{
		Name:            info.Name(),
		ContentType:     contentType,
		SizeBytes:       uint64(info.Size()),
		Prefix:          *prefix,
		OwnerID:         c.PersonaID,
		OwnerWrappedKey: wrappedDEK,
		Tags:            []string{},
		AccessList:      []objects.AccessEntry{},
	}

	// Tags and Prefix
	tags := strings.Split(*tagList, ",")
	searchSalt := crypto.DeriveSearchSalt(c.MasterKey, c.PersonaID)
	defer clear(searchSalt)
	// Get the hashed prefix
	hashedPrefix := crypto.HashTag(searchSalt, *prefix)
	// Make the list of hashed tags while cleaning the tags for storage in the metadata
	hashedTags := make([]string, 0, len(tags))
	for i := range tags {
		tags[i] = strings.ToLower(strings.TrimSpace(tags[i]))
		hashedTags = append(hashedTags, crypto.HashTag(searchSalt, tags[i]))
	}
	ufoMeta.AddTags(tags...)

	// Get Orbit
	url := c.ActivePersona.BaseURL + api.RouteOrbit
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = ServerScheme + url
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request %s", err)
	}

	timestamp := time.Now().Unix()
	if err = c.Sign(req, timestamp, false); err != nil {
		return fmt.Errorf("Error signing request: %w", err)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error executing request %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body %w", err)
	}

	var orbit []api.OrbitItem
	if err = json.Unmarshal(data, &orbit); err != nil {
		return fmt.Errorf("Error unmarshalling orbit data %w", err)
	}
	// make orbit into a map for faster searching for access list
	orbitMap := make(map[string]api.OrbitItem)
	for _, p := range orbit {
		orbitMap[p.PersonaID] = p
	}

	// Access List
	accessListMap := make(map[string][]byte)
	access := strings.Split(*accessList, ",")
	for i := range access {
		access[i] = strings.TrimSpace(access[i])
		if _, found := orbitMap[access[i]]; !found {
			fmt.Printf(
				"Skipping '%s' - persona is not in your orbit. Use 'ufos orbit add -u <Persona_ID>'",
				access[i],
			)
			continue
		}
		sharedSecret, err := crypto.GenerateSharedSecret(
			c.ActivePersona.PrivateExchangeKey,
			orbitMap[access[i]].ExchangeKey,
		)
		if err != nil {
			return fmt.Errorf("Error generating shared secret %w", err)
		}
		guestWrappingKey := crypto.DeriveWrappingKey(sharedSecret, access[i])
		wrappedDEK, err := ufoMeta.GrantAccess(access[i], guestWrappingKey, dek)
		clear(sharedSecret)
		clear(guestWrappingKey)
		accessListMap[access[i]] = wrappedDEK
	}

	// Encrypt the metadata
	metaBytes, err := json.Marshal(ufoMeta)
	if err != nil {
		return fmt.Errorf("Error marshalling metadata: %w", err)
	}
	metaBlob, err := crypto.Encrypt(c.MasterKey, metaBytes, crypto.CryptoSuiteV1)

	// Populate the UFOMetadataRequest struct
	UFOReqData := api.UFOMetadataRequest{
		PrefixHash: hashedPrefix,
		SizeBytes:  int64(ufoMeta.SizeBytes) + crypto.CryptoMetadataV1Size,
		Metadata:   metaBlob,
		TagHashes:  hashedTags,
		AccessList: accessListMap,
	}

}
