package model

import (
	"net/url"
	"strconv"

	"github.com/taythebot/archer/pkg/types"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Client struct {
	DB *gorm.DB
}

func ConnectToDB(cfg types.PostgresConfig) (*Client, error) {
	// Configure connection url
	dsn := url.URL{
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Scheme: "postgres",
		Host:   cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Path:   cfg.Database,
		RawQuery: (&url.Values{
			"sslmode":  []string{"disable"},
			"TimeZone": []string{"UTC"},
		}).Encode(),
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{Logger: newLogger()})
	if err != nil {
		return nil, err
	}

	return &Client{DB: db}, nil
}

// RunMigration syncs local model to the database
func (c *Client) RunMigration() error {
	return c.DB.AutoMigrate(&Scans{}, &Tasks{})
}
