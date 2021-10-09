package cashbill

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	prodURL = "https://pay.cashbill.pl/ws/rest"
	testURL = "https://pay.cashbill.pl/testws/rest"
)

type API interface {
	RequestPayment(ctx context.Context, newPayment NewPayment) (PaymentRequest, error)
	GetPayment(ctx context.Context, get GetPayment) (Payment, error)
}

func NewAPI(shopID, secret string) API {
	return &api{prodURL, shopID, secret}
}
func NewTestAPI(shopID, secret string) API {
	return &api{testURL, shopID, secret}
}

type api struct {
	URL    string
	ShopID string
	Secret string
}

type NewPayment struct {
	Title             string `url:"title"`
	Amount            string `url:"amount.value"`
	Currency          string `url:"amount.currencyCode"`
	Description       string `url:"description"`
	AdditionalData    string `url:"additionalData"`
	ReturnURL         string `url:"returnUrl"`
	NegativeReturnURL string `url:"negativeReturnUrl"`
	PaymentChannel    string `url:"paymentChannel"`
	LanguageCode      string `url:"languageCode"`
	Referer           string `url:"referer"`
}

func (n NewPayment) sign(secret string) string {
	hasher := sha1.New()
	hasher.Write([]byte(n.Title))
	hasher.Write([]byte(n.Amount))
	hasher.Write([]byte(n.Currency))
	hasher.Write([]byte(n.ReturnURL))
	hasher.Write([]byte(n.Description))
	hasher.Write([]byte(n.NegativeReturnURL))
	hasher.Write([]byte(n.AdditionalData))
	hasher.Write([]byte(n.PaymentChannel))
	hasher.Write([]byte(n.LanguageCode))
	hasher.Write([]byte(n.Referer))
	return hex.EncodeToString(hasher.Sum([]byte(secret)))
}

func (n NewPayment) valuesWithSign(secret string) url.Values {
	data := url.Values{}

	data.Set("title", n.Title)
	data.Set("amount.value", n.Amount)
	data.Set("amount.currencyCode", n.Currency)
	data.Set("description", n.Description)
	data.Set("additionalData", n.AdditionalData)
	data.Set("returnUrl", n.ReturnURL)
	data.Set("negativeReturnUrl", n.NegativeReturnURL)
	data.Set("paymentChannel", n.PaymentChannel)
	data.Set("languageCode", n.LanguageCode)
	data.Set("referer", n.Referer)
	data.Set("sign", n.sign(secret))
	return data

}

type PaymentRequest struct {
	ID        string `json:"id"`
	ReturnURL string `json:"returnUrl"`
}

func (api *api) RequestPayment(ctx context.Context, newPayment NewPayment) (PaymentRequest, error) {
	postForm := newPayment.valuesWithSign(api.Secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.URL+"/payment/"+api.ShopID, strings.NewReader(postForm.Encode()))
	if err != nil {
		return PaymentRequest{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return PaymentRequest{}, fmt.Errorf("failed to call cashbill: %w", err)
	}

	defer resp.Body.Close()

	var payment PaymentRequest
	err = json.NewDecoder(resp.Body).Decode(&payment)
	if err != nil {
		return PaymentRequest{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return payment, nil
}

type GetPayment struct {
	PaymentID string
	Signature string
}

func (g GetPayment) sign(secret string) string {
	h := sha1.New()
	h.Write([]byte(g.PaymentID))
	return hex.EncodeToString(h.Sum([]byte(secret)))
}

func (api *api) GetPayment(ctx context.Context, get GetPayment) (Payment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.URL+"/payment/"+api.ShopID+"/"+get.PaymentID+"?sign="+get.sign(api.Secret), nil)
	if err != nil {
		return Payment{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Payment{}, fmt.Errorf("failed to call cashbill: %w", err)
	}

	defer resp.Body.Close()

	var payment Payment
	err = json.NewDecoder(resp.Body).Decode(&payment)
	if err != nil {
		return Payment{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return payment, nil
}
