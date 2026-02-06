# ETS Properties - Enhancement Implementation Guide

This document outlines the implementation of three key enhancements requested:

1. Google OAuth Authentication ("Login with Google")
2. Timestamped Notes/Comments System
3. Property Maps Feature

---

## 1. Google OAuth Authentication

### Overview
Replace JWT-based authentication with Google OAuth 2.0 "Sign in with Google" flow.

### Implementation Steps

#### Backend Changes

**1.1 Update Dependencies (go.mod)**
```go
require (
    // Add these
    golang.org/x/oauth2 v0.20.0
    google.golang.org/api v0.177.0
)
```

**1.2 Create OAuth Configuration (internal/api/oauth.go)**
```go
package api

import (
    "context"
    "encoding/json"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
    Scopes: []string{
        "https://www.googleapis.com/auth/userinfo.email",
        "https://www.googleapis.com/auth/userinfo.profile",
    },
    Endpoint: google.Endpoint,
}

type GoogleUserInfo struct {
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    Picture       string `json:"picture"`
}

func (s *Server) handleGoogleLogin(c *gin.Context) {
    url := googleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
    c.JSON(http.StatusOK, gin.H{"url": url})
}

func (s *Server) handleGoogleCallback(c *gin.Context) {
    code := c.Query("code")

    token, err := googleOAuthConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to exchange token"})
        return
    }

    client := googleOAuthConfig.Client(context.Background(), token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get user info"})
        return
    }
    defer resp.Body.Close()

    var userInfo GoogleUserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to decode user info"})
        return
    }

    // Find or create user by email
    user, err := s.postgres.GetUserByEmail(context.Background(), userInfo.Email)
    if err != nil {
        // Create new user
        user = &models.User{
            Username: userInfo.Email,
            Email:    userInfo.Email,
            Role:     "user",
            Active:   true,
        }
        if err := s.postgres.CreateUser(context.Background(), user); err != nil {
            c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user"})
            return
        }
    }

    // Generate JWT token
    jwtToken, err := generateToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, models.LoginResponse{
        Token: jwtToken,
        User:  *user,
    })
}
```

**1.3 Add to Router (internal/api/router.go)**
```go
// Public routes
router.GET("/auth/google/login", s.handleGoogleLogin)
router.GET("/auth/google/callback", s.handleGoogleCallback)
```

**1.4 Add GetUserByEmail to PostgresStore**
```go
func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    u := &models.User{}
    query := `SELECT id, username, password, email, role, active, created_at, updated_at
        FROM users WHERE email = $1`
    err := s.db.QueryRowContext(ctx, query, email).Scan(
        &u.ID, &u.Username, &u.Password, &u.Email, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    return u, err
}
```

#### Frontend Changes

**1.5 Update LoginPage.tsx**
```tsx
import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { apiClient } from '../api/client'

export default function LoginPage() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  useEffect(() => {
    // Handle OAuth callback
    const code = searchParams.get('code')
    if (code) {
      handleOAuthCallback(code)
    }
  }, [searchParams])

  const handleOAuthCallback = async (code: string) => {
    setLoading(true)
    try {
      const response = await apiClient.googleCallback(code)
      apiClient.setToken(response.token)
      navigate('/')
    } catch (err: any) {
      setError(err.message || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  const handleGoogleLogin = async () => {
    setLoading(true)
    try {
      const { url } = await apiClient.googleLogin()
      window.location.href = url
    } catch (err: any) {
      setError(err.message || 'Failed to initiate Google login')
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        <h1 className="text-2xl font-bold text-center mb-6">
          ETS NOC (Network Operations Center)
        </h1>

        {error && (
          <div className="mb-4 text-red-600 text-sm text-center">{error}</div>
        )}

        <button
          onClick={handleGoogleLogin}
          disabled={loading}
          className="w-full flex items-center justify-center gap-3 bg-white border-2 border-gray-300 text-gray-700 py-3 px-4 rounded-md hover:bg-gray-50 disabled:bg-gray-100 disabled:cursor-not-allowed"
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
            <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
            <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
          </svg>
          {loading ? 'Signing in...' : 'Sign in with Google'}
        </button>
      </div>
    </div>
  )
}
```

