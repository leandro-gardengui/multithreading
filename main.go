package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Structs para respostas das APIs
type BrasilAPIResponse struct {
	Cep      string `json:"cep"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Street   string `json:"street"`
	Service  string `json:"service"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

// Response genérica para unificar as respostas
type APIResponse struct {
	CEP      string
	Street   string
	District string
	City     string
	State    string
	APIName  string
}

func fetchBrasilAPI(ctx context.Context, cep string, ch chan<- APIResponse) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var data BrasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	response := APIResponse{
		CEP:      data.Cep,
		Street:   data.Street,
		District: data.District,
		City:     data.City,
		State:    data.State,
		APIName:  "BrasilAPI",
	}

	ch <- response
}

func fetchViaCEP(ctx context.Context, cep string, ch chan<- APIResponse) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var data ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	response := APIResponse{
		CEP:      data.Cep,
		Street:   data.Logradouro,
		District: data.Bairro,
		City:     data.Localidade,
		State:    data.Uf,
		APIName:  "ViaCEP",
	}

	ch <- response
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <CEP>")
		fmt.Println("Exemplo: go run main.go 01153000")
		os.Exit(1)
	}

	cep := os.Args[1]

	// Context com timeout de 1 segundo
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Channel para receber as respostas
	ch := make(chan APIResponse, 2)

	// Executar as requisições simultaneamente
	go fetchBrasilAPI(ctx, cep, ch)
	go fetchViaCEP(ctx, cep, ch)

	// Aguardar a primeira resposta ou timeout
	select {
	case response := <-ch:
		fmt.Printf("API mais rápida: %s\n", response.APIName)
		fmt.Printf("CEP: %s\n", response.CEP)
		fmt.Printf("Logradouro: %s\n", response.Street)
		fmt.Printf("Bairro: %s\n", response.District)
		fmt.Printf("Cidade: %s\n", response.City)
		fmt.Printf("Estado: %s\n", response.State)
	case <-ctx.Done():
		fmt.Println("Erro: Timeout - nenhuma API respondeu em menos de 1 segundo")
		os.Exit(1)
	}
}
