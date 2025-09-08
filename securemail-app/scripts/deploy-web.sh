#!/bin/bash
# Web Deployment Script for SecureMail Frontend

echo "🚀 Deploying SecureMail Web Frontend..."

# Set environment variables for test backend
export EXPO_PUBLIC_API_URL=http://localhost:8080
export EXPO_PUBLIC_DEBUG=true
export EXPO_PUBLIC_PLACEHOLDER_PAYMENTS=true

# Build the web version
echo "📦 Building web version..."
npm run build:web

# Check if build was successful
if [ ! -d "dist" ]; then
    echo "❌ Build failed - dist directory not found"
    exit 1
fi

echo "✅ Web build successful!"

# Create deployment package
echo "📋 Creating deployment package..."
tar -czf securemail-web-deployment.tar.gz -C dist .

echo "🎉 Deployment package created: securemail-web-deployment.tar.gz"
echo ""
echo "📋 Deployment Instructions:"
echo "1. Upload securemail-web-deployment.tar.gz to your web server"
echo "2. Extract to web root: tar -xzf securemail-web-deployment.tar.gz"
echo "3. Configure web server for SPA routing (redirect all routes to index.html)"
echo "4. Set up SSL certificate for HTTPS"
echo "5. Configure CORS headers for API calls to backend"
echo ""
echo "🔧 Web Server Configuration:"
echo "- Serve static files from extracted directory"
echo "- Enable HTTPS (required for secure token storage)"
echo "- Set CORS headers: Access-Control-Allow-Origin: https://app.securesystem.email"
echo "- Configure SPA routing: redirect all routes to index.html"
echo ""
echo "🌐 Target URL: https://app.securesystem.email"
echo "🔗 Backend API: http://localhost:8080 (test mode)"