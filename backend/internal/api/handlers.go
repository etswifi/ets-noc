package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/etswifi/ets-noc/internal/gcs"
	"github.com/etswifi/ets-noc/internal/models"
	"github.com/etswifi/ets-noc/internal/monitor"
	"github.com/etswifi/ets-noc/internal/storage"
)

type Server struct {
	postgres *storage.PostgresStore
	redis    *storage.RedisStore
	gcs      *gcs.Client
}

func NewServer(postgres *storage.PostgresStore, redis *storage.RedisStore, gcsClient *gcs.Client) *Server {
	return &Server{
		postgres: postgres,
		redis:    redis,
		gcs:      gcsClient,
	}
}

// Health check
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Dashboard
func (s *Server) handleDashboard(c *gin.Context) {
	properties, err := s.postgres.ListProperties(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Get all property statuses from Redis
	propertyStatuses, err := s.redis.GetAllPropertyStatuses(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	var propertiesWithStatus []models.PropertyWithStatus
	redCount, yellowCount, greenCount := 0, 0, 0

	for _, prop := range properties {
		pws := models.PropertyWithStatus{
			Property: prop,
			Status:   "green",
		}

		if status, ok := propertyStatuses[prop.ID]; ok {
			pws.Status = status.Status
			pws.OnlineCount = status.OnlineCount
			pws.OfflineCount = status.OfflineCount
			pws.TotalCount = status.TotalCount
			pws.CriticalOffline = status.CriticalOffline
			pws.LastCheck = status.LastCheck.Format(time.RFC3339)

			switch status.Status {
			case "red":
				redCount++
			case "yellow":
				yellowCount++
			case "green":
				greenCount++
			}
		} else {
			greenCount++
		}

		propertiesWithStatus = append(propertiesWithStatus, pws)
	}

	response := models.DashboardResponse{
		Properties: propertiesWithStatus,
	}
	response.Summary.TotalProperties = len(properties)
	response.Summary.RedCount = redCount
	response.Summary.YellowCount = yellowCount
	response.Summary.GreenCount = greenCount

	c.JSON(http.StatusOK, response)
}

// Properties
func (s *Server) handleListProperties(c *gin.Context) {
	properties, err := s.postgres.ListProperties(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, properties)
}

func (s *Server) handleGetProperty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	property, err := s.postgres.GetProperty(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Property not found"})
		return
	}

	c.JSON(http.StatusOK, property)
}

func (s *Server) handleCreateProperty(c *gin.Context) {
	var property models.Property
	if err := c.ShouldBindJSON(&property); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := s.postgres.CreateProperty(context.Background(), &property); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, property)
}

func (s *Server) handleUpdateProperty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	var property models.Property
	if err := c.ShouldBindJSON(&property); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	property.ID = id
	if err := s.postgres.UpdateProperty(context.Background(), &property); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, property)
}

func (s *Server) handleDeleteProperty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	if err := s.postgres.DeleteProperty(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Property deleted"})
}

func (s *Server) handleGetPropertyStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	// Get property devices
	devices, err := s.postgres.ListDevicesForProperty(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Compute status
	statusComputer := monitor.NewStatusComputer(s.postgres, s.redis)
	status, err := statusComputer.ComputePropertyStatus(context.Background(), id, devices)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (s *Server) handleGetPropertyDevices(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	devices, err := s.postgres.ListDevicesForProperty(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// Contacts
func (s *Server) handleListContactsForProperty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	contacts, err := s.postgres.ListContactsForProperty(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, contacts)
}

func (s *Server) handleCreateContact(c *gin.Context) {
	propertyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	contact.PropertyID = propertyID
	if err := s.postgres.CreateContact(context.Background(), &contact); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, contact)
}

func (s *Server) handleGetContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid contact ID"})
		return
	}

	contact, err := s.postgres.GetContact(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Contact not found"})
		return
	}

	c.JSON(http.StatusOK, contact)
}

func (s *Server) handleUpdateContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid contact ID"})
		return
	}

	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	contact.ID = id
	if err := s.postgres.UpdateContact(context.Background(), &contact); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, contact)
}

