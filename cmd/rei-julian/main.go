package main

import (
	"log"
	"net/http"
	"rei-julian/internal/api"
	"rei-julian/internal/db"
	"rei-julian/internal/geo"
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

	tr := tracker.NewTracker()

	// 3. Inicia uma frota de rastreios com persistência
	go simulation.StartSimulation(tr, database, "julian", geo.DefaultStartLat, geo.DefaultStartLon)
	go simulation.StartSimulation(tr, database, "mort", geo.DefaultStartLat+0.0005, geo.DefaultStartLon+0.0005)
	go simulation.StartSimulation(tr, database, "maurice", geo.DefaultStartLat-0.0005, geo.DefaultStartLon-0.0005)
	go simulation.StartSimulation(tr, database, "clover", geo.DefaultStartLat+0.0008, geo.DefaultStartLon-0.0008)

	http.HandleFunc("/positions", api.PositionsHandler(tr))
	http.HandleFunc("/stream", api.SSEHandler(tr))

	// SERVIR FRONTEND
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", fs)

	log.Println("Servidor em http://localhost:8080")
	log.Println("Banco de Dados: reijulian.db pronto.")
	log.Println("GET /positions -> posições atuais em JSON")
	log.Println("GET /stream    -> stream SSE com atualizações")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
