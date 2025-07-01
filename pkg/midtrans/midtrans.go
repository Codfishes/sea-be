package midtrans

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type Interface interface {
	CreateSnapTransaction(req *SnapRequest) (*SnapResponse, error)

	ChargeTransaction(req *ChargeRequest) (*ChargeResponse, error)
	CheckTransactionStatus(orderID string) (*TransactionStatusResponse, error)
	CancelTransaction(orderID string) (*TransactionStatusResponse, error)
	ApproveTransaction(orderID string) (*TransactionStatusResponse, error)
	ExpireTransaction(orderID string) (*TransactionStatusResponse, error)
	RefundTransaction(orderID string, amount int64, reason string) (*RefundResponse, error)

	ValidateSignature(orderID, statusCode, grossAmount, serverKey string) string
	HandleNotification(notification *NotificationPayload) (*TransactionStatusResponse, error)

	IsProduction() bool
	GetServerKey() string
	GetClientKey() string
}

type Service struct {
	snapClient *snap.Client
	coreClient *coreapi.Client
	config     *Config
}

type Config struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
	Environment  midtrans.EnvironmentType
}

type SnapRequest struct {
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CustomerDetails    *CustomerDetails   `json:"customer_details,omitempty"`
	ItemDetails        []ItemDetail       `json:"item_details,omitempty"`
	Callbacks          *Callbacks         `json:"callbacks,omitempty"`
	Expiry             *Expiry            `json:"expiry,omitempty"`
	CustomField1       string             `json:"custom_field1,omitempty"`
	CustomField2       string             `json:"custom_field2,omitempty"`
	CustomField3       string             `json:"custom_field3,omitempty"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type CustomerDetails struct {
	FirstName       string   `json:"first_name,omitempty"`
	LastName        string   `json:"last_name,omitempty"`
	Email           string   `json:"email,omitempty"`
	Phone           string   `json:"phone,omitempty"`
	BillingAddress  *Address `json:"billing_address,omitempty"`
	ShippingAddress *Address `json:"shipping_address,omitempty"`
}

type Address struct {
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

type ItemDetail struct {
	ID       string `json:"id"`
	Price    int64  `json:"price"`
	Quantity int32  `json:"quantity"`
	Name     string `json:"name"`
	Brand    string `json:"brand,omitempty"`
	Category string `json:"category,omitempty"`
}

type Callbacks struct {
	Finish  string `json:"finish,omitempty"`
	Error   string `json:"error,omitempty"`
	Pending string `json:"pending,omitempty"`
}

type Expiry struct {
	StartTime string `json:"start_time,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Duration  int    `json:"duration,omitempty"`
}

type SnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type ChargeRequest struct {
	PaymentType        string             `json:"payment_type"`
	TransactionDetails TransactionDetails `json:"transaction_details"`
	CustomerDetails    *CustomerDetails   `json:"customer_details,omitempty"`
	ItemDetails        []ItemDetail       `json:"item_details,omitempty"`
	BankTransfer       *BankTransfer      `json:"bank_transfer,omitempty"`
	CreditCard         *CreditCard        `json:"credit_card,omitempty"`
	Gopay              *Gopay             `json:"gopay,omitempty"`
	ShopeePay          *ShopeePay         `json:"shopeepay,omitempty"`
}

type BankTransfer struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number,omitempty"`
}

type CreditCard struct {
	TokenID     string       `json:"token_id,omitempty"`
	Bank        string       `json:"bank,omitempty"`
	Installment *Installment `json:"installment,omitempty"`
	SaveTokenID bool         `json:"save_token_id,omitempty"`
}

type Installment struct {
	Required bool             `json:"required"`
	Terms    map[string][]int `json:"terms,omitempty"`
}

type Gopay struct {
	EnableCallback bool   `json:"enable_callback,omitempty"`
	CallbackURL    string `json:"callback_url,omitempty"`
}

type ShopeePay struct {
	CallbackURL string `json:"callback_url,omitempty"`
}