**1.6 Add OAuth Methods to API Client**
```typescript
async googleLogin() {
  return this.request<{ url: string }>('/auth/google/login')
}

async googleCallback(code: string) {
  return this.request<{ token: string; user: any }>('/auth/google/callback', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
}
```

#### Environment Variables
```bash
# Backend
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=https://properties.etsusa.com/auth/google/callback

# Get credentials from: https://console.cloud.google.com/apis/credentials
```

---

## 2. Timestamped Notes/Comments System

### Overview
Convert the `notes` field to a timestamped comment system showing which user made each comment.

### Implementation Steps

#### Database Changes

**2.1 Create Comments Table**
```sql
CREATE TABLE IF NOT EXISTS property_comments (
    id BIGSERIAL PRIMARY KEY,
    property_id BIGINT NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_property_comments_property_id ON property_comments(property_id);
CREATE INDEX idx_property_comments_created_at ON property_comments(created_at DESC);
```

**2.2 Migration Script**
```sql
-- Migrate existing notes to comments
INSERT INTO property_comments (property_id, user_id, comment, created_at)
SELECT
    id,
    1, -- admin user ID
    notes,
    created_at
FROM properties
WHERE notes IS NOT NULL AND notes != '';

-- Optionally drop notes column
-- ALTER TABLE properties DROP COLUMN notes;
```

#### Backend Changes

**2.3 Add Comment Model (models.go)**
```go
type PropertyComment struct {
    ID         int64     `json:"id"`
    PropertyID int64     `json:"property_id"`
    UserID     int64     `json:"user_id"`
    Username   string    `json:"username"`
    Comment    string    `json:"comment"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

**2.4 Add PostgresStore Methods**
```go
func (s *PostgresStore) CreatePropertyComment(ctx context.Context, pc *models.PropertyComment) error {
    query := `
        INSERT INTO property_comments (property_id, user_id, comment)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
    return s.db.QueryRowContext(ctx, query, pc.PropertyID, pc.UserID, pc.Comment).
        Scan(&pc.ID, &pc.CreatedAt, &pc.UpdatedAt)
}

func (s *PostgresStore) ListPropertyComments(ctx context.Context, propertyID int64) ([]models.PropertyComment, error) {
    query := `
        SELECT pc.id, pc.property_id, pc.user_id, u.username, pc.comment, pc.created_at, pc.updated_at
        FROM property_comments pc
        JOIN users u ON pc.user_id = u.id
        WHERE pc.property_id = $1
        ORDER BY pc.created_at DESC`

    rows, err := s.db.QueryContext(ctx, query, propertyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var comments []models.PropertyComment
    for rows.Next() {
        var pc models.PropertyComment
        if err := rows.Scan(&pc.ID, &pc.PropertyID, &pc.UserID, &pc.Username, &pc.Comment,
            &pc.CreatedAt, &pc.UpdatedAt); err != nil {
            return nil, err
        }
        comments = append(comments, pc)
    }
    return comments, rows.Err()
}

func (s *PostgresStore) DeletePropertyComment(ctx context.Context, id int64, userID int64) error {
    // Only allow users to delete their own comments, or admins to delete any
    query := `DELETE FROM property_comments WHERE id = $1 AND user_id = $2`
    result, err := s.db.ExecContext(ctx, query, id, userID)
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return fmt.Errorf("comment not found or permission denied")
    }
    return nil
}
```

**2.5 Add API Handlers**
```go
func (s *Server) handleListPropertyComments(c *gin.Context) {
    propertyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
        return
    }

    comments, err := s.postgres.ListPropertyComments(context.Background(), propertyID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, comments)
}

