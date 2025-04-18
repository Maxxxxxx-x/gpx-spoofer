-- name: GetRecords :many
SELECT * FROM Records LIMIT $1 OFFSET $2;

-- name: GetRecordById :one
SELECT * FROM Records WHERE id = $1 LIMIT 1;

-- name: InsertSpoofedRecord :exec
INSERT INTO spoofed_gpx (
    Id, Duration, Distance, HighestPoint, LowestPoint, ElevationDiff, Trails, RawData
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
