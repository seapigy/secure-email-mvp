#!/bin/bash
# Local Development Build Script for SecureMail Frontend

echo "🔧 Building SecureMail for Local Development..."

# Set environment variables for test backend
export EXPO_PUBLIC_API_URL=http://localhost:8080
export EXPO_PUBLIC_DEBUG=true
export EXPO_PUBLIC_PLACEHOLDER_PAYMENTS=true

# Install dependencies
echo "📦 Installing dependencies..."
npm install

# Start development server for iOS
echo "🍎 Starting iOS development server..."
echo "Run: npm run ios"
echo "This will open iOS Simulator (requires macOS)"

# Start development server for Android
echo "🤖 Starting Android development server..."
echo "Run: npm run android"
echo "This will open Android Emulator (requires Android Studio)"

# Start development server for Web
echo "🌐 Starting Web development server..."
echo "Run: npm run web"
echo "This will open web browser at http://localhost:19006"

echo ""
echo "🎯 Development Commands:"
echo "npm run ios     - iOS Simulator"
echo "npm run android - Android Emulator"
echo "npm run web     - Web Browser"
echo "npm start       - Expo Dev Tools"
echo ""
echo "🔗 Backend API: http://localhost:8080 (test mode)"
echo "📱 Test on multiple platforms simultaneously!"
