package shipping

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"terminal-cafe/internal/config"
)

type ShippingInfo struct {
	Service  string
	Price    float64
	Deadline int
}

type CorreiosProvider struct {
	CompanyCode string // Código da empresa nos Correios
	Password    string
	OriginCEP   string // CEP de origem (da loja)
}

type FreteRequest struct {
	Service        string // 04510 = PAC, 04014 = SEDEX
	OriginCEP      string
	DestinationCEP string
	Weight         float64 // em kg
	Length         float64 // em cm
	Height         float64
	Width          float64
}

type correiosResponse struct {
	XMLName xml.Name `xml:"Servicos"`
	Service struct {
		Value     float64 `xml:"Valor"`
		Deadline  int     `xml:"PrazoEntrega"`
		Available bool    `xml:"Erro" available:"0"`
	} `xml:"cServico"`
}

func NewCorreiosProvider(cfg *config.Config) *CorreiosProvider {
	return &CorreiosProvider{
		CompanyCode: cfg.CorreiosCode,
		Password:    cfg.CorreiosPassword,
		OriginCEP:   cfg.StoreCEP,
	}
}

func (c *CorreiosProvider) CalculateShipping(destCEP string) (*ShippingInfo, error) {
	// Configuração padrão para pacote de café
	request := FreteRequest{
		Service:        "04510", // PAC
		OriginCEP:      c.OriginCEP,
		DestinationCEP: destCEP,
		Weight:         0.5, // 500g
		Length:         16,  // cm
		Height:         8,   // cm
		Width:          12,  // cm
	}

	url := fmt.Sprintf(
		"http://ws.correios.com.br/calculador/CalcPrecoPrazo.aspx?"+
			"nCdEmpresa=%s&sDsSenha=%s&nCdServico=%s&"+
			"sCepOrigem=%s&sCepDestino=%s&"+
			"nVlPeso=%.2f&nCdFormato=1&"+
			"nVlComprimento=%.2f&nVlAltura=%.2f&"+
			"nVlLargura=%.2f&nVlDiametro=0&"+
			"sCdMaoPropria=N&nVlValorDeclarado=0&sCdAvisoRecebimento=N&"+
			"StrRetorno=xml",
		c.CompanyCode, c.Password, request.Service,
		request.OriginCEP, request.DestinationCEP,
		request.Weight,
		request.Length, request.Height,
		request.Width,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar Correios: %v", err)
	}
	defer resp.Body.Close()

	var result correiosResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if !result.Service.Available {
		return nil, fmt.Errorf("serviço indisponível para este CEP")
	}

	return &ShippingInfo{
		Price:    result.Service.Value,
		Deadline: result.Service.Deadline,
		Service:  "PAC",
	}, nil
}
