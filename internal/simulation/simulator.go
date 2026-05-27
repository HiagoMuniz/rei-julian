package simulation

import (
	"log"
	"math"
	"math/rand"
	"rei-julian/internal/db"
	"rei-julian/internal/geo"
	"rei-julian/internal/tracker"
	"time"
)

// Gera um destino aleatório dentro de um raio a partir de um ponto
func generateRandomDestination(rng *rand.Rand, baseLat, baseLon, radiusMeters float64) (float64, float64) {
	bearing := rng.Float64() * 2 * math.Pi
	distance := rng.Float64() * radiusMeters
	return geo.MovePoint(baseLat, baseLon, distance, bearing)
}

func StartSimulation(t *tracker.Tracker, database *db.DB, name string, startLat, startLon float64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(len(name))))

	driverID, err := database.EnsureDriver(name)
	if err != nil {
		log.Printf("Erro ao assegurar entregador %s no banco: %v", name, err)
		return
	}

	// Define um destino alvo em um raio de ~2km
	targetLat, targetLon := generateRandomDestination(rng, geo.DefaultStartLat, geo.DefaultStartLon, 2000.0)

	t.SetPosition(tracker.Position{
		ID:        name,
		Latitude:  startLat,
		Longitude: startLon,
		TargetLat: targetLat,
		TargetLon: targetLon,
		Timestamp: time.Now(),
		Step:      0,
	})

	// Salva posição inicial
	database.SavePosition(driverID, startLat, startLon)

	for {
		delay := time.Duration(3+rng.Float64()*2) * time.Second
		time.Sleep(delay)

		current, _ := t.GetPosition(name)
		
		// Direciona o movimento ligeiramente para o alvo
		bearingToTarget := math.Atan2(
			math.Sin(current.TargetLon-current.Longitude)*math.Cos(current.TargetLat),
			math.Cos(current.Latitude)*math.Sin(current.TargetLat)-math.Sin(current.Latitude)*math.Cos(current.TargetLat)*math.Cos(current.TargetLon-current.Longitude),
		)
		
		// Adiciona um pouco de "ruído" à direção para não ser uma linha perfeitamente reta
		noise := (rng.Float64() - 0.5) * (math.Pi / 4) // +/- 22.5 graus de ruído
		bearing := bearingToTarget + noise

		newLat, newLon := geo.MovePoint(current.Latitude, current.Longitude, geo.StepMeters * 2, bearing) // Passos um pouco maiores

		next := tracker.Position{
			ID:        name,
			Latitude:  newLat,
			Longitude: newLon,
			TargetLat: current.TargetLat, // Mantém o destino
			TargetLon: current.TargetLon,
			Timestamp: time.Now(),
			Step:      current.Step + 1,
		}

		t.SetPosition(next)

		if err := database.SavePosition(driverID, newLat, newLon); err != nil {
			log.Printf("[%s] Erro ao salvar no banco: %v", name, err)
		}
	}
}
