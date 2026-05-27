package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func OpenDB(filepath string) (*DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) InitSchema(schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}

func (db *DB) SavePosition(driverID int, lat, lon float64) error {
	_, err := db.Exec(
		"INSERT INTO position_history (driver_id, latitude, longitude) VALUES (?, ?, ?)",
		driverID, lat, lon,
	)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		"UPDATE drivers SET last_lat = ?, last_lon = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		lat, lon, driverID,
	)
	return err
}

func (db *DB) GetDriverIDByName(name string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM drivers WHERE name = ?", name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DB) EnsureDriver(name string) (int, error) {
	id, err := db.GetDriverIDByName(name)
	if err == sql.ErrNoRows {
		res, err := db.Exec("INSERT INTO drivers (name, status) VALUES (?, 'available')", name)
		if err != nil {
			return 0, err
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(lastID), nil
	}
	return id, err
}
