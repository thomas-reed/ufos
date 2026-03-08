package client

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomas-reed/ufos/internal/crypto"
)

type VaultVersion uint16

const (
	_ VaultVersion = iota
	VaultV1
	// VaultV2
)

const (
	vaultFilename = ".ufosvault.json"

	timeCost    uint32 = 3
	memoryCost  uint32 = 32 * 1024
	parallelism uint8  = 4
)

type Persona struct {
	Name       string             `json:"name"`        // user-provided descriptor
	BaseURL    string             `json:"base_url"`    // URL of the server
	PrivateKey ed25519.PrivateKey `json:"private_key"` // private key for this identity
}

type Vault struct {
	Version   VaultVersion `json:"version"`
	KDFSalt   []byte       `json:"kdf_salt"`
	KDFParams KDF          `json:"kdf_params"`
	Payload   []byte       `json:"payload"`
}

type KDF struct {
	TimeCost    uint32 `json:"time_cost"`
	MemoryCost  uint32 `json:"memory_cost"`
	Parallelism uint8  `json:"parallelism"`
}

func (c *Client) getVaultFilepath() error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Could not get homedir: %w", err)
	}
	c.VaultPath = filepath.Join(homedir, vaultFilename)
	return nil
}

func (c *Client) GetPersonaFromVault(password []byte) error {
	v, err := c.loadVault()
	if err != nil {
		return err
	}
	switch v.Version {
	case VaultV1:
		if err = c.getPersonaFromVaultV1(v, password); err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid vault version")
	}
	return nil
}

func (c *Client) AddPersonaToVault(baseURL string, password []byte) error {
	v, err := c.loadVault()
	if err != nil {
		return err
	}

	switch v.Version {
	case VaultV1:
		if err = c.addPersonaToVaultV1(baseURL, password); err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid vault version")
	}
	return nil
}

func (c *Client) CreateNewVault(baseURL string, password []byte) error {
	if _, err := os.Stat(c.VaultPath); err == nil {
		return fmt.Errorf("Vault already exists")
	}
	var v Vault
	var vaultKey []byte
	var err error

	v.Version = VaultV1
	v.KDFParams.TimeCost = timeCost
	v.KDFParams.MemoryCost = memoryCost
	v.KDFParams.Parallelism = parallelism
	vaultKey, v.KDFSalt, err = crypto.CreateVaultKey(
		password,
		v.KDFParams.TimeCost,
		v.KDFParams.MemoryCost,
		v.KDFParams.Parallelism,
	)
	defer clear(vaultKey)

	persona, err := c.buildPersona(baseURL)
	if err != nil {
		return err
	}
	defer clear(persona.PrivateKey)

	payload, err := json.Marshal([]Persona{persona})
	defer clear(payload)

	v.Payload, err = crypto.Encrypt(vaultKey, payload, crypto.CryptoSuiteV1)
	if err != nil {
		return fmt.Errorf("Error encrypting persona payload: %w", err)
	}

	return c.saveVault(v)
}

func (c *Client) getPersonaFromVaultV1(v Vault, password []byte) error {
	vaultKey, err := crypto.DeriveVaultKey(
		password,
		v.KDFSalt,
		v.KDFParams.TimeCost,
		v.KDFParams.MemoryCost,
		v.KDFParams.Parallelism,
	)
	defer clear(vaultKey)
	data, err := crypto.Decrypt(vaultKey, v.Payload)
	if err != nil {
		return err
	}
	defer clear(data)

	var personas []Persona
	err = json.Unmarshal(data, &personas)
	if err != nil {
		return fmt.Errorf("Error unmarshalling vault data: %w", err)
	}
	defer func() {
		for i := range personas {
			clear(personas[i].PrivateKey)
		}
	}()

	if err = c.resolvePersona(personas); err != nil {
		return err
	}
	for i := range personas {
		if c.PersonaData.Name == personas[i].Name &&
			c.PersonaData.BaseURL == personas[i].BaseURL {
			c.PersonaData.PrivateKey = make([]byte, len(personas[i].PrivateKey))
			copy(c.PersonaData.PrivateKey, personas[i].PrivateKey)
			c.PersonaID = crypto.DerivePersonaID(
				c.PersonaData.PrivateKey.Public().(ed25519.PublicKey),
				c.PersonaData.Name,
			)
			c.MasterKey = crypto.DeriveMasterKey(
				c.PersonaData.PrivateKey,
				c.PersonaID,
			)
		}
	}
	return nil
}

