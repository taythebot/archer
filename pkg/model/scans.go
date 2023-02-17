package model

import (
	"time"

	"github.com/taythebot/archer/internal/uid"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Scans struct {
	ID          string         `gorm:"type:varchar;primaryKey" json:"id"`
	Modules     pq.StringArray `gorm:"type:varchar[];not null" json:"modules"`
	Targets     pq.StringArray `gorm:"type:varchar[];not null" json:"targets"`
	Ports       pq.Int32Array  `gorm:"type:int[];not null" json:"ports"`
	Arguments   pq.StringArray `gorm:"type:varchar" json:"arguments"`
	Status      string         `gorm:"type:varchar;not null;default:pending" json:"status"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Tasks       []Tasks        `gorm:"foreignKey:ScanID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"tasks,omitempty"`
}

// BeforeCreate generates a unique ID for new records
func (s *Scans) BeforeCreate(_ *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID, err = uid.Generate()
		if err != nil {
			return err
		}
	}

	return
}
