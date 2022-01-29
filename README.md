Go package to calculate a night time for the flight

## General

It's a simplified way to calculate where/what time an airplane and sun can meet. It doesn't compensate sunset/sunrise times for airplane altitude.

The logic is simple. It finds a midpoint on the route, check the time when airplane will be there and compares it with sunset/sunrise at this point. 
If the difference is too big it takes front/rear part of the route and check the midpoint again and again...

## Known issues

It will not properly calculate a night time if the departure is before sunset and the arrival is after sunrise

## Usage

```golang
package main

import (
	nighttime "github.com/vsimakhin/go-nighttime"
	"time"
	"fmt"
)

func main() {

	// flight from LEPA to ESMX
	route := nighttime.Route{
		Departure: nighttime.Place{
			Lat:  39.551700592,
			Lon:  2.7388100624,
			Time: time.Date(2021, 12, 8, 5, 4, 0, 0, time.UTC),
		},
		Arrival: nighttime.Place{
			Lat:  56.9291000366,
			Lon:  14.7279996872,
			Time: time.Date(2021, 12, 8, 7, 53, 0, 0, time.UTC),
		},
	}

	fmt.Println("Flight time:", route.FlightTime())
	fmt.Println("Distance (nm):", route.Distance())
	fmt.Println("Night time:", route.NightTime())
}
```


```bash
$ go run main.go 
Flight time: 2h49m0s
Distance (nm): 1145.6113996245742
Night time: 1h35m0s

```

## Used modules

* sunrisesunset https://github.com/kelvins/sunrisesunset
