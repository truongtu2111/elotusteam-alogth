# Monitoring Stack Documentation

This directory contains a comprehensive monitoring solution for the microservices architecture, providing observability, alerting, and performance monitoring capabilities.

## 🏗️ Architecture Overview

The monitoring stack includes:

- **Prometheus** - Metrics collection and storage
- **Grafana** - Visualization and dashboards
- **Node Exporter** - System-level metrics
- **cAdvisor** - Container metrics
- **Alertmanager** - Alert management and routing
- **Jaeger** - Distributed tracing

## 🚀 Quick Start

### 1. Start the Monitoring Stack

```bash
# From the project root directory
./scripts/start-monitoring.sh
```

### 2. Start Your Microservices

```bash
docker-compose up -d
```

### 3. Access Dashboards

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093
- **Jaeger**: http://localhost:16686

## 📊 Grafana Dashboards

Pre-configured dashboards are available in Grafana:

### System Monitoring
- **System Overview** - CPU, Memory, Network, Disk metrics
- Node Exporter metrics for host system monitoring

### Application Monitoring
- **Microservices Overview** - Service health, request rates, response times
- Business metrics (file uploads, authentication, user actions)
- Error rates and service availability

### Performance Monitoring
- **Performance Profiling** - Container resource usage
- Go runtime metrics (goroutines, GC, memory)
- Application performance indicators

### Alerts & SLA
- **Alerts Dashboard** - Active alerts and alert history
- SLA tracking and uptime monitoring
- Service level indicators

## 🔔 Alerting Configuration

### Alert Rules

Alert rules are defined in `prometheus/alert_rules.yml`:

- **System Health**: High CPU/Memory usage, low disk space
- **Service Health**: Service down, high error rates, slow response times
- **Business Metrics**: Low file upload success, high auth failures
- **Container Health**: High resource usage, frequent restarts

### Alert Routing

Alertmanager configuration in `alertmanager/alertmanager.yml`:

- **Critical alerts** → PagerDuty + Slack
- **Infrastructure alerts** → Infrastructure team
- **Application alerts** → Backend team
- **Security alerts** → Security team

### Configuring Alert Channels

#### Slack Integration

1. Create a Slack webhook URL
2. Update `alertmanager/alertmanager.yml`:

```yaml
global:
  slack_api_url: 'YOUR_SLACK_WEBHOOK_URL'
```

#### Email Notifications

1. Configure SMTP settings in `alertmanager/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@yourcompany.com'
  smtp_auth_username: 'your-email@gmail.com'
  smtp_auth_password: 'your-app-password'
```

#### PagerDuty Integration

1. Get your PagerDuty integration key
2. Update the routing configuration:

```yaml
route:
  routes:
  - match:
      severity: critical
    receiver: 'pagerduty-critical'

receivers:
- name: 'pagerduty-critical'
  pagerduty_configs:
  - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
```

## 📈 Metrics Collection

### Application Metrics

Each microservice exposes metrics at `/metrics` endpoint:

- **HTTP Metrics**: Request count, duration, status codes
- **Business Metrics**: Service-specific counters
- **Go Runtime Metrics**: Goroutines, memory, GC stats

### System Metrics

- **Node Exporter**: CPU, memory, disk, network
- **cAdvisor**: Container resource usage
- **Prometheus**: Self-monitoring metrics

### Custom Metrics

To add custom metrics to your services:

```go
// Define metric
var customCounter = prometheus.NewCounter(
    prometheus.CounterOpts{
        Name: "custom_operations_total",
        Help: "Total number of custom operations",
    },
)

// Register metric
prometheus.MustRegister(customCounter)

// Use metric
customCounter.Inc()
```

## 🔍 Distributed Tracing

Jaeger provides distributed tracing capabilities:

1. **Access Jaeger UI**: http://localhost:16686
2. **View traces** across microservices
3. **Analyze performance** bottlenecks
4. **Debug distributed** requests

### Adding Tracing to Services

To instrument your services with tracing:

```go
import "github.com/opentracing/opentracing-go"

// Start a span
span := opentracing.StartSpan("operation-name")
defer span.Finish()

// Add tags
span.SetTag("user.id", userID)
span.SetTag("http.method", "GET")
```

## 🛠️ Troubleshooting

### Common Issues

#### Grafana Data Source Connection

```bash
# Check Prometheus connectivity
curl http://localhost:9090/api/v1/query?query=up
```

#### Missing Metrics

```bash
# Check service metrics endpoints
curl http://localhost:8081/metrics  # Auth service
curl http://localhost:8082/metrics  # File service
curl http://localhost:8083/metrics  # User service
curl http://localhost:8085/metrics  # Analytics service
```

#### Alert Not Firing

1. Check alert rules in Prometheus: http://localhost:9090/alerts
2. Verify Alertmanager configuration: http://localhost:9093
3. Check alert routing and receivers

### Logs and Debugging

```bash
# View service logs
docker-compose logs prometheus
docker-compose logs grafana
docker-compose logs alertmanager

# Check service status
docker-compose ps
```

## 📋 Maintenance

### Data Retention

- **Prometheus**: 15 days (configurable in `prometheus.yml`)
- **Grafana**: Persistent storage in `grafana/data`
- **Alertmanager**: 120 hours for alert history

### Backup

```bash
# Backup Grafana dashboards and data
tar -czf grafana-backup.tar.gz monitoring/grafana/data

# Backup Prometheus data
tar -czf prometheus-backup.tar.gz monitoring/prometheus/data
```

### Updates

```bash
# Update monitoring stack
docker-compose pull prometheus grafana alertmanager jaeger
docker-compose up -d
```

## 🔐 Security Considerations

1. **Change default passwords** for Grafana
2. **Configure authentication** for production
3. **Use HTTPS** in production environments
4. **Restrict network access** to monitoring ports
5. **Secure alert channels** with proper credentials

## 📚 Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Node Exporter Metrics](https://github.com/prometheus/node_exporter)
- [cAdvisor Metrics](https://github.com/google/cadvisor)

## 🤝 Contributing

To add new dashboards or modify existing ones:

1. Export dashboard JSON from Grafana
2. Save to appropriate directory in `monitoring/grafana/dashboards/`
3. Update dashboard provisioning configuration
4. Test with fresh Grafana instance

## 📞 Support

For monitoring-related issues:

1. Check this documentation
2. Review service logs
3. Verify configuration files
4. Test connectivity between services
5. Consult official documentation for specific tools