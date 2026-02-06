package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/etswifi/ets-noc/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string, password string, db int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}

// Device Status Keys
func deviceStatusKey(deviceID int64) string {
	return fmt.Sprintf("device:status:%d", deviceID)
}

func deviceHistoryKey(deviceID int64) string {
	return fmt.Sprintf("device:history:%d", deviceID)
}

func allDeviceStatusKey() string {
	return "all_device_status"
}

// Property Status Keys
func propertyStatusKey(propertyID int64) string {
	return fmt.Sprintf("property:status:%d", propertyID)
}

func allPropertyStatusKey() string {
	return "all_property_status"
}

func propertyLastNotificationKey(propertyID int64) string {
	return fmt.Sprintf("property:last_notification:%d", propertyID)
}

// Device Status Operations
func (r *RedisStore) SetDeviceStatus(ctx context.Context, status *models.DeviceStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()

	// Store individual device status
	pipe.Set(ctx, deviceStatusKey(status.DeviceID), data, 10*time.Minute)

	// Add to all devices hash for quick lookup
	pipe.HSet(ctx, allDeviceStatusKey(), strconv.FormatInt(status.DeviceID, 10), data)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetDeviceStatus(ctx context.Context, deviceID int64) (*models.DeviceStatus, error) {
	data, err := r.client.Get(ctx, deviceStatusKey(deviceID)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("device status not found")
	}
	if err != nil {
		return nil, err
	}

	var status models.DeviceStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *RedisStore) GetAllDeviceStatuses(ctx context.Context) (map[int64]*models.DeviceStatus, error) {
	data, err := r.client.HGetAll(ctx, allDeviceStatusKey()).Result()
	if err != nil {
		return nil, err
	}

	statuses := make(map[int64]*models.DeviceStatus)
	for deviceIDStr, statusJSON := range data {
		deviceID, err := strconv.ParseInt(deviceIDStr, 10, 64)
		if err != nil {
			continue
		}

		var status models.DeviceStatus
		if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
			continue
		}
		statuses[deviceID] = &status
	}
	return statuses, nil
}

// Device History Operations
func (r *RedisStore) AddDeviceHistory(ctx context.Context, deviceID int64, status string, responseTime float64) error {
	timestamp := time.Now().Unix()
	history := models.DeviceHistory{
		Timestamp:    timestamp,
		Status:       status,
		ResponseTime: responseTime,
	}

	data, err := json.Marshal(history)
	if err != nil {
		return err
	}

	// Use sorted set with timestamp as score for time-series data
	err = r.client.ZAdd(ctx, deviceHistoryKey(deviceID), redis.Z{
		Score:  float64(timestamp),
		Member: data,
	}).Err()
	if err != nil {
		return err
	}

	// Keep only last 90 days
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90).Unix()
	return r.client.ZRemRangeByScore(ctx, deviceHistoryKey(deviceID), "0", strconv.FormatInt(ninetyDaysAgo, 10)).Err()
}

func (r *RedisStore) GetDeviceHistory(ctx context.Context, deviceID int64, startTime, endTime time.Time) ([]models.DeviceHistory, error) {
	data, err := r.client.ZRangeByScore(ctx, deviceHistoryKey(deviceID), &redis.ZRangeBy{
		Min: strconv.FormatInt(startTime.Unix(), 10),
		Max: strconv.FormatInt(endTime.Unix(), 10),
	}).Result()
	if err != nil {
		return nil, err
	}

	var history []models.DeviceHistory
	for _, item := range data {
		var h models.DeviceHistory
		if err := json.Unmarshal([]byte(item), &h); err != nil {
			continue
		}
		history = append(history, h)
	}
	return history, nil
}

// Property Status Operations
func (r *RedisStore) SetPropertyStatus(ctx context.Context, status *models.PropertyStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()

	// Store individual property status
	pipe.Set(ctx, propertyStatusKey(status.PropertyID), data, 10*time.Minute)

	// Add to all properties hash for quick lookup
	pipe.HSet(ctx, allPropertyStatusKey(), strconv.FormatInt(status.PropertyID, 10), data)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetPropertyStatus(ctx context.Context, propertyID int64) (*models.PropertyStatus, error) {
	data, err := r.client.Get(ctx, propertyStatusKey(propertyID)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("property status not found")
	}
	if err != nil {
		return nil, err
	}

	var status models.PropertyStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *RedisStore) GetAllPropertyStatuses(ctx context.Context) (map[int64]*models.PropertyStatus, error) {
	data, err := r.client.HGetAll(ctx, allPropertyStatusKey()).Result()
	if err != nil {
		return nil, err
	}

	statuses := make(map[int64]*models.PropertyStatus)
	for propertyIDStr, statusJSON := range data {
		propertyID, err := strconv.ParseInt(propertyIDStr, 10, 64)
		if err != nil {
			continue
		}

		var status models.PropertyStatus
		if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
			continue
		}
		statuses[propertyID] = &status
	}
	return statuses, nil
}

// Notification Cooldown Operations
func (r *RedisStore) SetLastNotification(ctx context.Context, propertyID int64, eventType string) error {
	key := propertyLastNotificationKey(propertyID)
	now := time.Now().Unix()
	return r.client.HSet(ctx, key, eventType, now).Err()
}

func (r *RedisStore) GetLastNotification(ctx context.Context, propertyID int64, eventType string) (time.Time, error) {
	key := propertyLastNotificationKey(propertyID)
	timestamp, err := r.client.HGet(ctx, key, eventType).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}

func (r *RedisStore) ShouldNotify(ctx context.Context, propertyID int64, eventType string, cooldownSeconds int) (bool, error) {
	lastNotification, err := r.GetLastNotification(ctx, propertyID, eventType)
	if err != nil {
		return false, err
	}

	if lastNotification.IsZero() {
		return true, nil
	}

	elapsed := time.Since(lastNotification)
	return elapsed.Seconds() >= float64(cooldownSeconds), nil
}

// Cleanup Operations
func (r *RedisStore) CleanupOldHistory(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Unix()

	// Get all device history keys
	keys, err := r.client.Keys(ctx, "device:history:*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := r.client.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(cutoff, 10)).Err(); err != nil {
			return err
		}
	}
	return nil
}
