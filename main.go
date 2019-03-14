// +build !windows

package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var started = time.Now()

func main() {

	//run()

	out, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close

	// c, tr := get_term_size(out)

	// fmt.Println(c)
	// fmt.Println(tr)

	fmt.Fprintf(out, "lala")

}

func reversePositions(positions []position) []position {
	for i, j := 0, len(positions)-1; i < j; i, j = i+1, j-1 {
		positions[i], positions[j] = positions[j], positions[i]
	}
	return positions
}

var wg sync.WaitGroup

type vehicleRoute struct {
	vehicleId string
	route     route
}

func routeConsumer(id int, ch <-chan vehicleRoute, done <-chan bool) {
	defer wg.Done()

	count := 1

	for {
		PrintConsumer(id, count, "Waiting for channels...")
		select {
		case vr := <-ch:
			PrintConsumer(id, count, "Processing route %d: %s\n", count, vr.route.name)
			select {
			case <-time.After(time.Duration(seededRand.Intn(1000)) * time.Millisecond):
			case <-done:
				PrintConsumer(id, count, "Quitting...")
				return
			}
		case <-done:
			PrintConsumer(id, count, "Quitting...")
			return
		}
		count++
	}
}

func routeProducer(id int, ch chan<- vehicleRoute, done <-chan bool) {
	defer wg.Done()

	count := 1

	for {
		select {
		case ch <- generateVehicleRoute(id, count):
		case <-done:
			PrintProducer(id, count, "Quitting...")
			return
		}
		count++
	}
}

func generateVehicleRoute(id int, count int) vehicleRoute {
	vr := vehicleRoute{route: getRoute(), vehicleId: generateId("urn:external:vehicle")}
	PrintProducer(id, count, "Route %s for vehicle %s send...\n", vr.route.name, vr.vehicleId)
	return vr
}

func ClearScreen() {
	fmt.Print("\033[2J")
}

var mutex = &sync.Mutex{}

func PrintAt(x int, y int, format string, a ...interface{}) {
	mutex.Lock()
	fmt.Printf("\033[%d;%dH\033[K", y, x)
	fmt.Printf(format, a...)
	mutex.Unlock()
}

var status = make(chan string)

func statusBar(done <-chan bool) {
	for {
		ticker := time.NewTicker(500 * time.Millisecond).C

		select {
		case msg := <-status:
			PrintAt(1, statusOffset, "\033[30;42m%-196s\033[0m", msg)
		case <-ticker:
			printTOD()
		case <-done:
			return
		}
	}
}

