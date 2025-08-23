# =============================================================================
# Generate User Flow HTML Script
# =============================================================================
# Converts the complete user flow guide to HTML format for PDF conversion
# =============================================================================

Write-Host "🚀 Generating User Flow HTML..." -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Gray

# Define paths
$markdownFile = "docs/complete-user-flow-guide.md"
$outputDir = "docs/pdf"
$htmlFile = "$outputDir/complete-user-flow-guide.html"

# Create output directory if it doesn't exist
if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    Write-Host "✅ Created output directory: $outputDir" -ForegroundColor Green
}

# Check if markdown file exists
if (-not (Test-Path $markdownFile)) {
    Write-Host "❌ Markdown file not found: $markdownFile" -ForegroundColor Red
    exit 1
}

Write-Host "📄 Processing markdown file: $markdownFile" -ForegroundColor Cyan

# Create enhanced CSS for better PDF conversion
$cssContent = @"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Complete User Flow Guide - Secure Email System</title>
    <style>
        /* Reset and base styles */
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
            background-color: #ffffff;
        }
        
        /* Typography */
        h1, h2, h3, h4, h5, h6 {
            color: #2c3e50;
            margin-top: 1.5em;
            margin-bottom: 0.5em;
            page-break-after: avoid;
        }
        
        h1 {
            font-size: 2.5em;
            border-bottom: 3px solid #3498db;
            padding-bottom: 15px;
            margin-bottom: 30px;
            text-align: center;
            color: #1a252f;
        }
        
        h2 {
            font-size: 2em;
            border-bottom: 2px solid #ecf0f1;
            padding-bottom: 10px;
            margin-top: 40px;
            color: #2c3e50;
        }
        
        h3 {
            font-size: 1.5em;
            color: #34495e;
            margin-top: 30px;
        }
        
        h4 {
            font-size: 1.3em;
            color: #7f8c8d;
            margin-top: 25px;
        }
        
        /* Code styling */
        code {
            background-color: #f8f9fa;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Courier New', 'Consolas', monospace;
            font-size: 0.9em;
            color: #e74c3c;
        }
        
        pre {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            overflow-x: auto;
            border-left: 4px solid #3498db;
            margin: 20px 0;
            font-family: 'Courier New', 'Consolas', monospace;
            font-size: 0.9em;
            line-height: 1.4;
        }
        
        pre code {
            background-color: transparent;
            padding: 0;
            color: #333;
        }
        
        /* Blockquotes */
        blockquote {
            border-left: 4px solid #3498db;
            margin: 20px 0;
            padding: 15px 20px;
            background-color: #f8f9fa;
            color: #7f8c8d;
            font-style: italic;
            border-radius: 0 5px 5px 0;
        }
        
        /* Tables */
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 20px 0;
            font-size: 0.9em;
        }
        
        th, td {
            border: 1px solid #ddd;
            padding: 12px 15px;
            text-align: left;
            vertical-align: top;
        }
        
        th {
            background-color: #3498db;
            color: white;
            font-weight: bold;
            text-transform: uppercase;
            font-size: 0.85em;
            letter-spacing: 0.5px;
        }
        
        tr:nth-child(even) {
            background-color: #f2f2f2;
        }
        
        tr:hover {
            background-color: #e8f4f8;
        }
        
        /* Lists */
        ul, ol {
            padding-left: 25px;
            margin: 15px 0;
        }
        
        li {
            margin-bottom: 8px;
            line-height: 1.5;
        }
        
        /* Links */
        a {
            color: #3498db;
            text-decoration: none;
            border-bottom: 1px solid transparent;
            transition: border-bottom 0.3s ease;
        }
        
        a:hover {
            border-bottom: 1px solid #3498db;
        }
        
        /* Special styling for security features */
        .security-feature {
            background-color: #e8f5e8;
            border: 1px solid #4caf50;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .security-feature h4 {
            color: #2e7d32;
            margin-top: 0;
            margin-bottom: 15px;
        }
        
        /* Workflow steps */
        .workflow-step {
            background-color: #f0f8ff;
            border-left: 4px solid #2196f3;
            padding: 15px 20px;
            margin: 15px 0;
            border-radius: 0 5px 5px 0;
        }
        
        /* Code blocks with language highlighting */
        .language-go, .language-typescript, .language-javascript {
            background-color: #f4f4f4;
            border: 1px solid #ddd;
            border-radius: 5px;
            padding: 20px;
            overflow-x: auto;
            margin: 20px 0;
        }
        
        /* Emojis and icons */
        .emoji {
            font-size: 1.2em;
            margin-right: 8px;
        }
        
        /* Table of contents */
        .toc {
            background-color: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 20px;
            margin: 30px 0;
        }
        
        .toc h2 {
            border-bottom: none;
            margin-top: 0;
            margin-bottom: 15px;
        }
        
        .toc ul {
            list-style-type: none;
            padding-left: 0;
        }
        
        .toc li {
            margin-bottom: 5px;
        }
        
        .toc a {
            color: #495057;
            text-decoration: none;
            padding: 2px 0;
            display: block;
        }
        
        .toc a:hover {
            color: #3498db;
        }
        
        /* Print styles for PDF conversion */
        @media print {
            body {
                font-size: 12pt;
                line-height: 1.4;
                max-width: none;
                margin: 0;
                padding: 20px;
            }
            
            h1 { 
                font-size: 18pt; 
                page-break-before: always;
                page-break-after: avoid;
            }
            
            h2 { 
                font-size: 16pt; 
                page-break-after: avoid;
            }
            
            h3 { 
                font-size: 14pt; 
                page-break-after: avoid;
            }
            
            h4 { 
                font-size: 12pt; 
                page-break-after: avoid;
            }
            
            .no-print {
                display: none;
            }
            
            pre {
                white-space: pre-wrap;
                word-wrap: break-word;
                font-size: 10pt;
                page-break-inside: avoid;
            }
            
            table {
                page-break-inside: avoid;
                font-size: 10pt;
            }
            
            .security-feature, .workflow-step {
                page-break-inside: avoid;
            }
            
            /* Ensure proper page breaks */
            h1, h2, h3 {
                page-break-after: avoid;
            }
            
            /* Add page numbers */
            @page {
                margin: 1in;
                @bottom-center {
                    content: counter(page);
                }
            }
        }
        
        /* Header and footer for PDF */
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 2px solid #ecf0f1;
        }
        
        .header h1 {
            border-bottom: none;
            margin-bottom: 10px;
        }
        
        .header .subtitle {
            color: #7f8c8d;
            font-size: 1.1em;
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ecf0f1;
            text-align: center;
            color: #7f8c8d;
            font-size: 0.9em;
        }
        
        /* Responsive design */
        @media (max-width: 768px) {
            body {
                padding: 20px 15px;
            }
            
            h1 { font-size: 2em; }
            h2 { font-size: 1.7em; }
            h3 { font-size: 1.4em; }
            
            pre {
                padding: 15px;
                font-size: 0.8em;
            }
            
            table {
                font-size: 0.8em;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Complete User Flow Guide</h1>
        <div class="subtitle">Secure Email System</div>
        <div class="subtitle">Comprehensive User Experience Documentation</div>
    </div>
    
    <div class="toc">
        <h2>📋 Table of Contents</h2>
        <ul>
            <li><a href="#system-overview">System Overview</a></li>
            <li><a href="#user-types--access-levels">User Types & Access Levels</a></li>
            <li><a href="#authentication--onboarding">Authentication & Onboarding</a></li>
            <li><a href="#internal-user-workflows">Internal User Workflows</a></li>
            <li><a href="#external-user-workflows">External User Workflows</a></li>
            <li><a href="#administrative-workflows">Administrative Workflows</a></li>
            <li><a href="#security-features-reference">Security Features Reference</a></li>
            <li><a href="#troubleshooting--support">Troubleshooting & Support</a></li>
        </ul>
    </div>
"@

# Read the markdown content
$markdownContent = Get-Content $markdownFile -Raw -Encoding UTF8

# Convert markdown to HTML (basic conversion)
$htmlContent = $markdownContent -replace '^# (.+)$', '<h1 id="$1">$1</h1>' `
    -replace '^## (.+)$', '<h2 id="$1">$1</h2>' `
    -replace '^### (.+)$', '<h3 id="$1">$1</h3>' `
    -replace '^#### (.+)$', '<h4 id="$1">$1</h4>' `
    -replace '^##### (.+)$', '<h5 id="$1">$1</h5>' `
    -replace '^###### (.+)$', '<h6 id="$1">$1</h6>' `
    -replace '\*\*(.+?)\*\*', '<strong>$1</strong>' `
    -replace '\*(.+?)\*', '<em>$1</em>' `
    -replace '`(.+?)`', '<code>$1</code>' `
    -replace '^\s*-\s+(.+)$', '<li>$1</li>' `
    -replace '^\s*\d+\.\s+(.+)$', '<li>$1</li>' `
    -replace '^\s*$', '<br>' `
    -replace '^---$', '<hr>'

# Wrap lists properly
$htmlContent = $htmlContent -replace '(<li>.*?</li>)+', '<ul>$&</ul>'

# Add footer
$footer = @"
    <div class="footer">
        <p><strong>Secure Email System - Complete User Flow Guide</strong></p>
        <p>Version 1.0 | Last Updated: August 2025 | System Version: Secure Email MVP v1.0</p>
        <p>This document provides a comprehensive guide to the Secure Email system user flows.</p>
        <p>For technical implementation details, please refer to the developer documentation and API specifications.</p>
    </div>
</body>
</html>
"@

# Combine all content
$fullHtml = $cssContent + $htmlContent + $footer

# Save the HTML file
$fullHtml | Out-File -FilePath $htmlFile -Encoding UTF8

Write-Host "✅ HTML file generated successfully: $htmlFile" -ForegroundColor Green

# Create a simple conversion guide
$conversionGuide = @"
# PDF Conversion Guide

## Option 1: Browser Print to PDF (Recommended)
1. Open the HTML file in your web browser
2. Press Ctrl+P (or Cmd+P on Mac)
3. Select "Save as PDF" as the destination
4. Choose your preferred settings:
   - Page size: A4
   - Margins: Default or Minimum
   - Include background graphics: Yes
5. Click "Save" and choose your file location

## Option 2: Online Converters
1. Visit an online HTML to PDF converter
2. Upload the HTML file: $htmlFile
3. Convert and download the PDF

## Option 3: PDF Software
1. Open the HTML file in a PDF creation tool
2. Export as PDF with your preferred settings

## File Locations
- HTML File: $htmlFile
- Output Directory: $outputDir
"@

$conversionGuide | Out-File -FilePath "$outputDir/pdf-conversion-guide.txt" -Encoding UTF8

Write-Host "📄 Conversion guide created: $outputDir/pdf-conversion-guide.txt" -ForegroundColor Cyan

# Open the HTML file in default browser
Write-Host "🔄 Opening HTML file in browser..." -ForegroundColor Yellow
Start-Process $htmlFile

Write-Host "==================================================================" -ForegroundColor Gray
Write-Host "🎉 HTML Generation Complete!" -ForegroundColor Green
Write-Host "📄 HTML File: $htmlFile" -ForegroundColor Cyan
Write-Host "📁 Output Directory: $outputDir" -ForegroundColor Cyan
Write-Host "To convert to PDF: Open the HTML file and use browser Print to PDF" -ForegroundColor Yellow
Write-Host "==================================================================" -ForegroundColor Gray
