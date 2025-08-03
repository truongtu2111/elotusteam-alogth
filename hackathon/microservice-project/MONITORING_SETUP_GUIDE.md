# 📊 Complete Monitoring Setup Guide

This guide provides step-by-step instructions for setting up, configuring, and maintaining the comprehensive monitoring stack for the microservices architecture.

## 🎯 Overview

The monitoring solution includes:
- **Prometheus** - Metrics collection
- **Grafana** - Visualization dashboards
- **Alertmanager** - Alert management
- **Jaeger** - Distributed tracing
- **Node Exporter** - System metrics
- **cAdvisor** - Container metrics

---

## 🚀 Quick Start (5 Minutes)

### Step 1: Start Monitoring Stack
```bash
cd /path/to/microservice-project
./scripts/start-monitoring.sh
```

### Step 2: Access Dashboards
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093
- **Jaeger**: http://localhost:16686

### Step 3: Start Microservices (Optional)
```bash
docker-compose up -d
```

---

## 📋 Detailed Configuration Flow

### Phase 1: Infrastructure Setup

#### 1.1 Docker Compose Configuration
**File**: `docker-compose.yml`

**Services Added**:
```yaml
# Monitoring Services
prometheus:
  image: prom/prometheus:latest
  ports: ["9090:9090"]
  volumes:
    - ./monitoring/prometheus:/etc/prometheus
    - prometheus_data:/prometheus

grafana:
  image: grafana/grafana:latest
  ports: ["3000:3000"]
  volumes:
    - ./monitoring/grafana:/etc/grafana/provisioning
    - grafana_data:/var/lib/grafana

node-exporter:
  image: prom/node-exporter:latest
  ports: ["9100:9100"]

cadvisor:
  image: gcr.io/cadvisor/cadvisor:latest
  ports: ["8080:8080"]

alertmanager:
  image: prom/alertmanager:latest
  ports: ["9093:9093"]
  volumes:
    - ./monitoring/alertmanager:/etc/alertmanager

jaeger:
  image: jaegertracing/all-in-one:latest
  ports: ["16686:16686", "14268:14268"]
```

#### 1.2 Directory Structure Creation
```bash
mkdir -p monitoring/{prometheus,grafana,alertmanager}
mkdir -p monitoring/grafana/{provisioning/{datasources,dashboards},dashboards/{system,application,performance,alerts}}
```

### Phase 2: Prometheus Configuration

#### 2.1 Main Configuration
**File**: `monitoring/prometheus/prometheus.yml`

**Key Sections**:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  # System Metrics
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']
  
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
  
  # Microservice Metrics
  - job_name: 'auth-service'
    static_configs:
      - targets: ['auth-service:8081']
    metrics_path: '/metrics'
  
  - job_name: 'file-service'
    static_configs:
      - targets: ['file-service:8082']
    metrics_path: '/metrics'
  
  - job_name: 'user-service'
    static_configs:
      - targets: ['user-service:8083']
    metrics_path: '/metrics'
  
  - job_name: 'analytics-service'
    static_configs:
      - targets: ['analytics-service:8085']
    metrics_path: '/metrics'

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### 2.2 Alert Rules
**File**: `monitoring/prometheus/alert_rules.yml`

**Categories**:
- System Health Alerts
- Service Health Alerts
- Business Metric Alerts
- Container Health Alerts
- Network Health Alerts

### Phase 3: Grafana Configuration

#### 3.1 Data Sources
**File**: `monitoring/grafana/provisioning/datasources/datasources.yml`

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
  
  - name: Jaeger
    type: jaeger
    access: proxy
    url: http://jaeger:16686
```

#### 3.2 Dashboard Provisioning
**File**: `monitoring/grafana/provisioning/dashboards/dashboards.yml`

**Dashboard Categories**:
1. **System Monitoring** (`/system/`)
   - System Overview Dashboard
   - Node Exporter metrics

2. **Application Monitoring** (`/application/`)
   - Microservices Overview
   - Business metrics

3. **Performance Monitoring** (`/performance/`)
   - Container resource usage
   - Go runtime metrics

4. **Alerts & SLA** (`/alerts/`)
   - Alert status dashboard
   - SLA tracking

### Phase 4: Alertmanager Configuration

#### 4.1 Main Configuration
**File**: `monitoring/alertmanager/alertmanager.yml`

**Alert Routing Strategy**:
```yaml
route:
  group_by: ['alertname', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        team: infrastructure
      receiver: 'infrastructure-team'
    - match:
        team: backend
      receiver: 'backend-team'
```

### Phase 5: Microservice Instrumentation

#### 5.1 Prometheus Metrics Integration
**Files Modified**:
- `services/auth/main.go`
- `services/file/main.go`
- `services/user/main.go`
- `services/analytics/main.go`

**Metrics Added**:
```go
// HTTP Metrics
httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "endpoint", "status"},
)

httpRequestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "Duration of HTTP requests in seconds",
        Buckets: prometheus.DefBuckets,
    },
    []string{"method", "endpoint"},
)

// Business Metrics (service-specific)
// Auth: authAttemptsTotal
// File: fileUploadsTotal
// User: userActionsTotal
// Analytics: analyticsEventsTotal, analyticsReportsGenerated
```

#### 5.2 Metrics Endpoint
Each service exposes metrics at: `http://service:port/metrics`

---

## 🔧 Configuration Steps for Production

### Step 1: Security Configuration

#### 1.1 Grafana Authentication
```bash
# Change default admin password
docker exec -it microservice-project-grafana-1 grafana-cli admin reset-admin-password NEW_PASSWORD
```

#### 1.2 Enable HTTPS (Production)
**Update docker-compose.yml**:
```yaml
grafana:
  environment:
    - GF_SERVER_PROTOCOL=https
    - GF_SERVER_CERT_FILE=/etc/ssl/certs/grafana.crt
    - GF_SERVER_CERT_KEY=/etc/ssl/private/grafana.key
  volumes:
    - ./ssl:/etc/ssl
```

### Step 2: Alert Channel Configuration

#### 2.1 Slack Integration
**Update**: `monitoring/alertmanager/alertmanager.yml`

```yaml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'

receivers:
- name: 'slack-alerts'
  slack_configs:
  - channel: '#alerts'
    title: 'Alert: {{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
```

#### 2.2 Email Configuration
```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@yourcompany.com'
  smtp_auth_username: 'your-email@gmail.com'
  smtp_auth_password: 'your-app-password'

receivers:
- name: 'email-alerts'
  email_configs:
  - to: 'oncall@yourcompany.com'
    subject: 'Alert: {{ .GroupLabels.alertname }}'
    body: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

#### 2.3 PagerDuty Integration
```yaml
receivers:
- name: 'pagerduty-critical'
  pagerduty_configs:
  - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
    description: 'Critical Alert: {{ .GroupLabels.alertname }}'
```

### Step 3: Custom Metrics Addition

#### 3.1 Adding New Business Metrics
**Example**: Order processing metrics

```go
// Define metric
var orderProcessingTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "orders_processed_total",
        Help: "Total number of orders processed",
    },
    []string{"status", "payment_method"},
)

// Register metric
prometheus.MustRegister(orderProcessingTotal)

// Use metric
orderProcessingTotal.WithLabelValues("completed", "credit_card").Inc()
```

#### 3.2 Update Prometheus Configuration
**Add to**: `monitoring/prometheus/prometheus.yml`

```yaml
scrape_configs:
  - job_name: 'order-service'
    static_configs:
      - targets: ['order-service:8086']
    metrics_path: '/metrics'
```

### Step 4: Dashboard Customization

#### 4.1 Creating Custom Dashboards
1. **Access Grafana**: http://localhost:3000
2. **Create Dashboard**: Click "+" → Dashboard
3. **Add Panel**: Configure metrics and visualization
4. **Export JSON**: Settings → JSON Model
5. **Save to File**: `monitoring/grafana/dashboards/custom/my-dashboard.json`

#### 4.2 Dashboard Best Practices
- **Use consistent time ranges**
- **Group related metrics**
- **Add meaningful descriptions**
- **Use appropriate visualization types**
- **Set up drill-down capabilities**

### Step 5: Data Retention Configuration

#### 5.1 Prometheus Data Retention
**Update**: `docker-compose.yml`

```yaml
prometheus:
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--storage.tsdb.retention.time=30d'  # 30 days retention
    - '--storage.tsdb.retention.size=10GB'  # 10GB max size
```

#### 5.2 Grafana Data Source Settings
```yaml
datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090
    jsonData:
      timeInterval: "15s"
      queryTimeout: "60s"
```

---

## 🔍 Troubleshooting Guide

### Common Issues and Solutions

#### Issue 1: Grafana Can't Connect to Prometheus
**Symptoms**: "Bad Gateway" or connection errors

**Solutions**:
```bash
# Check Prometheus status
curl http://localhost:9090/api/v1/query?query=up

# Check Docker network
docker network ls
docker network inspect microservice-project_microservice-network

