package payment

import (
	"context"
	"fmt"
	"os"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"

	"terminal-cafe/internal/models"
)

type MercadoPagoProvider struct {
	client payment.Client
}

func NewMercadoPagoProvider() (*MercadoPagoProvider, error) {
	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("MP_ACCESS_TOKEN não configurado")
	}

	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("erro ao configurar cliente: %v", err)
	}

	client := payment.NewClient(cfg)
	return &MercadoPagoProvider{client: client}, nil
}

func (mp *MercadoPagoProvider) CreatePayment(order *models.Order) (*payment.Response, error) {
	request := &payment.Request{
		TransactionAmount: float64(order.Total()),
		Description:       "Terminal Café - Pedido",
		PaymentMethodID:   "pix",
		Payer: &payment.PayerRequest{
			Email: order.Customer.Email,
		},
	}

	response, err := mp.client.Create(context.Background(), *request)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pagamento: %v", err)
	}

	return response, nil
}

func (mp *MercadoPagoProvider) HandleWebhook(payload []byte) error {
	// Move webhook handling here if needed
	return nil
}
