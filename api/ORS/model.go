package open_route_service

type Query struct {
	Coordinates [][]float64 `json:"coordinates"`
	Profile     string      `json:"profile"`
	ProfileName string      `json:"profileName"`
	Format      string      `json:"format"`
}

type Engine struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GraphDate string `json:"graph_date"`
}

type Metadata struct {
	Attribution string `json:"attribution"`
	Serivce     string `json:"service"`
	Timestamp   uint64 `json:"timestamp"`
	Query       Query  `json:"query"`
	Engine      Engine `json:"engine"`
}

type Step struct {
	Distance    float32 `json:"distance"`
	Duration    float32 `json:"duration"`
	Type        int     `json:"type"`
	Instruction string  `json:"instruction"`
	Name        string  `json:"name"`
	WayPoints   []int   `json:"way_points"`
}

type Segment struct {
	Distnace float32 `json:"distance"`
	Duration float32 `json:"durations"`
	Steps    []Step  `json:"steps"`
}

type Summary struct {
	Distance float32 `json:"distance"`
	Duration float32 `json:"duration"`
}

type Properties struct {
	Segments  []Segment `json:"segments"`
	WayPoints []int     `json:"way_points"`
	Summary   Summary   `json:"summary"`
}

type Geometry struct {
	Coordinates [][]float64 `json:"coordinates"`
	Type        string      `json:"type"`
}

type Feature struct {
	BBox       []float64  `json:"bbox"`
	Type       string     `json:"type"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

