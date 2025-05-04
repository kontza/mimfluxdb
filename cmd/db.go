package cmd

import (
	"context"
	"errors"
	"fmt"
	"mimfluxdb/dao/model"
	"mimfluxdb/dao/query"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openDb() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Helsinki",
		appConfig.DatabaseConnection.Host,
		appConfig.DatabaseConnection.Username,
		appConfig.DatabaseConnection.Password,
		appConfig.DatabaseConnection.Database,
		appConfig.DatabaseConnection.Port)
	newLogger := logger.New(
		&log.Logger, // IO.writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: false,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,          // Disable color
		},
	)
	db, err := gorm.Open(postgres.Open(dsn),
		&gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database due to")
	}
	return db
}

func getDevice(ctx context.Context, deviceValue string, locationValue string) *model.Device {
	db := openDb()
	query := query.Use(db)
	device, err := query.Device.WithContext(ctx).Where(query.Device.Deviceid.Eq(deviceValue)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			device = &model.Device{Deviceid: deviceValue, Label: locationValue}
			query.Device.WithContext(ctx).Create(device)
			return getDevice(ctx, deviceValue, locationValue)
		} else {
			log.Panic().Err(err).Msg("Failed to get device due to")
		}
	}
	return device
}
