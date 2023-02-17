package types

import (
	"context"
)

type ElasticIndexer interface {
	Add(context.Context, string, interface{}, string) error
}
