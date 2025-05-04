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
	query := query.Use(db)
	device, err := query.Device.WithContext(ctx).Where(query.Device.Deviceid.Eq(deviceValue)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			device = &model.Device{Deviceid: deviceValue, Label: locationValue}
			query.Device.WithContext(ctx).Create(device)
			if created, err := query.Device.WithContext(ctx).
				Where(query.Device.Deviceid.Eq(deviceValue)).
				First(); err != nil {
				log.Error().Err(err).Msg("Failed to get the new device due to")
			} else {
				log.Info().Interface("device", created).Msg("Stored")
				return created
			}
		} else {
			log.Panic().Err(err).Msg("Failed to get device due to")
		}
	}
	log.Info().Interface("device", device).Msg("Existing")
	return device
}

func storeCount(ctx context.Context, timestampValue int64, device *model.Device, countValue int) {
	count := &model.Count{
		RecordedAt: time.Unix(0, timestampValue),
		Value:      int32(countValue),
		DeviceID:   device.ID,
	}
	query := query.Use(db)
	query.Count.WithContext(ctx).Create(count)
	if storedCount, err := query.Count.WithContext(ctx).
		Where(query.Count.DeviceID.Eq(device.ID)).
		Where(query.Count.RecordedAt.Eq(time.Unix(0, timestampValue))).
		First(); err != nil {
		log.Error().Err(err).Msg("Failed to get the new count due to")
	} else {
		log.Info().Interface("count", storedCount).Msg("Stored")
	}
}

func storeRssi(ctx context.Context, timestampValue int64, device *model.Device, rssiValue int) {
	rssi := &model.Rssi{
		RecordedAt: time.Unix(0, timestampValue),
		Value:      int32(rssiValue),
		DeviceID:   device.ID}
	query := query.Use(db)
	query.Rssi.WithContext(ctx).Create(rssi)
	if storedRssi, err := query.Rssi.WithContext(ctx).
		Where(query.Rssi.DeviceID.Eq(device.ID)).
		Where(query.Rssi.RecordedAt.Eq(time.Unix(0, timestampValue))).
		First(); err != nil {
		log.Error().Err(err).Msg("Failed to get the new rssi due to")
	} else {
		log.Info().Interface("rssi", storedRssi).Msg("Stored")
	}
}

func storeTemperature(ctx context.Context, timestampValue int64, device *model.Device, temperatureValue float64) {
	temperature := &model.Temperature{
		RecordedAt: time.Unix(0, timestampValue),
		Value:      temperatureValue,
		DeviceID:   device.ID}
	query := query.Use(db)
	query.Temperature.WithContext(ctx).Create(temperature)
	if storedTemperature, err := query.Temperature.WithContext(ctx).
		Where(query.Temperature.DeviceID.Eq(device.ID)).
		Where(query.Temperature.RecordedAt.Eq(time.Unix(0, timestampValue))).
		First(); err != nil {
		log.Error().Err(err).Msg("Failed to get the new temperature due to")
	} else {
		log.Info().Interface("temperature", storedTemperature).Msg("Stored")
	}
}
