package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db       *sql.DB
	dbDriver string
)

// Product represents a product in the system
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// Transaction represents a sales transaction
type Transaction struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

// Receipt represents a sales receipt
type Receipt struct {
	ID       int           `json:"id"`
	Items    []ReceiptItem `json:"items"`
	Total    float64       `json:"total"`
	Discount float64       `json:"discount"`
	Tax      float64       `json:"tax"`
	NetTotal float64       `json:"net_total"`
	DateTime time.Time     `json:"datetime"`
}

type ReceiptItem struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

// SalesReport represents daily sales data
type SalesReport struct {
	Date               string        `json:"date"`
	TotalTransactions  int           `json:"total_transactions"`
	TotalSalesAmount   float64       `json:"total_sales_amount"`
	TransactionDetails []Transaction `json:"transaction_details"`
}

func init() {
	if err := connectDatabase(); err != nil {
		log.Fatal(err)
	}

	createTables()
}

func connectDatabase() error {
	dbDriver = os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "mysql"
	}

	dsn := os.Getenv("DB_DSN")
	var err error

	switch dbDriver {
	case "mysql", "mysql2":
		if dsn == "" {
			dsn = "root:@tcp(127.0.0.1:3306)/posdb?parseTime=true"
		}
		db, err = sql.Open("mysql", dsn)
	case "sqlite3":
		if dsn == "" {
			dsn = "pos.db"
		}
		db, err = sql.Open("sqlite3", dsn)
	default:
		return fmt.Errorf("unsupported DB_DRIVER: %s", dbDriver)
	}

	if err != nil {
		return err
	}

	return db.Ping()
}

func createTables() {
	var schema string
	if dbDriver == "mysql" || dbDriver == "mysql2" {
		schema = `
	CREATE TABLE IF NOT EXISTS products(
		id INT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL,
		price DOUBLE NOT NULL,
		stock INT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS transactions(
		id INT PRIMARY KEY AUTO_INCREMENT,
		product_id INT NOT NULL,
		quantity INT NOT NULL,
		amount DOUBLE NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(product_id) REFERENCES products(id)
	);

	CREATE TABLE IF NOT EXISTS receipts(
		id INT PRIMARY KEY AUTO_INCREMENT,
		items TEXT NOT NULL,
		total DOUBLE NOT NULL,
		discount DOUBLE DEFAULT 0,
		tax DOUBLE DEFAULT 0,
		net_total DOUBLE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	} else {
		schema = `
	CREATE TABLE IF NOT EXISTS products(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price REAL NOT NULL,
		stock INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS transactions(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		amount REAL NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(product_id) REFERENCES products(id)
	);

	CREATE TABLE IF NOT EXISTS receipts(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		items TEXT NOT NULL,
		total REAL NOT NULL,
		discount REAL DEFAULT 0,
		tax REAL DEFAULT 0,
		net_total REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	}

	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			log.Fatal(err)
		}
	}
}

// API Handlers
func addProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO products(name, price, stock) VALUES(?, ?, ?)",
		p.Name, p.Price, p.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	p.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "product": p})
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, price, stock FROM products")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func updateProductStock(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ProductID int `json:"product_id"`
		Stock     int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE products SET stock = ? WHERE id = ?", data.Stock, data.ProductID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func removeProduct(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ProductID int `json:"product_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM products WHERE id = ?", data.ProductID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func processTransaction(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get product
	var price float64
	var stock int
	err := db.QueryRow("SELECT price, stock FROM products WHERE id = ?", data.ProductID).Scan(&price, &stock)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Check stock
	if stock < data.Quantity {
		http.Error(w, "Insufficient stock", http.StatusBadRequest)
		return
	}

	// Calculate amount
	amount := price * float64(data.Quantity)

	// Record transaction
	result, err := db.Exec("INSERT INTO transactions(product_id, quantity, amount) VALUES(?, ?, ?)",
		data.ProductID, data.Quantity, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update stock
	newStock := stock - data.Quantity
	db.Exec("UPDATE products SET stock = ? WHERE id = ?", newStock, data.ProductID)

	id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "success",
		"transaction_id":  id,
		"amount":          amount,
		"remaining_stock": newStock,
	})
}

