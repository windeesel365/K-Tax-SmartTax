package pgdb

import "database/sql"

type PersonalDeduction struct {
	ID                int
	PersonalDeduction float64
}

func CreateAdminDeductionsTable(db *sql.DB) error {
	//SQL statement เพื่อ create 'deductions' table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS deductions (
			id SERIAL PRIMARY KEY,
			personal_deduction INTEGER NOT NULL,
			k_receipt_deduction INTEGER NOT NULL
		);
	`
	// execute SQL statement ข้างบน
	_, err := db.Exec(createTableSQL)
	return err
}

func CreateDeduction(db *sql.DB, personalDeduction float64, kReceiptDeduction float64) (int, error) {
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

func GetLowestID(db *sql.DB, tableName string) (int, error) {
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

func GetPersonalDeduction(db *sql.DB, id int) (PersonalDeduction, error) {
	var deduc PersonalDeduction
	row := db.QueryRow(`SELECT id, personal_deduction FROM deductions WHERE id = $1;`, id)
	err := row.Scan(&deduc.ID, &deduc.PersonalDeduction)
	if err != nil {
		return PersonalDeduction{}, err
	}
	return deduc, nil
}

func UpdatePersonalDeduction(db *sql.DB, id int, personalDeduction float64) error {
	_, err := db.Exec(`UPDATE deductions SET personal_deduction = $1 WHERE id = $2;`, personalDeduction, id)
	return err
}
