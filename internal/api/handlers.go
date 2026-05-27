package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rei-julian/internal/tracker"
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
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming não suportado", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no")

		ch := t.Subscribe()
		defer t.Unsubscribe(ch)

		// envia todas as posições atuais imediatamente
		initialPositions := t.GetPositions()
		for _, pos := range initialPositions {
			initialJSON, _ := json.Marshal(pos)
			fmt.Fprintf(w, "data: %s\n\n", initialJSON)
		}
		flusher.Flush()

		ctx := r.Context()

		for {
			select {
			case <-ctx.Done():
				return
			case pos := <-ch:
				payload, err := json.Marshal(pos)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", payload)
				flusher.Flush()
			}
		}
	}
}
