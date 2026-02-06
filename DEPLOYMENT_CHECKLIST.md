# ETS NOC (Network Operations Center) System - Deployment Checklist

Use this checklist to ensure successful deployment of the ETS NOC (Network Operations Center) System.

---

## Pre-Deployment

### ☐ 1. Prerequisites Installed
- [ ] gcloud CLI installed and configured
- [ ] kubectl installed
- [ ] Docker installed (for local testing)
- [ ] PostgreSQL client (psql) installed
- [ ] Git installed

### ☐ 2. GCP Project Setup
- [ ] GCP project created
- [ ] Project ID noted: `_______________`
- [ ] Billing enabled
- [ ] Required APIs enabled:
  - [ ] Kubernetes Engine API
  - [ ] Cloud SQL Admin API
  - [ ] Cloud Storage API
  - [ ] Cloud Build API
  - [ ] Container Registry API

### ☐ 3. Permissions Verified
- [ ] You have Owner or Editor role on GCP project
- [ ] Service accounts can be created
- [ ] Can create GKE clusters
- [ ] Can create Cloud SQL instances

---

## Infrastructure Setup

### ☐ 4. Cloud SQL Database
- [ ] Cloud SQL instance created: `properties-db`
- [ ] PostgreSQL version: 15
- [ ] Region selected: `_______________`
- [ ] Tier selected: `db-g1-small` or higher
- [ ] Root password set and saved securely
- [ ] Database created: `ets_properties`
- [ ] Connection name noted: `_______________`

### ☐ 5. GCS Bucket
- [ ] Bucket created: `ets-noc-attachments`
- [ ] Region: Same as Cloud SQL
- [ ] Uniform access control enabled
- [ ] Service account has Storage Object Admin role

### ☐ 6. GKE Cluster
- [ ] Cluster created: `ets-noc`
- [ ] Region: Same as Cloud SQL and GCS
- [ ] Number of nodes: 3+
- [ ] Machine type: `n1-standard-2` or higher
- [ ] Autoscaling enabled
- [ ] Workload Identity enabled (optional but recommended)

### ☐ 7. Service Accounts
- [ ] Cloud SQL client service account created
- [ ] Service account has Cloud SQL Client role
- [ ] Workload Identity configured (if using)

---

## Database Initialization

### ☐ 8. Schema Applied
- [ ] Cloud SQL proxy downloaded
- [ ] Connected to database via proxy
- [ ] `schema.sql` executed successfully
- [ ] Tables verified: `SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';`
- [ ] Default admin user exists: `SELECT * FROM users WHERE username = 'admin';`
- [ ] Default settings exist: `SELECT * FROM settings;`

---

## Build & Push Images

### ☐ 9. Backend API Image
- [ ] Navigated to `/backend` directory
- [ ] `cloudbuild-api.yaml` updated with correct PROJECT_ID (if needed)
- [ ] Build submitted: `gcloud builds submit --config cloudbuild-api.yaml`
- [ ] Build succeeded
- [ ] Image available: `gcr.io/PROJECT_ID/ets-noc-api:latest`

### ☐ 10. Backend Worker Image
- [ ] Still in `/backend` directory
- [ ] `cloudbuild-worker.yaml` updated with correct PROJECT_ID (if needed)
- [ ] Build submitted: `gcloud builds submit --config cloudbuild-worker.yaml`
- [ ] Build succeeded
- [ ] Image available: `gcr.io/PROJECT_ID/ets-noc-worker:latest`

### ☐ 11. Frontend Image
- [ ] Navigated to `/frontend` directory
- [ ] `cloudbuild.yaml` updated with correct PROJECT_ID (if needed)
- [ ] Build submitted: `gcloud builds submit --config cloudbuild.yaml`
- [ ] Build succeeded
- [ ] Image available: `gcr.io/PROJECT_ID/ets-noc-frontend:latest`

---

## Kubernetes Configuration

### ☐ 12. Update Manifests
- [ ] All YAML files in `/k8s` directory updated with:
  - [ ] Correct PROJECT_ID in image references
  - [ ] Correct Cloud SQL connection name
  - [ ] Correct GCS bucket name
  - [ ] Correct region

