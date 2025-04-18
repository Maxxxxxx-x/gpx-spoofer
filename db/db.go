package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Maxxxxxx-x/gpx-spoofer/config"
	"github.com/Maxxxxxx-x/gpx-spoofer/sql/sqlc"
	"github.com/Maxxxxxx-x/gpx-spoofer/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	GetRecordsFromDatabase() ([]sqlc.Record, error)
	GetRecordFromDatabaseById(id string) (sqlc.Record, error)
}

type BaseDatabase struct {
	currentOffset int32
	limit         int32
	queries       *sqlc.Queries
}

func createCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Minute)
}

func (db BaseDatabase) GetRecordsFromDatabase() ([]sqlc.Record, error) {
	ctx, cancel := createCtx()
	defer cancel()

	queryParams := sqlc.GetRecordsParams{
		Limit: db.limit,
		Offset: db.currentOffset,
	}

	records, err := db.queries.GetRecords(ctx, queryParams)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (db BaseDatabase) GetRecordFromDatabaseById(id string) (sqlc.Record, error) {
	ctx, cancel := createCtx()
	defer cancel()

	record, err := db.queries.GetRecordById(ctx, id)
	if err != nil {
		return sqlc.Record{}, err
	}
	return record, nil
}


func (db BaseDatabase) InsertSpoofedRoute(duration, distance, highest, lowest float64, trailName, gpxData string) error {
	recordId, err := utils.GenerateULID()
	if err != nil {
		return err
	}

	elevationDiff := highest - lowest

	insertParam := sqlc.InsertSpoofedRecordParams{
		ID: recordId,
		Duration: &duration,
		Distance: &distance,
		Highestpoint: &highest,
		Lowestpoint: &lowest,
		Elevationdiff: &elevationDiff,
		Trails: &trailName,
		Rawdata: &gpxData,
	}

	ctx, cancel := createCtx()
	defer cancel()
	if err := db.queries.InsertSpoofedRecord(ctx, insertParam); err != nil {
		return err
	}
	return nil
}

func New(dbConn *pgxpool.Pool) Database {
	database := BaseDatabase{
		currentOffset: 0,
		queries:       sqlc.New(dbConn),
		limit: 50,
	}

	return database
}

func CreateDatabaseConnection(config config.Database) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.Login,
		config.Password,
		config.DatabaseName,
	)

	dbConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	dbConfig.MaxConns = 20
	dbConfig.MinConns = 5
	dbConfig.MaxConnIdleTime = 10 * time.Hour
	dbConfig.MaxConnLifetime = 10 * time.Hour
	dbConfig.MaxConnLifetimeJitter = 11 * time.Hour

	connPool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}
	return connPool, err
}
