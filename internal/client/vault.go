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
	Name               string `json:"name"`                 // user-provided descriptor
	BaseURL            string `json:"base_url"`             // URL of the server
	PrivateSigningKey  []byte `json:"private_signing_key"`  // Ed25519 private key (64bytes)
	PublicSigningKey   []byte `json:"public_signing_key"`   // Ed25519 public key
	PrivateExchangeKey []byte `json:"private_exchange_key"` // X25519 private key (32bytes)
	PublicExchangeKey  []byte `json:"public_exchange_key"`  // X25519 public key
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

var ErrPersonaNotFound = errors.New("Persona not found")

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
			clear(personas[i].PrivateSigningKey)
			clear(personas[i].PrivateExchangeKey)
		}
	}()

	if err = c.resolvePersona(personas, personaName); err != nil {
		return err
	}

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
		return ErrPersonaNotFound
	case 1:
		c.ActivePersona = &Persona{}
		c.ActivePersona.Name = matches[0].Name
		c.ActivePersona.BaseURL = matches[0].BaseURL
		c.ActivePersona.PrivateSigningKey = make([]byte, len(matches[0].PrivateSigningKey))
		copy(c.ActivePersona.PrivateSigningKey, matches[0].PrivateSigningKey)
		c.ActivePersona.PublicSigningKey = make([]byte, len(matches[0].PublicSigningKey))
		copy(c.ActivePersona.PublicSigningKey, matches[0].PublicSigningKey)
		c.ActivePersona.PrivateExchangeKey = make([]byte, len(matches[0].PrivateExchangeKey))
		copy(c.ActivePersona.PrivateExchangeKey, matches[0].PrivateExchangeKey)
		c.ActivePersona.PublicExchangeKey = make([]byte, len(matches[0].PublicExchangeKey))
		copy(c.ActivePersona.PublicExchangeKey, matches[0].PublicExchangeKey)
		// generate the ID and master key from the retrieved persona
		c.PersonaID = crypto.DerivePersonaID(
			ed25519.PublicKey(c.ActivePersona.PrivateSigningKey),
			c.ActivePersona.Name,
		)
		c.MasterKey = crypto.DeriveMasterKey(
			c.ActivePersona.PrivateSigningKey,
			c.PersonaID,
		)
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
			clear(personas[i].PrivateSigningKey)
			clear(personas[i].PrivateExchangeKey)
		}
	}()

	var persona Persona
	for _, p := range personas {
		if p.BaseURL == baseURL && p.Name == personaName {
			return fmt.Errorf("Persona '%s' already exists at domain '%s'", personaName, baseURL)
		}
		if p.BaseURL == baseURL {
			persona.Name = personaName
			persona.BaseURL = baseURL
			persona.PrivateSigningKey = make([]byte, len(p.PrivateSigningKey))
			copy(persona.PrivateSigningKey, p.PrivateSigningKey)
			persona.PublicSigningKey = make([]byte, len(p.PublicSigningKey))
			copy(persona.PublicSigningKey, p.PublicSigningKey)
			persona.PrivateExchangeKey = make([]byte, len(p.PrivateExchangeKey))
			copy(persona.PrivateExchangeKey, p.PrivateExchangeKey)
			persona.PublicExchangeKey = make([]byte, len(p.PublicExchangeKey))
			copy(persona.PublicExchangeKey, p.PublicExchangeKey)
			break
		}
	}
	if persona.PrivateSigningKey == nil && persona.PrivateExchangeKey == nil {
		persona, err = buildPersona(personaName, baseURL)
		if err != nil {
			return err
		}
	}
	defer clear(persona.PrivateSigningKey)
	defer clear(persona.PrivateExchangeKey)

	payload, err := json.Marshal(append(personas, persona))

	v.Payload, err = crypto.Encrypt(vaultKey, payload, crypto.CryptoSuiteV1)
	if err != nil {
		return fmt.Errorf("Error encrypting persona payload: %w", err)
	}

	return saveVault(v)
}

func (c *Client) RemovePersonaFromVault(personaName, baseURL string, password []byte) error {
	v, err := loadVault()
	if err != nil {
		return err
	}
	switch v.Version {
	case VaultV1:
		if err = c.removePersonaFromVaultV1(v, personaName, baseURL, password); err != nil {
			return err
		}

	default:
		return fmt.Errorf("Invalid vault version")
	}
	return nil
}

func (c *Client) removePersonaFromVaultV1(
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
			clear(personas[i].PrivateSigningKey)
			clear(personas[i].PrivateExchangeKey)
		}
	}()

	// Filter out the persona to remove
	newPersonas := make([]Persona, 0, len(personas))
	for i := range personas {
		defer clear(personas[i].PrivateSigningKey)
		defer clear(personas[i].PrivateExchangeKey)
		if personas[i].Name == personaName && personas[i].BaseURL == baseURL {
			continue
		}
		newPersonas = append(newPersonas, personas[i])
	}
	defer func() {
		for i := range newPersonas {
			clear(newPersonas[i].PrivateSigningKey)
			clear(newPersonas[i].PrivateExchangeKey)
		}
	}()

	payload, err := json.Marshal(newPersonas)

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
	defer clear(persona.PrivateSigningKey)
	defer clear(persona.PrivateExchangeKey)

	payload, err := json.Marshal([]Persona{persona})
	defer clear(payload)

	v.Payload, err = crypto.Encrypt(vaultKey, payload, crypto.CryptoSuiteV1)
	if err != nil {
		return fmt.Errorf("Error encrypting persona payload: %w", err)
	}

	return saveVault(&v)
}

func buildPersona(personaName, baseURL string) (Persona, error) {
	publicEd, privateEd, err := crypto.GenerateSigningKeyPair()
	if err != nil {
		return Persona{}, fmt.Errorf("Error generating private key: %w", err)
	}

	publicX, privateX, err := crypto.GenerateExchangeKeyPair()
	if err != nil {
		return Persona{}, fmt.Errorf("Error generating exchange key: %w", err)
	}

	return Persona{
		Name:               personaName,
		BaseURL:            baseURL,
		PrivateSigningKey:  privateEd,
		PublicSigningKey:   publicEd,
		PrivateExchangeKey: privateX,
		PublicExchangeKey:  publicX,
	}, nil
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