func (s *Server) handleCreatePropertyComment(c *gin.Context) {
    propertyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid property ID"})
        return
    }

    var req struct {
        Comment string `json:"comment" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
        return
    }

    userID, _ := c.Get("user_id")

    comment := &models.PropertyComment{
        PropertyID: propertyID,
        UserID:     userID.(int64),
        Comment:    req.Comment,
    }

    if err := s.postgres.CreatePropertyComment(context.Background(), comment); err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, comment)
}

func (s *Server) handleDeletePropertyComment(c *gin.Context) {
    commentID, err := strconv.ParseInt(c.Param("commentId"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid comment ID"})
        return
    }

    userID, _ := c.Get("user_id")
    role, _ := c.Get("role")

    // Admins can delete any comment
    if role == "admin" {
        if err := s.postgres.DeletePropertyComment(context.Background(), commentID, 0); err != nil {
            c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
            return
        }
    } else {
        if err := s.postgres.DeletePropertyComment(context.Background(), commentID, userID.(int64)); err != nil {
            c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Permission denied"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}
```

**2.6 Add Routes**
```go
api.GET("/properties/:id/comments", s.handleListPropertyComments)
api.POST("/properties/:id/comments", s.handleCreatePropertyComment)
api.DELETE("/comments/:commentId", s.handleDeletePropertyComment)
```

#### Frontend Changes

**2.7 Create CommentsList Component**
```tsx
import { useState } from 'react'
import { apiClient } from '../api/client'

interface CommentsListProps {
  comments: any[]
  propertyId: number
  currentUserId: number
  onUpdate: () => void
}

export default function CommentsList({ comments, propertyId, currentUserId, onUpdate }: CommentsListProps) {
  const [newComment, setNewComment] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newComment.trim()) return

    setSubmitting(true)
    try {
      await apiClient.createPropertyComment(propertyId, newComment)
      setNewComment('')
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (commentId: number) => {
    if (!confirm('Delete this comment?')) return
    try {
      await apiClient.deletePropertyComment(commentId)
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString()
  }

  return (
    <div className="space-y-4">
      <form onSubmit={handleSubmit} className="mb-4">
        <textarea
          value={newComment}
          onChange={(e) => setNewComment(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md"
          rows={3}
          placeholder="Add a comment..."
          disabled={submitting}
        />
        <button
          type="submit"
          disabled={submitting || !newComment.trim()}
          className="mt-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
        >
          {submitting ? 'Adding...' : 'Add Comment'}
        </button>
      </form>

      <div className="space-y-3">
        {comments.map((comment) => (
          <div key={comment.id} className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-medium text-sm">{comment.username}</span>
                  <span className="text-xs text-gray-500">
                    {formatDate(comment.created_at)}
                  </span>
                </div>
                <p className="text-gray-700 whitespace-pre-wrap">{comment.comment}</p>
              </div>
              {comment.user_id === currentUserId && (
                <button
                  onClick={() => handleDelete(comment.id)}
                  className="text-red-600 hover:text-red-700 text-sm"
                >
                  Delete
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {comments.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          No comments yet. Be the first to add one!
        </div>
      )}
    </div>
  )
}
```

**2.8 Add Comments Tab to PropertyDetailModal**
```tsx
// Add to tabs array
{ id: 'comments', label: 'Comments' }

// Add state
const [comments, setComments] = useState<any[]>([])

// Add load function
const loadComments = async () => {
  setLoading(true)
  try {
    const data = await apiClient.getPropertyComments(property.id)
    setComments(data)
  } catch (error) {
    console.error('Failed to load comments:', error)
  } finally {
    setLoading(false)
  }
}

// Add to useEffect
else if (activeTab === 'comments') {
  loadComments()
}

// Add to content rendering
{activeTab === 'comments' && (
  <CommentsList
    comments={comments}
    propertyId={property.id}
    currentUserId={user.id}
    onUpdate={loadComments}
  />
)}
```

---

## 3. Property Maps Feature

### Overview
Add a special category of attachments called "Property Maps" that are displayed prominently in the UI for quick glanceable information.

### Implementation Steps

#### Database Changes

**3.1 Add is_map Column to Attachments**
```sql
ALTER TABLE attachments ADD COLUMN is_map BOOLEAN DEFAULT false;
CREATE INDEX idx_attachments_is_map ON attachments(property_id, is_map) WHERE is_map = true;
```

#### Backend Changes

**3.2 Update Attachment Model**
```go
type Attachment struct {
    ID          int64     `json:"id"`
    PropertyID  int64     `json:"property_id"`
    Filename    string    `json:"filename"`
    Description string    `json:"description"`
    StorageType string    `json:"storage_type"`
    StoragePath string    `json:"storage_path"`
    FileSize    int64     `json:"file_size"`
    MimeType    string    `json:"mime_type"`
    UploadedBy  string    `json:"uploaded_by"`
    IsMap       bool      `json:"is_map"`      // NEW
    CreatedAt   time.Time `json:"created_at"`
}
```

**3.3 Add PostgresStore Method**
```go
func (s *PostgresStore) ListPropertyMaps(ctx context.Context, propertyID int64) ([]models.Attachment, error) {
    query := `SELECT id, property_id, filename, description, storage_type, storage_path,
              file_size, mime_type, uploaded_by, is_map, created_at
        FROM attachments
        WHERE property_id = $1 AND is_map = true
        ORDER BY created_at DESC`

    rows, err := s.db.QueryContext(ctx, query, propertyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var maps []models.Attachment
    for rows.Next() {
        var a models.Attachment
        if err := rows.Scan(&a.ID, &a.PropertyID, &a.Filename, &a.Description, &a.StorageType,
            &a.StoragePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.IsMap, &a.CreatedAt); err != nil {
            return nil, err
        }
        maps = append(maps, a)
    }
    return maps, rows.Err()
}
```

**3.4 Update Upload Handler**
```go
func (s *Server) handleUploadAttachment(c *gin.Context) {
    // ... existing code ...

    isMap := c.PostForm("is_map") == "true"

    attachment := &models.Attachment{
        PropertyID:  propertyID,
        Filename:    file.Filename,
        Description: description,
        StorageType: "gcs",
        StoragePath: objectName,
        FileSize:    file.Size,
        MimeType:    file.Header.Get("Content-Type"),
        UploadedBy:  username.(string),
        IsMap:       isMap,  // NEW
    }

    // ... rest of code ...
}
```

**3.5 Add Routes**
```go
api.GET("/properties/:id/maps", s.handleListPropertyMaps)
```

#### Frontend Changes

**3.6 Create PropertyMaps Component**
```tsx
import { useState } from 'react'
import { apiClient } from '../api/client'

interface PropertyMapsProps {
  maps: any[]
  propertyId: number
  onUpdate: () => void
}

export default function PropertyMaps({ maps, propertyId, onUpdate }: PropertyMapsProps) {
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [file, setFile] = useState<File | null>(null)
  const [description, setDescription] = useState('')
  const [selectedMap, setSelectedMap] = useState<any>(null)

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file) return

    setUploading(true)
    try {
      await apiClient.uploadAttachment(propertyId, file, description, true)
      setShowUploadModal(false)
      setFile(null)
      setDescription('')
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    } finally {
      setUploading(false)
    }
  }

  const handleView = async (map: any) => {
    try {
      const { url } = await apiClient.getAttachmentDownloadUrl(map.id)
      setSelectedMap({ ...map, url })
    } catch (error: any) {
      alert(error.message)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this map?')) return
    try {
      await apiClient.deleteAttachment(id)
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">Property Maps ({maps.length})</h3>
        <button
          onClick={() => setShowUploadModal(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Upload Map
        </button>
      </div>

      {/* Map Gallery */}
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        {maps.map((map) => (
          <div
            key={map.id}
            className="bg-gray-50 rounded-lg p-3 cursor-pointer hover:shadow-lg transition-shadow"
            onClick={() => handleView(map)}
          >
            <div className="aspect-video bg-gray-200 rounded mb-2 flex items-center justify-center">
              <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" />
              </svg>
            </div>
            <h4 className="font-medium text-sm truncate">{map.filename}</h4>
            {map.description && (
              <p className="text-xs text-gray-600 truncate mt-1">{map.description}</p>
            )}
            <button
              onClick={(e) => {
                e.stopPropagation()
                handleDelete(map.id)
              }}
              className="mt-2 text-xs text-red-600 hover:text-red-700"
            >
              Delete
            </button>
          </div>
        ))}
      </div>

      {maps.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No property maps uploaded yet
        </div>
      )}

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-xl font-bold mb-4">Upload Property Map</h3>
            <form onSubmit={handleUpload}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Image File (JPG, PNG, PDF)
                  </label>
                  <input
                    type="file"
                    accept="image/*,application/pdf"
                    onChange={(e) => setFile(e.target.files?.[0] || null)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Description
                  </label>
                  <input
                    type="text"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    placeholder="e.g., Floor 1 Network Layout"
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <button
                  type="button"
                  onClick={() => {
                    setShowUploadModal(false)
                    setFile(null)
                    setDescription('')
                  }}
                  className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
                  disabled={uploading}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
                  disabled={uploading}
                >
                  {uploading ? 'Uploading...' : 'Upload'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Map Viewer Modal */}
      {selectedMap && (
        <div className="fixed inset-0 bg-black bg-opacity-90 flex items-center justify-center z-50 p-4">
          <div className="max-w-6xl w-full">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-white text-xl font-bold">{selectedMap.filename}</h3>
              <button
                onClick={() => setSelectedMap(null)}
                className="text-white text-2xl hover:text-gray-300"
              >
                ×
              </button>
            </div>
            <img
              src={selectedMap.url}
              alt={selectedMap.filename}
              className="w-full h-auto max-h-[80vh] object-contain rounded"
            />
          </div>
        </div>
      )}
    </div>
  )
}
```

**3.7 Add Maps Tab to PropertyDetailModal**
```tsx
// Add to tabs
{ id: 'maps', label: 'Maps' }

// Add state
const [maps, setMaps] = useState<any[]>([])

// Add load function
const loadMaps = async () => {
  setLoading(true)
  try {
    const data = await apiClient.getPropertyMaps(property.id)
    setMaps(data)
  } catch (error) {
    console.error('Failed to load maps:', error)
  } finally {
    setLoading(false)
  }
}

// Add to useEffect
else if (activeTab === 'maps') {
  loadMaps()
}

// Add to content
{activeTab === 'maps' && (
  <PropertyMaps
    maps={maps}
    propertyId={property.id}
    onUpdate={loadMaps}
  />
)}
```

**3.8 Add Maps Preview to Property Card**
```tsx
{/* Show first map as thumbnail if available */}
{property.has_maps && (
  <div className="mt-2">
    <div className="text-xs text-gray-500 mb-1">📍 Property Map Available</div>
  </div>
)}
```

---

## Testing Checklist

### Google OAuth
- [ ] Google OAuth consent screen configured
- [ ] Credentials created with correct redirect URI
- [ ] Login redirects to Google
- [ ] Callback creates or finds user by email
- [ ] JWT token generated and stored
- [ ] User can access protected routes

### Timestamped Notes
- [ ] Comments table created
- [ ] Existing notes migrated
- [ ] New comments show username and timestamp
- [ ] Users can add comments
- [ ] Users can delete their own comments
- [ ] Admins can delete any comment
- [ ] Comments ordered by date (newest first)

### Property Maps
- [ ] Maps upload with is_map flag
- [ ] Maps displayed in grid layout
- [ ] Maps clickable to view full size
- [ ] Maps deletable
- [ ] Only image/PDF files allowed
- [ ] Property card shows map indicator

---

## Deployment

After implementing these features:

1. Update database schema in production
2. Add new environment variables to k8s secrets
3. Rebuild and deploy updated images
4. Test each feature in production
5. Update user documentation

---

## Future Considerations

- **Map Annotations**: Allow users to annotate maps with device locations
- **Comment Mentions**: @mention users in comments for notifications
- **Comment Reactions**: Add emoji reactions to comments
- **Map Versioning**: Track changes to property maps over time
- **Mobile App**: Native mobile app with offline map viewing