### ☐ 13. Get Cluster Credentials
- [ ] `gcloud container clusters get-credentials ets-noc --region=REGION`
- [ ] `kubectl cluster-info` works
- [ ] Context set to correct cluster

### ☐ 14. Create Namespace
- [ ] `kubectl apply -f k8s/namespace.yaml`
- [ ] Namespace exists: `kubectl get namespace ets-noc`

### ☐ 15. Create Secrets
- [ ] PostgreSQL connection string prepared
- [ ] Secrets created:
```bash
kubectl create secret generic ets-noc-secrets \
  --namespace=ets-noc \
  --from-literal=postgres-url="CONNECTION_STRING"
```
- [ ] Secrets verified: `kubectl get secrets -n ets-noc`

### ☐ 16. Create ConfigMap
- [ ] `kubectl apply -f k8s/configmap.yaml`
- [ ] ConfigMap verified: `kubectl get configmap -n ets-noc`

---

## Deploy Applications

### ☐ 17. Deploy Redis
- [ ] `kubectl apply -f k8s/redis.yaml`
- [ ] PVC created: `kubectl get pvc -n ets-noc`
- [ ] Deployment created: `kubectl get deployment ets-noc-redis -n ets-noc`
- [ ] Pod running: `kubectl get pods -n ets-noc -l app=ets-noc-redis`
- [ ] Service created: `kubectl get svc ets-noc-redis -n ets-noc`
- [ ] Redis working: `kubectl exec -n ets-noc deployment/ets-noc-redis -- redis-cli PING`

### ☐ 18. Deploy API
- [ ] `kubectl apply -f k8s/api.yaml`
- [ ] Deployment created
- [ ] 2 pods running: `kubectl get pods -n ets-noc -l app=ets-noc-api`
- [ ] Both containers running (api + cloudsql-proxy)
- [ ] Service created
- [ ] HPA created: `kubectl get hpa -n ets-noc`
- [ ] Health check: `kubectl exec -n ets-noc deployment/ets-noc-api -- wget -qO- localhost:8080/health`

### ☐ 19. Deploy Worker
- [ ] `kubectl apply -f k8s/worker.yaml`
- [ ] Deployment created
- [ ] 1 pod running: `kubectl get pods -n ets-noc -l app=ets-noc-worker`
- [ ] Both containers running (worker + cloudsql-proxy)
- [ ] NET_RAW capability granted
- [ ] Worker logs show pinger started: `kubectl logs -n ets-noc deployment/ets-noc-worker --tail=50`

### ☐ 20. Deploy Frontend
- [ ] `kubectl apply -f k8s/frontend.yaml`
- [ ] Deployment created
- [ ] 2 pods running: `kubectl get pods -n ets-noc -l app=ets-noc-frontend`
- [ ] Service created

### ☐ 21. Deploy Load Balancer
- [ ] `kubectl apply -f k8s/loadbalancer.yaml`
- [ ] LoadBalancer service created
- [ ] External IP assigned (may take 2-5 minutes)
- [ ] External IP noted: `_______________`

---

## Verification

### ☐ 22. All Pods Running
```bash
kubectl get pods -n ets-noc
```
Expected output:
- [ ] ets-noc-api-xxxxx (2/2 Running)
- [ ] ets-noc-worker-xxxxx (2/2 Running)
- [ ] ets-noc-frontend-xxxxx (1/1 Running)
- [ ] ets-noc-redis-xxxxx (1/1 Running)

### ☐ 23. Services Healthy
```bash
kubectl get svc -n ets-noc
```
- [ ] ets-noc-api (ClusterIP)
- [ ] ets-noc-frontend (ClusterIP)
- [ ] ets-noc-redis (ClusterIP)
- [ ] ets-noc-lb (LoadBalancer with EXTERNAL-IP)

### ☐ 24. Health Checks Passing
- [ ] API health: `curl http://EXTERNAL_IP/health`
- [ ] Frontend health: `curl http://EXTERNAL_IP`
- [ ] Response codes are 200 OK

### ☐ 25. Database Connectivity
```bash
kubectl logs -n ets-noc deployment/ets-noc-api --tail=20
```
- [ ] Logs show "Connected to PostgreSQL"
- [ ] Logs show "Connected to Redis"
- [ ] No connection errors

