// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"context"
)

type Querier interface {
	GetRecordById(ctx context.Context, id string) (Record, error)
	GetRecords(ctx context.Context, arg GetRecordsParams) ([]Record, error)
	InsertSpoofedRecord(ctx context.Context, arg InsertSpoofedRecordParams) error
}

var _ Querier = (*Queries)(nil)
