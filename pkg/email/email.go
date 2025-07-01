package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

type Interface interface {
	SendEmail(to []string, subject, body string, isHTML bool) error
	SendEmailWithTemplate(to []string, subject, templateName string, data interface{}) error
	SendWelcomeEmail(to, name string) error
	SendOTPEmail(to, name, otp string) error
	SendPasswordResetEmail(to, name, resetLink string) error
	SendSubscriptionConfirmationEmail(to, name string, subscription *SubscriptionDetails) error
	SendSubscriptionCancellationEmail(to, name string) error
	SendOrderConfirmationEmail(to, name string, order *OrderDetails) error
	SendPaymentConfirmationEmail(to, name string, payment *PaymentDetails) error
	TestConnection() error
}

type Service struct {
	config    *Config
	dialer    *gomail.Dialer
	templates map[string]*template.Template
}

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromName     string
	FromEmail    string
	UseTLS       bool
	UseSSL       bool
	Timeout      time.Duration
}

type SubscriptionDetails struct {
	PlanName     string
	MealTypes    []string
	DeliveryDays []string
	TotalPrice   float64
	StartDate    time.Time
	NextDelivery time.Time
}

type OrderDetails struct {
	OrderID      string
	Items        []OrderItem
	TotalAmount  float64
	DeliveryDate time.Time
	DeliveryTime string
	Address      string
}

type OrderItem struct {
	Name     string
	Quantity int
	Price    float64
}

type PaymentDetails struct {
	PaymentID     string
	Amount        float64
	Method        string
	TransactionID string
	Date          time.Time
}

func LoadConfig() *Config {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}

	port := 587
	if envPort := os.Getenv("SMTP_PORT"); envPort != "" {
		if parsed, err := strconv.Atoi(envPort); err == nil {
			port = parsed
		}
	}

	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "SEA Catering"
	}

	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@seacatering.com"
	}

	return &Config{
		SMTPHost:     host,
		SMTPPort:     port,
		SMTPUsername: username,
		SMTPPassword: password,
		FromName:     fromName,
		FromEmail:    fromEmail,
		UseTLS:       true,
		UseSSL:       false,
		Timeout:      30 * time.Second,
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

	dialer := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)
	dialer.TLSConfig = nil
	dialer.SSL = config.UseSSL

	service := &Service{
		config:    config,
		dialer:    dialer,
		templates: make(map[string]*template.Template),
	}

	service.loadTemplates()

	return service
}

func (s *Service) SendEmail(to []string, subject, body string, isHTML bool) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	return s.dialer.DialAndSend(m)
}

func (s *Service) SendEmailWithTemplate(to []string, subject, templateName string, data interface{}) error {
	tmpl, exists := s.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String(), true)
}

func (s *Service) SendWelcomeEmail(to, name string) error {
	data := struct {
		Name string
		Year int
	}{
		Name: name,
		Year: time.Now().Year(),
	}

	subject := "Welcome to SEA Catering!"
	return s.SendEmailWithTemplate([]string{to}, subject, "welcome", data)
}

func (s *Service) SendOTPEmail(to, name, otp string) error {
	data := struct {
		Name string
		OTP  string
		Year int
	}{
		Name: name,
		OTP:  otp,
		Year: time.Now().Year(),
	}

	subject := "Your OTP Verification Code"
	return s.SendEmailWithTemplate([]string{to}, subject, "otp", data)
}

func (s *Service) SendPasswordResetEmail(to, name, resetLink string) error {
	data := struct {
		Name      string
		ResetLink string
		Year      int
	}{
		Name:      name,
		ResetLink: resetLink,
		Year:      time.Now().Year(),
	}

	subject := "Reset Your Password"
	return s.SendEmailWithTemplate([]string{to}, subject, "password_reset", data)
}

func (s *Service) SendSubscriptionConfirmationEmail(to, name string, subscription *SubscriptionDetails) error {
	data := struct {
		Name         string
		Subscription *SubscriptionDetails
		Year         int
	}{
		Name:         name,
		Subscription: subscription,
		Year:         time.Now().Year(),
	}

	subject := "Subscription Confirmed - Welcome to SEA Catering!"
	return s.SendEmailWithTemplate([]string{to}, subject, "subscription_confirmation", data)
}

