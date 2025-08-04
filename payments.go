package cashbill

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/subtle"
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

func NewAPI(shopID, secret string, opts ...Option) *Client {
	a := &Client{
		url:    prodURL,
		shopID: shopID,
		secret: secret,
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func NewTestAPI(shopID, secret string, opts ...Option) *Client {
	a := &Client{
		url:    testURL,
		shopID: shopID,
		secret: secret,
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

type Client struct {
	url    string
	shopID string
	secret string
	client *http.Client
}

// Option is a function that configures the API
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client for the API
func WithHTTPClient(client *http.Client) Option {
	return func(a *Client) {
		a.client = client
	}
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
	ID          string `json:"id"`
	RedirectURL string `json:"redirectUrl"`
}

func (api *Client) RequestPayment(ctx context.Context, newPayment NewPayment) (PaymentRequest, error) {
	postForm := newPayment.valuesWithSign(api.secret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.url+"/payment/"+api.shopID, bytes.NewBufferString(postForm.Encode()))
	if err != nil {
		return PaymentRequest{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := api.client.Do(req)
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

func (api *Client) GetPayment(ctx context.Context, paymentID string) (Payment, error) {
	h := sha1.New()
	h.Write([]byte(paymentID))
	h.Write([]byte(api.secret))
	sign := hex.EncodeToString(h.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.url+"/payment/"+api.shopID+"/"+paymentID+"?sign="+sign, nil)
	if err != nil {
		return Payment{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := api.client.Do(req)
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

func (api *Client) VerifyCallback(r *http.Request) (Payment, error) {
	q := r.URL.Query()

	cmd := q.Get("cmd")
	arg := q.Get("args")
	sig := q.Get("sign")

	local := md5Hash(cmd + arg + api.secret)

	if subtle.ConstantTimeCompare(local, []byte(sig)) != 0 {
		return Payment{}, fmt.Errorf("invalid signature")
	}

	return api.GetPayment(r.Context(), arg)
}

func md5Hash(s string) []byte {
	hash := md5.New()
	return hash.Sum([]byte(s))
}
