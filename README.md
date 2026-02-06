# ETS NOC (Network Operations Center) System

A property-centric infrastructure monitoring system for 179 ETS properties with ~3,600 devices (WAPs, switches, routers).

## Architecture

- **Backend API**: Go/Gin REST API (port 8080)
- **Worker**: ICMP pinger with property status rollup
- **Frontend**: React/TypeScript SPA with Tailwind CSS
- **Database**: PostgreSQL (Cloud SQL) for metadata
- **Cache**: Redis for real-time status + 90-day history
- **Storage**: Google Cloud Storage for file attachments
- **Deployment**: Google Kubernetes Engine (GKE)

## Status Logic

- **RED**: All devices offline OR any `is_critical` device offline
- **YELLOW**: Some (not all) devices offline, no critical devices offline
- **GREEN**: All devices online

## Project Structure

```
ets-noc/
├── backend/
│   ├── cmd/
│   │   ├── api/main.go           # API server entry point
│   │   └── worker/main.go        # Worker entry point
│   ├── internal/
│   │   ├── models/models.go      # Data models
│   │   ├── storage/              # PostgreSQL & Redis
│   │   ├── api/                  # HTTP handlers & routing
│   │   ├── monitor/              # Pinger & status computer
│   │   └── gcs/                  # GCS client
│   ├── schema.sql                # Database schema
│   ├── Dockerfile.api
│   ├── Dockerfile.worker
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/client.ts         # API client
│   │   ├── contexts/AuthContext.tsx
│   │   ├── pages/                # Login & Dashboard
│   │   └── components/           # UI components
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
└── k8s/                          # Kubernetes manifests
    ├── namespace.yaml
    ├── configmap.yaml
    ├── redis.yaml
    ├── api.yaml
    ├── worker.yaml
    ├── frontend.yaml
    └── loadbalancer.yaml
```

## Setup Instructions

### 1. Create GCP Resources

```bash
# Set project
export PROJECT_ID=your-project-id
gcloud config set project $PROJECT_ID

# Create Cloud SQL instance
gcloud sql instances create properties-db \
  --database-version=POSTGRES_15 \
  --tier=db-g1-small \
  --region=us-central1

# Create database
gcloud sql databases create ets_properties --instance=properties-db

# Create GCS bucket
gsutil mb -l us-central1 gs://ets-noc-attachments

# Create service account for GCS
gcloud iam service-accounts create ets-noc-sa
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:ets-noc-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"
```

### 2. Initialize Database

```bash
# Get Cloud SQL connection string
POSTGRES_URL="postgres://postgres:PASSWORD@/ets_properties?host=/cloudsql/PROJECT_ID:REGION:properties-db"

# Apply schema
psql "$POSTGRES_URL" < backend/schema.sql
```

### 3. Build and Deploy

```bash
# Build images
cd backend
gcloud builds submit --config cloudbuild-api.yaml
gcloud builds submit --config cloudbuild-worker.yaml

cd ../frontend
gcloud builds submit --config cloudbuild.yaml

# Create GKE cluster (if needed)
gcloud container clusters create ets-noc \
  --region us-central1 \
  --num-nodes 3 \
  --machine-type n1-standard-2

# Get credentials
gcloud container clusters get-credentials ets-noc --region us-central1

# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create secrets
kubectl create secret generic ets-noc-secrets \
  --namespace=ets-noc \
  --from-literal=postgres-url="$POSTGRES_URL"

# Deploy all resources
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/api.yaml
kubectl apply -f k8s/worker.yaml
kubectl apply -f k8s/frontend.yaml
kubectl apply -f k8s/loadbalancer.yaml
```

### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -n ets-noc

# Check services
kubectl get svc -n ets-noc

# Get load balancer IP
kubectl get svc ets-noc-lb -n ets-noc

# View logs
kubectl logs -n ets-noc deployment/ets-noc-api -f
kubectl logs -n ets-noc deployment/ets-noc-worker -f
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - Login with username/password
- `GET /api/v1/auth/me` - Get current user

### Dashboard
- `GET /api/v1/dashboard` - Get all properties with status

### Properties
- `GET /api/v1/properties` - List all properties
- `POST /api/v1/properties` - Create property
- `GET /api/v1/properties/:id` - Get property details
- `PUT /api/v1/properties/:id` - Update property
- `DELETE /api/v1/properties/:id` - Delete property
- `GET /api/v1/properties/:id/status` - Get property status
- `GET /api/v1/properties/:id/devices` - List property devices