type ChargeResponse struct {
	StatusCode        string     `json:"status_code"`
	StatusMessage     string     `json:"status_message"`
	TransactionID     string     `json:"transaction_id"`
	OrderID           string     `json:"order_id"`
	MerchantID        string     `json:"merchant_id"`
	GrossAmount       string     `json:"gross_amount"`
	Currency          string     `json:"currency"`
	PaymentType       string     `json:"payment_type"`
	TransactionTime   string     `json:"transaction_time"`
	TransactionStatus string     `json:"transaction_status"`
	FraudStatus       string     `json:"fraud_status,omitempty"`
	VANumbers         []VANumber `json:"va_numbers,omitempty"`
	Actions           []Action   `json:"actions,omitempty"`
}

type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number"`
}

type Action struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type TransactionStatusResponse struct {
	StatusCode             string     `json:"status_code"`
	StatusMessage          string     `json:"status_message"`
	TransactionID          string     `json:"transaction_id"`
	MaskedCard             string     `json:"masked_card,omitempty"`
	OrderID                string     `json:"order_id"`
	PaymentType            string     `json:"payment_type"`
	TransactionTime        string     `json:"transaction_time"`
	TransactionStatus      string     `json:"transaction_status"`
	FraudStatus            string     `json:"fraud_status,omitempty"`
	ApprovalCode           string     `json:"approval_code,omitempty"`
	SignatureKey           string     `json:"signature_key,omitempty"`
	Bank                   string     `json:"bank,omitempty"`
	GrossAmount            string     `json:"gross_amount"`
	ChannelResponseCode    string     `json:"channel_response_code,omitempty"`
	ChannelResponseMessage string     `json:"channel_response_message,omitempty"`
	CardType               string     `json:"card_type,omitempty"`
	PaymentOption          string     `json:"payment_option,omitempty"`
	VANumbers              []VANumber `json:"va_numbers,omitempty"`
	BillerCode             string     `json:"biller_code,omitempty"`
	BillKey                string     `json:"bill_key,omitempty"`
}

type RefundResponse struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message"`
	TransactionID string `json:"transaction_id"`
	OrderID       string `json:"order_id"`
	RefundAmount  string `json:"refund_amount"`
	RefundTime    string `json:"refund_time"`
	RefundKey     string `json:"refund_key"`
}

type NotificationPayload struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
	ApprovalCode      string `json:"approval_code"`
	CardType          string `json:"card_type"`
	Bank              string `json:"bank"`
	PaymentOption     string `json:"payment_option"`
}

func LoadConfig() *Config {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	clientKey := os.Getenv("MIDTRANS_CLIENT_KEY")

	isProduction := false
	if envProd := os.Getenv("MIDTRANS_IS_PRODUCTION"); envProd != "" {
		if parsed, err := strconv.ParseBool(envProd); err == nil {
			isProduction = parsed
		}
	}

	var environment midtrans.EnvironmentType
	if isProduction {
		environment = midtrans.Production
	} else {
		environment = midtrans.Sandbox
	}

	return &Config{
		ServerKey:    serverKey,
		ClientKey:    clientKey,
		IsProduction: isProduction,
		Environment:  environment,
	}
}

func New() Interface {
	config := LoadConfig()
	return NewWithConfig(config)
}

func NewWithConfig(config *Config) Interface {
	if config == nil {
		config = LoadConfig()
	}

	snapClient := snap.Client{}
	snapClient.New(config.ServerKey, config.Environment)

	coreClient := coreapi.Client{}
	coreClient.New(config.ServerKey, config.Environment)

	return &Service{
		snapClient: &snapClient,
		coreClient: &coreClient,
		config:     config,
	}
}

