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
	"regexp"
	"strings"
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

// initial exemptions and limits
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

	e.POST("/tax/calculations", handleTaxCalculation)

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

// Handle tax calculation
func handleTaxCalculation(c echo.Context) error {
	// Read body to a variable
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	defer c.Request().Body.Close()

	// bind JSON to struct และ check error
	req := new(TaxRequest)
	if err := json.Unmarshal(body, &req); err != nil {
		// Provide a more detailed error message if JSON is incorrect
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	// expected key order ที่ถูกต้อง เพื่อใช้ validate JSON order
	expectedKeys := []string{"totalIncome", "wht", "allowances"}

	// validate JSON top-level keys count
	count, err := JsonRootLevelKeyCount(string(body))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	if count != len(expectedKeys) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format, ensure input just totalIncome, wht and allowances")
	}

	// validate JSON order
	if err := checkJSONOrder(body, expectedKeys); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// รวมการ Validate amount ของ struct req  values
	if err := validateTaxRequestAmount(*req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// allowance 3 types เริ่มมาจากค่าเริ่มต้น
	personalExemption := initialPersonalExemption
	donations := initialdonations
	kReceipts := initialkReceipts

	// loop และแยก allowance 3 types แล้วเทียบ เพื่อได้ค่าที่นำไปใช้ได้
	for _, allowance := range req.Allowances {
		if allowance.AllowanceType == "personal" {
			if allowance.Amount <= 10000 {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The personal exemption must be more than 10,000 THB.  Please update the amount and try again."})

			} else {
				if allowance.Amount > personalExemptionUpperLimit {
					personalExemption = personalExemptionUpperLimit
				}
			}
		}

		if allowance.AllowanceType == "donation" {
			if allowance.Amount >= 0 {
				donations += allowance.Amount
				if donations > donationsUpperLimit {
					donations = donationsUpperLimit
				}
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The donation must be more than 0 THB. Please enter a positive amount and try again."})
			}
		}

		if allowance.AllowanceType == "k-receipt" {
			if allowance.Amount > 0 {
				kReceipts += allowance.Amount
				if kReceipts > kReceiptsUpperLimit {
					kReceipts = kReceiptsUpperLimit
				}
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "The kReceipts must be more than 0 THB. Please enter a positive amount and try again."})
			}
		}
	}

	// หา taxable income
	taxableIncome := caltaxableIncome(req.TotalIncome, personalExemption, donations, kReceipts)

	// หา taxPayable, taxRefund
	taxPayable, taxRefund := CalculateTaxPayableAndRefund(taxableIncome, req.WHT)

	response := TaxResponse{Tax: taxPayable, TaxRefund: taxRefund}

	// เปลี่ยน response เป็น map ดึง tax มาจาก response
	responseMap := map[string]interface{}{
		"tax": response.Tax,
	}

	// รวม taxRefund เข้า map ถ้า taxRefund ไม่เป็นzero
	if response.TaxRefund > 0 {
		responseMap["taxRefund"] = response.TaxRefund
	}

	//pending
	return c.JSON(http.StatusOK, responseMap)
}

// checkJSONOrder checks if the keys in the provided JSON body are in the expected order.
func checkJSONOrder(body []byte, expectedKeys []string) error {

	keys, err := GetOrderedKeysFromJSON(body)
	if err != nil {
		fmt.Println("Error:", err)
	}

	keys = keys[0:len(expectedKeys)]

	expectedKeysMembers := strings.Join(expectedKeys, ", ")

	if len(keys) != len(expectedKeys) {
		return fmt.Errorf("please enter just %d key(s) ordered by: %s. Then process again", len(expectedKeys), expectedKeysMembers)
	}

	for i, key := range keys {
		if key != expectedKeys[i] {
			//if key != expectedKeys[i] && !unexpectedform {
			return fmt.Errorf("please enter data in correct order, key name: %s. Then process again", expectedKeysMembers)
		}
	}

	// clear keys after processing
	keys = nil

	return nil
}

// Function to extract keys from JSON string preserving order
func GetOrderedKeysFromJSON(jsonStr []byte) ([]string, error) {
	var keys []string

	// สร้างตัวถอดรหัส JSON
	decoder := json.NewDecoder(strings.NewReader(string(jsonStr)))

	// อ่านโทเคนจาก JSON  โดยต่อเนื่อง
	for {
		// เรียก Token + จัดการข้อผิดพลาด
		token, err := decoder.Token()
		if err != nil {
			break //หยุดลูปถ้าเกิดผิดพลาด
		}

		// ถ้า token เป็น key ให้ append เข้า keys slice
		if key, ok := token.(string); ok {
			keys = append(keys, key)
			// Skip the value token
			_, err := decoder.Token()
			if err != nil {
				break
			}
		}
	}

	return keys, nil
}

// validate taxRequest struct
func validateTaxRequestAmount(req TaxRequest) error {

	// check if TotalIncome value is not number
	if IsNotNumber(req.TotalIncome) {
		return fmt.Errorf("totalIncome must be a non-negative value")
	}

	// check if TotalIncome is a positive value
	if req.TotalIncome < 0 {
		return fmt.Errorf("totalIncome must be a non-negative value")
	}

	// Check if wht value is not number
	if IsNotNumber(req.WHT) {
		return fmt.Errorf("wht must be a non-negative value")
	}

	// Check if WHT is positive value
	if req.WHT < 0 {
		return fmt.Errorf("wht must be a non-negative value")
	}

	if req.WHT > req.TotalIncome {
		return fmt.Errorf("please ensure that Withholding Tax(WHT) not exceed your total income. Let us know if you need any help")
	}

	// check if allowances array is not empty
	if len(req.Allowances) == 0 {
		return fmt.Errorf("at least one allowance must be provided")
	}

	// check each allowance
	for _, allowance := range req.Allowances {
		// Check if AllowanceType is not empty
		if allowance.AllowanceType == "" ||
			(allowance.AllowanceType != "donation" &&
				allowance.AllowanceType != "k-receipt" &&
				allowance.AllowanceType != "personalDeduction") {
			return fmt.Errorf("please ensure that allowanceType inputed correctly")
		}
		// check if Amount is a positive value
		if allowance.Amount < 0 {
			return fmt.Errorf("amount for %s must be a non-negative value", allowance.AllowanceType)
		}
	}

	// no validation errors  return nil
	return nil
}

func caltaxableIncome(TotalIncome, personalExemption, donations, kReceipts float64) float64 {
	taxableIncome := TotalIncome - (personalExemption + donations + kReceipts)
	if taxableIncome < 0 {
		taxableIncome = 0
	}
	return taxableIncome
}