### Contacts
- `GET /api/v1/properties/:id/contacts` - List contacts
- `POST /api/v1/properties/:id/contacts` - Create contact
- `GET /api/v1/contacts/:id` - Get contact
- `PUT /api/v1/contacts/:id` - Update contact
- `DELETE /api/v1/contacts/:id` - Delete contact

### Attachments
- `GET /api/v1/properties/:id/attachments` - List attachments
- `POST /api/v1/properties/:id/attachments` - Upload file
- `GET /api/v1/attachments/:id/download` - Download file
- `DELETE /api/v1/attachments/:id` - Delete attachment

### Devices
- `GET /api/v1/devices` - List all devices
- `POST /api/v1/devices` - Create device
- `GET /api/v1/devices/:id` - Get device
- `PUT /api/v1/devices/:id` - Update device
- `DELETE /api/v1/devices/:id` - Delete device
- `GET /api/v1/devices/:id/status` - Get device status
- `GET /api/v1/devices/:id/history` - Get device history

### Admin (Admin role required)
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `GET /api/v1/settings` - Get settings
- `PUT /api/v1/settings` - Update settings

## Default Credentials

```
Username: admin
Password: changeme
```

**IMPORTANT**: Change the default password after first login!

## Configuration

### Environment Variables (API)
- `POSTGRES_URL` - PostgreSQL connection string
- `REDIS_ADDR` - Redis address (default: localhost:6379)
- `REDIS_PASSWORD` - Redis password (optional)
- `GCS_BUCKET` - GCS bucket name for attachments
- `PORT` - API server port (default: 8080)

### Environment Variables (Worker)
- `POSTGRES_URL` - PostgreSQL connection string
- `REDIS_ADDR` - Redis address (default: localhost:6379)
- `REDIS_PASSWORD` - Redis password (optional)

### Settings (Configurable via API)
- `max_concurrent_pings` - Max concurrent ICMP pings (default: 150)
- `default_check_interval` - Device check interval in seconds (default: 60)
- `default_retries` - Ping retries (default: 3)
- `default_timeout` - Ping timeout in ms (default: 10000)
- `history_retention_days` - Redis history retention (default: 90)
- `notification_cooldown` - Notification cooldown in seconds (default: 300)

## Monitoring

### Health Checks
- API: `GET /health` - Returns 200 OK
- Frontend: `GET /health` - Returns 200 OK

### Metrics
Monitor these key metrics:
- Worker ping rate and success rate
- API response times
- Redis memory usage
- Property status distribution (red/yellow/green)
- Device online/offline counts

### Logs
```bash
# API logs
kubectl logs -n ets-noc deployment/ets-noc-api --tail=100 -f

# Worker logs
kubectl logs -n ets-noc deployment/ets-noc-worker --tail=100 -f

# Redis logs
kubectl logs -n ets-noc deployment/ets-noc-redis --tail=100 -f
```

## Performance Considerations

- **Worker**: Single replica only (no distributed coordination)
- **Concurrency**: 150 max concurrent pings for 3,600 devices
- **Check Interval**: 60 seconds per device
- **History**: 90 days stored in Redis
- **Attachments**: Max 50MB per file

## Security

- JWT-based authentication with 24-hour expiration
- Passwords hashed with bcrypt
- Role-based access control (admin/user)
- GCS signed URLs for secure file downloads (1-hour expiration)
- Cloud SQL proxy for secure database connections

## Development

### Local Development

```bash
# Backend
cd backend
export POSTGRES_URL="postgres://localhost/ets_properties"
export REDIS_ADDR="localhost:6379"
export GCS_BUCKET="ets-noc-attachments"
go run cmd/api/main.go

# Worker
go run cmd/worker/main.go

# Frontend
cd frontend
npm install
npm run dev
```

### Testing

```bash
# Test API
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"changeme"}'

# Test dashboard
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/dashboard
```

## Troubleshooting

### Worker not pinging devices
- Check NET_RAW capability in worker deployment
- Verify devices table has active=true
- Check worker logs for errors

### Property status not updating
- Verify Redis is running and accessible
- Check that devices are assigned to properties
- Review worker logs for status computation errors

### Attachments not uploading
- Verify GCS bucket exists and is accessible
- Check service account permissions
- Ensure file size is under 50MB

## Future Enhancements

- [ ] Google OAuth authentication (replacing JWT)
- [ ] Timestamped notes/comments feature
- [ ] Property maps as special attachments
- [ ] Notification system (Slack, email)
- [ ] Device detail modal with charts
- [ ] Bulk device import from CSV
- [ ] Mobile-responsive improvements

## License

Proprietary - ETS USA

## Support

For issues or questions, contact the development team.
