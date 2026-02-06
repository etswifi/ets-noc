package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/etswifi/ets-noc/internal/models"
	"github.com/etswifi/ets-noc/internal/storage"
)

type Pinger struct {
	postgres     *storage.PostgresStore
	redis        *storage.RedisStore
	maxConcurrent int
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

func NewPinger(postgres *storage.PostgresStore, redis *storage.RedisStore, maxConcurrent int) *Pinger {
	return &Pinger{
		postgres:     postgres,
		redis:        redis,
		maxConcurrent: maxConcurrent,
		stopChan:     make(chan struct{}),
	}
}

func (p *Pinger) Start(ctx context.Context) error {
	log.Printf("Pinger started with max concurrent pings: %d", p.maxConcurrent)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Pinger stopping...")
			p.wg.Wait()
			return ctx.Err()
		case <-p.stopChan:
			log.Println("Pinger stopped")
			p.wg.Wait()
			return nil
		case <-ticker.C:
			if err := p.checkDevices(ctx); err != nil {
				log.Printf("Error checking devices: %v", err)
			}
		}
	}
}

func (p *Pinger) Stop() {
	close(p.stopChan)
}

func (p *Pinger) checkDevices(ctx context.Context) error {
	devices, err := p.postgres.ListActiveDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	if len(devices) == 0 {
		return nil
	}

	log.Printf("Checking %d devices", len(devices))

	// Create semaphore for concurrency control
	sem := make(chan struct{}, p.maxConcurrent)
	var wg sync.WaitGroup

	// Group devices by property for status computation
	devicesByProperty := make(map[int64][]models.Device)
	for _, device := range devices {
		devicesByProperty[device.PropertyID] = append(devicesByProperty[device.PropertyID], device)
	}

	// Check each device
	for _, device := range devices {
		wg.Add(1)
		go func(d models.Device) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()

				status := p.pingDevice(ctx, &d)
				if err := p.redis.SetDeviceStatus(ctx, status); err != nil {
					log.Printf("Failed to set device status for %s: %v", d.Name, err)
				}

				// Store history
				if err := p.redis.AddDeviceHistory(ctx, d.ID, status.Status, status.ResponseTime); err != nil {
					log.Printf("Failed to add device history for %s: %v", d.Name, err)
				}
			}
		}(device)
	}

	wg.Wait()

	// Compute property statuses
	statusComputer := NewStatusComputer(p.postgres, p.redis)
	for propertyID, propertyDevices := range devicesByProperty {
		propertyStatus, err := statusComputer.ComputePropertyStatus(ctx, propertyID, propertyDevices)
		if err != nil {
			log.Printf("Failed to compute property status for property %d: %v", propertyID, err)
			continue
		}

		if err := p.redis.SetPropertyStatus(ctx, propertyStatus); err != nil {
			log.Printf("Failed to set property status for property %d: %v", propertyID, err)
		}
	}

	return nil
}

func (p *Pinger) pingDevice(ctx context.Context, device *models.Device) *models.DeviceStatus {
	status := &models.DeviceStatus{
		DeviceID:  device.ID,
		LastCheck: time.Now(),
	}

	pinger, err := probing.NewPinger(device.Hostname)
	if err != nil {
		status.Status = "offline"
		status.Message = fmt.Sprintf("Failed to create pinger: %v", err)
		return status
	}

	pinger.SetPrivileged(true)
	pinger.Count = device.Retries
	pinger.Timeout = time.Duration(device.Timeout) * time.Millisecond

	err = pinger.Run()
	if err != nil {
		status.Status = "offline"
		status.Message = fmt.Sprintf("Ping failed: %v", err)
		return status
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		status.Status = "online"
		status.ResponseTime = float64(stats.AvgRtt.Milliseconds())
		status.Message = "OK"
	} else {
		status.Status = "offline"
		status.Message = fmt.Sprintf("No packets received (%d sent)", stats.PacketsSent)
	}

	return status
}
