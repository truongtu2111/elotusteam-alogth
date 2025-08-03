# 🚀 Monitoring Quick Reference Card

## 🎯 Essential Commands

### Start/Stop Services
```bash
# Start monitoring stack
./scripts/start-monitoring.sh

# Start with Docker Compose
cd monitoring && docker-compose up -d

# Stop all services
docker-compose down

# Restart specific service
docker-compose restart prometheus
```

### Health Checks
```bash
# Check all services status
docker-compose ps

# Check specific service logs
docker-compose logs -f grafana

# Test endpoints
curl http://localhost:9090/api/v1/query?query=up  # Prometheus
curl http://localhost:3000/api/health             # Grafana
curl http://localhost:8081/metrics                # Auth service
```

## 🌐 Access Points

| Service | URL | Default Credentials |
|---------|-----|--------------------|
| Grafana | http://localhost:3000 | admin/admin |
| Prometheus | http://localhost:9090 | - |
| Alertmanager | http://localhost:9093 | - |
| Jaeger | http://localhost:16686 | - |
| cAdvisor | http://localhost:8080 | - |
| Node Exporter | http://localhost:9100 | - |

## 📊 Key Metrics Endpoints

```bash
# Microservice metrics
http://localhost:8081/metrics  # Auth Service
http://localhost:8082/metrics  # File Service
http://localhost:8083/metrics  # User Service
http://localhost:8085/metrics  # Analytics Service

# System metrics
http://localhost:9100/metrics  # Node Exporter
http://localhost:8080/metrics  # cAdvisor
```

## 🔍 Common Prometheus Queries

```promql
# Service availability
up

# Request rate (per second)
rate(http_requests_total[5m])

# Error rate percentage
(rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])) * 100

# 95th percentile response time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Memory usage by service
container_memory_usage_bytes{name=~".*auth-service.*"}

# CPU usage percentage
rate(container_cpu_usage_seconds_total[5m]) * 100
```

## 🚨 Alert Status Check

```bash
# Check active alerts
curl http://localhost:9093/api/v1/alerts

# Check Prometheus alert rules
curl http://localhost:9090/api/v1/rules

# Silence an alert (example)
curl -X POST http://localhost:9093/api/v1/silences \
  -H "Content-Type: application/json" \
  -d '{"matchers":[{"name":"alertname","value":"ServiceDown"}],"startsAt":"2024-01-01T00:00:00Z","endsAt":"2024-01-01T01:00:00Z","comment":"Maintenance window"}'
```

## 🔧 Troubleshooting Commands

```bash
# Check Docker network
docker network ls
docker network inspect microservice-project_microservice-network

# Check service connectivity
docker exec -it microservice-project-prometheus-1 wget -qO- http://grafana:3000/api/health

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check Grafana data sources
curl -u admin:admin http://localhost:3000/api/datasources

# View container resource usage
docker stats --no-stream
```

## 📁 Important File Locations

```
monitoring/
├── prometheus/
│   ├── prometheus.yml      # Main Prometheus config
│   └── alert_rules.yml     # Alert definitions
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/    # Data source configs
│   │   └── dashboards/     # Dashboard provisioning
│   └── dashboards/         # Dashboard JSON files
└── alertmanager/
    └── alertmanager.yml    # Alert routing config
```

## 🔄 Configuration Reload

```bash
# Reload Prometheus config (without restart)
curl -X POST http://localhost:9090/-/reload

# Reload Alertmanager config
curl -X POST http://localhost:9093/-/reload

# Restart Grafana to pick up new dashboards
docker-compose restart grafana
```

## 📈 Performance Optimization

```bash
# Check Prometheus storage usage
du -sh /var/lib/docker/volumes/microservice-project_prometheus_data

# Clean up old data (adjust retention)
docker-compose exec prometheus \
  promtool tsdb create-blocks-from openmetrics \
  --retention.time=7d /prometheus

# Check Grafana performance
curl -u admin:admin http://localhost:3000/api/admin/stats
```

## 🛠️ Backup Commands

```bash
# Backup Prometheus data
docker run --rm -v microservice-project_prometheus_data:/data \
  -v $(pwd):/backup alpine tar czf /backup/prometheus-backup.tar.gz /data

# Backup Grafana data
docker run --rm -v microservice-project_grafana_data:/data \
  -v $(pwd):/backup alpine tar czf /backup/grafana-backup.tar.gz /data

# Export Grafana dashboards
curl -u admin:admin http://localhost:3000/api/search?type=dash-db | \
  jq -r '.[] | .uid' | \
  xargs -I {} curl -u admin:admin http://localhost:3000/api/dashboards/uid/{} > dashboard-{}.json
```

## 🔐 Security Commands

```bash
# Change Grafana admin password
docker exec -it microservice-project-grafana-1 \
  grafana-cli admin reset-admin-password newpassword

# Create Grafana API key
curl -X POST -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{"name":"monitoring-api","role":"Viewer"}' \
  http://localhost:3000/api/auth/keys

# Check Prometheus security
curl http://localhost:9090/api/v1/status/flags
```

## 📱 Mobile/Remote Access

```bash
# Tunnel for remote access (SSH)
ssh -L 3000:localhost:3000 -L 9090:localhost:9090 user@your-server

# Or use ngrok for temporary access
ngrok http 3000  # For Grafana
ngrok http 9090  # For Prometheus
```

---

**💡 Pro Tips:**
- Use `docker-compose logs -f service-name` for real-time log monitoring
- Bookmark the Grafana dashboards for quick access
- Set up browser bookmarks for all monitoring URLs
- Use Grafana mobile app for on-the-go monitoring
- Create custom Grafana playlists for automated dashboard rotation

**🆘 Emergency Contacts:**
- Prometheus not responding: Check Docker daemon and restart services
- Grafana login issues: Reset admin password using CLI command above
- No metrics: Verify microservice `/metrics` endpoints are accessible
- High resource usage: Check retention settings and reduce scrape frequency