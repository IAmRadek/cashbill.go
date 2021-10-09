package cashbill

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	prodURL = "https://pay.cashbill.pl/ws/rest"
	testURL = "https://pay.cashbill.pl/testws/rest"
)

type API interface {
	RequestPayment(ctx context.Context, newPayment NewPayment) (PaymentRequest, error)
	GetPayment(ctx context.Context, paymentID string) (Payment, error)
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
	h := sha1.New()
	h.Write([]byte(n.Title))
	h.Write([]byte(n.Amount))
	h.Write([]byte(n.Currency))
	h.Write([]byte(n.ReturnURL))
	h.Write([]byte(n.Description))
	h.Write([]byte(n.NegativeReturnURL))
	h.Write([]byte(n.AdditionalData))
	h.Write([]byte(n.PaymentChannel))
	h.Write([]byte(n.LanguageCode))
	h.Write([]byte(n.Referer))
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum(nil))
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

	fmt.Println(postForm.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.URL+"/payment/"+api.ShopID, bytes.NewBufferString(postForm.Encode()))
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

func (api *api) GetPayment(ctx context.Context, paymentID string) (Payment, error) {
	h := sha1.New()
	h.Write([]byte(paymentID))
	h.Write([]byte(api.Secret))
	sign := hex.EncodeToString(h.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.URL+"/payment/"+api.ShopID+"/"+paymentID+"?sign="+sign, nil)
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
