package geo

import (
	"math"
)

const (
	EarthRadiusMeters = 6371000.0
)

// MovePoint desloca um ponto geográfico por certa distância em uma direção dada.
// Fórmula de navegação sobre esfera.
func MovePoint(latDeg, lonDeg, distanceMeters, bearingRad float64) (float64, float64) {
	lat1 := degreesToRadians(latDeg)
	lon1 := degreesToRadians(lonDeg)
	angularDistance := distanceMeters / EarthRadiusMeters

	lat2 := math.Asin(
		math.Sin(lat1)*math.Cos(angularDistance) +
			math.Cos(lat1)*math.Sin(angularDistance)*math.Cos(bearingRad),
	)

	lon2 := lon1 + math.Atan2(
		math.Sin(bearingRad)*math.Sin(angularDistance)*math.Cos(lat1),
		math.Cos(angularDistance)-math.Sin(lat1)*math.Sin(lat2),
	)

	return radiansToDegrees(lat2), NormalizeLongitude(radiansToDegrees(lon2))
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180.0
}

func radiansToDegrees(r float64) float64 {
	return r * 180.0 / math.Pi
}

func NormalizeLongitude(lon float64) float64 {
	for lon > 180.0 {
		lon -= 360.0
	}
	for lon < -180.0 {
		lon += 360.0
	}
	return lon
}
