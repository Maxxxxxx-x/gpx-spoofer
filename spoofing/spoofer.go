package spoofing

import (
	"bytes"
	"sync"
	"time"

	open_elevation "github.com/Maxxxxxx-x/gpx-spoofer/api/OE"
	open_route_service "github.com/Maxxxxxx-x/gpx-spoofer/api/ORS"
	"github.com/Maxxxxxx-x/gpx-spoofer/db"
	"github.com/Maxxxxxx-x/gpx-spoofer/models"
	"github.com/Maxxxxxx-x/gpx-spoofer/sql/sqlc"
	"github.com/twpayne/go-gpx"
)

/*
workflow:
get path
process path (start end pt)
generate path
get elevation
shove into db
*/

const (
	TEST_ID   = "01JE60J5TSM3F99832MPGPF3NQ"
	MAX_SPEED = 1
	MIN_SPEED = 0.5
)

type ProcessedRecord struct {
	trailId         string
	trailName       string
	preprocessData  string
	duration        float64
	distance        float64
	startPoint      models.Position
	startTimestamp  string
	endPoint        models.Position
	postprocessData string
	error           error
}

type Spoofer interface {
	Start() error
	TestStart() error
}

type baseSpoofer struct {
	routeApi open_route_service.RouteAPI
	elevApi  open_elevation.ElevateAPI
	db       db.Database
}

func prepareRecords(records []sqlc.Record) ([]ProcessedRecord, error) {
	var processedRecords []ProcessedRecord
	for _, record := range records {
		processedRecord := ProcessedRecord{
			trailId:        record.ID,
			trailName:      *record.Trails,
			preprocessData: *record.Rawdata,
			duration:       *record.Duration,
			distance:       *record.Distance,
		}

		gpxBytes := bytes.NewBufferString(*record.Rawdata)
		parsed, err := gpx.Read(gpxBytes)
		if err != nil {
			processedRecord.error = err
			processedRecords = append(processedRecords, processedRecord)
			continue
		}

		processedRecord.startPoint = models.Position{
			Lat: parsed.Wpt[0].Lat,
			Lon: parsed.Wpt[0].Lon,
			Elv: parsed.Wpt[0].Ele,
		}

		processedRecord.startTimestamp = parsed.Wpt[0].Time.String()

		endIdx := len(parsed.Wpt) - 1
		processedRecord.endPoint = models.Position{
			Lat: parsed.Wpt[endIdx].Lat,
			Lon: parsed.Wpt[endIdx].Lon,
			Elv: parsed.Wpt[endIdx].Ele,
		}
		processedRecord.error = nil
		processedRecords = append(processedRecords, processedRecord)
	}

	return processedRecords, nil
}

func (spoofer baseSpoofer) generatePath(prepared []ProcessedRecord) ([]models.Position, error) {
	errChan := make(chan error)
	doneChan := make(chan int)
	pathChan := make(chan []models.Position)

	defer close(errChan)
	defer close(doneChan)
	defer close(pathChan)
	var wg sync.WaitGroup

	for _, record := range prepared {
		wg.Add(1)
		go func() {
			defer wg.Done()
			positions := []models.Position{
				record.startPoint,
				record.endPoint,
			}
			path, err := spoofer.routeApi.GetRouteForPositions(positions)
			if err != nil {
				errChan <- err
				wg.Done()
			}
			newPath, err := spoofer.elevApi.GetElevations(path)
			if err != nil {
				errChan <- err
				wg.Done()
			}

			pathChan <- newPath
		}()
	}

	generatedPaths := []models.Position{}
	for newPath := range pathChan {
		waypoints := []*gpx.WptType{}
		starTime := time.Now()
		for i, pos := range newPath {
			wayPoint := *gpx.WptType{
			}

			if i == 0 {

			}
			distance := getDistance(pos, newPath[])
			waypoints = append(waypoints, *gpx.WptType{
				Lat: pos.Lat,
				Lon: pos.Lon,
				Ele: pos.Elv,
				Time: time.
			})
		}

		newGpx := &gpx.GPX{
			Version: "1.1",
			Wpt: waypoints,
		}
	}
}

func getDistance(pos1 models.Position, pos2 models.Position) float64 {
	phi1 := pos1.Lat * math.Pi / 180
	phi2 := pos2.Lat * math.Pi / 180
	deltaPhi := (pos2.Lat - pos1.Lat) * math.Pi / 180
	deltaLambda := (pos2.Lon- pos1.Lon) * math.Pi / 180

	R = 6371e3

    a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
        math.Cos(phi1)*math.Cos(phi2)*
            math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
    c := 2 * math.Asin(math.Sqrt(a))
    dist := R * c

    diffElevation := pos2.Elv - pos1.Elv
    dist3D := math.Sqrt(dist*dist + diffElevation*diffElevation)
	return dist3D
}

func (spoofer baseSpoofer) Start() error {
	records, err := spoofer.db.GetRecordsFromDatabase()
	if err != nil {
		return err
	}

	prepared, err := prepareRecords(records)
	if err != nil {
		return err
	}

	return nil
}

func (spoofer baseSpoofer) TestStart() error {
	record, err := spoofer.db.GetRecordFromDatabaseById(TEST_ID)
	if err != nil {
		return err
	}

	records := []sqlc.Record{record}

	prepared, err := prepareRecords(records)
	if err != nil {
		return err
	}

	return nil
}

func New(routeApi open_route_service.RouteAPI, elevApi open_elevation.ElevateAPI, db db.Database) Spoofer {
	spoofer := baseSpoofer{
		routeApi: routeApi,
		elevApi:  elevApi,
		db:       db,
	}

	return spoofer
}
