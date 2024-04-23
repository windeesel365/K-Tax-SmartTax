package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// data structure pattern ที่ user client request
type TaxRequest struct {
	TotalIncome float64 `json:"totalIncome"`
	WHT         float64 `json:"wht"`
	Allowances  []struct {
		AllowanceType string  `json:"allowanceType"`
		Amount        float64 `json:"amount"`
	} `json:"allowances"`
}

// CustomFloat64 เป็น float64 ที่ custom ใหม่
type CustomFloat64 float64

// customizes ตัวเลขการเงิน เพื่อ output decimal places ตามต้องการ
func (cf CustomFloat64) MarshalJSON() ([]byte, error) {
	d := decimal.NewFromFloat(float64(cf))     //จาก github.com/shopspring/decimal
	formatted := d.RoundBank(1).StringFixed(1) // RoundBank เพื่อ banker's rounding // StringFixed ตำแหน่งทศนิยมในข้อมูลที่จะแสดงผล
	return []byte(formatted), nil
}

type TaxResponse struct {
	Tax       CustomFloat64 `json:"tax"`
	TaxRefund CustomFloat64 `json:"taxRefund,omitempty"`
}

// initialize value
var initialPersonalExemption float64 = 60000.0
var initialdonations float64 = 0.0
var initialkReceipts float64 = 0.0

// initial exemptions กับค่า limits
var personalExemptionUpperLimit float64 = 100000.0
var donationsUpperLimit float64 = 100000.0
var kReceiptsUpperLimit float64 = 50000.0

// declare สำหรับ ref database
var db *sql.DB

//pending

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get port number from the environment variable 'PORT'
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable not set.")
	}

	//define pattern regex สำหรับเช็คportว่า 4-digit number
	pattern := regexp.MustCompile(`^\d{4}$`)
	if !pattern.MatchString(port) {
		log.Fatal("before starting server, please ensure that PORT environment variable must in 4-digit number.")
	}

	// Postgresql preparation part
	// Retrieve the DATABASE_URL from the environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("DATABASE_URL is not set")
		return
	}

	// สร้าง connection กับ postgresql
	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PostgreSQL: Connected successfully.")

	e.POST("/tax/calculations", HandleTaxCalculation)

	//graceful shutdown //start server in goroutine
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// รอ interrupt signal เพื่อ gracefully shutdown server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	e.Logger.Print("shutting down the server")

	// context เพื่อ timeout shutdown after 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// shutdown server
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	//pending
}