func printTOD() {
	d := time.Since(started)
	t := time.Now()

	PrintAt(197, statusOffset, "\033[30;42m\u2502 Elapsed: %5ds \u2502 %d-%02d-%02d %02d:%02d:%02d \033[0m",
		int(d.Seconds()),
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

func PrintStatusBar(format string, a ...interface{}) {
	PrintAt(1, statusOffset, strings.Repeat("\033[30;42m \033[0m", 200 /*terminalSize.cols*/))
	status <- fmt.Sprintf(format, a...)
}

func PrintTitle(x int, y int, w int, format string, a ...interface{}) {

	wf := fmt.Sprintf("\033[41m%%-%ds\033[0m", w)
	t := fmt.Sprintf(format, a...)

	PrintAt(x, y, wf, t)
	PrintAt(x, y+1, strings.Repeat("-", w))

}

func PrintConsumer(id int, cnt int, format string, a ...interface{}) {
	PrintAt(1, consumerOffset+1+id, "%3d \u2502 %5d \u2502", id, cnt)
	PrintAt(15, consumerOffset+1+id, format, a...)
}

func PrintProducer(id int, cnt int, format string, a ...interface{}) {
	PrintAt(1, producerOffset+1+id, "%3d \u2502 %5d \u2502", id, cnt)
	PrintAt(15, producerOffset+1+id, format, a...)
}

var consumers = 5
var producers = 5

var statusOffset = 1

var consumerOffset = statusOffset + 1
var producerOffset = consumerOffset + 1 + consumers + 2

func run() {
	routeCh := make(chan vehicleRoute)
	done := make(chan bool)

	fmt.Printf("\033[?25l")

	go statusBar(done)

	ClearScreen()
	PrintStatusBar("Simulatron")
	PrintTitle(1, consumerOffset, 80, "Consumers")
	for i := 1; i <= consumers; i++ {
		PrintConsumer(i, 0, "")
	}
	PrintTitle(1, producerOffset, 80, "Producers")
	for i := 1; i <= producers; i++ {
		PrintProducer(i, 0, "")
	}

	for i := 1; i <= consumers; i++ {
		wg.Add(1)
		go routeConsumer(i, routeCh, done)
	}
	for i := 1; i <= producers; i++ {
		wg.Add(1)
		go routeProducer(i, routeCh, done)
	}

	//fmt.Println("main thread sleeping")

	time.Sleep(10 * time.Second)

	//fmt.Println("done sleeping, sending exit signal")

	fmt.Printf("\033[?25h")

	close(done)

	wg.Wait()

	PrintAt(1, producerOffset+1+producers+2, "Done")

	//	fmt.Println("Done")

	//awayRoute := getRandomRoute()
	//homeRoute := getRandomRoute()

	//vehicleId := generateId("urn:external:vehicle")

	//fmt.Printf("Going away %s\n", awayRoute.name)

	//for _, p := range awayRoute.positions {
	//
	//	j := getMessage(vehicleId, p)
	//
	//	fmt.Println(string(j))
	//
	//}

	//fmt.Printf("Going home %s\n", homeRoute.name)

	//for _, p := range reversePositions(homeRoute.positions) {
	//
	//	j := getMessage(vehicleId, p)
	//
	//	fmt.Println(string(j))
	//
	//}

}

func getMessage(vehicleId string, p position) []byte {
	coordinates := getCoordinates(p)

	point := NewPointGeometry(coordinates)

	msg := Message{
		MessageId: generateId("urn:external:position"),
		VehicleId: vehicleId,
		Timestamp: time.Now(),
		Position:  point,
	}

	j, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return j
}

/**********************************************************************************************************************
 *
 * route
 *
 **********************************************************************************************************************/

var allRoutes = loadRoutes()

type route struct {
	name      string
	positions []position
}

type position struct {
	longitude float64
	latitude  float64
}

func getRoute() route {
	return allRoutes[seededRand.Int()%len(allRoutes)]
}

func loadRoutes() []route {

	routesFiles, err := filepath.Glob("route_*.csv")
	if err != nil {
		panic(err)
	}
	if len(routesFiles) < 1 {
		panic("No routes found!")
	}

	routes := make([]route, len(routesFiles))

	r := regexp.MustCompile(`route_(?P<name>.*)\.csv`)

	for i, rf := range routesFiles {
		routes[i] = route{
			name:      r.ReplaceAllString(rf, "$name"),
			positions: loadRoutePositions(rf),
		}

	}

	return routes
}

func loadRoutePositions(route string) []position {
	f, err := os.Open(route)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var positions []position

	reader := csv.NewReader(bufio.NewReader(f))
	lineNum := 0

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		lineNum++

		if lineNum == 1 {
			continue
		}

		positions = append(positions, position{
			longitude: convertFloat64(line[0]),
			latitude:  convertFloat64(line[1]),
		})
	}

	return positions
}

func convertFloat64(s string) float64 {
	lon, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return lon
}

func getCoordinates(position position) []float64 {
	return []float64{position.longitude, position.latitude}
}

/**********************************************************************************************************************
 *
 * generateId
 *
 **********************************************************************************************************************/

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateId(prefix string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 7)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	s := string(b)

	return fmt.Sprintf("%s:%s", prefix, s)
}

/**********************************************************************************************************************
 *
 * geoJSON
 *
 **********************************************************************************************************************/

type GeometryType string

const (
	GeometryPoint GeometryType = "Point"
)

type Message struct {
	MessageId string    `json:"messageId"`
	VehicleId string    `json:"vehicleId"`
	Timestamp time.Time `json:"timestamp"`
	Position  Geometry  `json:"position"`
}

type Geometry struct {
	Type  GeometryType `json:"type"`
	Point []float64
}

func NewPointGeometry(coordinate []float64) Geometry {
	return Geometry{
		Type:  GeometryPoint,
		Point: coordinate,
	}
}

func (g Geometry) MarshalJSON() ([]byte, error) {
	type geometry struct {
		Type        GeometryType `json:"type"`
		Coordinates interface{}  `json:"coordinates,omitempty"`
	}

	geo := &geometry{
		Type: g.Type,
	}

	switch g.Type {
	case GeometryPoint:
		geo.Coordinates = g.Point
	}

	return json.Marshal(geo)
}
