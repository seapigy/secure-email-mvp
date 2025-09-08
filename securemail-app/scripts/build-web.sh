#!/bin/bash
# Build script for web deployment

echo "Building SecureMail for web deployment..."

# Install dependencies
npm install

# Build for web
npm run build:web

# Create deployment directory
mkdir -p dist/web

# Copy built files
cp -r dist/* dist/web/

echo "Web build complete! Files are in dist/web/"
echo "Deploy the contents of dist/web/ to your web server."
