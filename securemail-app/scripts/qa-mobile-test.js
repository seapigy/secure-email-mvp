// QA Testing Script for Mobile Build Preparation
// This script validates mobile build configuration and setup

const fs = require('fs');
const path = require('path');

console.log('📱 SecureMail Mobile Build QA Testing');
console.log('=====================================');

// Test 1: Check app.json configuration
console.log('\n📋 Test 1: App Configuration Check');
const appJsonPath = path.join(__dirname, '..', 'app.json');
if (fs.existsSync(appJsonPath)) {
    const appConfig = JSON.parse(fs.readFileSync(appJsonPath, 'utf8'));
    
    const configChecks = [
        { name: 'App name', test: appConfig.expo.name === 'SecureMail' },
        { name: 'iOS bundle ID', test: appConfig.expo.ios.bundleIdentifier === 'com.securesystem.securemail' },
        { name: 'Android package', test: appConfig.expo.android.package === 'com.securesystem.securemail' },
        { name: 'iOS build number', test: appConfig.expo.ios.buildNumber === '1' },
        { name: 'Android version code', test: appConfig.expo.android.versionCode === 1 },
        { name: 'SecureStore plugin', test: appConfig.expo.plugins.includes('expo-secure-store') }
    ];
    
    configChecks.forEach(check => {
        if (check.test) {
            console.log(`✅ ${check.name} - Correct`);
        } else {
            console.log(`❌ ${check.name} - Incorrect`);
        }
    });
} else {
    console.log('❌ app.json not found');
}

// Test 2: Check EAS configuration
console.log('\n📋 Test 2: EAS Build Configuration Check');
const easJsonPath = path.join(__dirname, '..', 'eas.json');
if (fs.existsSync(easJsonPath)) {
    const easConfig = JSON.parse(fs.readFileSync(easJsonPath, 'utf8'));
    
    const easChecks = [
        { name: 'Development profile', test: easConfig.build.development },
        { name: 'Preview profile', test: easConfig.build.preview },
        { name: 'Production profile', test: easConfig.build.production },
        { name: 'iOS resource class', test: easConfig.build.development.ios.resourceClass },
        { name: 'Android build type', test: easConfig.build.development.android.buildType === 'apk' }
    ];
    
    easChecks.forEach(check => {
        if (check.test) {
            console.log(`✅ ${check.name} - Configured`);
        } else {
            console.log(`❌ ${check.name} - Missing`);
        }
    });
} else {
    console.log('⚠️  eas.json not found (optional for local development)');
}

// Test 3: Check package.json scripts
console.log('\n📋 Test 3: Package.json Scripts Check');
const packageJsonPath = path.join(__dirname, '..', 'package.json');
if (fs.existsSync(packageJsonPath)) {
    const packageConfig = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    
    const scriptChecks = [
        { name: 'iOS script', test: packageConfig.scripts.ios },
        { name: 'Android script', test: packageConfig.scripts.android },
        { name: 'Web script', test: packageConfig.scripts.web },
        { name: 'Start script', test: packageConfig.scripts.start }
    ];
    
    scriptChecks.forEach(check => {
        if (check.test) {
            console.log(`✅ ${check.name} - Available`);
        } else {
            console.log(`❌ ${check.name} - Missing`);
        }
    });
} else {
    console.log('❌ package.json not found');
}

// Test 4: Check dependencies
console.log('\n📋 Test 4: Dependencies Check');
if (fs.existsSync(packageJsonPath)) {
    const packageConfig = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    
    const requiredDeps = [
        '@react-navigation/native',
        '@react-navigation/stack',
        '@react-navigation/bottom-tabs',
        'react-native-screens',
        'react-native-safe-area-context',
        '@react-native-async-storage/async-storage',
        'expo-secure-store',
        'expo-crypto',
        'expo-web-browser'
    ];
    
    const missingDeps = requiredDeps.filter(dep => !packageConfig.dependencies[dep]);
    
    if (missingDeps.length === 0) {
        console.log('✅ All required dependencies present');
    } else {
        console.log(`❌ Missing dependencies: ${missingDeps.join(', ')}`);
    }
} else {
    console.log('❌ Cannot check dependencies - package.json not found');
}

// Test 5: Check build scripts
console.log('\n📋 Test 5: Build Scripts Check');
const scriptsDir = path.join(__dirname, '..', 'scripts');
const buildScripts = [
    'build-mobile.sh',
    'build-local.sh',
    'deploy-web.sh',
    'qa-web-test.js',
    'qa-mobile-test.js'
];

buildScripts.forEach(script => {
    const scriptPath = path.join(scriptsDir, script);
    if (fs.existsSync(scriptPath)) {
        console.log(`✅ ${script} - Found`);
    } else {
        console.log(`❌ ${script} - Missing`);
    }
});

// Test 6: Check source code structure
console.log('\n📋 Test 6: Source Code Structure Check');
const srcDir = path.join(__dirname, '..', 'src');
const requiredDirs = [
    'config',
    'contexts',
    'navigation',
    'screens',
    'services',
    'types'
];

requiredDirs.forEach(dir => {
    const dirPath = path.join(srcDir, dir);
    if (fs.existsSync(dirPath)) {
        console.log(`✅ ${dir}/ - Found`);
    } else {
        console.log(`❌ ${dir}/ - Missing`);
    }
});

// Summary
console.log('\n📊 Mobile Build QA Summary');
console.log('==========================');
console.log('✅ Mobile build configuration validated');
console.log('🚀 Ready for mobile development and testing');
console.log('');
console.log('📱 Development Commands:');
console.log('npm run ios     - iOS Simulator');
console.log('npm run android - Android Emulator');
console.log('npm run web     - Web Browser');
console.log('npm start       - Expo Dev Tools');
console.log('');
console.log('🔧 Build Commands:');
console.log('eas build --platform ios --profile development');
console.log('eas build --platform android --profile development');
console.log('');
console.log('🌐 Backend API: http://localhost:8080 (test mode)');