# Restart services
docker-compose restart prometheus grafana
```

#### Issue 2: No Metrics from Microservices
**Symptoms**: Empty dashboards, no service metrics

**Solutions**:
```bash
# Check service metrics endpoints
curl http://localhost:8081/metrics  # Auth service
curl http://localhost:8082/metrics  # File service
curl http://localhost:8083/metrics  # User service
curl http://localhost:8085/metrics  # Analytics service

# Check Prometheus targets
# Visit: http://localhost:9090/targets

# Verify service health
docker-compose ps
docker-compose logs auth-service
```

#### Issue 3: Alerts Not Firing
**Symptoms**: No alerts despite threshold breaches

**Solutions**:
```bash
# Check alert rules
# Visit: http://localhost:9090/alerts

# Verify Alertmanager
curl http://localhost:9093/api/v1/alerts

# Check alert routing
# Visit: http://localhost:9093/#/status

# Test alert rule
# Visit: http://localhost:9090/graph
# Query: up == 0
```

#### Issue 4: High Resource Usage
**Symptoms**: Slow performance, high CPU/memory

**Solutions**:
```bash
# Check resource usage
docker stats

# Reduce scrape frequency
# Edit: monitoring/prometheus/prometheus.yml
# Change: scrape_interval: 30s

# Limit data retention
# Edit: docker-compose.yml
# Add: --storage.tsdb.retention.time=7d
```

---

## 📊 Monitoring Checklist

### Daily Operations
- [ ] Check Grafana dashboards for anomalies
- [ ] Review active alerts in Alertmanager
- [ ] Verify all services are up in Prometheus targets
- [ ] Check system resource usage

### Weekly Maintenance
- [ ] Review alert rules effectiveness
- [ ] Update dashboard configurations
- [ ] Check data retention and cleanup
- [ ] Verify backup procedures

### Monthly Reviews
- [ ] Analyze SLA metrics and trends
- [ ] Review and update alert thresholds
- [ ] Optimize dashboard performance
- [ ] Update monitoring documentation

---

## 🚀 Advanced Configuration

### High Availability Setup

#### 1. Prometheus HA
```yaml
prometheus-1:
  image: prom/prometheus:latest
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--web.external-url=http://prometheus-1:9090'

prometheus-2:
  image: prom/prometheus:latest
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--web.external-url=http://prometheus-2:9090'
```

#### 2. Grafana HA with Load Balancer
```yaml
nginx:
  image: nginx:alpine
  ports:
    - "3000:80"
  volumes:
    - ./nginx-grafana.conf:/etc/nginx/nginx.conf

grafana-1:
  image: grafana/grafana:latest
  environment:
    - GF_DATABASE_TYPE=postgres
    - GF_DATABASE_HOST=postgres:5432

grafana-2:
  image: grafana/grafana:latest
  environment:
    - GF_DATABASE_TYPE=postgres
    - GF_DATABASE_HOST=postgres:5432
```

### External Integrations

#### 1. AWS CloudWatch Integration
```yaml
# Add to prometheus.yml
scrape_configs:
  - job_name: 'cloudwatch'
    ec2_sd_configs:
      - region: us-west-2
        port: 9100
```

#### 2. Kubernetes Integration
```yaml
# For Kubernetes deployment
scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
```

---

## 📚 Additional Resources

### Documentation Links
- [Prometheus Configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
- [Grafana Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)
- [Alertmanager Configuration](https://prometheus.io/docs/alerting/latest/configuration/)
- [Jaeger Deployment](https://www.jaegertracing.io/docs/1.35/deployment/)

### Useful Queries

#### Prometheus Queries
```promql
# Service availability
up{job="auth-service"}

# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# Response time percentiles
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Memory usage
container_memory_usage_bytes{name=~".*auth-service.*"}

# CPU usage
rate(container_cpu_usage_seconds_total{name=~".*auth-service.*"}[5m])
```

### Alert Rule Examples
```yaml
# High error rate
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High error rate detected"

# Service down
- alert: ServiceDown
  expr: up == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Service {{ $labels.job }} is down"
```

---

## 🎯 Next Steps

1. **Immediate Actions**:
   - Configure alert channels (Slack, Email, PagerDuty)
   - Set up authentication for production
   - Create custom dashboards for your specific needs

2. **Short Term (1-2 weeks)**:
   - Implement distributed tracing in microservices
   - Set up log aggregation with ELK stack
   - Configure backup and disaster recovery

3. **Long Term (1-3 months)**:
   - Implement SLI/SLO monitoring
   - Set up capacity planning dashboards
   - Integrate with CI/CD pipeline for deployment monitoring
   - Implement chaos engineering with monitoring

---

**📞 Support**: For issues or questions, refer to the troubleshooting section or consult the official documentation links provided above.