### ☐ 26. Worker Functionality
```bash
kubectl logs -n ets-noc deployment/ets-noc-worker --tail=50
```
- [ ] Logs show "Pinger started with max concurrent pings: 150"
- [ ] Logs show "Checking X devices"
- [ ] No permission errors (NET_RAW)

---

## First Access

### ☐ 27. Login Test
- [ ] Navigate to: `http://EXTERNAL_IP`
- [ ] Login page loads
- [ ] Login with:
  - Username: `admin`
  - Password: `changeme`
- [ ] Login successful
- [ ] Dashboard loads

### ☐ 28. Change Admin Password
- [ ] Navigate to user settings/profile
- [ ] Change password from `changeme` to secure password
- [ ] Log out and log back in with new password
- [ ] Secure password saved in password manager

---

## Data Migration

### ☐ 29. Import Properties
Method: Manual or Migration Tool
- [ ] 179 properties imported
- [ ] Property names correct
- [ ] Addresses populated
- [ ] ISP information added (if available)

### ☐ 30. Import Devices
Method: Bulk CSV import or API
- [ ] ~3,600 devices imported
- [ ] Property assignments correct
- [ ] Hostnames/IPs valid
- [ ] Device types set (WAP/switch/router)
- [ ] Critical devices marked

### ☐ 31. Add Contacts
- [ ] Key contacts added for critical properties
- [ ] Contact information verified
- [ ] Roles assigned

### ☐ 32. Verify Monitoring
- [ ] Wait 2-3 minutes for first ping cycle
- [ ] Check dashboard for status updates
- [ ] Verify property statuses computing
- [ ] Check device statuses in Redis:
```bash
kubectl exec -n ets-noc deployment/ets-noc-redis -- \
  redis-cli HLEN all_device_status
```
- [ ] Count matches number of active devices

---

## Security Hardening

### ☐ 33. Network Security
- [ ] Cloud SQL configured for private IP only (optional)
- [ ] GKE cluster has network policies (optional)
- [ ] Firewall rules reviewed
- [ ] Load balancer has health check configured

### ☐ 34. TLS/HTTPS Setup
- [ ] Domain name configured (if applicable)
- [ ] cert-manager installed (optional)
- [ ] TLS certificate issued
- [ ] HTTPS redirect enabled
- [ ] HTTP to HTTPS upgrade tested

### ☐ 35. Access Control
- [ ] GCP IAM roles reviewed
- [ ] Kubernetes RBAC configured
- [ ] Service accounts have minimum required permissions
- [ ] Default admin password changed

### ☐ 36. Secrets Management
- [ ] All secrets in Kubernetes secrets (not ConfigMaps)
- [ ] Secrets not in version control
- [ ] Database password rotated from default
- [ ] JWT secret set (if not using default)

---

## Monitoring & Logging

### ☐ 37. Cloud Logging
- [ ] Logs flowing to Cloud Logging
- [ ] Log filters created for errors
- [ ] Log-based metrics created (optional)

### ☐ 38. Cloud Monitoring
- [ ] Workspaces created (if needed)
- [ ] Uptime checks configured:
  - [ ] API health endpoint
  - [ ] Frontend availability
- [ ] Alerting policies created:
  - [ ] API pods down
  - [ ] Worker pod down
  - [ ] High error rate
  - [ ] High latency

### ☐ 39. Dashboards
- [ ] Cloud Monitoring dashboard created
- [ ] Key metrics added:
  - [ ] Request rate
  - [ ] Error rate
  - [ ] Latency
  - [ ] Pod CPU/Memory
  - [ ] Device status counts

---

## Performance Testing

### ☐ 40. Load Testing
- [ ] Dashboard loads in <2 seconds
- [ ] API responds in <100ms average
- [ ] Can handle 179 property cards
- [ ] Search/filter responsive
- [ ] Modal opens quickly

### ☐ 41. Worker Performance
- [ ] Monitoring all 3,600 devices
- [ ] Check interval: 60 seconds
- [ ] No dropped pings
- [ ] No timeout errors
- [ ] CPU usage acceptable (<50%)
- [ ] Memory usage acceptable (<1GB)

