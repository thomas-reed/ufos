package contacts

type ContactMetadata struct {
	PersonaID string    `json:"persona_id"`
	Domain    string    `json:"domain"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Company   string    `json:"company"`
	Phones    []Phone   `json:"phones"`
	Addresses []Address `json:"addresses"`
	Notes     string    `json:"notes"`
}

type Address struct {
	Address1    string `json:"address1"`     // address line 1
	Address2    string `json:"address2"`     // address line 2
	City        string `json:"city"`         // city
	StateRegion string `json:"state_region"` // state/region
	Zip         string `json:"zip"`          // Zip/Postal code
	Country     string `json:"country"`      // country
}

type Phone struct {
	PhoneType string `json:"phone_type"`
	Number    string `json:"number"`
}
