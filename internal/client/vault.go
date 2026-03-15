package client

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
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

func (c *Client) GetPersonaFromVault(personaName string, password []byte) error {
	v, err := loadVault()
	if err != nil {
		return err
	}
	switch v.Version {
	case VaultV1:
		if err = c.getPersonaFromVaultV1(v, personaName, password); err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid vault version")
	}
	return nil
}

func (c *Client) getPersonaFromVaultV1(
	v *Vault,
	personaName string,
	password []byte,
) error {
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

	if err = c.resolvePersona(personas, personaName); err != nil {
		return err
	}
	// generate the ID and master key from the retrieved persona
	c.PersonaID = crypto.DerivePersonaID(
		c.ActivePersona.PrivateKey.Public().(ed25519.PublicKey),
		c.ActivePersona.Name,
	)
	c.MasterKey = crypto.DeriveMasterKey(
		c.ActivePersona.PrivateKey,
		c.PersonaID,
	)
	return nil
}

func (c *Client) resolvePersona(personas []Persona, personaName string) error {
	parts := strings.Split(personaName, "@")
	if len(parts) > 2 {
		return fmt.Errorf("Persona name invalid. Usage: <name>[@<domain>]")
	}
	name := parts[0]
	domain := ""
	if len(parts) == 2 {
		domain = parts[1]
	}

	var matches []*Persona
	for i := range personas {
		if domain != "" && domain == personas[i].BaseURL && name == personas[i].Name {
			matches = append(matches, &personas[i])
			break
		}
		if domain == "" && name == personas[i].Name {
			matches = append(matches, &personas[i])
			continue
		}
	}

	switch len(matches) {
	case 0:
		return fmt.Errorf("Persona '%s' not found", name)
	case 1:
		c.ActivePersona = &Persona{}
		c.ActivePersona.Name = matches[0].Name
		c.ActivePersona.BaseURL = matches[0].BaseURL
		c.ActivePersona.PrivateKey = make([]byte, len(matches[0].PrivateKey))
		copy(c.ActivePersona.PrivateKey, matches[0].PrivateKey)
		return nil
	default: // matches more than 1
		var domains []string
		for _, p := range matches {
			domains = append(domains, p.BaseURL)
		}
		return fmt.Errorf(
			"Use '%s@<domain>' - multiple domains found: %q",
			personaName,
			domains,
		)
	}
}

func (c *Client) AddPersonaToVault(
	personaName, baseURL string,
	password []byte,
) error {
	v, err := loadVault()
	if err != nil {
		return err
	}

	switch v.Version {
	case VaultV1:
		if err = c.addPersonaToVaultV1(v, personaName, baseURL, password); err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid vault version")
	}
	return nil
}

func (c *Client) addPersonaToVaultV1(
	v *Vault,
	personaName, baseURL string,
	password []byte,
) error {
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
	for _, p := range personas {
		if p.BaseURL == baseURL {
			persona.Name = personaName
			persona.BaseURL = baseURL
			persona.PrivateKey = make([]byte, len(p.PrivateKey))
			copy(persona.PrivateKey, p.PrivateKey)
		}
	}
	if persona.PrivateKey == nil {
		persona, err = buildPersona(personaName, baseURL)
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

	return saveVault(v)
}

func getVaultFilepath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Could not get homedir: %w", err)
	}
	return filepath.Join(homedir, vaultFilename), nil
}

func CreateNewVault(personaName, baseURL string, password []byte) error {
	vaultPath, err := getVaultFilepath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(vaultPath); err == nil {
		return fmt.Errorf("Vault already exists")
	}
	var v Vault
	var vaultKey []byte

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

	persona, err := buildPersona(personaName, baseURL)
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

	return saveVault(&v)
}

func buildPersona(personaName, baseURL string) (Persona, error) {
	persona := Persona{
		Name:    personaName,
		BaseURL: baseURL,
	}
	_, privateKey, err := crypto.GenerateAsymPrivateKey()
	if err != nil {
		return Persona{}, fmt.Errorf("Error generating private key: %w", err)
	}
	persona.PrivateKey = privateKey
	return persona, nil
}

func saveVault(v *Vault) error {
	vaultPath, err := getVaultFilepath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	// 0600 ensures only the server process can read/write this file
	return os.WriteFile(vaultPath, data, 0600)
}

func loadVault() (vault *Vault, err error) {
	vaultPath, err := getVaultFilepath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(vaultPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("Vault not found, Run 'ufos init' to create a vault.")
	}
	if err != nil {
		return nil, fmt.Errorf("Could not read vault file: %w", err)
	}
	if err := json.Unmarshal(data, &vault); err != nil {
		return nil, fmt.Errorf("Error unmarshalling vault data: %w", err)
	}
	return vault, nil
}