### ☐ 42. Scaling Test
- [ ] API HPA triggers at 70% CPU
- [ ] Additional pods created when needed
- [ ] Pods scale back down after load decreases

---

## Backup & Recovery

### ☐ 43. Database Backups
- [ ] Automated backups enabled on Cloud SQL
- [ ] Backup retention period set: ___ days
- [ ] Point-in-time recovery enabled
- [ ] Test backup restore (optional but recommended)

### ☐ 44. Disaster Recovery Plan
- [ ] Recovery Time Objective (RTO) defined: ___
- [ ] Recovery Point Objective (RPO) defined: ___
- [ ] Backup restore procedure documented
- [ ] Incident response plan created

---

## Documentation

### ☐ 45. Internal Documentation
- [ ] Deployment process documented
- [ ] Architecture diagrams created
- [ ] Runbook created for common tasks:
  - [ ] Adding properties/devices
  - [ ] User management
  - [ ] Troubleshooting
  - [ ] Scaling
- [ ] Incident response procedures documented

### ☐ 46. User Training
- [ ] User guide created
- [ ] Training session scheduled
- [ ] Admin training completed
- [ ] User support process defined

---

## Post-Deployment

### ☐ 47. User Acceptance Testing
- [ ] Key stakeholders have access
- [ ] Feedback collected
- [ ] Issues logged and prioritized
- [ ] Success criteria met

### ☐ 48. Performance Monitoring (Week 1)
- [ ] Monitor CPU/memory usage daily
- [ ] Check for errors in logs
- [ ] Verify device monitoring accuracy
- [ ] Collect user feedback
- [ ] Document any issues

### ☐ 49. Optimization (Week 2-4)
- [ ] Identify performance bottlenecks
- [ ] Optimize slow queries
- [ ] Adjust resource limits if needed
- [ ] Tune HPA settings
- [ ] Implement user-requested features

---

## Enhancements (Future)

### ☐ 50. Google OAuth Integration
See `ENHANCEMENTS.md` Section 1
- [ ] Google OAuth credentials obtained
- [ ] Backend OAuth handler implemented
- [ ] Frontend login page updated
- [ ] Tested and deployed

### ☐ 51. Timestamped Comments
See `ENHANCEMENTS.md` Section 2
- [ ] Comments table created
- [ ] Existing notes migrated
- [ ] Comment UI implemented
- [ ] Tested and deployed

### ☐ 52. Property Maps
See `ENHANCEMENTS.md` Section 3
- [ ] is_map column added
- [ ] Map upload functionality implemented
- [ ] Map gallery UI implemented
- [ ] Tested and deployed

---

## Sign-Off

**Deployment completed by**: _______________

**Date**: _______________

**Production URL**: _______________

**Notes**:
```
[Add any deployment-specific notes, issues, or deviations from plan]
```

**Stakeholder Approval**:
- [ ] Technical Lead: _______________
- [ ] Product Owner: _______________
- [ ] Operations: _______________

---

## Emergency Contacts

**Development Team**: _______________

**GCP Support**: _______________

**On-Call Rotation**: _______________

---

## Useful Commands Reference

### View Logs
```bash
# API logs
kubectl logs -n ets-noc deployment/ets-noc-api -f

# Worker logs
kubectl logs -n ets-noc deployment/ets-noc-worker -f

# Frontend logs
kubectl logs -n ets-noc deployment/ets-noc-frontend -f
```

### Restart Deployment
```bash
kubectl rollout restart deployment/ets-noc-api -n ets-noc
kubectl rollout restart deployment/ets-noc-worker -n ets-noc
```

### Scale Deployment
```bash
kubectl scale deployment ets-noc-api --replicas=5 -n ets-noc
```

### Check Status
```bash
kubectl get all -n ets-noc
kubectl top pods -n ets-noc
kubectl get events -n ets-noc --sort-by='.lastTimestamp'
```

### Database Access
```bash
# Via Cloud SQL Proxy
cloud_sql_proxy -instances=CONNECTION_NAME=tcp:5432 &
psql "host=127.0.0.1 port=5432 dbname=ets_properties user=postgres"
```

---

**Congratulations! Your ETS NOC (Network Operations Center) System is now deployed! 🎉**
