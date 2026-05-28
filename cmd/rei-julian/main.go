package main

import (
	"log"
	"net/http"
	"rei-julian/internal/api"
	"rei-julian/internal/db"
	"rei-julian/internal/simulation"
	"rei-julian/internal/tracker"
)

func main() {
	// 1. Inicializa o Banco de Dados
	database, err := db.OpenDB("reijulian.db")
	if err != nil {
		log.Fatal("Erro ao abrir banco de dados:", err)
	}

	// 2. Inicializa o Schema
	if err := database.InitSchema("scripts/schema.sql"); err != nil {
		log.Fatal("Erro ao inicializar schema:", err)
	}

	// 3. Limpa rastros de execuções anteriores para não poluir o mapa
	database.Exec("DELETE FROM position_history")
	database.Exec("DELETE FROM orders")
	database.Exec("UPDATE drivers SET status = 'offline', last_lat = -31.770687, last_lon = -52.341350")

	tr := tracker.NewTracker()

	// 1. Garante que os padrão existem no banco
	defaultDrivers := []string{"Julian", "Mort", "Maurice"}
	for _, name := range defaultDrivers {
		database.EnsureDriver(name)
	}

	// 2. Inicia os workers para TODOS os entregadores registrados no banco
	rows, err := database.Query("SELECT id, name FROM drivers")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				simulation.EnsureDriverWorker(tr, database, name, id)
			}
		}
	}

	http.HandleFunc("/positions", api.PositionsHandler(tr))
	http.HandleFunc("/stream", api.SSEHandler(tr))
	http.HandleFunc("/api/drivers", api.DriversHandler(database))
	http.HandleFunc("/api/history/", api.HistoryHandler(database))
	http.HandleFunc("/api/order/", api.OrderHandler(database))
	http.HandleFunc("/api/location", api.LocationHandler(tr, database))
	http.HandleFunc("/api/simulation/start", api.DispatchHandler(tr, database))
	http.HandleFunc("/api/drivers/create", api.CreateDriverHandler(tr, database))

	// SERVIR FRONTEND
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", fs)

	log.Println("Servidor em http://localhost:8080")
	log.Println("Banco de Dados: reijulian.db pronto.")
	log.Println("GET /api/drivers -> lista de motoristas")
	log.Println("GET /api/history/{id} -> historico do motorista")
	log.Println("POST /api/location -> atualiza localizacao")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

