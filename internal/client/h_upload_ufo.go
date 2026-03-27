package client

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
		fmt.Print("Enter local filepath for file to upload > ")
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

	// Generate data encryption key
	dek, err := crypto.GenerateKey()
	if err != nil {
		return err
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
		return fmt.Errorf("Error getting file data: %w", err)
	}

	fileHash, err := crypto.HashFile(file)
	if err != nil {
		return err
	}
	defer clear(fileHash)

	iv, err := crypto.GenerateIV()
	if err != nil {
		return err
	}
	defer clear(iv)

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
		IV:              iv,
		PlaintextHash:   fileHash,
	}

	// Tags and Prefix
	searchSalt := crypto.DeriveSearchSalt(c.MasterKey, c.PersonaID)
	defer clear(searchSalt)
	// Get the hashed prefix
	hashedPrefix := crypto.HashTag(searchSalt, *prefix)
	// Make the list of hashed tags while cleaning the tags for storage in the metadata
	ufoMeta.UserTags = strings.Split(*tagList, ",")
	ufoMeta.SyncTags()
	hashedTags := make([]string, 0, len(ufoMeta.Tags))
	for i := range ufoMeta.Tags {
		hashedTags = append(hashedTags, crypto.HashTag(searchSalt, ufoMeta.Tags[i]))
	}

	// Get Orbit
	orbitUrl := c.ActivePersona.BaseURL + api.RouteOrbit
	orbit, err := ufoSignedRequest[[]api.OrbitItem](c, http.MethodGet, orbitUrl, nil, nil)

	// make orbit into a map for faster searching for access list
	orbitMap := make(map[string]api.OrbitItem)
	for _, p := range orbit {
		orbitMap[p.PersonaID] = p
	}

	// Access List
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
		envelope := make([]byte, 0, len(iv)+len(ufoMeta.PlaintextHash)+len(wrappedDEK))
		envelope = append(envelope, iv...)
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

	// Encrypt the metadata
	metaBytes, err := json.Marshal(ufoMeta)
	if err != nil {
		return fmt.Errorf("Error marshalling metadata: %w", err)
	}
	metaBlob, err := crypto.Encrypt(c.MasterKey, metaBytes, crypto.CryptoSuiteV1)

	// Populate the UFOMetadataRequest struct
	UFOReqData := api.UFOMetadataRequest{
		PrefixHash: hashedPrefix,
		SizeBytes:  int64(ufoMeta.SizeBytes),
		Metadata:   metaBlob,
		TagHashes:  hashedTags,
		AccessList: accessListMap,
	}

	// Send the request to create the UFO database entry
	url := c.ActivePersona.BaseURL + api.RouteUFOs
	res, err := ufoSignedRequest[api.CreateUFOResponse](
		c,
		http.MethodPost,
		url,
		UFOReqData,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error sending createUFO request: %w", err)
	}

	// Send the file data for storage
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		// Set up the Block Cipher (AES-CTR)
		block, err := aes.NewCipher(dek)
		if err != nil {
			pr.CloseWithError(err)
			return
		}

		stream := cipher.NewCTR(block, iv)

		// Set up the writer to encrypt and write streamed bytes
		writer := &crypto.EncryptingWriter{W: pw, Stream: stream}

		// Copy the bytes from disk into the pipe
		if _, err := io.Copy(writer, file); err != nil {
			pr.CloseWithError(err)
			return
		}
	}()

	streamUrl := url + "/" + res.ID
	err = ufoStreamRequest(c, "PUT", streamUrl, pr, UFOReqData.SizeBytes, nil)
	if err != nil {
		return fmt.Errorf("Error sending UFO data: %w", err)
	}

	fmt.Printf("UFO %s uploaded successfully.\n", res.ID)
	return nil
}
