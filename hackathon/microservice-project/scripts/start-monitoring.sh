#!/bin/bash

# Start Monitoring Stack Script
# This script starts the complete monitoring infrastructure including:
# - Prometheus for metrics collection
# - Grafana for visualization
# - Node Exporter for system metrics
# - cAdvisor for container metrics
# - Alertmanager for alert management
# - Jaeger for distributed tracing

set -e

echo "🚀 Starting Monitoring Stack..."

# Check if Docker and Docker Compose are installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create necessary directories if they don't exist
echo "📁 Creating monitoring directories..."
mkdir -p monitoring/prometheus/data
mkdir -p monitoring/grafana/data
mkdir -p monitoring/alertmanager/data

# Set proper permissions for Grafana
echo "🔐 Setting permissions..."
sudo chown -R 472:472 monitoring/grafana/data 2>/dev/null || echo "Warning: Could not set Grafana permissions"

# Start the monitoring stack
echo "🐳 Starting monitoring services..."
docker-compose up -d prometheus grafana node-exporter cadvisor alertmanager jaeger

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 10

# Check service health
echo "🔍 Checking service health..."

services=(
    "prometheus:9090"
    "grafana:3000"
    "node-exporter:9100"
    "cadvisor:8080"
    "alertmanager:9093"
    "jaeger:16686"
)

for service in "${services[@]}"; do
    name=$(echo $service | cut -d':' -f1)
    port=$(echo $service | cut -d':' -f2)
    
    if curl -s http://localhost:$port > /dev/null 2>&1; then
        echo "✅ $name is running on port $port"
    else
        echo "❌ $name is not responding on port $port"
    fi
done

echo ""
echo "🎉 Monitoring Stack Started Successfully!"
echo ""
echo "📊 Access URLs:"
echo "  • Grafana Dashboard: http://localhost:3000 (admin/admin)"
echo "  • Prometheus: http://localhost:9090"
echo "  • Alertmanager: http://localhost:9093"
echo "  • Jaeger Tracing: http://localhost:16686"
echo "  • Node Exporter: http://localhost:9100"
echo "  • cAdvisor: http://localhost:8080"
echo ""
echo "📈 Microservice Metrics Endpoints:"
echo "  • Auth Service: http://localhost:8081/metrics"
echo "  • File Service: http://localhost:8082/metrics"
echo "  • User Service: http://localhost:8083/metrics"
echo "  • Analytics Service: http://localhost:8085/metrics"
echo "  • API Gateway: http://localhost:8080/metrics"
echo ""
echo "🔔 Alert Configuration:"
echo "  • Edit monitoring/alertmanager/alertmanager.yml for alert routing"
echo "  • Edit monitoring/prometheus/alert_rules.yml for alert rules"
echo ""
echo "📋 Next Steps:"
echo "  1. Start your microservices: docker-compose up -d"
echo "  2. Generate some traffic to see metrics"
echo "  3. Check Grafana dashboards for system and application metrics"
echo "  4. Configure alert channels (Slack, Email, PagerDuty) in Alertmanager"
echo ""
echo "🛑 To stop monitoring: docker-compose down"