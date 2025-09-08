# SecureMail Multi-Platform App

A complete, production-ready signup/login frontend for SecureMail that works on web, iOS, and Android. This frontend integrates with the existing Phase 3 backend APIs and supports Free, Premium, and Enterprise accounts with placeholder payments.

## Features

### Multi-Platform Support
- **Web**: Deployable on subdomain (app.securesystem.email)
- **iOS**: Buildable via Expo/React Native
- **Android**: Buildable via Expo/React Native
- **Cross-platform**: Same codebase for all platforms

### Authentication Flow
- **Signup**: Account type selection (Free/Premium/Enterprise)
- **Login**: Secure authentication with session management
- **Email Verification**: Integration with backend verification system
- **MFA Support**: Hooks for TOTP integration (backend ready)

### Account Management
- **Trial Management**: 1-month placeholder trials for Premium/Enterprise
- **Organization Support**: Enterprise user organization management
- **Role-Based Access**: Different features based on account type

### Security & Privacy
- **Secure Storage**: Encrypted token storage on mobile, secure cookies on web
- **Blind Server**: No sensitive data stored unencrypted on client
- **AES-256-GCM**: Ready for encrypted payloads from backend

## Project Structure

```
securemail-app/
├── src/
│   ├── config/          # API configuration
│   ├── contexts/        # React contexts (Auth)
│   ├── navigation/      # Navigation components
│   ├── screens/         # App screens
│   ├── services/        # API and storage services
│   ├── types/           # TypeScript types
│   └── __tests__/       # Test files
├── scripts/             # Build and deployment scripts
├── App.tsx              # Main app component
└── package.json         # Dependencies and scripts
```

## Setup Instructions

### Prerequisites
- Node.js 18+ 
- npm or yarn
- Expo CLI (`npm install -g @expo/cli`)
- iOS Simulator (for iOS development)
- Android Studio (for Android development)

### Installation

1. **Clone and install dependencies:**
   ```bash
   cd securemail-app
   npm install
   ```

2. **Configure API endpoints:**
   Edit `src/config/api.ts` to set your backend URL:
   ```typescript
   BASE_URL: __DEV__ 
     ? 'http://localhost:8080'  // Development backend
     : 'https://api.securesystem.email',  // Production backend
   ```

3. **Start development server:**
   ```bash
   npm start
   ```

### Development

#### Web Development
```bash
npm run web
```
Opens the app in your web browser at `http://localhost:19006`

#### iOS Development
```bash
npm run ios
```
Opens the app in iOS Simulator (requires macOS)

#### Android Development
```bash
npm run android
```
Opens the app in Android Emulator

### Building for Production

#### Web Build
```bash
npm run build:web
```
Creates static files in `dist/` directory ready for web deployment.

#### Mobile Builds
```bash
# For iOS (requires EAS CLI)
npm install -g eas-cli
eas build --platform ios

# For Android
eas build --platform android
```

## Backend Integration

This frontend integrates with the existing Phase 3 backend APIs:

### Authentication Endpoints
- `POST /api/auth/signup` - User registration
- `POST /api/auth/login` - User authentication
- `POST /api/auth/verify-email` - Email verification
- `POST /api/auth/resend-verification` - Resend verification

### Inbox Endpoints
- `GET /api/inbox/folders` - Get user folders
- `GET /api/inbox/messages` - Get messages in folder

### Trial Management
- `GET /api/trial/warning` - Check trial expiration
- `POST /api/trial/extend` - Extend trial (testing)

### Account Management
- `POST /api/auth/upgrade-account` - Upgrade account
- `POST /api/auth/downgrade-account` - Downgrade account

### Organization Management
- `POST /api/org/create` - Create organization
- `POST /api/org/add-user` - Add user to organization
- `GET /api/org/list-users` - List organization users

## Deployment

### Web Deployment (app.securesystem.email)

1. **Build the app:**
   ```bash
   npm run build:web
   ```

2. **Deploy to web server:**
   ```bash
   # Upload dist/ contents to your web server
   # Configure SSL certificate
   # Set up CORS for API calls
   ```

3. **Configure web server:**
   - Serve static files from `dist/`
   - Configure HTTPS
   - Set up CORS headers for API calls
   - Configure routing for SPA (redirect all routes to index.html)

### Mobile Deployment

1. **iOS App Store:**
   ```bash
   eas build --platform ios --profile production
   eas submit --platform ios
   ```

2. **Google Play Store:**
   ```bash
   eas build --platform android --profile production
   eas submit --platform android
   ```

## Testing

Run the test suite:
```bash
npm test
```

## Configuration

### Environment Variables
Create a `.env` file for environment-specific configuration:
```env
EXPO_PUBLIC_API_URL=https://api.securesystem.email
EXPO_PUBLIC_APP_NAME=SecureMail
```

### API Configuration
Modify `src/config/api.ts` to customize:
- API endpoints
- Request timeouts
- Default headers
- Account types

## Security Considerations

1. **Token Storage**: Uses SecureStore on mobile, localStorage on web
2. **API Security**: All API calls include proper authentication headers
3. **Data Encryption**: Ready for AES-256-GCM encrypted payloads
4. **Privacy**: No sensitive data logged or stored unencrypted

## Troubleshooting

### Common Issues

1. **Metro bundler issues:**
   ```bash
   npx expo start --clear
   ```

2. **iOS build issues:**
   ```bash
   cd ios && pod install && cd ..
   ```

3. **Android build issues:**
   ```bash
   npx expo run:android --clear
   ```

### Debug Mode
Enable debug mode by setting `__DEV__ = true` in your environment.

## Contributing

1. Follow the existing code structure
2. Add tests for new features
3. Ensure cross-platform compatibility
4. Update documentation for API changes

## License

This project is part of the SecureMail system and follows the same licensing terms.
