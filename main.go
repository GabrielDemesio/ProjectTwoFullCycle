package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Adress struct {
	Logradouro string `json:street`
	Bairro     string `json:neighborhood`
	Cidade     string `json:city`
	Estado     string `json:state`
	Cep        string `json:cep`
}

type ViaCEPENDERECO struct {
	Logradouro string `json:logradouro`
	Bairro     string `json:bairro`
	Localidade string `json:localidade`
	UF         string `json:uf`
	Cep        string `json:cep`
}
type ResultAPI struct {
	Fonte  string
	Adress *Adress
}

func fetchBasilAPI(cep string, ch chan<- *ResultAPI) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	var adress Adress
	if err := json.NewDecoder(resp.Body).Decode(&adress); err != nil {
		return
	}
	ch <- &ResultAPI{Fonte: "BrasilAPI", Adress: &adress}
}
func fetchViaCepAPI(cep string, ch chan<- *ResultAPI) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	var viaCepResp ViaCEPENDERECO
	if err := json.NewDecoder(resp.Body).Decode(&viaCepResp); err != nil {
		return
	}
	adress := &Adress{
		Logradouro: viaCepResp.Logradouro,
		Bairro:     viaCepResp.Bairro,
		Cidade:     viaCepResp.Localidade,
		Estado:     viaCepResp.UF,
		Cep:        viaCepResp.Cep,
	}
	ch <- &ResultAPI{Fonte: "ViaCepAPI", Adress: adress}
}

func main() {
	fmt.Println("Digite o Cep para consulta: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	cep := scanner.Text()
	if cep == "" {
		fmt.Println("Cep inválido")
		return
	}
	fmt.Printf("\nBuscando CEP: %s\n nas Apis... \n\n", cep)
	ch := make(chan *ResultAPI)
	go fetchBasilAPI(cep, ch)
	go fetchViaCepAPI(cep, ch)
	select {
	case result := <-ch:
		fmt.Printf("Api most quick", result.Fonte)
		fmt.Println("----------------------------------------")
		fmt.Printf("CEP: %s\n", result.Adress.Cep)
		fmt.Printf("Logradouro: %s\n", result.Adress.Logradouro)
		fmt.Printf("Bairro: %s\n", result.Adress.Bairro)
		fmt.Printf("Cidade: %s\n", result.Adress.Cidade)
		fmt.Printf("Estado: %s\n", result.Adress.Estado)
		fmt.Println("----------------------------------------")
	case <-time.After(1 * time.Second):
		fmt.Println("Error: Timeout. Nenhuma api respondeu em 1 segundo")
	}
}
