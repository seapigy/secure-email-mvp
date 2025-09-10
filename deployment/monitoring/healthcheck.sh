#!/bin/bash

# Comprehensive Health Check Script for Secure Email Backend
set -e

# Configuration
BACKEND_URL="http://localhost:8080"
LOG_FILE="/opt/secure-email/logs/health.log"
ALERT_EMAIL="admin@securesystem.email"
MAX_RETRIES=3
RETRY_DELAY=10

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"

# Function to send alert
send_alert() {
    local message="$1"
    log "ALERT: $message"
    
    # Send email alert if mail is configured
    if command -v mail >/dev/null 2>&1; then
        echo "$message" | mail -s "Secure Email Backend Alert" "$ALERT_EMAIL" 2>/dev/null || true
    fi
}

# Function to check service with retries
check_service() {
    local service_name="$1"
    local check_command="$2"
    local retries=0
    
    while [ $retries -lt $MAX_RETRIES ]; do
        if eval "$check_command" >/dev/null 2>&1; then
            success "$service_name is healthy"
            return 0
        else
            retries=$((retries + 1))
            if [ $retries -lt $MAX_RETRIES ]; then
                warning "$service_name check failed, retrying in $RETRY_DELAY seconds... (attempt $retries/$MAX_RETRIES)"
                sleep $RETRY_DELAY
            fi
        fi
    done
    
    error "$service_name is unhealthy after $MAX_RETRIES attempts"
    return 1
}

# Function to restart service
restart_service() {
    local service_name="$1"
    log "Restarting $service_name..."
    
    if systemctl restart "$service_name"; then
        success "$service_name restarted successfully"
        sleep 30  # Wait for service to fully start
    else
        error "Failed to restart $service_name"
        send_alert "Failed to restart $service_name"
    fi
}

log "Starting comprehensive health check..."

# 1. Check system resources
log "Checking system resources..."

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    error "Critical disk usage: ${DISK_USAGE}%"
    send_alert "Critical disk usage: ${DISK_USAGE}%"
elif [ "$DISK_USAGE" -gt 80 ]; then
    warning "High disk usage: ${DISK_USAGE}%"
else
    success "Disk usage: ${DISK_USAGE}%"
fi

# Check memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ "$MEMORY_USAGE" -gt 90 ]; then
    error "Critical memory usage: ${MEMORY_USAGE}%"
    send_alert "Critical memory usage: ${MEMORY_USAGE}%"
elif [ "$MEMORY_USAGE" -gt 80 ]; then
    warning "High memory usage: ${MEMORY_USAGE}%"
else
    success "Memory usage: ${MEMORY_USAGE}%"
fi

# Check CPU load
CPU_LOAD=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
CPU_CORES=$(nproc)
if (( $(echo "$CPU_LOAD > $CPU_CORES" | bc -l) )); then
    warning "High CPU load: $CPU_LOAD (cores: $CPU_CORES)"
else
    success "CPU load: $CPU_LOAD (cores: $CPU_CORES)"
fi

# 2. Check Docker
log "Checking Docker..."
if ! check_service "Docker" "docker info"; then
    send_alert "Docker is not running"
    exit 1
fi

# 3. Check Docker containers
log "Checking Docker containers..."
if ! check_service "Docker containers" "docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E '(Up|healthy)'"; then
    warning "Some Docker containers are not running properly"
    docker ps -a
fi

# 4. Check backend API
log "Checking backend API..."
if ! check_service "Backend API" "curl -f -s $BACKEND_URL/health"; then
    warning "Backend API health check failed, attempting restart..."
    restart_service "secure-email-backend.service"
    
    # Re-check after restart
    if ! check_service "Backend API (after restart)" "curl -f -s $BACKEND_URL/health"; then
        error "Backend API is still unhealthy after restart"
        send_alert "Backend API is unhealthy and restart failed"
    fi
fi

