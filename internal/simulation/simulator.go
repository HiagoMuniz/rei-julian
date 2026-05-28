package simulation

import (
	"log"
	"math"
	"rei-julian/internal/db"
	"rei-julian/internal/geo"
	"rei-julian/internal/osrm"
	"rei-julian/internal/tracker"
	"sync"
	"time"
)

type Job struct {
	DestLat float64
	DestLon float64
	Token   string
	OrderID int
}

var (
	driverQueues = make(map[string]chan Job)
	queuesMutex  sync.Mutex
)

// Coordenadas fixas do restaurante (ponto de origem)
const RestaurantLat = -31.770687
const RestaurantLon = -52.341350

// Distância Euclidiana plana (simplificada para pequenos segmentos)
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	dlat := lat1 - lat2
	dlon := lon1 - lon2
	return math.Sqrt(dlat*dlat + dlon*dlon)
}



// EnsureDriverWorker garante que existe uma goroutine rodando para o entregador
func EnsureDriverWorker(t *tracker.Tracker, database *db.DB, name string, driverID int) {
	queuesMutex.Lock()
	defer queuesMutex.Unlock()

	if _, exists := driverQueues[name]; !exists {
		ch := make(chan Job, 50) // Buffer para 50 pedidos
		driverQueues[name] = ch
		go driverWorker(t, database, name, driverID, ch)
	}
}

// DispatchJob adiciona um pedido na fila do entregador
func DispatchJob(name string, job Job) {
	queuesMutex.Lock()
	ch, exists := driverQueues[name]
	queuesMutex.Unlock()

	if exists {
		ch <- job
	} else {
		log.Printf("Aviso: Tentativa de despachar para entregador sem worker rodando: %s", name)
	}
}

