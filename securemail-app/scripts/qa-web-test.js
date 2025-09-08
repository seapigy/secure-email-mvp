// QA Testing Script for Web Deployment
// This script tests the basic functionality of the web frontend

const fs = require('fs');
const path = require('path');

console.log('🧪 SecureMail Web Frontend QA Testing');
console.log('=====================================');

// Test 1: Check if build files exist
console.log('\n📋 Test 1: Build Files Check');
const distPath = path.join(__dirname, '..', 'dist');
const requiredFiles = [
    'index.html',
    'favicon.ico',
    'metadata.json',
    '_expo/static/js/web/AppEntry-281fe7f36af6f927bcb8d412b1a81fe1.js'
];

let buildFilesPass = true;
requiredFiles.forEach(file => {
    const filePath = path.join(distPath, file);
    if (fs.existsSync(filePath)) {
        console.log(`✅ ${file} - Found`);
    } else {
        console.log(`❌ ${file} - Missing`);
        buildFilesPass = false;
    }
});

// Test 2: Check HTML content
console.log('\n📋 Test 2: HTML Content Check');
const indexPath = path.join(distPath, 'index.html');
if (fs.existsSync(indexPath)) {
    const htmlContent = fs.readFileSync(indexPath, 'utf8');
    
    const htmlChecks = [
        { name: 'HTML5 DOCTYPE', test: htmlContent.includes('<!DOCTYPE html>') },
        { name: 'Title tag', test: htmlContent.includes('<title>') },
        { name: 'Meta viewport', test: htmlContent.includes('viewport') },
        { name: 'App entry script', test: htmlContent.includes('AppEntry') },
        { name: 'Favicon', test: htmlContent.includes('favicon.ico') }
    ];
    
    htmlChecks.forEach(check => {
        if (check.test) {
            console.log(`✅ ${check.name} - Present`);
        } else {
            console.log(`❌ ${check.name} - Missing`);
            buildFilesPass = false;
        }
    });
} else {
    console.log('❌ index.html not found');
    buildFilesPass = false;
}

// Test 3: Check JavaScript bundle
console.log('\n📋 Test 3: JavaScript Bundle Check');
const jsPath = path.join(distPath, '_expo/static/js/web/AppEntry-281fe7f36af6f927bcb8d412b1a81fe1.js');
if (fs.existsSync(jsPath)) {
    const stats = fs.statSync(jsPath);
    const sizeKB = Math.round(stats.size / 1024);
    console.log(`✅ JavaScript bundle found - ${sizeKB} KB`);
    
    if (sizeKB > 500) {
        console.log('✅ Bundle size is reasonable');
    } else {
        console.log('⚠️  Bundle size seems small, may indicate build issues');
    }
} else {
    console.log('❌ JavaScript bundle not found');
    buildFilesPass = false;
}

// Test 4: Check assets
console.log('\n📋 Test 4: Assets Check');
const assetsPath = path.join(distPath, 'assets');
if (fs.existsSync(assetsPath)) {
    const assets = fs.readdirSync(assetsPath, { recursive: true });
    console.log(`✅ Assets directory found with ${assets.length} files`);
    
    // Check for font files
    const fontFiles = assets.filter(file => file.toString().endsWith('.ttf'));
    if (fontFiles.length > 0) {
        console.log(`✅ Font files found: ${fontFiles.length}`);
    } else {
        console.log('⚠️  No font files found');
    }
} else {
    console.log('❌ Assets directory not found');
    buildFilesPass = false;
}

// Test 5: Configuration check
console.log('\n📋 Test 5: Configuration Check');
const configPath = path.join(__dirname, '..', 'deployment-config.json');
if (fs.existsSync(configPath)) {
    console.log('✅ Deployment configuration found');
    const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
    
    if (config.environments.development.apiUrl) {
        console.log(`✅ Development API URL configured: ${config.environments.development.apiUrl}`);
    }
    
    if (config.web.baseUrl) {
        console.log(`✅ Web base URL configured: ${config.web.baseUrl}`);
    }
} else {
    console.log('⚠️  Deployment configuration not found');
}

// Summary
console.log('\n📊 QA Test Summary');
console.log('==================');
if (buildFilesPass) {
    console.log('✅ All critical tests passed!');
    console.log('🚀 Web frontend is ready for deployment');
    console.log('');
    console.log('📋 Next Steps:');
    console.log('1. Deploy dist/ contents to web server');
    console.log('2. Configure HTTPS and CORS');
    console.log('3. Test signup/login flows with backend');
    console.log('4. Verify trial warning system');
} else {
    console.log('❌ Some tests failed!');
    console.log('🔧 Please fix the issues before deployment');
}

console.log('\n🌐 Target Deployment URL: https://app.securesystem.email');
console.log('🔗 Backend API: http://localhost:8080 (test mode)');
