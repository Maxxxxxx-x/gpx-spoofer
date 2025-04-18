package open_route_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Maxxxxxx-x/gpx-spoofer/models"
)

type RouteAPI interface {
	GetRouteForPositions(positions []models.Position) ([]models.Position, error)
}

type routeAPI struct {
	baseUrl string
}

type requestBody struct {
	Coordinates [][]float64 `json:"coordinates"`
}

type responseBody struct {
	Type     string    `json:"type"`
	Bbox     []float64 `json:"bbox"`
	Features []Feature `json:"features"`
	Metadata Metadata  `json:"metadata"`
}

// lon lat

func parseDecodedBody(reqBody requestBody, respBody responseBody) []models.Position {
	startPos := models.Position{
		Lon: reqBody.Coordinates[0][0],
		Lat: reqBody.Coordinates[0][1],
	}

	endPos := models.Position{
		Lon: reqBody.Coordinates[1][0],
		Lat: reqBody.Coordinates[1][1],
	}

	var parsedPos []models.Position
	parsedPos = append(parsedPos, startPos)

	for _, pos := range respBody.Features[0].Geometry.Coordinates {
		currentPos := models.Position{
			Lon: pos[0],
			Lat: pos[1],
		}
		parsedPos = append(parsedPos, currentPos)
	}

	parsedPos = append(parsedPos, endPos)

	return parsedPos
}

func (api routeAPI) GetRouteForPositions(positions []models.Position) ([]models.Position, error) {
	reqBody := requestBody{}
	for _, position := range positions {
		reqBody.Coordinates = append(reqBody.Coordinates, []float64{position.Lon, position.Lat})
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/ors/v2/directions/foot-hiking/geojson", api.baseUrl)
	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decodedBody responseBody
	if err := json.Unmarshal(respBody, &decodedBody); err != nil {
		return nil, err
	}
	// fmt.Println(decodedBody)
	return parseDecodedBody(reqBody, decodedBody), nil
}

func New(routeUrl string) RouteAPI {
	routeApi := routeAPI{
		baseUrl: routeUrl,
	}

	return routeApi
}
