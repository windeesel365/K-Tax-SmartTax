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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/windeesel365/assessment-tax/handlefileupload"
	"github.com/windeesel365/assessment-tax/handletax"
	"github.com/windeesel365/assessment-tax/pgdb"
	"github.com/windeesel365/assessment-tax/sharedvars"
	"github.com/windeesel365/assessment-tax/validityguard"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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

type TaxLevel struct {
	Level string        `json:"level"`
	Tax   CustomFloat64 `json:"tax"`
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

type KReceiptDeduction struct {
	ID                        int
	UpperLimKReceiptDeduction float64
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// get port number from the environment variable 'PORT'
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
	// Retrieve DATABASE_URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("DATABASE_URL is not set")
		return
	}

	// สร้าง connection กับ postgresql
	var err error
	sharedvars.Db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer sharedvars.Db.Close()

	// Check the connection
	err = sharedvars.Db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PostgreSQL: Connected successfully.")

	// create 'deductions' table เพื่อเตรียมสำหรับ admin ปรับค่า
	err = pgdb.CreateAdminDeductionsTable(sharedvars.Db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PostgreSQL: Created admin deductions table successfully.")

	// count คือจำนวน row ใน table ณ จุดเวลา
	count, err := pgdb.CountRows(sharedvars.Db, "deductions")
	if err != nil {
		log.Fatal(err)
	}

	// ถ้า count น้อยกว่า 1, ถึงยอมให้สร้าง row ใหม่
	if count < 1 {
		id, err := pgdb.CreateDeduction(sharedvars.Db, sharedvars.InitialPersonalExemption, sharedvars.KReceiptsUpperLimit)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("PostgreSQL: Created admin deductions row with id: %d, initialPersonalExemption: %f\n kReceiptsUpperLimit: %f\n", id, sharedvars.InitialPersonalExemption, sharedvars.KReceiptsUpperLimit)
	} //else ไปเช็คค่าใน db

	// ใช้ row แรกสุด คือ id แรกสุด ในการ update และ read
	lowestID, err := pgdb.GetLowestID(sharedvars.Db, "deductions")
	if err != nil {
		panic(err.Error())
	}

	sharedvars.Id = lowestID // id เป็น id สำหรับ ref ข้อมูล postgresql

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

	e.POST("/tax/calculations", handletax.HandleTaxCalculation)
	e.POST("/tax/calculations/upload-csv", handlefileupload.HandleFileUpload)

	adminGroup := e.Group("/admin")
	adminGroup.Use(basicAuthMiddleware)

	adminGroup.POST("/login", login)

	adminGroup.POST("/deductions/personal", setPersonalDeduction)
	adminGroup.POST("/deductions/k-receipt", setKReceiptDeduction)

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

// POST: /admin/deductions/personal
func setPersonalDeduction(c echo.Context) error {
	// read Body ให้เป็น variable
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	defer c.Request().Body.Close()

	// validation function process
	if err := validityguard.ValidatePersonalInput(body); err != nil {
		return err
	}

	// หลังจากการ validation
	// bind JSON to struct
	d := new(Deduction)
	if err := json.Unmarshal(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	// set change to initialPersonalExemption
	sharedvars.InitialPersonalExemption = d.Amount

	// update PersonalDeduction to postgres db
	err = pgdb.UpdatePersonalDeduction(sharedvars.Db, sharedvars.Id, sharedvars.InitialPersonalExemption)
	if err != nil {
		log.Fatal(err)
	}
	// read หลังการ update ทำ log ในระบบ
	adminPDeductions, err := pgdb.GetPersonalDeduction(sharedvars.Db, sharedvars.Id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("***********\nAdmin updated initialPersonalExemption validated then updated postgresql row: %+v\n", adminPDeductions)

	//respond client(admin)
	return c.JSON(http.StatusOK, map[string]CustomFloat64{"personalDeduction": CustomFloat64(d.Amount)})
}

// POST: /admin/deductions/k-receipt
func setKReceiptDeduction(c echo.Context) error {
	// Read body to a variable
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}
	defer c.Request().Body.Close()

	// function to combine validation
	if err := validityguard.ValidateInputsetKReceipt(body); err != nil {
		return err
	}

	// หลังจากการ validation
	// Bind JSON to struct
	d := new(Deduction)
	if err := json.Unmarshal(body, d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input format: "+err.Error())
	}

	// Set change to kReceiptsUpperLimit
	sharedvars.KReceiptsUpperLimit = d.Amount

	// postgresql part
	// update KReceiptDeduction to postgres db
	err = pgdb.UpdateKReceiptDeduction(sharedvars.Db, sharedvars.Id, sharedvars.KReceiptsUpperLimit)
	if err != nil {
		log.Fatal(err)
	}
	// read หลังการ update ทำ log ในระบบ
	adminKDeductions, err := pgdb.GetKReceiptDeduction(sharedvars.Db, sharedvars.Id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("***********\nAdmin updated kReceiptsUpperLimit validated then updated postgresql row: %+v\n", adminKDeductions)

	//respond to client(admin)
	return c.JSON(http.StatusOK, map[string]CustomFloat64{
		"kReceipt": CustomFloat64(d.Amount)})

}
