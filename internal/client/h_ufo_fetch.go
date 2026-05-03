package client

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha3"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"golang.org/x/term"
)

func (c *Client) HandleFetchUFO(cmd Command) error {
	// Set up flags and parse
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	name := fs.String("name", "", "The name of the persona you wish to use. Specify '@<domain>' if you have use the same persona name for multiple domains)")
	fs.StringVar(name, "n", "", "alias for --name")
	url := fs.String("url", "", "The direct url to the UFO you want to fetch from someone else's server")
	fs.StringVar(url, "u", "", "alias for --url")
	host := fs.String("host", "", "The Persona ID of the user hosting the UFO")
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

	// If url wasn't in Args, error out
	if *url == "" {
		return fmt.Errorf("Enter url of UFO you wish to fetch using '--url' or '-u'")
	}

	// If host wasn't in Args, error out
	if *host == "" {
		return fmt.Errorf("Enter Host's Persona ID using '--host' or '-h'")
	}

	// Get vault password to decrypt vault, find persona
	fmt.Print("Enter your vault password: ")
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

	headers := map[string]string{
		api.HeaderHost: *host,
	}

	res, err := ufoDownloadStream(c, *url, headers)
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

	// Get response header
	wrappedKeyBase64 := res.Header.Get(api.HeaderWrappedKey)

	if wrappedKeyBase64 != "" {
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
		var hostPublicKey []byte
		orbitUrl := serverScheme + c.ActivePersona.BaseURL + api.RouteOrbit + "/" + *host
		if personaData, status, err := ufoSignedRequest[api.Satellite](c, http.MethodGet, orbitUrl, nil, nil); err == nil && status == http.StatusOK {
			hostPublicKey = personaData.ExchangeKey
		} else {
			// Host must not be in the Guest's orbit for some reason, so fetch public keys from host's server
			log.Printf("Failed to get host's public keys from orbit (%d): %s\nFetching directly from host server\n", status, err)
			hostURL, _, ok := strings.Cut(*url, api.RouteUFOs)
			if !ok {
				// This should never happen, because the downloadStream would have failed
				return fmt.Errorf("Invalid URL")
			}
			hostURL = hostURL + api.RoutePersonas + "/" + *host
			keys, keyStatus, err := ufoPublicRequest[api.PersonaKeysResponse](c, http.MethodGet, hostURL, nil, nil)
			if err != nil || keyStatus != http.StatusOK {
				return fmt.Errorf("Failed to fetch Host's public key (%d): %w", keyStatus, err)
			}
			hostPublicKey = keys.ExchangeKey
		}
		sharedSecret, err := crypto.GenerateSharedSecret(
			c.ActivePersona.PrivateExchangeKey,
			hostPublicKey,
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
		return fmt.Errorf("Server returned no wrapped key!")
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