func generateReceipt(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Items    []ReceiptItem `json:"items"`
		Discount float64       `json:"discount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(data.Items) == 0 {
		http.Error(w, "No items provided", http.StatusBadRequest)
		return
	}

	// Calculate totals
	total := 0.0
	for _, item := range data.Items {
		total += item.Amount
	}

	discountAmount := data.Discount
	afterDiscount := total - discountAmount
	if afterDiscount < 0 {
		afterDiscount = 0
	}
	tax := afterDiscount * 0.12 // 12% tax
	netTotal := afterDiscount + tax

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range data.Items {
		var stock int
		if err := tx.QueryRow("SELECT stock FROM products WHERE id = ?", item.ProductID).Scan(&stock); err != nil {
			tx.Rollback()
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		if stock < item.Quantity {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Insufficient stock for product %d", item.ProductID), http.StatusBadRequest)
			return
		}

		newStock := stock - item.Quantity
		if _, err := tx.Exec("UPDATE products SET stock = ? WHERE id = ?", newStock, item.ProductID); err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := tx.Exec("INSERT INTO transactions(product_id, quantity, amount) VALUES(?, ?, ?)",
			item.ProductID, item.Quantity, item.Amount); err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	itemsJSON, _ := json.Marshal(data.Items)
	result, err := tx.Exec("INSERT INTO receipts(items, total, discount, tax, net_total) VALUES(?, ?, ?, ?, ?)",
		string(itemsJSON), total, discountAmount, tax, netTotal)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	receipt := Receipt{
		ID:       int(id),
		Items:    data.Items,
		Total:    total,
		Discount: discountAmount,
		Tax:      tax,
		NetTotal: netTotal,
		DateTime: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

func getDailySalesReport(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	rows, err := db.Query(`SELECT id, product_id, quantity, amount, timestamp 
		FROM transactions 
		WHERE DATE(timestamp) = ?
		ORDER BY timestamp DESC`, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []Transaction
	totalAmount := 0.0

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.ProductID, &t.Quantity, &t.Amount, &t.Timestamp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, t)
		totalAmount += t.Amount
	}

	report := SalesReport{
		Date:               date,
		TotalTransactions:  len(transactions),
		TotalSalesAmount:   totalAmount,
		TransactionDetails: transactions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func getTotalSales(w http.ResponseWriter, r *http.Request) {
	var totalSales float64
	err := db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions").Scan(&totalSales)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total_sales": totalSales})
}

func getReceiptHistory(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, items, total, discount, tax, net_total, created_at 
		FROM receipts 
		ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var receipts []Receipt
	for rows.Next() {
		var receipt Receipt
		var itemsJSON string
		if err := rows.Scan(&receipt.ID, &itemsJSON, &receipt.Total, &receipt.Discount, &receipt.Tax, &receipt.NetTotal, &receipt.DateTime); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var items []ReceiptItem
		if err := json.Unmarshal([]byte(itemsJSON), &items); err != nil {
			log.Printf("Error parsing receipt items: %v", err)
		}
		receipt.Items = items
		receipts = append(receipts, receipt)
	}

	w.Header().Set("Content-Type", "application/json")
	if receipts == nil {
		receipts = []Receipt{}
	}
	json.NewEncoder(w).Encode(receipts)
}

// Serve static files
func serveStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	defer db.Close()

	// API Routes
	http.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getProducts(w, r)
		case "POST":
			addProduct(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/products/remove", removeProduct)
	http.HandleFunc("/api/products/stock", updateProductStock)
	http.HandleFunc("/api/transaction", processTransaction)
	http.HandleFunc("/api/receipt", generateReceipt)
	http.HandleFunc("/api/sales/daily", getDailySalesReport)
	http.HandleFunc("/api/sales/total", getTotalSales)
	http.HandleFunc("/api/receipts/history", getReceiptHistory)

	// Static files
	http.HandleFunc("/", serveStatic)
	http.HandleFunc("/index.html", serveStatic)
	fs := http.FileServer(http.Dir("."))
	http.Handle("/static/", fs)

	fmt.Println("POS System running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
