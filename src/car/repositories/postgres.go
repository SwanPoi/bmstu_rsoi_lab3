package repositories

import (
	"fmt"
	"net/url"

	"gorm.io/gorm"
	"gorm.io/driver/postgres"
)


type DatabaseConfig struct {
	Host 		string 	`mapstructure:"DB_HOST"`
	Port 		int64 	`mapstructure:"DB_PORT"`
	User 		string 	`mapstructure:"DB_USER"`
	Password 	string 	`mapstructure:"DB_PASSWORD"`
	Database 	string 	`mapstructure:"DB_NAME"`
}

func GetConnectionString(cfg *DatabaseConfig) (connStr string) {
	 dsn := url.URL{
        Scheme:   "postgres",
        User:     url.UserPassword(cfg.User, cfg.Password),
        Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Path:     cfg.Database,
    }

	q := dsn.Query()
	q.Set("sslmode", "disable")
	dsn.RawQuery = q.Encode()

	return dsn.String()
}

func InitDb(url string) (db *gorm.DB, err error) {
	db, err = gorm.Open(postgres.Open(url), &gorm.Config{})

	return db, err
}