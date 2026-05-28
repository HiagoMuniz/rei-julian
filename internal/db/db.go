package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

func (db *DB) SavePosition(driverID int, orderID int, lat, lon float64) error {
	var err error
	if orderID > 0 {
		_, err = db.Exec(
			"INSERT INTO position_history (driver_id, order_id, latitude, longitude) VALUES (?, ?, ?, ?)",
			driverID, orderID, lat, lon,
		)
	} else {
		_, err = db.Exec(
			"INSERT INTO position_history (driver_id, latitude, longitude) VALUES (?, ?, ?)",
			driverID, lat, lon,
		)
	}
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
		res, err := db.Exec("INSERT INTO drivers (name, status, last_lat, last_lon) VALUES (?, 'available', -31.770687, -52.341350)", name)
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

type Order struct {
	ID        int
	DriverID  int
	Client    string
	Token     string
	Status    string
	CreatedAt string
}

type PositionHistory struct {
	Lat       float64 `json:"latitude"`
	Lon       float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}



func (db *DB) GenerateToken() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (db *DB) CreateOrder(driverID int, client string) (string, int, error) {
	token := db.GenerateToken()
	res, err := db.Exec("INSERT INTO orders (driver_id, client_name, tracking_token, status) VALUES (?, ?, ?, 'preparing')", driverID, client, token)
	if err != nil {
		return "", 0, err
	}
	id, err := res.LastInsertId()
	return token, int(id), err
}

func (db *DB) GetOrderByToken(token string) (Order, error) {
	var o Order
	err := db.QueryRow("SELECT id, driver_id, client_name, tracking_token, status, created_at FROM orders WHERE tracking_token = ?", token).Scan(&o.ID, &o.DriverID, &o.Client, &o.Token, &o.Status, &o.CreatedAt)
	return o, err
}

func (db *DB) UpdateOrderStatus(orderID int, status string) error {
	_, err := db.Exec("UPDATE orders SET status = ? WHERE id = ?", status, orderID)
	return err
}

func (db *DB) GetOrderHistory(token string, limit int) ([]PositionHistory, error) {
	o, err := db.GetOrderByToken(token)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		"SELECT latitude, longitude, timestamp FROM position_history WHERE order_id = ? ORDER BY timestamp DESC LIMIT ?",
		o.ID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []PositionHistory
	for rows.Next() {
		var p PositionHistory
		if err := rows.Scan(&p.Lat, &p.Lon, &p.Timestamp); err != nil {
			return nil, err
		}
		history = append(history, p)
	}

	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

func (db *DB) GetDriverHistory(name string, limit int) ([]PositionHistory, error) {
	driverID, err := db.GetDriverIDByName(name)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		"SELECT latitude, longitude, timestamp FROM position_history WHERE driver_id = ? ORDER BY timestamp DESC LIMIT ?",
		driverID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []PositionHistory
	for rows.Next() {
		var p PositionHistory
		if err := rows.Scan(&p.Lat, &p.Lon, &p.Timestamp); err != nil {
			return nil, err
		}
		history = append(history, p)
	}

	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

type Driver struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	LastLat   float64 `json:"last_lat"`
	LastLon   float64 `json:"last_lon"`
	UpdatedAt string  `json:"updated_at"`
}

func (db *DB) GetAllDrivers() ([]Driver, error) {
	rows, err := db.Query("SELECT id, name, status, COALESCE(last_lat, 0), COALESCE(last_lon, 0), updated_at FROM drivers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []Driver
	for rows.Next() {
		var d Driver
		if err := rows.Scan(&d.ID, &d.Name, &d.Status, &d.LastLat, &d.LastLon, &d.UpdatedAt); err != nil {
			return nil, err
		}
		drivers = append(drivers, d)
	}
	return drivers, nil
}
