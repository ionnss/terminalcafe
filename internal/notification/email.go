package notification

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"terminal-cafe/internal/models"
)

type EmailNotifier struct {
	from     string
	password string
	to       string
	smtpHost string
	smtpPort string
}

func NewEmailNotifier(from, password, to, smtpHost, smtpPort string) *EmailNotifier {
	return &EmailNotifier{
		from:     from,
		password: password,
		to:       to,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
	}
}

func (e *EmailNotifier) NotifyNewOrder(order *models.Order) error {
	subject := fmt.Sprintf("Novo Pedido #%d - Terminal Café", time.Now().Unix())

	body := fmt.Sprintf(`
Data: %s

DADOS DO CLIENTE
---------------
Email: %s
CPF: %s
Telefone: %s
Endereço: %s, %s
CEP: %s
%s%s
%s

ITENS DO PEDIDO
--------------
`,
		time.Now().Format("02/01/2006 15:04:05"),
		order.Customer.Email,
		order.Customer.CPF,
		order.Customer.Phone,
		order.Customer.Address,
		order.Customer.Number,
		order.Customer.CEP,
		func() string {
			if order.Customer.Type == "apartamento" {
				return fmt.Sprintf("Apartamento: %s\n", order.Customer.Unit)
			}
			return "Casa\n"
		}(),
		func() string {
			if order.Customer.Complement != "" {
				return fmt.Sprintf("Complemento: %s\n", order.Customer.Complement)
			}
			return ""
		}(),
		strings.Repeat("-", 40),
	)

	var total float64
	for _, item := range order.Items {
		subtotal := item.Product.Price * float64(item.Quantity)
		total += subtotal
		body += fmt.Sprintf("%dx %s (R$ %.2f cada) = R$ %.2f\n",
			item.Quantity,
			item.Product.Name,
			item.Product.Price,
			subtotal)
	}

	body += fmt.Sprintf("\n%s\nTOTAL: R$ %.2f\n", strings.Repeat("-", 40), total)

	msg := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", e.to, subject, body)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)
	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)

	if err := smtp.SendMail(addr, auth, e.from, []string{e.to}, []byte(msg)); err != nil {
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	return nil
}
