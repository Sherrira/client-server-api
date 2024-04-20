package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Bid float64 `json:"bid"`
}

func getDollarExchangeRate() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return 0, fmt.Errorf("Erro ao criar requisiçao: %s", err.Error())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Erro ao obter resposta da requisicao: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Erro: %d", resp.StatusCode)
	}

	var data Response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return data.Bid, nil
}

func saveToFile(exchangeRate float64) {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println("Erro: Falha ao criar arquivo")
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %f", exchangeRate))
	if err != nil {
		fmt.Println("Erro: Falha ao escrever no arquivo")
		return
	}
}

func main() {
	exchangeRate, err := getDollarExchangeRate()
	if err != nil {
		fmt.Printf("Erro: Falha ao obter a cotação do dólar: %s", err.Error())
		return
	}

	saveToFile(exchangeRate)
}
