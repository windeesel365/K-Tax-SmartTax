package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// pattern ที่ admin input request
type Deduction struct {
	Amount float64 `json:"amount"`
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

type TaxLevel struct {
	Level string        `json:"level"`
	Tax   CustomFloat64 `json:"tax"`
}

type IncomewithTaxResponse struct {
	Totalincome CustomFloat64 `json:"totalIncome"`
	Tax         CustomFloat64 `json:"tax"`
	TaxRefund   CustomFloat64 `json:"taxRefund,omitempty"`
}

type jwtCustomClaims struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.RegisteredClaims
}

type PersonalDeduction struct {
	ID                int
	PersonalDeduction float64
}

// initialize value
var initialPersonalExemption float64 = 60000.0
var initialdonations float64 = 0.0
var initialkReceipts float64 = 0.0

// initial exemptions กับค่า limits
var personalExemptionUpperLimit float64 = 100000.0
var donationsUpperLimit float64 = 100000.0
var kReceiptsUpperLimit float64 = 50000.0

// declare สำหรับ ref database และ idข้อมูล postgresql
var db *sql.DB
var id int

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

	// create 'deductions' table เพื่อเตรียมสำหรับ admin ปรับค่า
	err = createAdminDeductionsTable(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PostgreSQL: Created admin deductions table successfully.")

	// count คือจำนวน row ใน table ณ จุดเวลา
	count, err := CountRows(db, "deductions")
	if err != nil {
		log.Fatal(err)
	}

	// ถ้า count น้อยกว่า 1, ถึงยอมให้สร้าง row ใหม่
	if count < 1 {
		id, err := createDeduction(db, initialPersonalExemption, kReceiptsUpperLimit)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("PostgreSQL: Created admin deductions row with id: %d, initialPersonalExemption: %f\n kReceiptsUpperLimit: %f\n", id, initialPersonalExemption, kReceiptsUpperLimit)
	} //else ไปเช็คค่าใน db

	// ใช้ row แรกสุด คือ id แรกสุด ในการ update และ read
	lowestID, err := getLowestID(db, "deductions")
	if err != nil {
		panic(err.Error())
	}

	id = lowestID // id เป็น id สำหรับ ref ข้อมูล postgresql

	//เช็ค authentication ของ admin
	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	basicAuthMiddleware := middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		isAuthenticated := username == adminUsername && password == adminPassword
		if !isAuthenticated {
			// log เพื่อ notice failed attempt
			log.Printf("Failed login attempt for username: %s", username)
			// ส่ง customize response message to client
			return false, echo.NewHTTPError(http.StatusUnauthorized, "There was a problem logging in. Check your username and password.")
		}
		return isAuthenticated, nil
	})

	e.POST("/tax/calculations", HandleTaxCalculation)

	adminGroup := e.Group("/admin")
	adminGroup.Use(basicAuthMiddleware)

	adminGroup.POST("/login", login)

	adminGroup.POST("/deductions/personal", setPersonalDeduction)

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

}

func login(c echo.Context) error {
	//username and password จาก client request form-data
	username := c.FormValue("username")
	password := c.FormValue("password")

	// check against env variables
	if username != os.Getenv("ADMIN_USERNAME") || password != os.Getenv("ADMIN_PASSWORD") {
		return echo.ErrUnauthorized
	}

	claims := &jwtCustomClaims{
		username,
		true,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 10)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

func createAdminDeductionsTable(db *sql.DB) error {
	// SQL statement เพื่อ create 'deductions' table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS deductions (
			id SERIAL PRIMARY KEY,
			personal_deduction INTEGER NOT NULL,
			k_receipt_deduction INTEGER NOT NULL
		);
	`
	// Execute SQL statement ข้างบน
	_, err := db.Exec(createTableSQL)
	return err
}

func createDeduction(db *sql.DB, personalDeduction float64, kReceiptDeduction float64) (int, error) {
	var id int
	err := db.QueryRow(`INSERT INTO deductions(personal_deduction, k_receipt_deduction) VALUES($1, $2) RETURNING id;`, personalDeduction, kReceiptDeduction).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CountRows นับจำนวน row ของ table ใน database
func CountRows(db *sql.DB, tableName string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getLowestID(db *sql.DB, tableName string) (int, error) {
	var lowestID int
	// query select lowest ID จาก table
	query := "SELECT MIN(id) FROM " + tableName

	// query database
	err := db.QueryRow(query).Scan(&lowestID)
	if err != nil {
		return 0, err
	}
	return lowestID, nil
}

// POST: /admin/deductions/personal
func setPersonalDeduction(c echo.Context) error {
	// Read body to a variable
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	defer c.Request().Body.Close()

	// validation function process
	if err := validatePersonalInput(body); err != nil {
		return err
	}

	// after validation
	// Bind JSON to struct
	d := new(Deduction)
	if err := json.Unmarshal(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	// Set change to initialPersonalExemption
	initialPersonalExemption = d.Amount

	// update PersonalDeduction to postgres db
	err = updatePersonalDeduction(db, id, initialPersonalExemption)
	if err != nil {
		log.Fatal(err)
	}
	// read after above update
	adminPDeductions, err := getPersonalDeduction(db, id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("***********\nAdmin updated initialPersonalExemption validated then updated postgresql row: %+v\n", adminPDeductions)

	//respond client(admin)
	return c.JSON(http.StatusOK, map[string]CustomFloat64{"personalDeduction": CustomFloat64(d.Amount)})
}

func getPersonalDeduction(db *sql.DB, id int) (PersonalDeduction, error) {
	var deduc PersonalDeduction
	row := db.QueryRow(`SELECT id, personal_deduction FROM deductions WHERE id = $1;`, id)
	err := row.Scan(&deduc.ID, &deduc.PersonalDeduction)
	if err != nil {
		return PersonalDeduction{}, err
	}
	return deduc, nil
}

func updatePersonalDeduction(db *sql.DB, id int, personalDeduction float64) error {
	_, err := db.Exec(`UPDATE deductions SET personal_deduction = $1 WHERE id = $2;`, personalDeduction, id)
	return err
}

// Validation   input data for personal deductions
func validatePersonalInput(body []byte) error {
	//validate raw JSON not empty
	if len(body) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide input data")
	}

	//validate raw JSON root-level key count ว่าmatch  key count of correct pattern
	expectedKeys := []string{"amount"}
	count, err := JsonRootLevelKeyCount(string(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if count != len(expectedKeys) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input. Please ensure you enter only one amount, corresponding to setting value of personal deduction.")
	}

	//validate raw JSON root-level key count order
	if err := CheckJSONOrder(body, expectedKeys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//validate struct และ amount
	d := new(Deduction)
	if err := json.Unmarshal(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	if err := validateFields(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format. Please check the input format again")
	}

	if d.Amount > 100000.0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please ensure Personal Deduction amount does not exceed THB 100,000.")
	}

	if d.Amount <= 10000.0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please ensure Personal Deduction must be more than THB 10000.")
	}

	return nil
}

// validateFields matches the number of JSON keys to the number of struct fields
func validateFields(data []byte, d *Deduction) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Get number of fields in Deduction struct
	t := reflect.TypeOf(*d)
	deductionFieldCount := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			deductionFieldCount++
		}
	}

	// Check if numbers of fields match
	if len(raw) != deductionFieldCount {
		return fmt.Errorf("number of fields in JSON (%d) does not match number of fields in Deduction (%d)", len(raw), deductionFieldCount)
	}

	return nil
}
