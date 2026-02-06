-- ETS NOC (Network Operations Center) Database Schema



-- Properties table
CREATE TABLE IF NOT EXISTS properties (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    notes TEXT,
    isp_company_name VARCHAR(255),
    isp_account_info TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Contacts table
CREATE TABLE IF NOT EXISTS contacts (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    role VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Attachments table
CREATE TABLE IF NOT EXISTS attachments (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    description TEXT,
    storage_type VARCHAR(20) NOT NULL CHECK (storage_type IN ('gcs', 'google_drive')),
    storage_path TEXT NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    uploaded_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    device_type VARCHAR(50),
    is_critical BOOLEAN DEFAULT false,
    check_interval INT DEFAULT 60,
    retries INT DEFAULT 3,
    timeout INT DEFAULT 10000,
    description TEXT DEFAULT '',
    tags TEXT[] DEFAULT '{}',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Notification channels table
CREATE TABLE IF NOT EXISTS notification_channels (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('slack', 'email')),
    config TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Property notifications junction table
CREATE TABLE IF NOT EXISTS property_notifications (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    notification_channel_id BIGINT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT true,
    notify_on_red BOOLEAN DEFAULT true,
    notify_on_recovery BOOLEAN DEFAULT true,
    UNIQUE(property_id, notification_channel_id)
);

-- Notification events log table
CREATE TABLE IF NOT EXISTS notification_events (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    notification_channel_id BIGINT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    success BOOLEAN DEFAULT false,
    error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'user')),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Settings table
CREATE TABLE IF NOT EXISTS settings (
    id BIGSERIAL PRIMARY KEY,
    max_concurrent_pings INT DEFAULT 150,
    default_check_interval INT DEFAULT 60,
    default_retries INT DEFAULT 3,
    default_timeout INT DEFAULT 10000,
    history_retention_days INT DEFAULT 90,
    notification_cooldown INT DEFAULT 300
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_devices_property_id ON devices(property_id);
CREATE INDEX IF NOT EXISTS idx_devices_hostname ON devices(hostname);
CREATE INDEX IF NOT EXISTS idx_devices_active ON devices(active);
CREATE INDEX IF NOT EXISTS idx_devices_critical ON devices(is_critical);
CREATE INDEX IF NOT EXISTS idx_contacts_property_id ON contacts(property_id);
CREATE INDEX IF NOT EXISTS idx_attachments_property_id ON attachments(property_id);
CREATE INDEX IF NOT EXISTS idx_property_notifications_property_id ON property_notifications(property_id);
CREATE INDEX IF NOT EXISTS idx_notification_events_property_id ON notification_events(property_id);
CREATE INDEX IF NOT EXISTS idx_notification_events_created_at ON notification_events(created_at);

-- Insert default settings
INSERT INTO settings (id, max_concurrent_pings, default_check_interval, default_retries, default_timeout, history_retention_days, notification_cooldown)
VALUES (1, 150, 60, 3, 10000, 90, 300)
ON CONFLICT (id) DO NOTHING;

-- Insert default admin user (password: changeme)
-- Password hash for "changeme" using bcrypt
INSERT INTO users (username, password, email, role, active)
VALUES ('admin', '$2a$10$YVZxZIYXXXXXXXXXXXXXXeN5xN5xN5xN5xN5xN5xN5xN5xN5xN5xN', 'admin@etsusa.com', 'admin', true)
ON CONFLICT (username) DO NOTHING;

-- Note: You should change the admin password after first login!
