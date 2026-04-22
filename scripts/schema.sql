-- Schema for Rei Julian Logistics Tracking
-- Database: SQLite

-- 1. Table for Drivers (Entregadores)
CREATE TABLE IF NOT EXISTS drivers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    status TEXT CHECK(status IN ('available', 'busy', 'offline')) DEFAULT 'offline',
    last_lat REAL,
    last_lon REAL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. Table for Orders (Pedidos)
CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_name TEXT NOT NULL,
    delivery_address TEXT,
    status TEXT CHECK(status IN ('preparing', 'in_route', 'delivered', 'cancelled')) DEFAULT 'preparing',
    driver_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (driver_id) REFERENCES drivers(id)
);

-- 3. Table for Position History (Histórico de Movimentação)
-- This allows for path reconstruction and delivery auditing.
CREATE TABLE IF NOT EXISTS position_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    driver_id INTEGER NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (driver_id) REFERENCES drivers(id)
);

-- Initial Seed Data for Testing
INSERT INTO drivers (name, status) VALUES ('João Motoboy', 'available');
INSERT INTO drivers (name, status) VALUES ('Maria Entrega', 'offline');

INSERT INTO orders (client_name, delivery_address, status, driver_id) 
VALUES ('Restaurante Central', 'Rua XV de Novembro, Pelotas', 'preparing', 1);
