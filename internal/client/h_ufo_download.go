package client

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha3"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
	"golang.org/x/term"
)

func (c *Client) HandleDownloadUFO(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet("download", flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	id := fs.String("id", "", "The id of the file you want to update")
	fs.StringVar(id, "i", "", "alias for --id")
	host := fs.String("host", "", "The Persona ID of the user hosting the UFO when downloading from someone else's server")
	fs.StringVar(host, "h", "", "alias for --host")
	to := fs.String("to", "", "The local path where you want to save the file you are downloading. Defaults to the current directory")
	fs.StringVar(to, "t", "", "alias for --to")

	if err := fs.Parse(cmd.Args); err != nil {
		return err
	}

	// If name wasn't in Args, prompt
	if *name == "" {
		n, err := getInput("your persona name", true)
		if err != nil {
			return err
		}
		name = &n
	}

	// If id wasn't in Args, error out
	if *id == "" {
		return fmt.Errorf("Enter id of UFO you wish to download using '--id' or '-i'")
	}

	// Get master password to decrypt vault, find persona
	fmt.Print("Enter master password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("Error reading password: %w", err)
	}
	defer clear(password)
	fmt.Println()
	err = c.GetPersonaFromVault(*name, password)
	if err != nil {
		return err
	}
	defer clear(c.ActivePersona.PrivateSigningKey)
	defer clear(c.ActivePersona.PrivateExchangeKey)
	defer clear(c.MasterKey)

	// Determine Target Host and Server
	targetServer := c.ActivePersona.BaseURL
	hostPersonaID := c.PersonaID // Assume local Owner by default
	headers := make(map[string]string)

	if *host != "" {
		hostPersonaID, targetServer, ok := strings.Cut(*host, "@")
		if !ok {
			return fmt.Errorf("Invalid host identity. Expected '<persona_id@<domain>'")
		}

		// Construct the foreign server URL
		if !strings.HasPrefix(targetServer, "http") {
			targetServer = serverScheme + targetServer
		}
		targetServer = strings.TrimSuffix(targetServer, "/")
		headers[api.HeaderHost] = hostPersonaID
	}

	url := serverScheme + targetServer + api.RouteUFOs + "/" + *id
	res, err := ufoDownloadStream(c, url, headers)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Variable Initialization
	var dek, iv, hash []byte
	var filename string
	defer clear(dek)
	defer clear(iv)
	defer clear(hash)

	// Get response headers to appropriately handle the download
	metadataBase64 := res.Header.Get(api.HeaderMetadata)
	wrappedKeyBase64 := res.Header.Get(api.HeaderWrappedKey)

	if metadataBase64 != "" {
		encryptedMetadata, err := base64.StdEncoding.DecodeString(metadataBase64)
		if err != nil {
			return fmt.Errorf("Error decoding base64 metadata: %w", err)
		}
		metadataBytes, err := crypto.Decrypt(c.MasterKey, encryptedMetadata)
		if err != nil {
			return err
		}
		defer clear(metadataBytes)
		var metadata objects.ObjectMetadata
		if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("Error unmarshalling decrypted metadata: %w", err)
		}
		iv = metadata.IV
		hash = metadata.PlaintextHash
		filename = metadata.Name
		dek, err = crypto.Decrypt(
			crypto.DeriveWrappingKey(c.MasterKey, c.PersonaID),
			metadata.OwnerWrappedKey,
		)
		if err != nil {
			return err
		}
	} else if wrappedKeyBase64 != "" {
		envelope, err := base64.StdEncoding.DecodeString(wrappedKeyBase64)
		if err != nil {
			return fmt.Errorf("Error decoding base64 envelope: %w", err)
		}
		if len(envelope) < 49 {
			return fmt.Errorf("Malformed envelope: header missing")
		}
		iv = envelope[:16]
		hash = envelope[16:48]
		nameLen := int(envelope[48])

		// Get filename and wrapped key
		minRequired := 49 + nameLen + crypto.CryptoMetadataV1Size
		if len(envelope) < minRequired {
			return fmt.Errorf("Malformed envelope: filename or key truncated")
		}
		filename = string(envelope[49 : 49+nameLen])
		wrappedKey := envelope[49+nameLen:]

		// Get Persona data from Orbit
		orbitUrl := c.ActivePersona.BaseURL + api.RouteOrbit + "/" + hostPersonaID
		personaData, _, err := ufoSignedRequest[api.Satellite](c, http.MethodGet, orbitUrl, nil, nil)
		if err != nil {
			return err
		}
		sharedSecret, err := crypto.GenerateSharedSecret(
			c.ActivePersona.PrivateExchangeKey,
			personaData.ExchangeKey,
		)
		if err != nil {
			return err
		}
		guestWrappingKey := crypto.DeriveWrappingKey(sharedSecret, c.PersonaID)
		clear(sharedSecret)
		dek, err = crypto.Decrypt(guestWrappingKey, wrappedKey)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Server returned no key material!")
	}

	// Setup the Block Cipher
	block, err := aes.NewCipher(dek)
	if err != nil {
		return err
	}
	stream := cipher.NewCTR(block, iv)

	// Wrap the network body with the decryptor
	dr := &crypto.DecryptingReader{
		R:      res.Body,
		Stream: stream,
	}

	// Check file integrity
	h := sha3.New256()
	tee := io.TeeReader(dr, h)

	// Determine the download path
	downloadPath := *to
	if downloadPath == "" {
		downloadPath = filename
	} else {
		// Check if the provided path is a directory
		info, err := os.Stat(downloadPath)
		if err == nil && info.IsDir() {
			// If it is a directory, append the original filename
			downloadPath = filepath.Join(downloadPath, filename)
		}
	}

	// Create the local file
	outFile, err := os.Create(downloadPath)
	if err != nil {
		return fmt.Errorf("could not create local file: %w", err)
	}
	defer outFile.Close()

	// Copy the raw file to the outFile
	_, err = io.Copy(outFile, tee)
	if err != nil {
		return fmt.Errorf("Copy error: %w", err)
	}

	// Validate the hash
	if !bytes.Equal(h.Sum(nil), hash) {
		// Delete the downloaded file if the hashes don't match
		os.Remove(downloadPath)
		return fmt.Errorf("Hash verification failed. Downloaded file has been removed.")
	}

	fmt.Println("UFO downloaded and decrypted successfully!")
	return nil
}
