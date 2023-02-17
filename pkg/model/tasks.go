package model

import (
	"time"

	"github.com/taythebot/archer/internal/uid"

	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type Tasks struct {
	ID          string        `gorm:"type:varchar;primaryKey" json:"id"`
	ScanID      string        `gorm:"type:varchar;not null" json:"scan_id"`
	Module      string        `gorm:"type:varchar;not null" json:"module"`
	Payload     *pgtype.JSONB `gorm:"type:jsonb" json:"payload,omitempty"`
	Results     *int          `gorm:"int" json:"results,omitempty"`
	Status      string        `gorm:"type:varchar;not null;default:pending" json:"status"`
	WorkerID    *string       `gorm:"type:varchar" json:"worker_id,omitempty"`
	StartedAt   *time.Time    `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	Scan        Scans         `gorm:"foreignKey:ScanID;references:ID" json:"-"`
}

// BeforeCreate generates a unique ID for new records
func (t *Tasks) BeforeCreate(_ *gorm.DB) (err error) {
	// Generate ID
	if t.ID == "" {
		t.ID, err = uid.Generate()
		if err != nil {
			return err
		}
	}

	// Set Payload to null JSONB
	if t.Payload == nil {
		t.Payload = &pgtype.JSONB{Status: pgtype.Null}
	}

	return
}
