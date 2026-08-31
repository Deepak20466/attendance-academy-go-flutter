package attendance

import "math"

const earthRadiusMeters = 6371000

// HaversineDistanceMeters computes the great-circle distance between two
// lat/lng points, used to verify a coach is physically at the assigned
// location before their check-in/out counts as geofence-verified.
func HaversineDistanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }

	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
