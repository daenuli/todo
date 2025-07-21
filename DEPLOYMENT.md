# Todo API Deployment Guide

This guide explains how to deploy the Todo REST API to your VPS using Docker.

## Prerequisites

- VPS with Docker and Docker Compose installed
- SSH access to your VPS
- Domain name (optional, for production)

## Quick Deployment

### 1. Clone the repository to your VPS

```bash
git clone <your-repo-url>
cd todo
```

### 2. Deploy with Docker Compose

```bash
# Build and start the services
docker-compose up -d

# Check if services are running
docker-compose ps

# View logs
docker-compose logs -f
```

### 3. Test the API

```bash
# Test the API endpoint
curl http://your-vps-ip:8080/api/v1/todos

# Create a new todo
curl -X POST http://your-vps-ip:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Todo","description":"Testing deployment"}'
```

## Production Configuration

### Environment Variables

Create a `.env` file for production settings:

```bash
# MongoDB Configuration
MONGO_INITDB_ROOT_USERNAME=your_admin_user
MONGO_INITDB_ROOT_PASSWORD=your_secure_password
MONGODB_URI=mongodb://your_admin_user:your_secure_password@mongodb:27017/todoapp?authSource=admin

# API Configuration
PORT=8080
```

### Update docker-compose.yml for production

```yaml
version: '3.8'

services:
  mongodb:
    image: mongo:7.0
    container_name: todo-mongodb
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: ${MONGO_INITDB_ROOT_USERNAME}
      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_INITDB_ROOT_PASSWORD}
      MONGO_INITDB_DATABASE: todoapp
    volumes:
      - mongodb_data:/data/db
      - ./init-mongo.js:/docker-entrypoint-initdb.d/init-mongo.js:ro
    networks:
      - todo-network
    # Remove port mapping for security (only internal access)
    # ports:
    #   - "27017:27017"

  todo-api:
    build: .
    container_name: todo-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=${MONGODB_URI}
      - PORT=${PORT}
    depends_on:
      - mongodb
    networks:
      - todo-network

volumes:
  mongodb_data:

networks:
  todo-network:
    driver: bridge
```

## Reverse Proxy Setup (Nginx)

For production, use Nginx as a reverse proxy:

### 1. Install Nginx

```bash
sudo apt update
sudo apt install nginx
```

### 2. Create Nginx configuration

```bash
sudo nano /etc/nginx/sites-available/todo-api
```

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### 3. Enable the site

```bash
sudo ln -s /etc/nginx/sites-available/todo-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## SSL Certificate (Let's Encrypt)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

## Useful Commands

```bash
# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build

# View logs
docker-compose logs todo-api
docker-compose logs mongodb

# Access MongoDB shell
docker-compose exec mongodb mongosh -u admin -p password123

# Backup database
docker-compose exec mongodb mongodump --uri="mongodb://admin:password123@localhost:27017/todoapp?authSource=admin" --out=/data/backup

# Update application
git pull
docker-compose up -d --build
```

## Security Considerations

1. **Change default passwords** in production
2. **Use environment variables** for sensitive data
3. **Enable firewall** and only open necessary ports
4. **Regular updates** of Docker images
5. **Monitor logs** for suspicious activity
6. **Backup database** regularly

## Monitoring

Add health checks and monitoring:

```bash
# Check service health
curl http://localhost:8080/api/v1/todos

# Monitor resource usage
docker stats

# Check disk usage
df -h
docker system df
```

## Troubleshooting

### Common Issues

1. **Port already in use**: Change port in docker-compose.yml
2. **MongoDB connection failed**: Check MongoDB logs and credentials
3. **CORS issues**: Verify CORS headers in the API
4. **Out of disk space**: Clean up Docker images and volumes

```bash
# Clean up Docker
docker system prune -a
docker volume prune
```