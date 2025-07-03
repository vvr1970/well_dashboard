package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// InitDB инициализирует подключение к PostgreSQL
func InitDB() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Создание таблицы, если её нет
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS wells (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			depth DECIMAL(10,2),
			location VARCHAR(255),
			status VARCHAR(50),
			productivity DECIMAL(10,2),
			drilling_date DATE,
			field VARCHAR(100),
			operator VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}

// AllWells возвращает все скважины из БД
func AllWells(db *sql.DB) ([]Well, error) {
	rows, err := db.Query(`
		SELECT id, name, depth, location, status, productivity, 
		       drilling_date, field, operator 
		FROM wells 
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wells []Well
	for rows.Next() {
		var w Well
		err := rows.Scan(
			&w.ID, &w.Name, &w.Depth, &w.Location, &w.Status,
			&w.Productivity, &w.DrillingDate, &w.Field, &w.Operator,
		)
		if err != nil {
			return nil, err
		}
		wells = append(wells, w)
	}

	return wells, nil
}

// GetWell возвращает скважину по ID
func GetWell(db *sql.DB, id int) (*Well, error) {
	var w Well
	err := db.QueryRow(`
		SELECT id, name, depth, location, status, productivity, 
		       drilling_date, field, operator 
		FROM wells 
		WHERE id = $1
	`, id).Scan(
		&w.ID, &w.Name, &w.Depth, &w.Location, &w.Status,
		&w.Productivity, &w.DrillingDate, &w.Field, &w.Operator,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// UpdateWell обновляет данные скважины
func UpdateWell(db *sql.DB, w *Well) error {
	_, err := db.Exec(`
		UPDATE wells 
		SET name = $1, depth = $2, location = $3, status = $4, 
		    productivity = $5, field = $6, operator = $7, 
		    updated_at = CURRENT_TIMESTAMP 
		WHERE id = $8
	`, w.Name, w.Depth, w.Location, w.Status,
		w.Productivity, w.Field, w.Operator, w.ID)
	return err
}

func CreateWell(db *sql.DB, well *Well) error {
	err := db.QueryRow(`
        INSERT INTO wells (name, depth, location, status, productivity,
                          drilling_date, field, operator)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, created_at, updated_at
    `, well.Name, well.Depth, well.Location, well.Status,
		well.Productivity, well.DrillingDate, well.Field, well.Operator).
		Scan(&well.ID, &well.CreatedAt, &well.UpdatedAt)

	return err
}
