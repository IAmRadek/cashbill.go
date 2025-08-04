package cashbill

type Status string

const (
	// PreStart means Payment has been started. Customer has not yet choosen payment channel.
	PreStart Status = "PreStart"
	// Start means Payment has been started. Customer has not yet paid.
	Start Status = "Start"
	// NegativeAuthorization means Payment Channel has refused payment (ie. insufficient funds).
	NegativeAuthorization Status = "NegativeAuthorization"
	// Abort means Customer has aborted the payment.
	Abort Status = "Abort"
	// Fraud means Payment Channel has refused payment and classified is as fraudant. This is a final status and cannot change.
	Fraud Status = "Fraud"
	// PositiveAuthorization means Payment Channel has accepted transaction for processing.
	PositiveAuthorization Status = "PositiveAuthorization"
	// PositiveFinish means Payment Channel has confirmed transfer of funds. This is a final status and cannot change.
	PositiveFinish Status = "PositiveFinish"
	// NegativeFinish means Payment Channel has refused transfer of funds. This is a final status and cannot change.
	NegativeFinish Status = "NegativeFinish"
)

type Payment struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	Status          Status       `json:"status"`
	PaymentChannel  string       `json:"paymentChannel"`
	Description     string       `json:"description"`
	AdditionalData  string       `json:"additionalData"`
	Amount          Amount       `json:"amount"`
	RequestedAmount Amount       `json:"requestedAmount"`
	PersonalData    PersonalData `json:"personalData"`
}

func (p Payment) IsPaid() bool {
	return p.Status == PositiveFinish
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
