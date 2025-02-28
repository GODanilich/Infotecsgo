package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// generateWalletAddress генерирует случайный адрес кошелька.
// Возвращает строку длиной 64 символа, являющуюся шестнадцатеричным значением,
// сгенерированным на основе 32 случайных байтов.
func generateWalletAddress() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes)
}

// initDB инициализирует SQLite.
// Создает таблицы wallets и transactions, если они еще не существуют.
// При первом запуске создает 10 кошельков с балансом 100.0 у.е.
func initDB(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS wallets (
		address TEXT PRIMARY KEY,
		balance REAL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_address TEXT,
		to_address TEXT,
		amount REAL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM wallets").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		for i := 0; i < 10; i++ {
			address := generateWalletAddress()
			_, err := db.Exec("INSERT INTO wallets (address, balance) VALUES (?, ?)", address, 100.0)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Println("Created 10 initial wallets with 100.0 balance each")
	}
}

// sendHandler обрабатывает POST /api/send для транзакций между кошельками.
// Принимает JSON-объект с полями from (отправитель), to (получатель) и amount (сумма).
// Проверяет кошельки на существование и наличие средств, выполняет транзакцию в атомарном блоке.
// Возвращает сообщение об успехе или ошибку.
func sendHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			From   string  `json:"from"`
			To     string  `json:"to"`
			Amount float64 `json:"amount"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if req.From == "" || req.To == "" || req.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters: 'from', 'to', and 'amount' must be valid"})
			return
		}

		var fromBalance float64
		err := db.QueryRow("SELECT balance FROM wallets WHERE address = ?", req.From).Scan(&fromBalance)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet 'from' not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var toExists int
		err = db.QueryRow("SELECT 1 FROM wallets WHERE address = ?", req.To).Scan(&toExists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet 'to' not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if fromBalance < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		_, err = tx.Exec("UPDATE wallets SET balance = balance - ? WHERE address = ?", req.Amount, req.From)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		_, err = tx.Exec("UPDATE wallets SET balance = balance + ? WHERE address = ?", req.Amount, req.To)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		_, err = tx.Exec("INSERT INTO transactions (from_address, to_address, amount) VALUES (?, ?, ?)", req.From, req.To, req.Amount)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if err = tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transaction successful"})
	}
}

// getLastHandler обрабатывает GET /api/transactions?count=N для получения N последних транзакций.
// Принимает параметр count, возвращает массив JSON-объектов с информацией о транзакциях,
// отсортированных по убыванию времени, либо ошибку.
func getLastHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		countStr := c.Query("count")
		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'count' parameter: must be a positive integer"})
			return
		}

		rows, err := db.Query("SELECT from_address, to_address, amount, timestamp FROM transactions ORDER BY timestamp DESC LIMIT ?", count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		var transactions []map[string]interface{}
		for rows.Next() {
			var from, to, timestamp string
			var amount float64
			if err := rows.Scan(&from, &to, &amount, &timestamp); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
			transactions = append(transactions, map[string]interface{}{
				"from":      from,
				"to":        to,
				"amount":    amount,
				"timestamp": timestamp,
			})
		}

		c.JSON(http.StatusOK, transactions)
	}
}

// getBalanceHandler обрабатывает GET /api/wallet/{address}/balance для отображения баланса.
// Извлекает адрес кошелька из пути запроса, возвращает JSON-объект с текущим балансом.
// Если кошелек не найден,то возвращает ошибку 404.
func getBalanceHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Address is required"})
			return
		}

		var balance float64
		err := db.QueryRow("SELECT balance FROM wallets WHERE address = ?", address).Scan(&balance)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balance": balance})
	}
}

// main запускает HTTP-сервер приложения.
// Инициализирует БД SQLite, настраивается маршутизация и запускается сервер на порт 8080.
func main() {
	db, err := sql.Open("sqlite3", "./wallet.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	initDB(db)

	r := gin.Default()

	r.POST("/api/send", sendHandler(db))
	r.GET("/api/transactions", getLastHandler(db))
	r.GET("/api/wallet/:address/balance", getBalanceHandler(db))

	log.Println("Server started on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
