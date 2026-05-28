package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rei-julian/internal/db"
	"rei-julian/internal/simulation"
	"rei-julian/internal/tracker"
	"strings"
	"time"
)

func PositionsHandler(t *tracker.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")

		pos := t.GetPositions()
		if err := json.NewEncoder(w).Encode(pos); err != nil {
			http.Error(w, "erro ao serializar posições", http.StatusInternalServerError)
			return
		}
	}
}

func SSEHandler(t *tracker.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming não suportado", http.StatusInternalServerError)
			return
		}

		ch := t.Subscribe()
		defer t.Unsubscribe(ch)

		for {
			select {
			case <-r.Context().Done():
				return
			case p := <-ch:
				data, err := json.Marshal(p)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}

// DriversHandler retorna a lista de todos os entregadores
func DriversHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		drivers, err := database.GetAllDrivers()
		if err != nil {
			http.Error(w, "erro ao buscar entregadores", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		json.NewEncoder(w).Encode(drivers)
	}
}

// HistoryHandler retorna o histórico de posições de um pedido ou entregador
func HistoryHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "Token ou ID não fornecido", http.StatusBadRequest)
			return
		}
		idOrToken := pathParts[3]

		// 1. Tenta buscar como um Token de Pedido (Para a Aba do Cliente)
		history, err := database.GetOrderHistory(idOrToken, 100)
		if err != nil || len(history) == 0 {
			// 2. Se falhar ou estiver vazio, tenta buscar como Nome do Entregador (Para a Aba Admin)
			history, err = database.GetDriverHistory(idOrToken, 200)
			if err != nil {
				http.Error(w, "Nenhum histórico encontrado", http.StatusNotFound)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}
}

// OrderHandler retorna detalhes de um pedido pelo token
func OrderHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 4 {
			http.Error(w, "Token não fornecido", http.StatusBadRequest)
			return
		}
		token := pathParts[3]

		order, err := database.GetOrderByToken(token)
		if err != nil {
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}

type LocationPayload struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
}

func LocationHandler(t *tracker.Tracker, database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var payload LocationPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Payload inválido", http.StatusBadRequest)
			return
		}

		driverID, err := database.EnsureDriver(payload.ID)
		if err != nil {
			http.Error(w, "Erro ao registrar driver", http.StatusInternalServerError)
			return
		}

		current, exists := t.GetPosition(payload.ID)
		step := 0
		if exists {
			step = current.Step + 1
		}

		status := payload.Status
		if status == "" && exists {
			status = current.Status
		}

		pos := tracker.Position{
			ID:        payload.ID,
			Latitude:  payload.Latitude,
			Longitude: payload.Longitude,
			TargetLat: current.TargetLat,
			TargetLon: current.TargetLon,
			Status:    status,
			Timestamp: time.Now(),
			Step:      step,
		}

		t.SetPosition(pos)

		if err := database.SavePosition(driverID, 0, payload.Latitude, payload.Longitude); err != nil {
			http.Error(w, "Erro ao salvar posição", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

type DispatchPayload struct {
	DriverName string  `json:"driver_name"`
	DestLat    float64 `json:"dest_lat"`
	DestLon    float64 `json:"dest_lon"`
}

// DispatchHandler inicia uma nova simulação OSRM dinamicamente
func DispatchHandler(t *tracker.Tracker, database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var payload DispatchPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Payload inválido", http.StatusBadRequest)
			return
		}

		if payload.DriverName == "" || payload.DestLat == 0 || payload.DestLon == 0 {
			http.Error(w, "Parâmetros incompletos", http.StatusBadRequest)
			return
		}

		driverID, err := database.EnsureDriver(payload.DriverName)
		if err != nil {
			http.Error(w, "Erro ao registrar driver no banco", http.StatusInternalServerError)
			return
		}

		simulation.EnsureDriverWorker(t, database, payload.DriverName, driverID)

		token, orderID, err := database.CreateOrder(driverID, "Cliente Genérico")
		if err != nil {
			http.Error(w, "Erro ao criar pedido", http.StatusInternalServerError)
			return
		}

		job := simulation.Job{
			DestLat: payload.DestLat,
			DestLon: payload.DestLon,
			Token:   token,
			OrderID: orderID,
		}

		simulation.DispatchJob(payload.DriverName, job)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "dispatched", "driver": payload.DriverName, "token": token})
	}
}

type CreateDriverPayload struct {
	Name string `json:"name"`
}

func CreateDriverHandler(t *tracker.Tracker, database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var payload CreateDriverPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name == "" {
			http.Error(w, "Payload inválido", http.StatusBadRequest)
			return
		}

		driverID, err := database.EnsureDriver(payload.Name)
		if err != nil {
			http.Error(w, "Erro ao registrar driver no banco", http.StatusInternalServerError)
			return
		}

		// Inicializa a goroutine e joga o driver pra "available" na base
		simulation.EnsureDriverWorker(t, database, payload.Name, driverID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "created", "name": payload.Name, "id": driverID})
	}
}