func (s *Server) handleDeleteContact(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid contact ID"})
		return
	}

	if err := s.postgres.DeleteContact(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted"})
}

// Attachments
func (s *Server) handleListAttachmentsForProperty(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	attachments, err := s.postgres.ListAttachmentsForProperty(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, attachments)
}

func (s *Server) handleUploadAttachment(c *gin.Context) {
	propertyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "No file provided"})
		return
	}

	description := c.PostForm("description")
	username, _ := c.Get("username")

	// Check file size (max 50MB)
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "File too large (max 50MB)"})
		return
	}

	// Generate unique filename
	objectName := fmt.Sprintf("properties/%d/%d-%s", propertyID, time.Now().Unix(), file.Filename)

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to read file"})
		return
	}
	defer fileReader.Close()

	// Upload to GCS
	if err := s.gcs.UploadFile(context.Background(), objectName, fileReader, file.Header.Get("Content-Type")); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Failed to upload: %v", err)})
		return
	}

	// Create attachment record
	attachment := &models.Attachment{
		PropertyID:  propertyID,
		Filename:    file.Filename,
		Description: description,
		StorageType: "gcs",
		StoragePath: objectName,
		FileSize:    file.Size,
		MimeType:    file.Header.Get("Content-Type"),
		UploadedBy:  username.(string),
	}

	if err := s.postgres.CreateAttachment(context.Background(), attachment); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

func (s *Server) handleDownloadAttachment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid attachment ID"})
		return
	}

	attachment, err := s.postgres.GetAttachment(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Attachment not found"})
		return
	}

	if attachment.StorageType == "gcs" {
		// Generate signed URL (valid for 1 hour)
		url, err := s.gcs.GetSignedURL(context.Background(), attachment.StoragePath, time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate download URL"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": url})
	} else if attachment.StorageType == "google_drive" {
		// Return the Google Drive link directly
		c.JSON(http.StatusOK, gin.H{"url": attachment.StoragePath})
	} else {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Unsupported storage type"})
	}
}

func (s *Server) handleDeleteAttachment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid attachment ID"})
		return
	}

	attachment, err := s.postgres.GetAttachment(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Attachment not found"})
		return
	}

	// Delete from GCS if applicable
	if attachment.StorageType == "gcs" {
		if err := s.gcs.DeleteFile(context.Background(), attachment.StoragePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete file"})
			return
		}
	}

	// Delete database record
	if err := s.postgres.DeleteAttachment(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted"})
}

// Devices
func (s *Server) handleListDevices(c *gin.Context) {
	devices, err := s.postgres.ListDevices(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, devices)
}

func (s *Server) handleGetDevice(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device ID"})
		return
	}

	device, err := s.postgres.GetDevice(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (s *Server) handleCreateDevice(c *gin.Context) {
	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := s.postgres.CreateDevice(context.Background(), &device); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func (s *Server) handleUpdateDevice(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device ID"})
		return
	}

	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	device.ID = id
	if err := s.postgres.UpdateDevice(context.Background(), &device); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (s *Server) handleDeleteDevice(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device ID"})
		return
	}

	if err := s.postgres.DeleteDevice(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted"})
}

func (s *Server) handleGetDeviceStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device ID"})
		return
	}

	status, err := s.redis.GetDeviceStatus(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Device status not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (s *Server) handleGetDeviceHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device ID"})
		return
	}

	// Default to last 24 hours
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}

	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	history, err := s.redis.GetDeviceHistory(context.Background(), id, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// Users
func (s *Server) handleListUsers(c *gin.Context) {
	users, err := s.postgres.ListUsers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (s *Server) handleCreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
		return
	}
	user.Password = hashedPassword

	if err := s.postgres.CreateUser(context.Background(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (s *Server) handleUpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	user.ID = id
	if err := s.postgres.UpdateUser(context.Background(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (s *Server) handleDeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	if err := s.postgres.DeleteUser(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// Settings
func (s *Server) handleGetSettings(c *gin.Context) {
	settings, err := s.postgres.GetSettings(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (s *Server) handleUpdateSettings(c *gin.Context) {
	var settings models.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := s.postgres.UpdateSettings(context.Background(), &settings); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