// driverWorker processa a fila de entregas do entregador
func driverWorker(t *tracker.Tracker, database *db.DB, name string, driverID int, ch chan Job) {
	currentLat := RestaurantLat
	currentLon := RestaurantLon
	step := 0

	t.SetPosition(tracker.Position{
		ID:        name,
		Token:     "",
		Latitude:  currentLat,
		Longitude: currentLon,
		TargetLat: currentLat,
		TargetLon: currentLon,
		Status:    "available",
		Timestamp: time.Now(),
		Step:      step,
	})
	database.SavePosition(driverID, 0, currentLat, currentLon)

	for {
		job := <-ch

	processJobs:
		for {
			// Atualiza status para in_route com destino novo
			pos, _ := t.GetPosition(name)
			currentLat = pos.Latitude
			currentLon = pos.Longitude
			step = pos.Step + 1

			t.SetPosition(tracker.Position{
				ID:        name,
				Token:     job.Token,
				Latitude:  currentLat,
				Longitude: currentLon,
				TargetLat: job.DestLat,
				TargetLon: job.DestLon,
				Status:    "in_route",
				Timestamp: time.Now(),
				Step:      step,
			})
			database.SavePosition(driverID, job.OrderID, currentLat, currentLon)
			log.Printf("[%s] Iniciando entrega %d...", name, job.OrderID)

			// Pega Rota do OSRM
			route, err := osrm.GetRoute(currentLat, currentLon, job.DestLat, job.DestLon)
			if err != nil || len(route) == 0 {
				log.Printf("[%s] Aviso: OSRM falhou (%v), usando linha reta", name, err)
				route = [][]float64{{job.DestLon, job.DestLat}}
			}
			routeIdx := 0

			// Espera motorista chegar via GPS Real (ou botão concluído) ou simula movimento
			ticker := time.NewTicker(2 * time.Second)
			sub := t.Subscribe()
		deliveryLoop:
			for {
				select {
				case p := <-sub:
					if p.ID == name {
						currentLat = p.Latitude
						currentLon = p.Longitude
						step = p.Step
						if p.Status == "delivered" || distance(currentLat, currentLon, job.DestLat, job.DestLon) < 0.0005 {
							break deliveryLoop
						}
					}
				case <-ticker.C:
					if distance(currentLat, currentLon, job.DestLat, job.DestLon) < 0.0005 {
						break deliveryLoop
					}
					
					// Avançar nos pontos da rota OSRM
					if routeIdx < len(route) {
						targetLon := route[routeIdx][0]
						targetLat := route[routeIdx][1]
						
						distToNode := distance(currentLat, currentLon, targetLat, targetLon)
						if distToNode < 0.0005 {
							routeIdx++
							if routeIdx < len(route) {
								targetLon = route[routeIdx][0]
								targetLat = route[routeIdx][1]
							}
						}
						
						if routeIdx < len(route) {
							bearing := math.Atan2(
								math.Sin(geo.DegreesToRadians(targetLon-currentLon))*math.Cos(geo.DegreesToRadians(targetLat)),
								math.Cos(geo.DegreesToRadians(currentLat))*math.Sin(geo.DegreesToRadians(targetLat))-math.Sin(geo.DegreesToRadians(currentLat))*math.Cos(geo.DegreesToRadians(targetLat))*math.Cos(geo.DegreesToRadians(targetLon-currentLon)),
							)
							
							currentLat, currentLon = geo.MovePoint(currentLat, currentLon, 50.0, bearing)
							step++
							
							t.SetPosition(tracker.Position{
								ID:        name,
								Token:     job.Token,
								Latitude:  currentLat,
								Longitude: currentLon,
								TargetLat: job.DestLat,
								TargetLon: job.DestLon,
								Status:    "in_route",
								Timestamp: time.Now(),
								Step:      step,
							})
							database.SavePosition(driverID, job.OrderID, currentLat, currentLon)
						}
					}
				}
			}
			ticker.Stop()
			t.Unsubscribe(sub)

			// Chegou no destino
			step++
			t.SetPosition(tracker.Position{
				ID:        name,
				Token:     job.Token,
				Latitude:  currentLat,
				Longitude: currentLon,
				TargetLat: currentLat,
				TargetLon: currentLon,
				Status:    "delivered",
				Timestamp: time.Now(),
				Step:      step,
			})
			database.SavePosition(driverID, job.OrderID, currentLat, currentLon)
			database.UpdateOrderStatus(job.OrderID, "delivered")
			log.Printf("[%s] Entrega concluída no pedido %d.", name, job.OrderID)

			time.Sleep(3 * time.Second) // Breve pausa

			// Retornando à base
			step++
			t.SetPosition(tracker.Position{
				ID:        name,
				Token:     "",
				Latitude:  currentLat,
				Longitude: currentLon,
				TargetLat: RestaurantLat,
				TargetLon: RestaurantLon,
				Status:    "returning",
				Timestamp: time.Now(),
				Step:      step,
			})
			log.Printf("[%s] Retornando à base...", name)

			// Pega Rota do OSRM para a base
			routeRet, errRet := osrm.GetRoute(currentLat, currentLon, RestaurantLat, RestaurantLon)
			if errRet != nil || len(routeRet) == 0 {
				routeRet = [][]float64{{RestaurantLon, RestaurantLat}}
			}
			routeIdxRet := 0

			// Espera chegar na base ou simula movimento
			tickerRet := time.NewTicker(2 * time.Second)
			subRet := t.Subscribe()
		returnLoop:
			for {
				select {
				case p := <-subRet:
					if p.ID == name {
						currentLat = p.Latitude
						currentLon = p.Longitude
						step = p.Step
						if p.Status == "available" || distance(currentLat, currentLon, RestaurantLat, RestaurantLon) < 0.0005 {
							break returnLoop
						}
					}
				case <-tickerRet.C:
					if distance(currentLat, currentLon, RestaurantLat, RestaurantLon) < 0.0005 {
						break returnLoop
					}
					
					if routeIdxRet < len(routeRet) {
						targetLon := routeRet[routeIdxRet][0]
						targetLat := routeRet[routeIdxRet][1]
						
						distToNode := distance(currentLat, currentLon, targetLat, targetLon)
						if distToNode < 0.0005 {
							routeIdxRet++
							if routeIdxRet < len(routeRet) {
								targetLon = routeRet[routeIdxRet][0]
								targetLat = routeRet[routeIdxRet][1]
							}
						}
						
						if routeIdxRet < len(routeRet) {
							bearing := math.Atan2(
								math.Sin(geo.DegreesToRadians(targetLon-currentLon))*math.Cos(geo.DegreesToRadians(targetLat)),
								math.Cos(geo.DegreesToRadians(currentLat))*math.Sin(geo.DegreesToRadians(targetLat))-math.Sin(geo.DegreesToRadians(currentLat))*math.Cos(geo.DegreesToRadians(targetLat))*math.Cos(geo.DegreesToRadians(targetLon-currentLon)),
							)
							
							currentLat, currentLon = geo.MovePoint(currentLat, currentLon, 50.0, bearing)
							step++
							
							t.SetPosition(tracker.Position{
								ID:        name,
								Token:     "",
								Latitude:  currentLat,
								Longitude: currentLon,
								TargetLat: RestaurantLat,
								TargetLon: RestaurantLon,
								Status:    "returning",
								Timestamp: time.Now(),
								Step:      step,
							})
							database.SavePosition(driverID, 0, currentLat, currentLon)
						}
					}
				}
			}
			tickerRet.Stop()
			t.Unsubscribe(subRet)

			// Chegou no Restaurante
			step++
			currentLat, currentLon = RestaurantLat, RestaurantLon
			t.SetPosition(tracker.Position{
				ID:        name,
				Token:     "",
				Latitude:  currentLat,
				Longitude: currentLon,
				TargetLat: currentLat,
				TargetLon: currentLon,
				Status:    "available",
				Timestamp: time.Now(),
				Step:      step,
			})
			database.SavePosition(driverID, 0, currentLat, currentLon)
			log.Printf("[%s] Na base. Disponível.", name)

			// Verifica se tem mais trabalho na fila
			select {
			case nextJob := <-ch:
				log.Printf("[%s] Novo pedido recebido na base! Saindo...", name)
				job = nextJob
				continue processJobs
			default:
				log.Printf("[%s] Sem entregas na fila. Aguardando...", name)
				break processJobs
			}
		}
	}
}
