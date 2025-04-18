package open_elevation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Maxxxxxx-x/gpx-spoofer/models"
)

type ElevateAPI interface {
	GetElevations(positions []models.Position) ([]models.Position, error)
}

type elevateAPI struct {
	baseUrl string
}

type requestBody struct {
	Locations []models.Position `json:"locations"`
}

type responseBody struct {
	Results []models.Position `json:"results"`
}

func (api elevateAPI) GetElevations(positions []models.Position) ([]models.Position, error) {
	req := requestBody{
		Locations: positions,
	}
	bodyByte, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/lookup", api.baseUrl)

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyByte))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decodedBody := responseBody{}

	if err := json.Unmarshal(body, &decodedBody); err != nil {
		return nil, err
	}

	return decodedBody.Results, nil
}

func New(elevateUrl string) ElevateAPI {
	elevateAPI := elevateAPI{
		baseUrl: elevateUrl,
	}

	return elevateAPI
}
