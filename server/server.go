package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Response struct {
	Bid float64 `json:"bid"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "cotacao.db")
	if err != nil {
		log.Fatal("Erro: Falha ao conectar com banco de dados")
	}
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS cotacao (id INTEGER PRIMARY KEY, rate REAL, created_at TEXT)")
	res, err := statement.Exec()
	fmt.Println(res)
	if err != nil {
		log.Fatal("Erro: Falha ao criar tabela cotacao")
	}
}

func handleCotacao(w http.ResponseWriter, r *http.Request) {
	exchangeRate, err := getExchangeRate()
	if err != nil {
		http.Error(w, "Erro: Falha ao obter a cotação do dólar", http.StatusInternalServerError)
		fmt.Printf("Erro %s", err.Error())
		return
	}

	if err := saveToDB(exchangeRate); err != nil {
		http.Error(w, "Erro: Falha ao salvar no banco de dados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Bid: exchangeRate})
}

func getExchangeRate() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error in http.DefaultClient.Do: %s", err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Erro: %d", resp.StatusCode)
	}
	var data map[string]map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	str := data["USDBRL"]["bid"].(string)
	bid, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Printf("Erro ao converter para float64: %s", err.Error())
		return 0, fmt.Errorf("Erro: campo 'bid' não encontrado")
	}

	return bid, nil
}

func saveToDB(rate float64) error {
	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDB()

	_, err := db.ExecContext(ctxDB, "INSERT INTO cotacao (rate, created_at) VALUES (?, ?)", rate, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	http.HandleFunc("/cotacao", handleCotacao)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
