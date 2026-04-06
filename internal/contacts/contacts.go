package contacts

type ContactMetadata struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Nickname  string    `json:"nickname"`
	Company   string    `json:"company"`
	Phones    []Phone   `json:"phones"`
	Addresses []Address `json:"addresses"`
}

type Address struct {
	Address1 string `json:"address1"` // address line 1
	Address2 string `json:"address2"` // address line 2
	City     string `json:"city"`     // city
	State    string `json:"state"`    // state/region
	Zip      string `json:"zip"`      // Zip/Postal code
	Country  string `json:"country"`  // country
}

type Phone struct {
	PhoneType   PhoneType `json:"phone_type"`
	CountryCode string    `json:"country_code"`
	Number      string    `json:"number"`
}

type PhoneType int

const (
	Mobile PhoneType = iota
	Home
	Main
	Work
	School
	Other
)
