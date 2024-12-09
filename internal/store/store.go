package store

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"terminal-cafe/internal/config"
	"terminal-cafe/internal/models"
	"terminal-cafe/internal/notification"
	"terminal-cafe/internal/payment"
	"terminal-cafe/internal/shipping"
)

type Store struct {
	Products []models.Product
	notifier *notification.EmailNotifier
	shipping *shipping.CorreiosProvider
}

func NewStore(cfg *config.Config) *Store {
	notifier := notification.NewEmailNotifier(
		cfg.EmailFrom,
		cfg.EmailPassword,
		cfg.EmailTo,
		cfg.SMTPHost,
		cfg.SMTPPort,
	)

	shipping := shipping.NewCorreiosProvider(cfg)

	return &Store{
		Products: make([]models.Product, 0),
		notifier: notifier,
		shipping: shipping,
	}
}

func (s *Store) LoadProductsFromMD(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentCategory string
	var currentProduct *models.Product
	productID := 1 // Contador para IDs sequenciais

	// Regex para extrair preço
	priceRegex := regexp.MustCompile(`R\$ (\d+,?\d*)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Ignora linhas vazias
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Detecta categoria
		if strings.HasPrefix(line, "## ") {
			currentCategory = strings.TrimPrefix(line, "## ")
			continue
		}

		// Detecta nome do produto
		if strings.HasPrefix(line, "### ") {
			if currentProduct != nil {
				s.Products = append(s.Products, *currentProduct)
			}
			currentProduct = &models.Product{
				ID:       productID,
				Name:     strings.TrimPrefix(line, "### "),
				Category: currentCategory,
			}
			productID++ // Incrementa o ID para o próximo produto
			continue
		}

		// Processa preço e descrição
		if currentProduct != nil {
			if strings.HasPrefix(line, "- Preço:") {
				matches := priceRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					priceStr := strings.Replace(matches[1], ",", ".", 1)
					price, err := strconv.ParseFloat(priceStr, 64)
					if err == nil {
						currentProduct.Price = price
					}
				}
			} else if strings.HasPrefix(line, "- Descrição:") {
				currentProduct.Description = strings.TrimPrefix(line, "- Descrição: ")
			}
		}
	}

	// Adiciona o último produto
	if currentProduct != nil {
		s.Products = append(s.Products, *currentProduct)
	}

	return scanner.Err()
}

func (s *Store) DisplayMenu(out io.Writer) {
	fmt.Fprintf(out, "\n=== Terminal Café ===\n\n")

	var currentCategory string
	for _, product := range s.Products {
		if product.Category != currentCategory {
			currentCategory = product.Category
			fmt.Fprintf(out, "\n%s\n%s\n", currentCategory, strings.Repeat("-", len(currentCategory)))
		}

		fmt.Fprintf(out, "\n[%d] %s\nR$ %.2f\n%s\n",
			product.ID,
			product.Name,
			product.Price,
			product.Description)
	}
	fmt.Fprintf(out, "\n==================\n")
}

func (s *Store) ProcessOrder(in io.Reader, out io.Writer, errOut io.Writer) error {
	var order models.Order
	var customer models.Customer
	var input string
	var err error
	reader := bufio.NewReader(in)

	fmt.Fprintf(out, "\nBem-vindo ao Terminal Café!\nPara fazer seu pedido, siga as instruções:\n")

	for {
		s.DisplayMenu(out)

		fmt.Fprintf(out, "\nDigite o número do produto (ou 0 para finalizar): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("erro ao ler entrada: %v", err)
		}
		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Fprintf(errOut, "Entrada inválida\n")
			continue
		}

		if choice == 0 {
			break
		}

		if choice < 1 || choice > len(s.Products) {
			fmt.Fprintln(errOut, "Produto inválido!")
			continue
		}

		fmt.Fprintf(out, "Quantidade: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("erro ao ler entrada: %v", err)
		}
		qty, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Fprintf(errOut, "Quantidade inválida\n")
			continue
		}

		if qty < 1 {
			fmt.Fprintln(errOut, "Quantidade inválida!")
			continue
		}

		order.Items = append(order.Items, models.OrderItem{
			Product:  s.Products[choice-1],
			Quantity: qty,
		})

		fmt.Fprintf(out, "\nProduto adicionado! Deseja mais algum? (0 para finalizar)\n")
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("nenhum item selecionado")
	}

	// Coleta dados do cliente
	fmt.Fprint(out, "\n=== Dados para Entrega ===\n")

	fmt.Fprint(out, "\nEmail: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler email: %v", err)
	}
	customer.Email = strings.TrimSpace(input)
	if !strings.Contains(customer.Email, "@") {
		return fmt.Errorf("email inválido")
	}

	fmt.Fprint(out, "CPF (apenas números): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler CPF: %v", err)
	}
	customer.CPF = strings.TrimSpace(input)
	if len(customer.CPF) != 11 {
		return fmt.Errorf("CPF inválido")
	}

	fmt.Fprint(out, "Telefone (com DDD): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler telefone: %v", err)
	}
	customer.Phone = strings.TrimSpace(input)

	fmt.Fprint(out, "CEP: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler CEP: %v", err)
	}
	customer.CEP = strings.TrimSpace(input)
	if len(customer.CEP) != 8 {
		return fmt.Errorf("CEP inválido")
	}

	// Calcula frete
	shipping, err := s.shipping.CalculateShipping(customer.CEP)
	if err != nil {
		return fmt.Errorf("erro ao calcular frete: %v", err)
	}

	fmt.Fprintf(out, "\nFrete via %s: R$ %.2f (entrega em %d dias)\n",
		shipping.Service,
		shipping.Price,
		shipping.Deadline)

	fmt.Fprint(out, "Endereço (rua/avenida): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler endereço: %v", err)
	}
	customer.Address = strings.TrimSpace(input)

	fmt.Fprint(out, "Número: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler número: %v", err)
	}
	customer.Number = strings.TrimSpace(input)

	fmt.Fprint(out, "Tipo (casa/apartamento): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler tipo: %v", err)
	}
	customer.Type = strings.TrimSpace(input)

	if customer.Type == "apartamento" {
		fmt.Fprint(out, "Número do Apartamento: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("erro ao ler apartamento: %v", err)
		}
		customer.Unit = strings.TrimSpace(input)
	}

	fmt.Fprint(out, "Complemento (opcional - pressione Enter para pular): ")
	input, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("erro ao ler complemento: %v", err)
	}
	customer.Complement = strings.TrimSpace(input)

	order.Customer = customer

	// Inicializa provedor de pagamento
	mp, err := payment.NewMercadoPagoProvider()
	if err != nil {
		return fmt.Errorf("erro ao inicializar pagamento: %v", err)
	}

	// Exibe resumo do pedido
	fmt.Fprintf(out, "\n=== Resumo do Pedido ===\n")
	for _, item := range order.Items {
		fmt.Fprintf(out, "%dx %s (R$ %.2f cada) = R$ %.2f\n",
			item.Quantity,
			item.Product.Name,
			item.Product.Price,
			float64(item.Quantity)*item.Product.Price)
	}
	fmt.Fprintf(out, "\nSubtotal: R$ %.2f", order.Total())
	fmt.Fprintf(out, "\nFrete: R$ %.2f", shipping.Price)
	fmt.Fprintf(out, "\nTotal com frete: R$ %.2f\n", order.Total()+shipping.Price)
	fmt.Fprintf(out, "\nEndereço de entrega: %s, %s - %s\n",
		order.Customer.Address,
		order.Customer.Number,
		order.Customer.CEP)

	// Cria pagamento
	response, err := mp.CreatePayment(&order)
	if err != nil {
		return fmt.Errorf("erro ao processar pagamento: %v", err)
	}

	if response.PointOfInteraction.TransactionData.QRCode != "" {
		fmt.Fprint(out, "\nPara pagar, escaneie o QR Code abaixo:\n")
		fmt.Fprintln(out, response.PointOfInteraction.TransactionData.QRCode)
		fmt.Fprint(out, "\nOu copie e cole o código PIX:\n")
		fmt.Fprintln(out, response.PointOfInteraction.TransactionData.QRCodeBase64)

		// Notifica a empresa sobre o novo pedido
		if err := s.notifier.NotifyNewOrder(&order); err != nil {
			log.Printf("Erro ao notificar pedido: %v", err)
		}
	} else {
		return fmt.Errorf("dados do PIX não disponíveis na resposta")
	}

	return nil
}
