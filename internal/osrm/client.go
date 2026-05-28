package osrm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RouteResponse is the root structure returned by OSRM API
type RouteResponse struct {
	Code   string  `json:"code"`
	Routes []Route `json:"routes"`
}

type Route struct {
	Geometry Geometry `json:"geometry"`
}

type Geometry struct {
	Coordinates [][]float64 `json:"coordinates"` // [lon, lat] pairs
}

// GetRoute fetches a driving route from start to end coordinates using OSRM
func GetRoute(startLat, startLon, endLat, endLon float64) ([][]float64, error) {
	// Format: {lon},{lat};{lon},{lat}
	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		startLon, startLat, endLon, endLat)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to OSRM: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM returned status %d", resp.StatusCode)
	}

	var routeResp RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return nil, fmt.Errorf("failed to decode OSRM response: %v", err)
	}

	if routeResp.Code != "Ok" || len(routeResp.Routes) == 0 {
		return nil, fmt.Errorf("no routes found in OSRM response")
	}

	// Returns an array of [lon, lat] points that makeup the route polyline
	return routeResp.Routes[0].Geometry.Coordinates, nil
}