func (c *Client) resolvePersona(personas []Persona) error {
	parts := strings.Split(c.PersonaData.Name, "@")
	if len(parts) > 2 {
		return fmt.Errorf("Persona name invalid. Usage: <name>[@<domain>]")
	}
	personaName := parts[0]
	domains := []string{}
	for _, persona := range personas {
		if persona.Name == personaName {
			domains = append(domains, persona.BaseURL)
		}
	}

	switch len(domains) {
	case 0:
		return fmt.Errorf("Persona '%s' not found", personaName)
	case 1:
		// If the user was explicit (name@domain), it should match the one returned
		if len(parts) == 2 && domains[0] != parts[1] {
			return fmt.Errorf(
				"Persona '%s' not found for domain '%s'",
				personaName,
				parts[1],
			)
		}
		c.PersonaData.Name = personaName
		c.PersonaData.BaseURL = domains[0]
		return nil
	default: // domains >= 2
		// If the user was explicit, find the domain they specified
		if len(parts) == 2 {
			for _, domain := range domains {
				if domain == parts[1] {
					c.PersonaData.Name = personaName
					c.PersonaData.BaseURL = domain
					return nil
				}
			}
			return fmt.Errorf(
				"Persona '%s' not found for domain '%s'",
				personaName,
				parts[1],
			)
		}
		return fmt.Errorf(
			"Use '%s@<domain>' - multiple domains found: %q",
			personaName,
			domains,
		)
	}
}

func (c *Client) loadVault() (vault Vault, err error) {
	data, err := os.ReadFile(c.VaultPath)
	if err != nil {
		return Vault{}, fmt.Errorf("Could not read vault file: %w", err)
	}
	if err := json.Unmarshal(data, &vault); err != nil {
		return Vault{}, fmt.Errorf("Error unmarshalling vault data: %w", err)
	}
	return vault, nil
}

func (c *Client) saveVault(v Vault) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	// 0600 ensures only the server process can read/write this file
	return os.WriteFile(c.VaultPath, data, 0600)
}

func (c *Client) addPersonaToVaultV1(baseURL string, password []byte) error {
	v, err := c.loadVault()
	if err != nil {
		return err
	}
	vaultKey, err := crypto.DeriveVaultKey(
		password,
		v.KDFSalt,
		v.KDFParams.TimeCost,
		v.KDFParams.MemoryCost,
		v.KDFParams.Parallelism,
	)
	defer clear(vaultKey)
	data, err := crypto.Decrypt(vaultKey, v.Payload)
	if err != nil {
		return err
	}
	defer clear(data)

	var personas []Persona
	err = json.Unmarshal(data, &personas)
	if err != nil {
		return fmt.Errorf("Error unmarshalling vault data: %w", err)
	}
	defer func() {
		for i := range personas {
			clear(personas[i].PrivateKey)
		}
	}()

	var persona Persona
	for i := range personas {
		if personas[i].BaseURL == baseURL {
			persona.Name = c.PersonaData.Name
			persona.BaseURL = baseURL
			persona.PrivateKey = make([]byte, len(personas[i].PrivateKey))
			copy(persona.PrivateKey, personas[i].PrivateKey)
		}
	}
	if persona.PrivateKey == nil {
		persona, err = c.buildPersona(baseURL)
		if err != nil {
			return err
		}
	}
	defer clear(persona.PrivateKey)

	payload, err := json.Marshal(append(personas, persona))

	v.Payload, err = crypto.Encrypt(vaultKey, payload, crypto.CryptoSuiteV1)
	if err != nil {
		return fmt.Errorf("Error encrypting persona payload: %w", err)
	}

	return c.saveVault(v)
}

func (c *Client) buildPersona(baseURL string) (Persona, error) {
	persona := Persona{
		Name:    c.PersonaData.Name,
		BaseURL: baseURL,
	}
	_, privateKey, err := crypto.GenerateAsymPrivateKey()
	if err != nil {
		return Persona{}, fmt.Errorf("Error generating private key: %w", err)
	}
	persona.PrivateKey = privateKey
	return persona, nil
}