func (s *Service) SendSubscriptionCancellationEmail(to, name string) error {
	data := struct {
		Name string
		Year int
	}{
		Name: name,
		Year: time.Now().Year(),
	}

	subject := "Subscription Cancelled"
	return s.SendEmailWithTemplate([]string{to}, subject, "subscription_cancellation", data)
}

func (s *Service) SendOrderConfirmationEmail(to, name string, order *OrderDetails) error {
	data := struct {
		Name  string
		Order *OrderDetails
		Year  int
	}{
		Name:  name,
		Order: order,
		Year:  time.Now().Year(),
	}

	subject := fmt.Sprintf("Order Confirmation - #%s", order.OrderID)
	return s.SendEmailWithTemplate([]string{to}, subject, "order_confirmation", data)
}

func (s *Service) SendPaymentConfirmationEmail(to, name string, payment *PaymentDetails) error {
	data := struct {
		Name    string
		Payment *PaymentDetails
		Year    int
	}{
		Name:    name,
		Payment: payment,
		Year:    time.Now().Year(),
	}

	subject := "Payment Confirmation"
	return s.SendEmailWithTemplate([]string{to}, subject, "payment_confirmation", data)
}

func (s *Service) TestConnection() error {
	return s.dialer.DialAndSend()
}

func (s *Service) loadTemplates() {

	s.templates["welcome"] = template.Must(template.New("welcome").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to SEA Catering</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Welcome to SEA Catering, {{.Name}}!</h1>
        <p>Thank you for joining SEA Catering! We're excited to help you maintain a healthy lifestyle with our customizable meal plans.</p>
        <p>Our mission is to provide you with delicious, nutritious meals delivered right to your door, anywhere in Indonesia.</p>
        <h3>What's Next?</h3>
        <ul>
            <li>Browse our meal plans and find the perfect fit for your lifestyle</li>
            <li>Customize your meals based on your preferences and dietary needs</li>
            <li>Set up your delivery schedule</li>
            <li>Enjoy healthy, delicious meals!</li>
        </ul>
        <p>If you have any questions, feel free to contact our support team.</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))

	s.templates["otp"] = template.Must(template.New("otp").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>OTP Verification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Verification Code</h1>
        <p>Hello {{.Name}},</p>
        <p>Your verification code is:</p>
        <div style="background: #f4f4f4; padding: 20px; text-align: center; margin: 20px 0;">
            <h2 style="color: #2c5530; font-size: 32px; margin: 0; letter-spacing: 5px;">{{.OTP}}</h2>
        </div>
        <p>This code will expire in 10 minutes. If you didn't request this code, please ignore this email.</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))

	s.templates["password_reset"] = template.Must(template.New("password_reset").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Reset Your Password</h1>
        <p>Hello {{.Name}},</p>
        <p>We received a request to reset your password. Click the button below to reset it:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetLink}}" style="background: #2c5530; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
        </div>
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">{{.ResetLink}}</p>
        <p>This link will expire in 1 hour. If you didn't request a password reset, please ignore this email.</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))

	s.templates["subscription_confirmation"] = template.Must(template.New("subscription_confirmation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Subscription Confirmed</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Subscription Confirmed!</h1>
        <p>Hello {{.Name}},</p>
        <p>Great news! Your subscription has been confirmed. Here are the details:</p>
        <div style="background: #f9f9f9; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Subscription Details</h3>
            <p><strong>Plan:</strong> {{.Subscription.PlanName}}</p>
            <p><strong>Meal Types:</strong> {{range $index, $meal := .Subscription.MealTypes}}{{if $index}}, {{end}}{{$meal}}{{end}}</p>
            <p><strong>Delivery Days:</strong> {{range $index, $day := .Subscription.DeliveryDays}}{{if $index}}, {{end}}{{$day}}{{end}}</p>
            <p><strong>Monthly Total:</strong> Rp{{printf "%.2f" .Subscription.TotalPrice}}</p>
            <p><strong>Start Date:</strong> {{.Subscription.StartDate.Format "January 2, 2006"}}</p>
            <p><strong>Next Delivery:</strong> {{.Subscription.NextDelivery.Format "January 2, 2006"}}</p>
        </div>
        <p>We'll send you a reminder before each delivery. You can manage your subscription anytime through your account dashboard.</p>
        <p>Thank you for choosing SEA Catering!</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))

	s.templates["subscription_cancellation"] = template.Must(template.New("subscription_cancellation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Subscription Cancelled</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Subscription Cancelled</h1>
        <p>Hello {{.Name}},</p>
        <p>We're sorry to see you go! Your subscription has been successfully cancelled.</p>
        <p>You will continue to receive deliveries until the end of your current billing period. After that, no further charges will be made.</p>
        <p>We'd love to have you back anytime! If you have any feedback about your experience, please don't hesitate to reach out.</p>
        <p>Thank you for being part of the SEA Catering family.</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))

	s.templates["order_confirmation"] = template.Must(template.New("order_confirmation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Order Confirmation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Order Confirmation</h1>
        <p>Hello {{.Name}},</p>
        <p>Thank you for your order! Here are the details:</p>
        <div style="background: #f9f9f9; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Order #{{.Order.OrderID}}</h3>
            <table style="width: 100%; border-collapse: collapse;">
                <thead>
                    <tr>
                        <th style="text-align: left; padding: 8px; border-bottom: 1px solid #ddd;">Item</th>
                        <th style="text-align: center; padding: 8px; border-bottom: 1px solid #ddd;">Qty</th>
                        <th style="text-align: right; padding: 8px; border-bottom: 1px solid #ddd;">Price</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Order.Items}}
                    <tr>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">{{.Name}}</td>
                        <td style="text-align: center; padding: 8px; border-bottom: 1px solid #eee;">{{.Quantity}}</td>
                        <td style="text-align: right; padding: 8px; border-bottom: 1px solid #eee;">Rp{{printf "%.2f" .Price}}</td>
                    </tr>
                    {{end}}
                </tbody>
                <tfoot>
                    <tr>
                        <th colspan="2" style="text-align: right; padding: 8px; border-top: 2px solid #333;">Total:</th>
                        <th style="text-align: right; padding: 8px; border-top: 2px solid #333;">Rp{{printf "%.2f" .Order.TotalAmount}}</th>
                    </tr>
                </tfoot>
            </table>
            <p><strong>Delivery Date:</strong> {{.Order.DeliveryDate.Format "January 2, 2006"}}</p>
            <p><strong>Delivery Time:</strong> {{.Order.DeliveryTime}}</p>
            <p><strong>Delivery Address:</strong> {{.Order.Address}}</p>
        </div>
        <p>We'll notify you when your order is on its way!</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))

	s.templates["payment_confirmation"] = template.Must(template.New("payment_confirmation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Payment Confirmation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5530;">Payment Confirmed</h1>
        <p>Hello {{.Name}},</p>
        <p>Your payment has been successfully processed. Here are the details:</p>
        <div style="background: #f9f9f9; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Payment Details</h3>
            <p><strong>Payment ID:</strong> {{.Payment.PaymentID}}</p>
            <p><strong>Amount:</strong> Rp{{printf "%.2f" .Payment.Amount}}</p>
            <p><strong>Payment Method:</strong> {{.Payment.Method}}</p>
            <p><strong>Transaction ID:</strong> {{.Payment.TransactionID}}</p>
            <p><strong>Date:</strong> {{.Payment.Date.Format "January 2, 2006 at 3:04 PM"}}</p>
        </div>
        <p>Thank you for your payment! A receipt has been generated for your records.</p>
        <p>Best regards,<br>The SEA Catering Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} SEA Catering. All rights reserved.</p>
    </div>
</body>
</html>
	`))
}

func FormatPrice(amount float64) string {
	return fmt.Sprintf("Rp%.2f", amount)
}

func FormatMealTypes(mealTypes []string) string {
	return strings.Join(mealTypes, ", ")
}

func FormatDeliveryDays(days []string) string {
	return strings.Join(days, ", ")
}
