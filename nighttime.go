package gonighttime

import (
	"math"
	"time"

	"github.com/kelvins/sunrisesunset"
	"github.com/ringsaturn/tzf"
	tzfrel "github.com/ringsaturn/tzf-rel"
	"github.com/ringsaturn/tzf/pb"
	"google.golang.org/protobuf/proto"
)

type Place struct {
	Lat  float64
	Lon  float64
	Time time.Time
}

type Route struct {
	Departure Place
	Arrival   Place
}

func deg2rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Midpoint calculates a middle point between two coordinates
func Midpoint(start Place, end Place) Place {
	lat1 := deg2rad(start.Lat)
	lon1 := deg2rad(start.Lon)
	lat2 := deg2rad(end.Lat)
	lon2 := deg2rad(end.Lon)

	dlon := lon2 - lon1
	Bx := math.Cos(lat2) * math.Cos(dlon)
	By := math.Cos(lat2) * math.Sin(dlon)
	lat := math.Atan2(math.Sin(lat1)+math.Sin(lat2),
		math.Sqrt((math.Cos(lat1)+Bx)*(math.Cos(lat1)+Bx)+By*By))
	lon := lon1 + math.Atan2(By, (math.Cos(lat1)+Bx))

	lat = (lat * 180) / math.Pi
	lon = (lon * 180) / math.Pi

	return Place{
		Lat: lat,
		Lon: lon,
	}
}

// distance calculates a distance between 2 points
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 = deg2rad(lat1)
	lon1 = deg2rad(lon1)
	lat2 = deg2rad(lat2)
	lon2 = deg2rad(lon2)

	r := 6378100.0
	h := hsin(lat2-lat1) + math.Cos(lat1)*math.Cos(lat2)*hsin(lon2-lon1)

	return 2 * r * math.Asin(math.Sqrt(h)) / 1000 / 1.852 // nautical miles
}

// Distance returns the route distance
func (route *Route) Distance() float64 {
	return distance(route.Departure.Lat, route.Departure.Lon, route.Arrival.Lat, route.Arrival.Lon)
}

// FlightTime calculates total flight time
func (route *Route) FlightTime() time.Duration {
	return route.Arrival.Time.Sub(route.Departure.Time)
}

// Speed calculates average speed in knots
func (route *Route) Speed() float64 {
	return route.Distance() / route.FlightTime().Hours()
}

// GetZone returns time location and UTC offset base on coordinates
func (place *Place) GetZone() (*time.Location, float64) {
	// default location and zone in case of errors
	zone, offset := time.Now().Zone()
	defaultLocation, _ := time.LoadLocation(zone)
	defaultOffset := float64(offset / 3600)

	// get data
	input := &pb.Timezones{}

	dataFile := tzfrel.LiteData
	if err := proto.Unmarshal(dataFile, input); err != nil {
		return defaultLocation, defaultOffset
	}
	finder, _ := tzf.NewFinderFromPB(input)

	// get time zone name by coordinates
	timeZone := finder.GetTimezoneName(place.Lon, place.Lat)

	// get place location
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return defaultLocation, defaultOffset
	}

	_, offset = time.Now().In(location).Zone()

	return location, float64(offset / 3600)
}

// SunriseSunset returns sunrise and sunset times
func (place *Place) SunriseSunset() (time.Time, time.Time) {
	location, offset := place.GetZone()

	params := sunrisesunset.Parameters{
		Latitude:  place.Lat,
		Longitude: place.Lon,
		UtcOffset: offset,
		Date:      place.Time.In(location),
	}

	sunrise, sunset, _ := params.GetSunriseSunset()

	return sunrise, sunset
}

// Sunrise returns aviation sunrise time (-30 minutes from apparent sunrise)
func (place *Place) Sunrise() time.Time {
	s, _ := place.SunriseSunset()
	return s.Add(time.Duration(-30) * time.Minute)
}

// Sunset returns aviation sunset time (+30 minutes from apparent sunset)
func (place *Place) Sunset() time.Time {
	_, s := place.SunriseSunset()
	return s.Add(time.Duration(30) * time.Minute)
}

// MeetWithSun finds the point on the route where airplane meets with Sun (rised or set)
func (route *Route) MeetWithSun(target string) Place {
	maxIterations := 20   // max iteratons, in case some error just not to iterate infinite
	maxDiffMinutes := 1.0 // tolerance in minutes, where we agreed we got the sunset/sunrise

	iter := 0

	var xPoint Place
	diff := time.Duration(0)

	startPoint := route.Departure
	endPoint := route.Arrival

	speed := route.Speed()

	for iter < maxIterations {
		iter++

		xPoint = Midpoint(startPoint, endPoint)

		distance := distance(route.Departure.Lat, route.Departure.Lon, xPoint.Lat, xPoint.Lon)
		flightTime := distance / speed * 60

		xPoint.Time = route.Departure.Time.Add(time.Duration(flightTime) * time.Minute)

		if target == "sunrise" {
			diff = xPoint.Time.Sub(xPoint.Sunrise())
		} else {
			diff = xPoint.Time.Sub(xPoint.Sunset())
		}

		if math.Abs(diff.Minutes()) > maxDiffMinutes {
			if diff.Minutes() > 0 {
				endPoint = xPoint
			} else {
				startPoint = xPoint
			}
		} else {
			break
		}
	}

	return xPoint
}

// NightTime returns a calculated night time
func (route *Route) NightTime() time.Duration {
	nightTime := time.Duration(0)

	rdsr := route.Departure.Sunrise()
	rdss := route.Departure.Sunset()
	rasr := route.Arrival.Sunrise()
	rass := route.Arrival.Sunset()

	if (route.Departure.Time.After(rdsr) && route.Departure.Time.Before(rdss)) &&
		(route.Arrival.Time.After(rasr) && route.Arrival.Time.Before(rass)) {
		// full day flight
		nightTime = time.Duration(0)

	} else if route.Departure.Time.After(rdsr) && route.Departure.Time.Before(rdss) {
		// flight from day to night, night landing
		point := route.MeetWithSun("sunset")
		nightTime = route.Arrival.Time.Sub(point.Time)

	} else if route.Arrival.Time.After(rasr) && route.Arrival.Time.Before(rass) {
		// flight from night to day, day landing
		point := route.MeetWithSun("sunrise")
		nightTime = point.Time.Sub(route.Departure.Time)

	} else {
		// full night time
		nightTime = route.FlightTime()

	}

	return nightTime
}
