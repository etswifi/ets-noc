package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/etswifi/ets-noc/internal/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// Properties
func (s *PostgresStore) CreateProperty(ctx context.Context, p *models.Property) error {
	query := `
		INSERT INTO properties (name, address, notes, isp_company_name, isp_account_info)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	err := s.db.QueryRowContext(ctx, query, p.Name, p.Address, p.Notes, p.ISPCompanyName, p.ISPAccountInfo).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	// Auto-calculate subnet based on property ID: 10.(99 + floor(ID/256)).(ID%256).0/24
	subnetQuery := `
		UPDATE properties
		SET subnet = '10.' || (99 + (id / 256))::text || '.' || (id % 256)::text || '.0/24'
		WHERE id = $1
		RETURNING subnet`
	return s.db.QueryRowContext(ctx, subnetQuery, p.ID).Scan(&p.Subnet)
}

func (s *PostgresStore) GetProperty(ctx context.Context, id int64) (*models.Property, error) {
	p := &models.Property{}
	query := `SELECT id, name, address, subnet, notes, isp_company_name, isp_account_info,
		pfsense_host, pfsense_port, pfsense_username, pfsense_password, created_at, updated_at
		FROM properties WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Address, &p.Subnet, &p.Notes, &p.ISPCompanyName, &p.ISPAccountInfo,
		&p.PfSenseHost, &p.PfSensePort, &p.PfSenseUsername, &p.PfSensePassword,
		&p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("property not found")
	}
	return p, err
}