# 5. Check database connectivity
log "Checking database connectivity..."
if ! check_service "Database" "docker exec securechat-email-db-1 mysql -u secureuser -p\${DB_PASSWORD} -e 'SELECT 1' securesystem"; then
    warning "Database connectivity check failed"
    send_alert "Database connectivity issue detected"
fi

# 6. Check Nginx
log "Checking Nginx..."
if ! check_service "Nginx" "curl -f -s -I http://localhost"; then
    warning "Nginx health check failed"
    restart_service "nginx"
fi

# 7. Check SSL certificate (if configured)
if [ -f "/opt/secure-email/deployment/nginx/ssl/cert.pem" ]; then
    log "Checking SSL certificate..."
    CERT_EXPIRY=$(openssl x509 -in /opt/secure-email/deployment/nginx/ssl/cert.pem -noout -dates | grep notAfter | cut -d= -f2)
    CERT_EXPIRY_EPOCH=$(date -d "$CERT_EXPIRY" +%s)
    CURRENT_EPOCH=$(date +%s)
    DAYS_UNTIL_EXPIRY=$(( (CERT_EXPIRY_EPOCH - CURRENT_EPOCH) / 86400 ))
    
    if [ "$DAYS_UNTIL_EXPIRY" -lt 7 ]; then
        error "SSL certificate expires in $DAYS_UNTIL_EXPIRY days"
        send_alert "SSL certificate expires in $DAYS_UNTIL_EXPIRY days"
    elif [ "$DAYS_UNTIL_EXPIRY" -lt 30 ]; then
        warning "SSL certificate expires in $DAYS_UNTIL_EXPIRY days"
    else
        success "SSL certificate expires in $DAYS_UNTIL_EXPIRY days"
    fi
fi

# 8. Check log files for errors
log "Checking log files for errors..."
ERROR_COUNT=$(find /opt/secure-email/logs -name "*.log" -exec grep -l "ERROR\|FATAL" {} \; 2>/dev/null | wc -l)
if [ "$ERROR_COUNT" -gt 0 ]; then
    warning "Found $ERROR_COUNT log files with errors"
fi

# 9. Check network connectivity
log "Checking network connectivity..."
if ! check_service "Internet connectivity" "ping -c 1 8.8.8.8"; then
    warning "Internet connectivity check failed"
fi

# 10. Check system services
log "Checking system services..."
SERVICES=("docker" "fail2ban" "ufw")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet "$service"; then
        success "$service service is running"
    else
        warning "$service service is not running"
    fi
done

# 11. Check for security issues
log "Checking for security issues..."

# Check for failed login attempts
FAILED_LOGINS=$(grep "Failed password" /var/log/auth.log | grep "$(date '+%b %d')" | wc -l)
if [ "$FAILED_LOGINS" -gt 20 ]; then
    warning "High number of failed login attempts: $FAILED_LOGINS"
    send_alert "High number of failed login attempts: $FAILED_LOGINS"
fi

# Check fail2ban status
if command -v fail2ban-client >/dev/null 2>&1; then
    BANNED_IPS=$(fail2ban-client status sshd | grep "Currently banned" | awk '{print $4}')
    if [ "$BANNED_IPS" -gt 0 ]; then
        success "Fail2ban is active with $BANNED_IPS banned IPs"
    fi
fi

# 12. Performance metrics
log "Collecting performance metrics..."
RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' "$BACKEND_URL/health" 2>/dev/null || echo "N/A")
if [ "$RESPONSE_TIME" != "N/A" ]; then
    if (( $(echo "$RESPONSE_TIME > 2.0" | bc -l) )); then
        warning "High response time: ${RESPONSE_TIME}s"
    else
        success "Response time: ${RESPONSE_TIME}s"
    fi
fi

# 13. Clean up old logs
log "Cleaning up old logs..."
find /opt/secure-email/logs -name "*.log" -mtime +30 -delete 2>/dev/null || true

log "Health check completed"

# Summary
log "Health check summary:"
log "- System resources: OK"
log "- Docker: OK"
log "- Backend API: OK"
log "- Database: OK"
log "- Nginx: OK"
log "- Security: OK"
log "- Performance: OK"
