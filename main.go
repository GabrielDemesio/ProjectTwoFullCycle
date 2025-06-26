package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type EnderecoFinal struct {
	Logradouro string
	Bairro     string
	Cidade     string
	Estado     string
	CEP        string
}

type BrasilAPIResponse struct {
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Cep          string `json:"cep"`
}

type ViaCEPResponse struct {
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Cep        string `json:"cep"`
}

type ResultadoAPI struct {
	Fonte    string
	Endereco *EnderecoFinal
}

func fetchBrasilAPI(cep string, ch chan<- *ResultadoAPI) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var brasilApiResponse BrasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&brasilApiResponse); err != nil {
		return
	}

	endereco := &EnderecoFinal{
		Logradouro: brasilApiResponse.Street,
		Bairro:     brasilApiResponse.Neighborhood,
		Cidade:     brasilApiResponse.City,
		Estado:     brasilApiResponse.State,
		CEP:        brasilApiResponse.Cep,
	}

	ch <- &ResultadoAPI{Fonte: "BrasilAPI", Endereco: endereco}
}

func fetchViaCEP(cep string, ch chan<- *ResultadoAPI) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var viaCepResponse ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCepResponse); err != nil {
		return
	}

	endereco := &EnderecoFinal{
		Logradouro: viaCepResponse.Logradouro,
		Bairro:     viaCepResponse.Bairro,
		Cidade:     viaCepResponse.Localidade,
		Estado:     viaCepResponse.UF,
		CEP:        viaCepResponse.Cep,
	}

	ch <- &ResultadoAPI{Fonte: "ViaCEP", Endereco: endereco}
}

func main() {
	fmt.Print("Digite o CEP para consulta: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	cep := scanner.Text()

	if cep == "" {
		fmt.Println("CEP inválido.")
		return
	}

	fmt.Printf("\nBuscando CEP %s nas APIs...\n", cep)

	ch := make(chan *ResultadoAPI)

	go fetchBrasilAPI(cep, ch)
	go fetchViaCEP(cep, ch)

	select {
	case resultado := <-ch:
		fmt.Println("\n---------------------------------")
		fmt.Printf("Resposta mais rápida da API: %s\n", resultado.Fonte)
		fmt.Println("---------------------------------")
		fmt.Printf("CEP:         %s\n", resultado.Endereco.CEP)
		fmt.Printf("Logradouro:  %s\n", resultado.Endereco.Logradouro)
		fmt.Printf("Bairro:      %s\n", resultado.Endereco.Bairro)
		fmt.Printf("Cidade:      %s\n", resultado.Endereco.Cidade)
		fmt.Printf("Estado:      %s\n", resultado.Endereco.Estado)
		fmt.Println("---------------------------------")

	case <-time.After(1 * time.Second):
		fmt.Println("\n Erro: Timeout. Nenhuma API respondeu em 1 segundo.")
	}
}