func (s *Service) CreateSnapTransaction(req *SnapRequest) (*SnapResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	if req.TransactionDetails.OrderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	if req.TransactionDetails.GrossAmount <= 0 {
		return nil, fmt.Errorf("gross amount must be positive")
	}

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.TransactionDetails.OrderID,
			GrossAmt: req.TransactionDetails.GrossAmount,
		},
	}

	if req.CustomerDetails != nil {
		snapReq.CustomerDetail = &midtrans.CustomerDetails{
			FName: req.CustomerDetails.FirstName,
			LName: req.CustomerDetails.LastName,
			Email: req.CustomerDetails.Email,
			Phone: req.CustomerDetails.Phone,
		}

		if req.CustomerDetails.BillingAddress != nil {
			snapReq.CustomerDetail.BillAddr = &midtrans.CustomerAddress{
				FName: req.CustomerDetails.BillingAddress.FirstName,
				LName: req.CustomerDetails.BillingAddress.LastName,

				Phone:       req.CustomerDetails.BillingAddress.Phone,
				Address:     req.CustomerDetails.BillingAddress.Address,
				City:        req.CustomerDetails.BillingAddress.City,
				Postcode:    req.CustomerDetails.BillingAddress.PostalCode,
				CountryCode: req.CustomerDetails.BillingAddress.CountryCode,
			}
		}

		if req.CustomerDetails.ShippingAddress != nil {
			snapReq.CustomerDetail.ShipAddr = &midtrans.CustomerAddress{
				FName: req.CustomerDetails.ShippingAddress.FirstName,
				LName: req.CustomerDetails.ShippingAddress.LastName,

				Phone:       req.CustomerDetails.ShippingAddress.Phone,
				Address:     req.CustomerDetails.ShippingAddress.Address,
				City:        req.CustomerDetails.ShippingAddress.City,
				Postcode:    req.CustomerDetails.ShippingAddress.PostalCode,
				CountryCode: req.CustomerDetails.ShippingAddress.CountryCode,
			}
		}
	}

	if len(req.ItemDetails) > 0 {
		items := make([]midtrans.ItemDetails, len(req.ItemDetails))
		for i, item := range req.ItemDetails {
			items[i] = midtrans.ItemDetails{
				ID:       item.ID,
				Price:    item.Price,
				Qty:      item.Quantity,
				Name:     item.Name,
				Brand:    item.Brand,
				Category: item.Category,
			}
		}
		snapReq.Items = &items
	}

	resp, err := s.snapClient.CreateTransaction(snapReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return &SnapResponse{
		Token:       resp.Token,
		RedirectURL: resp.RedirectURL,
	}, nil
}

func (s *Service) ChargeTransaction(req *ChargeRequest) (*ChargeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := s.getAPIURL() + "/charge"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+s.getAuthHeader())

	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var chargeResp ChargeResponse
	if err := json.Unmarshal(respBody, &chargeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chargeResp, nil
}

func (s *Service) CheckTransactionStatus(orderID string) (*TransactionStatusResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	resp, err := s.coreClient.CheckTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check transaction status: %w", err)
	}

	return s.convertToTransactionStatusResponse(resp), nil
}

func (s *Service) CancelTransaction(orderID string) (*TransactionStatusResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	resp, err := s.coreClient.CancelTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel transaction: %w", err)
	}

	return &TransactionStatusResponse{
		StatusCode:        resp.StatusCode,
		StatusMessage:     resp.StatusMessage,
		TransactionID:     resp.TransactionID,
		OrderID:           resp.OrderID,
		PaymentType:       resp.PaymentType,
		TransactionTime:   resp.TransactionTime,
		TransactionStatus: resp.TransactionStatus,
		FraudStatus:       resp.FraudStatus,
		GrossAmount:       resp.GrossAmount,
	}, nil
}

func (s *Service) ApproveTransaction(orderID string) (*TransactionStatusResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	resp, err := s.coreClient.ApproveTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to approve transaction: %w", err)
	}

	return &TransactionStatusResponse{
		StatusCode:        resp.StatusCode,
		StatusMessage:     resp.StatusMessage,
		TransactionID:     resp.TransactionID,
		OrderID:           resp.OrderID,
		PaymentType:       resp.PaymentType,
		TransactionTime:   resp.TransactionTime,
		TransactionStatus: resp.TransactionStatus,
		FraudStatus:       resp.FraudStatus,
		GrossAmount:       resp.GrossAmount,
	}, nil
}

func (s *Service) ExpireTransaction(orderID string) (*TransactionStatusResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	resp, err := s.coreClient.ExpireTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to expire transaction: %w", err)
	}

	return &TransactionStatusResponse{
		StatusCode:        resp.StatusCode,
		StatusMessage:     resp.StatusMessage,
		TransactionID:     resp.TransactionID,
		OrderID:           resp.OrderID,
		PaymentType:       resp.PaymentType,
		TransactionTime:   resp.TransactionTime,
		TransactionStatus: resp.TransactionStatus,
		FraudStatus:       resp.FraudStatus,
		GrossAmount:       resp.GrossAmount,
	}, nil
}

