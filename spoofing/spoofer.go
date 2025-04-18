package spoofing

import (
	"log"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	open_elevation "github.com/Maxxxxxx-x/gpx-spoofer/api/OE"
	open_route_service "github.com/Maxxxxxx-x/gpx-spoofer/api/ORS"
	"github.com/Maxxxxxx-x/gpx-spoofer/db"
	"github.com/Maxxxxxx-x/gpx-spoofer/models"
	"github.com/Maxxxxxx-x/gpx-spoofer/sql/sqlc"
	"github.com/tkrajina/gpxgo/gpx"
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
	TEST_ID   = "01JE5XVK96N6FFDKXPK49R13HH"
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

		gpxBytes := []byte(*record.Rawdata)

		parsed, err := gpx.ParseBytes(gpxBytes)
		if err != nil {
			processedRecord.error = err
			processedRecords = append(processedRecords, processedRecord)
			continue
		}

		startPoint := parsed.Tracks[0].Segments[0].Points[0]
		processedRecord.startPoint = models.Position{
			Lat: startPoint.GetLatitude(),
			Lon: startPoint.GetLongitude(),
			Elv: startPoint.Elevation.Value(),
		}

		processedRecord.startTimestamp = startPoint.Timestamp.String()

		points := parsed.Tracks[0].Segments[0].Points
		endIdx := len(points) - 1
		processedRecord.endPoint = models.Position{
			Lat: points[endIdx].GetLatitude(),
			Lon: points[endIdx].GetLongitude(),
			Elv: points[endIdx].Elevation.Value(),
		}
		processedRecord.error = nil
		processedRecords = append(processedRecords, processedRecord)
	}

	return processedRecords, nil
}

func (spoofer baseSpoofer) generatePath(prepared []ProcessedRecord) error {
	errChan := make(chan error)

	defer close(errChan)
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
				return
			}

			newPath, err := spoofer.elevApi.GetElevations(path)
			if err != nil {
				errChan <- err
				wg.Done()
				return
			}

			var gpxFile gpx.GPX
			gpxFile.Version = "1.1"
			startTime := time.Now()
			currentTime := startTime
			totalDistance := 0.0
			highest := newPath[0].Elv
			lowest := newPath[0].Elv
			var points []gpx.GPXPoint
			for i, pos := range newPath {
				if i == 0 {
					currentPoint := gpx.GPXPoint{}
					currentPoint.Latitude = pos.Lat
					currentPoint.Longitude = pos.Lon
					currentPoint.Elevation.SetValue(pos.Elv)
					currentPoint.Timestamp = startTime
					points = append(points, currentPoint)
					continue
				}

				distance := getDistance(pos, newPath[i-1])
				totalDistance += distance
				speed := MIN_SPEED + rand.Float64()*(MAX_SPEED-MIN_SPEED)
				duration := distance / speed
				currentPoint := gpx.GPXPoint{}
				currentPoint.Latitude = pos.Lat
				currentPoint.Longitude = pos.Lon
				currentPoint.Elevation.SetValue(pos.Elv)
				currentTime = currentTime.Add(time.Duration(duration * 1000000000))
				currentPoint.Timestamp = currentTime
				if highest < pos.Elv {
					highest = pos.Elv
				}
				if lowest > pos.Elv {
					lowest = pos.Elv
				}
				points = append(points, currentPoint)
			}

			var segments []gpx.GPXTrackSegment
			segments = append(segments, gpx.GPXTrackSegment{
				Points: points,
			})
			var tracks []gpx.GPXTrack
			tracks = append(tracks, gpx.GPXTrack{
				Segments: segments,
			})
			gpxFile.Tracks = tracks
			gpxFile.Creator = ""
			xmlBytes, err := gpxFile.ToXml(gpx.ToXmlParams{
				Version: "1.1",
				Indent:  true,
			})
			if err != nil {
				errChan <- err
				wg.Done()
				return
			}
			duration := currentTime.Sub(startTime)
			log.Println("Inserting...")
			if err := spoofer.db.InsertSpoofedRoute(
				duration.Seconds(),
				totalDistance,
				highest,
				lowest,
				record.trailName,
				string(xmlBytes),
			); err != nil {
				errChan <- err
				log.Printf("Error: %v", err.Error())
				wg.Done()
				return
			}
			log.Printf("Inserted spoofed %s", record.trailId)
		}()
		wg.Wait()
	}

	return nil
}

func getDistance(pos1 models.Position, pos2 models.Position) float64 {
	phi1 := pos1.Lat * math.Pi / 180
	phi2 := pos2.Lat * math.Pi / 180
	deltaPhi := (pos2.Lat - pos1.Lat) * math.Pi / 180
	deltaLambda := (pos2.Lon - pos1.Lon) * math.Pi / 180

	R := 6371e3

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
	const total_records = 511788
	loops := int(math.Round(total_records / 50))

	for range loops {
		log.Println("Fetching records...")
		records, err := spoofer.db.GetRecordsFromDatabase()
		if err != nil {
			log.Printf("Error: %v", err.Error())
		}

		prepared, err := prepareRecords(records)
		if err != nil {
			log.Printf("Error: %v", err.Error())
		}

		err = spoofer.generatePath(prepared)
		if err != nil {
			log.Printf("Error: %v", err.Error())
		}
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

	err = spoofer.generatePath(prepared)
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
