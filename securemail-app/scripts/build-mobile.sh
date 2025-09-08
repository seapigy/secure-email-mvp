#!/bin/bash
# Mobile Build Script for SecureMail Frontend

echo "📱 Building SecureMail Mobile Apps..."

# Set environment variables for test backend
export EXPO_PUBLIC_API_URL=http://localhost:8080
export EXPO_PUBLIC_DEBUG=true
export EXPO_PUBLIC_PLACEHOLDER_PAYMENTS=true

# Check if EAS CLI is installed
if ! command -v eas &> /dev/null; then
    echo "❌ EAS CLI not found. Installing..."
    npm install -g eas-cli
fi

# Login to EAS (if not already logged in)
echo "🔐 Checking EAS authentication..."
if ! eas whoami &> /dev/null; then
    echo "⚠️  Please login to EAS:"
    eas login
fi

# Build for iOS (development)
echo "🍎 Building iOS development version..."
eas build --platform ios --profile development --non-interactive

# Check if iOS build was successful
if [ $? -eq 0 ]; then
    echo "✅ iOS development build successful!"
else
    echo "❌ iOS build failed!"
    exit 1
fi

# Build for Android (development)
echo "🤖 Building Android development version..."
eas build --platform android --profile development --non-interactive

# Check if Android build was successful
if [ $? -eq 0 ]; then
    echo "✅ Android development build successful!"
else
    echo "❌ Android build failed!"
    exit 1
fi

echo "🎉 Mobile builds completed successfully!"
echo ""
echo "📋 Build Summary:"
echo "- iOS development build: Ready for testing"
echo "- Android development build: Ready for testing"
echo ""
echo "📱 Next Steps:"
echo "1. Download builds from EAS dashboard"
echo "2. Install on iOS Simulator/Android Emulator"
echo "3. Test signup/login flows"
echo "4. Verify trial warning system"
echo "5. Test organization features (Enterprise)"
echo ""
echo "🔗 EAS Dashboard: https://expo.dev/accounts/[your-account]/projects/securemail-app"
