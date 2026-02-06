package models

import "time"

// Property represents a physical property location
type Property struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Subnet          string    `json:"subnet"`
	Notes           string    `json:"notes"`
	ISPCompanyName  string    `json:"isp_company_name"`
	ISPAccountInfo  string    `json:"isp_account_info"`
	PfSenseHost     string    `json:"pfsense_host"`
	PfSensePort     int       `json:"pfsense_port"`
	PfSenseUsername string    `json:"pfsense_username"`
	PfSensePassword string    `json:"pfsense_password,omitempty"` // omitempty for security
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PropertyWithStatus includes computed status
type PropertyWithStatus struct {
	Property
	Status          string `json:"status"`
	OnlineCount     int    `json:"online_count"`
	OfflineCount    int    `json:"offline_count"`
	TotalCount      int    `json:"total_count"`
	CriticalOffline bool   `json:"critical_offline"`
	LastCheck       string `json:"last_check"`
}

// PropertyStatus represents the computed rollup status
type PropertyStatus struct {
	PropertyID      int64     `json:"property_id"`
	Status          string    `json:"status"` // red, yellow, green
	OnlineCount     int       `json:"online_count"`
	OfflineCount    int       `json:"offline_count"`
	TotalCount      int       `json:"total_count"`
	CriticalOffline bool      `json:"critical_offline"`
	LastCheck       time.Time `json:"last_check"`
}

// Contact represents a contact for a property
type Contact struct {
	ID         int64     `json:"id"`
	PropertyID int64     `json:"property_id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Attachment represents a file attachment for a property
type Attachment struct {
	ID          int64     `json:"id"`
	PropertyID  int64     `json:"property_id"`
	Filename    string    `json:"filename"`
	Description string    `json:"description"`
	StorageType string    `json:"storage_type"` // gcs or google_drive
	StoragePath string    `json:"storage_path"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// Device represents a network device to monitor
type Device struct {
	ID            int64    `json:"id"`
	PropertyID    int64    `json:"property_id"`
	Name          string   `json:"name"`
	Hostname      string   `json:"hostname"`
	DeviceType    string   `json:"device_type"`
	IsCritical    bool     `json:"is_critical"`
	CheckInterval int      `json:"check_interval"`
	Retries       int      `json:"retries"`
	Timeout       int      `json:"timeout"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Active        bool     `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DeviceStatus represents the current status of a device
type DeviceStatus struct {
	DeviceID     int64     `json:"device_id"`
	Status       string    `json:"status"` // online or offline
	ResponseTime float64   `json:"response_time"`
	LastCheck    time.Time `json:"last_check"`
	Message      string    `json:"message"`
}

// DeviceHistory represents historical status data point
type DeviceHistory struct {
	Timestamp    int64   `json:"timestamp"`
	Status       string  `json:"status"`
	ResponseTime float64 `json:"response_time"`
	Message      string  `json:"message,omitempty"`
}

// NotificationChannel represents a notification destination
type NotificationChannel struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // slack, email
	Config      string    `json:"config"` // JSON config
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PropertyNotification links properties to notification channels
type PropertyNotification struct {
	ID                    int64 `json:"id"`
	PropertyID            int64 `json:"property_id"`
	NotificationChannelID int64 `json:"notification_channel_id"`
	Enabled               bool  `json:"enabled"`
	NotifyOnRed           bool  `json:"notify_on_red"`
	NotifyOnRecovery      bool  `json:"notify_on_recovery"`
}

// NotificationEvent tracks notification history
type NotificationEvent struct {
	ID                    int64     `json:"id"`
	PropertyID            int64     `json:"property_id"`
	NotificationChannelID int64     `json:"notification_channel_id"`
	EventType             string    `json:"event_type"` // property_down, property_recovery
	Message               string    `json:"message"`
	Success               bool      `json:"success"`
	Error                 string    `json:"error"`
	CreatedAt             time.Time `json:"created_at"`
}

// User represents a system user
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // admin, user
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Settings represents system-wide settings
type Settings struct {
	ID                     int64  `json:"id"`
	MaxConcurrentPings     int    `json:"max_concurrent_pings"`
	DefaultCheckInterval   int    `json:"default_check_interval"`
	DefaultRetries         int    `json:"default_retries"`
	DefaultTimeout         int    `json:"default_timeout"`
	HistoryRetentionDays   int    `json:"history_retention_days"`
	NotificationCooldown   int    `json:"notification_cooldown"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse contains JWT token
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// DashboardResponse contains all properties with status
type DashboardResponse struct {
	Properties []PropertyWithStatus `json:"properties"`
	Summary    struct {
		TotalProperties int `json:"total_properties"`
		RedCount        int `json:"red_count"`
		YellowCount     int `json:"yellow_count"`
		GreenCount      int `json:"green_count"`
	} `json:"summary"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
}
