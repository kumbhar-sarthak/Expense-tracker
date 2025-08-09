package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Expense struct {
	ID          int
	Description string
	Amount      string
	Category    string
	Ptype       string
}

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "models/expenseTracker.sqlite")

	if err != nil {
		log.Fatal("Database not created or open!!")
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS users (
		 google_id TEXT UNIQUE PRIMARY KEY NOT NULL,
		 name TEXT NOT NULL,
		 email TEXT UNIQUE NOT NULL,
		 picture TEXT NOT NULL
	)`)

	if err != nil {
		log.Fatal("user not created",err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS expense (
		 id INTEGER PRIMARY KEY AUTOINCREMENT,
		 user_id TEXT NOT NULL,
		 description TEXT,
		 amount INTEGER,
		 category TEXT,
		 ptype TEXT,
		 FOREIGN KEY(user_id) REFERENCES users(google_id)
	)`)

	if err != nil {
		log.Fatal("Table did not created"," " ,err)
	}

	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign key constraints:", err)
	}
}

func RegisterOrGetUser(googleID, email, name, picture string) (string, error) {
	var id string

	err := DB.QueryRow(`SELECT google_id FROM users WHERE google_id = ?`, googleID).Scan(&id)

	if err == sql.ErrNoRows {
		_, err := DB.Exec(`INSERT INTO users (google_id, email, name, picture) VALUES (?, ?, ?, ?)`, googleID, email, name, picture)
		if err != nil {
			log.Println("Unable to add the new user:", err)
			return "", err
		}
		return googleID, nil
	} else if err != nil {
		log.Println("Some error in DB while looking up user:", err)
		return "", err
	}
	return id, nil
}


func Addexpense(user_id string, des string, amount string, category string, ptype string) error {
	stemp := `INSERT INTO expense (user_id,description,amount,category,ptype) VALUES (?, ?, ?, ?, ?)`

	_, err := DB.Exec(stemp, user_id, des, amount, category, ptype)

	if err != nil {
		log.Println("Can't insert into db", err)
	}
	return err
}

func GetData(user_id string) ([]Expense, error) {

	rows, err := DB.Query(`SELECT id,description,amount,category,ptype FROM expense WHERE user_id = ?`,user_id)

	if err != nil {
		log.Println("Did get user data", " ",err)
	}
	defer rows.Close()

	var expense []Expense
	for rows.Next() {
		var e Expense
		err := rows.Scan(&e.ID, &e.Description, &e.Amount, &e.Category, &e.Ptype)
		if err != nil {
			return nil, err
		}
		expense = append(expense, e)
	}

	return expense, err
}

func Delete(user_id string,expense_id int) error {
	_, err := DB.Exec(`DELETE FROM expense WHERE id = ? AND user_id = ?`,expense_id,user_id)
	return err
}