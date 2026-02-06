package monitor

import (
	"context"
	"time"

	"github.com/etswifi/ets-noc/internal/models"
	"github.com/etswifi/ets-noc/internal/storage"
)

type StatusComputer struct {
	postgres *storage.PostgresStore
	redis    *storage.RedisStore
}

func NewStatusComputer(postgres *storage.PostgresStore, redis *storage.RedisStore) *StatusComputer {
	return &StatusComputer{
		postgres: postgres,
		redis:    redis,
	}
}

// ComputePropertyStatus computes the rollup status for a property based on device statuses
func (sc *StatusComputer) ComputePropertyStatus(ctx context.Context, propertyID int64, devices []models.Device) (*models.PropertyStatus, error) {
	if len(devices) == 0 {
		return &models.PropertyStatus{
			PropertyID: propertyID,
			Status:     "green",
			LastCheck:  time.Now(),
		}, nil
	}

	// Get all device statuses
	deviceStatuses := make(map[int64]*models.DeviceStatus)
	for _, d := range devices {
		status, err := sc.redis.GetDeviceStatus(ctx, d.ID)
		if err == nil && status != nil {
			deviceStatuses[d.ID] = status
		}
	}

	online, offline := 0, 0
	criticalOffline := false

	for _, device := range devices {
		if status, ok := deviceStatuses[device.ID]; ok && status.Status == "online" {
			online++
		} else {
			offline++
			if device.IsCritical {
				criticalOffline = true
			}
		}
	}

	propertyStatus := &models.PropertyStatus{
		PropertyID:      propertyID,
		OnlineCount:     online,
		OfflineCount:    offline,
		TotalCount:      len(devices),
		CriticalOffline: criticalOffline,
		LastCheck:       time.Now(),
	}

	// Status logic: red > yellow > green
	if offline == len(devices) || criticalOffline {
		propertyStatus.Status = "red"
	} else if offline > 0 {
		propertyStatus.Status = "yellow"
	} else {
		propertyStatus.Status = "green"
	}

	return propertyStatus, nil
}

// ComputeAllPropertyStatuses computes status for all properties
func (sc *StatusComputer) ComputeAllPropertyStatuses(ctx context.Context) error {
	properties, err := sc.postgres.ListProperties(ctx)
	if err != nil {
		return err
	}

	for _, property := range properties {
		devices, err := sc.postgres.ListDevicesForProperty(ctx, property.ID)
		if err != nil {
			continue
		}

		propertyStatus, err := sc.ComputePropertyStatus(ctx, property.ID, devices)
		if err != nil {
			continue
		}

		if err := sc.redis.SetPropertyStatus(ctx, propertyStatus); err != nil {
			continue
		}
	}

	return nil
}