func (s *Service) RefundTransaction(orderID string, amount int64, reason string) (*RefundResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	refundReq := &coreapi.RefundReq{
		Amount: amount,
		Reason: reason,
	}

	resp, err := s.coreClient.RefundTransaction(orderID, refundReq)
	if err != nil {
		return nil, fmt.Errorf("failed to refund transaction: %w", err)
	}

	return &RefundResponse{
		StatusCode:    resp.StatusCode,
		StatusMessage: resp.StatusMessage,
		TransactionID: resp.TransactionID,
		OrderID:       resp.OrderID,
		RefundAmount:  resp.RefundAmount,

		RefundKey: resp.RefundKey,
	}, nil
}

func (s *Service) ValidateSignature(orderID, statusCode, grossAmount, serverKey string) string {
	input := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func (s *Service) HandleNotification(notification *NotificationPayload) (*TransactionStatusResponse, error) {
	if notification == nil {
		return nil, fmt.Errorf("notification is nil")
	}

	expectedSignature := s.ValidateSignature(
		notification.OrderID,
		notification.StatusCode,
		notification.GrossAmount,
		s.config.ServerKey,
	)

	if notification.SignatureKey != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	return s.CheckTransactionStatus(notification.OrderID)
}

func (s *Service) IsProduction() bool {
	return s.config.IsProduction
}

func (s *Service) GetServerKey() string {
	return s.config.ServerKey
}

func (s *Service) GetClientKey() string {
	return s.config.ClientKey
}

func (s *Service) getAPIURL() string {
	if s.config.IsProduction {
		return "https://api.midtrans.com/v2"
	}
	return "https://api.sandbox.midtrans.com/v2"
}

func (s *Service) getAuthHeader() string {
	return fmt.Sprintf("%s:", s.config.ServerKey)
}

func (s *Service) convertToTransactionStatusResponse(resp *coreapi.TransactionStatusResponse) *TransactionStatusResponse {
	var vaNumbers []VANumber

	return &TransactionStatusResponse{
		StatusCode:             resp.StatusCode,
		StatusMessage:          resp.StatusMessage,
		TransactionID:          resp.TransactionID,
		MaskedCard:             resp.MaskedCard,
		OrderID:                resp.OrderID,
		PaymentType:            resp.PaymentType,
		TransactionTime:        resp.TransactionTime,
		TransactionStatus:      resp.TransactionStatus,
		FraudStatus:            resp.FraudStatus,
		ApprovalCode:           resp.ApprovalCode,
		SignatureKey:           resp.SignatureKey,
		GrossAmount:            resp.GrossAmount,
		ChannelResponseCode:    resp.ChannelResponseCode,
		ChannelResponseMessage: resp.ChannelResponseMessage,
		CardType:               resp.CardType,
		VANumbers:              vaNumbers,
		BillerCode:             resp.BillerCode,
		BillKey:                resp.BillKey,
	}
}

const (
	PaymentTypeCreditCard      = "credit_card"
	PaymentTypeBankTransfer    = "bank_transfer"
	PaymentTypeEchannel        = "echannel"
	PaymentTypeBCAKlikPay      = "bca_klikpay"
	PaymentTypeBCAKlikBCA      = "bca_klikbca"
	PaymentTypeMandiriClickPay = "mandiri_clickpay"
	PaymentTypeCimbClicks      = "cimb_clicks"
	PaymentTypeDanamonOnline   = "danamon_online"
	PaymentTypeBRIEpay         = "bri_epay"
	PaymentTypeIndomaret       = "cstore"
	PaymentTypeAlfamart        = "cstore"
	PaymentTypeGopay           = "gopay"
	PaymentTypeShopeePay       = "shopeepay"
	PaymentTypeQris            = "qris"
)

const (
	TransactionStatusCapture       = "capture"
	TransactionStatusSettlement    = "settlement"
	TransactionStatusPending       = "pending"
	TransactionStatusDeny          = "deny"
	TransactionStatusCancel        = "cancel"
	TransactionStatusExpire        = "expire"
	TransactionStatusFailure       = "failure"
	TransactionStatusRefund        = "refund"
	TransactionStatusPartialRefund = "partial_refund"
)

const (
	FraudStatusAccept    = "accept"
	FraudStatusDeny      = "deny"
	FraudStatusChallenge = "challenge"
)
