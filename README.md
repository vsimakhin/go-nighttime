## Arhived

The repository is archived. For the flight night-time calculation check https://github.com/vsimakhin/web-logbook/blob/main/internal/nighttime/nighttime.go

## General

It's a simplified way to calculate where/what time an airplane and the sun can meet. It doesn't compensate for sunset/sunrise times for airplane altitude.

The logic is simple. It finds a midpoint on the route, checks the time when the airplane will be there and compares it with sunset/sunrise at this point.
If the difference is too big it takes the front/rear part of the route and checks the midpoint again and again...

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

	// flight from EBBR to LKPR
	route := nighttime.Route{
		Departure: nighttime.Place{
			Lat:  50.9014015198,
			Lon:  4.4844398499,
			Time: time.Date(2022, 6, 3, 18, 53, 0, 0, time.UTC),
		},
		Arrival: nighttime.Place{
			Lat:  50.1007995605,
			Lon:  14.2600002289,
			Time: time.Date(2022, 6, 3, 20, 16, 0, 0, time.UTC),
		},
	}

	fmt.Println("Flight time:", route.FlightTime())
	fmt.Println("Distance (nm):", route.Distance())
	fmt.Println("Night time:", route.NightTime())
}
```


```bash
$ go run main.go
Flight time: 1h23m0s
Distance (nm): 376.5411972344908
Night time: 27m0s

```

## Used modules

* go-sunrise https://github.com/nathan-osman/go-sunrise
