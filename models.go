package cashbill

type Payment struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	Status          string       `json:"status"`
	PaymentChannel  string       `json:"paymentChannel"`
	Description     string       `json:"description"`
	AdditionalData  string       `json:"additionalData"`
	Amount          Amount       `json:"amount"`
	RequestedAmount Amount       `json:"requestedAmount"`
	PersonalData    PersonalData `json:"personalData"`
}

type Amount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type PersonalData struct {
	FirstName string `json:"firstName"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
	City      string `json:"city"`
	House     string `json:"house"`
	Flat      string `json:"flat"`
	Street    string `json:"street"`
	Postcode  string `json:"postcode"`
	Country   string `json:"country"`
}