func (s *PostgresStore) ListProperties(ctx context.Context) ([]models.Property, error) {
	query := `SELECT id, name, address, subnet, notes, isp_company_name, isp_account_info,
		pfsense_host, pfsense_port, pfsense_username, pfsense_password, created_at, updated_at
		FROM properties ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []models.Property
	for rows.Next() {
		var p models.Property
		if err := rows.Scan(&p.ID, &p.Name, &p.Address, &p.Subnet, &p.Notes, &p.ISPCompanyName, &p.ISPAccountInfo,
			&p.PfSenseHost, &p.PfSensePort, &p.PfSenseUsername, &p.PfSensePassword,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}
	return properties, rows.Err()
}

func (s *PostgresStore) UpdateProperty(ctx context.Context, p *models.Property) error {
	query := `
		UPDATE properties
		SET name = $1, address = $2, notes = $3, isp_company_name = $4, isp_account_info = $5,
		    pfsense_host = $6, pfsense_port = $7, pfsense_username = $8, pfsense_password = $9, updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at`
	return s.db.QueryRowContext(ctx, query, p.Name, p.Address, p.Notes, p.ISPCompanyName, p.ISPAccountInfo,
		p.PfSenseHost, p.PfSensePort, p.PfSenseUsername, p.PfSensePassword, p.ID).
		Scan(&p.UpdatedAt)
}

func (s *PostgresStore) DeleteProperty(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM properties WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("property not found")
	}
	return nil
}

// Contacts
func (s *PostgresStore) CreateContact(ctx context.Context, c *models.Contact) error {
	query := `
		INSERT INTO contacts (property_id, name, phone, email, role, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(ctx, query, c.PropertyID, c.Name, c.Phone, c.Email, c.Role, c.Notes).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (s *PostgresStore) GetContact(ctx context.Context, id int64) (*models.Contact, error) {
	c := &models.Contact{}
	query := `SELECT id, property_id, name, phone, email, role, notes, created_at, updated_at
		FROM contacts WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.PropertyID, &c.Name, &c.Phone, &c.Email, &c.Role, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contact not found")
	}
	return c, err
}

func (s *PostgresStore) ListContactsForProperty(ctx context.Context, propertyID int64) ([]models.Contact, error) {
	query := `SELECT id, property_id, name, phone, email, role, notes, created_at, updated_at
		FROM contacts WHERE property_id = $1 ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var c models.Contact
		if err := rows.Scan(&c.ID, &c.PropertyID, &c.Name, &c.Phone, &c.Email, &c.Role, &c.Notes,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

func (s *PostgresStore) UpdateContact(ctx context.Context, c *models.Contact) error {
	query := `
		UPDATE contacts
		SET name = $1, phone = $2, email = $3, role = $4, notes = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at`
	return s.db.QueryRowContext(ctx, query, c.Name, c.Phone, c.Email, c.Role, c.Notes, c.ID).
		Scan(&c.UpdatedAt)
}

func (s *PostgresStore) DeleteContact(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM contacts WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("contact not found")
	}
	return nil
}

// Attachments
func (s *PostgresStore) CreateAttachment(ctx context.Context, a *models.Attachment) error {
	query := `
		INSERT INTO attachments (property_id, filename, description, storage_type, storage_path, file_size, mime_type, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, a.PropertyID, a.Filename, a.Description, a.StorageType,
		a.StoragePath, a.FileSize, a.MimeType, a.UploadedBy).Scan(&a.ID, &a.CreatedAt)
}

func (s *PostgresStore) GetAttachment(ctx context.Context, id int64) (*models.Attachment, error) {
	a := &models.Attachment{}
	query := `SELECT id, property_id, filename, description, storage_type, storage_path, file_size, mime_type, uploaded_by, created_at
		FROM attachments WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.PropertyID, &a.Filename, &a.Description, &a.StorageType, &a.StoragePath,
		&a.FileSize, &a.MimeType, &a.UploadedBy, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("attachment not found")
	}
	return a, err
}

func (s *PostgresStore) ListAttachmentsForProperty(ctx context.Context, propertyID int64) ([]models.Attachment, error) {
	query := `SELECT id, property_id, filename, description, storage_type, storage_path, file_size, mime_type, uploaded_by, created_at
		FROM attachments WHERE property_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []models.Attachment
	for rows.Next() {
		var a models.Attachment
		if err := rows.Scan(&a.ID, &a.PropertyID, &a.Filename, &a.Description, &a.StorageType,
			&a.StoragePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}

func (s *PostgresStore) DeleteAttachment(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM attachments WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("attachment not found")
	}
	return nil
}

// Devices
func (s *PostgresStore) CreateDevice(ctx context.Context, d *models.Device) error {
	query := `
		INSERT INTO devices (property_id, name, hostname, device_type, is_critical, check_interval, retries, timeout, description, tags, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(ctx, query, d.PropertyID, d.Name, d.Hostname, d.DeviceType, d.IsCritical,
		d.CheckInterval, d.Retries, d.Timeout, d.Description, pq.Array(d.Tags), d.Active).
		Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (s *PostgresStore) GetDevice(ctx context.Context, id int64) (*models.Device, error) {
	d := &models.Device{}
	query := `SELECT id, property_id, name, hostname, device_type, is_critical, check_interval, retries, timeout, description, tags, active, created_at, updated_at
		FROM devices WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.PropertyID, &d.Name, &d.Hostname, &d.DeviceType, &d.IsCritical, &d.CheckInterval,
		&d.Retries, &d.Timeout, &d.Description, pq.Array(&d.Tags), &d.Active, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	return d, err
}

func (s *PostgresStore) ListDevices(ctx context.Context) ([]models.Device, error) {
	query := `SELECT id, property_id, name, hostname, device_type, is_critical, check_interval, retries, timeout, description, tags, active, created_at, updated_at
		FROM devices ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.PropertyID, &d.Name, &d.Hostname, &d.DeviceType, &d.IsCritical,
			&d.CheckInterval, &d.Retries, &d.Timeout, &d.Description, pq.Array(&d.Tags), &d.Active,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *PostgresStore) ListDevicesForProperty(ctx context.Context, propertyID int64) ([]models.Device, error) {
	query := `SELECT id, property_id, name, hostname, device_type, is_critical, check_interval, retries, timeout, description, tags, active, created_at, updated_at
		FROM devices WHERE property_id = $1 ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.PropertyID, &d.Name, &d.Hostname, &d.DeviceType, &d.IsCritical,
			&d.CheckInterval, &d.Retries, &d.Timeout, &d.Description, pq.Array(&d.Tags), &d.Active,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *PostgresStore) ListActiveDevices(ctx context.Context) ([]models.Device, error) {
	query := `SELECT id, property_id, name, hostname, device_type, is_critical, check_interval, retries, timeout, description, tags, active, created_at, updated_at
		FROM devices WHERE active = true ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.PropertyID, &d.Name, &d.Hostname, &d.DeviceType, &d.IsCritical,
			&d.CheckInterval, &d.Retries, &d.Timeout, &d.Description, pq.Array(&d.Tags), &d.Active,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *PostgresStore) UpdateDevice(ctx context.Context, d *models.Device) error {
	query := `
		UPDATE devices
		SET property_id = $1, name = $2, hostname = $3, device_type = $4, is_critical = $5,
		    check_interval = $6, retries = $7, timeout = $8, description = $9, tags = $10, active = $11, updated_at = NOW()
		WHERE id = $12
		RETURNING updated_at`
	return s.db.QueryRowContext(ctx, query, d.PropertyID, d.Name, d.Hostname, d.DeviceType, d.IsCritical,
		d.CheckInterval, d.Retries, d.Timeout, d.Description, pq.Array(d.Tags), d.Active, d.ID).
		Scan(&d.UpdatedAt)
}

func (s *PostgresStore) DeleteDevice(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM devices WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

// Notification Channels
func (s *PostgresStore) CreateNotificationChannel(ctx context.Context, nc *models.NotificationChannel) error {
	query := `
		INSERT INTO notification_channels (name, type, config, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(ctx, query, nc.Name, nc.Type, nc.Config, nc.Enabled).
		Scan(&nc.ID, &nc.CreatedAt, &nc.UpdatedAt)
}

func (s *PostgresStore) GetNotificationChannel(ctx context.Context, id int64) (*models.NotificationChannel, error) {
	nc := &models.NotificationChannel{}
	query := `SELECT id, name, type, config, enabled, created_at, updated_at
		FROM notification_channels WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&nc.ID, &nc.Name, &nc.Type, &nc.Config, &nc.Enabled, &nc.CreatedAt, &nc.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification channel not found")
	}
	return nc, err
}

func (s *PostgresStore) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	query := `SELECT id, name, type, config, enabled, created_at, updated_at
		FROM notification_channels ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.NotificationChannel
	for rows.Next() {
		var nc models.NotificationChannel
		if err := rows.Scan(&nc.ID, &nc.Name, &nc.Type, &nc.Config, &nc.Enabled,
			&nc.CreatedAt, &nc.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, nc)
	}
	return channels, rows.Err()
}

func (s *PostgresStore) UpdateNotificationChannel(ctx context.Context, nc *models.NotificationChannel) error {
	query := `
		UPDATE notification_channels
		SET name = $1, type = $2, config = $3, enabled = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`
	return s.db.QueryRowContext(ctx, query, nc.Name, nc.Type, nc.Config, nc.Enabled, nc.ID).
		Scan(&nc.UpdatedAt)
}

func (s *PostgresStore) DeleteNotificationChannel(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM notification_channels WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification channel not found")
	}
	return nil
}

// Property Notifications
func (s *PostgresStore) CreatePropertyNotification(ctx context.Context, pn *models.PropertyNotification) error {
	query := `
		INSERT INTO property_notifications (property_id, notification_channel_id, enabled, notify_on_red, notify_on_recovery)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	return s.db.QueryRowContext(ctx, query, pn.PropertyID, pn.NotificationChannelID, pn.Enabled,
		pn.NotifyOnRed, pn.NotifyOnRecovery).Scan(&pn.ID)
}

func (s *PostgresStore) ListPropertyNotifications(ctx context.Context, propertyID int64) ([]models.PropertyNotification, error) {
	query := `SELECT id, property_id, notification_channel_id, enabled, notify_on_red, notify_on_recovery
		FROM property_notifications WHERE property_id = $1`
	rows, err := s.db.QueryContext(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.PropertyNotification
	for rows.Next() {
		var pn models.PropertyNotification
		if err := rows.Scan(&pn.ID, &pn.PropertyID, &pn.NotificationChannelID, &pn.Enabled,
			&pn.NotifyOnRed, &pn.NotifyOnRecovery); err != nil {
			return nil, err
		}
		notifications = append(notifications, pn)
	}
	return notifications, rows.Err()
}

func (s *PostgresStore) UpdatePropertyNotification(ctx context.Context, pn *models.PropertyNotification) error {
	query := `
		UPDATE property_notifications
		SET enabled = $1, notify_on_red = $2, notify_on_recovery = $3
		WHERE id = $4`
	_, err := s.db.ExecContext(ctx, query, pn.Enabled, pn.NotifyOnRed, pn.NotifyOnRecovery, pn.ID)
	return err
}

func (s *PostgresStore) DeletePropertyNotification(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM property_notifications WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("property notification not found")
	}
	return nil
}

// Notification Events
func (s *PostgresStore) CreateNotificationEvent(ctx context.Context, ne *models.NotificationEvent) error {
	query := `
		INSERT INTO notification_events (property_id, notification_channel_id, event_type, message, success, error)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, ne.PropertyID, ne.NotificationChannelID, ne.EventType,
		ne.Message, ne.Success, ne.Error).Scan(&ne.ID, &ne.CreatedAt)
}

func (s *PostgresStore) ListNotificationEvents(ctx context.Context, propertyID int64, limit int) ([]models.NotificationEvent, error) {
	query := `SELECT id, property_id, notification_channel_id, event_type, message, success, error, created_at
		FROM notification_events WHERE property_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := s.db.QueryContext(ctx, query, propertyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.NotificationEvent
	for rows.Next() {
		var ne models.NotificationEvent
		if err := rows.Scan(&ne.ID, &ne.PropertyID, &ne.NotificationChannelID, &ne.EventType,
			&ne.Message, &ne.Success, &ne.Error, &ne.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ne)
	}
	return events, rows.Err()
}

// Users
func (s *PostgresStore) CreateUser(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (username, password, email, role, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(ctx, query, u.Username, u.Password, u.Email, u.Role, u.Active).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (s *PostgresStore) GetUser(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, username, password, email, role, active, created_at, updated_at
		FROM users WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Password, &u.Email, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (s *PostgresStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, username, password, email, role, active, created_at, updated_at
		FROM users WHERE username = $1`
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Password, &u.Email, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (s *PostgresStore) ListUsers(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, username, password, email, role, active, created_at, updated_at
		FROM users ORDER BY username`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Email, &u.Role, &u.Active,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *PostgresStore) UpdateUser(ctx context.Context, u *models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, role = $3, active = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`
	return s.db.QueryRowContext(ctx, query, u.Username, u.Email, u.Role, u.Active, u.ID).
		Scan(&u.UpdatedAt)
}

func (s *PostgresStore) UpdateUserPassword(ctx context.Context, userID int64, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, hashedPassword, userID)
	return err
}

func (s *PostgresStore) DeleteUser(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// Settings
func (s *PostgresStore) GetSettings(ctx context.Context) (*models.Settings, error) {
	settings := &models.Settings{}
	query := `SELECT id, max_concurrent_pings, default_check_interval, default_retries,
		default_timeout, history_retention_days, notification_cooldown
		FROM settings LIMIT 1`
	err := s.db.QueryRowContext(ctx, query).Scan(
		&settings.ID, &settings.MaxConcurrentPings, &settings.DefaultCheckInterval,
		&settings.DefaultRetries, &settings.DefaultTimeout, &settings.HistoryRetentionDays,
		&settings.NotificationCooldown)
	if err == sql.ErrNoRows {
		// Return defaults
		return &models.Settings{
			MaxConcurrentPings:   150,
			DefaultCheckInterval: 60,
			DefaultRetries:       3,
			DefaultTimeout:       10000,
			HistoryRetentionDays: 90,
			NotificationCooldown: 300,
		}, nil
	}
	return settings, err
}

func (s *PostgresStore) UpdateSettings(ctx context.Context, settings *models.Settings) error {
	query := `
		UPDATE settings
		SET max_concurrent_pings = $1, default_check_interval = $2, default_retries = $3,
		    default_timeout = $4, history_retention_days = $5, notification_cooldown = $6
		WHERE id = $7`
	_, err := s.db.ExecContext(ctx, query, settings.MaxConcurrentPings, settings.DefaultCheckInterval,
		settings.DefaultRetries, settings.DefaultTimeout, settings.HistoryRetentionDays,
		settings.NotificationCooldown, settings.ID)
	return err
}

// Helper to unmarshal JSON config
func unmarshalConfig(configJSON string, v interface{}) error {
	return json.Unmarshal([]byte(configJSON), v)
}
