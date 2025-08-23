# Simple HTML Generator for User Flow Guide
Write-Host "Generating User Flow HTML..." -ForegroundColor Green

# Define paths
$markdownFile = "docs/complete-user-flow-guide.md"
$outputDir = "docs/pdf"
$htmlFile = "$outputDir/complete-user-flow-guide.html"

# Create output directory if it doesn't exist
if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    Write-Host "Created output directory: $outputDir" -ForegroundColor Green
}

# Check if markdown file exists
if (-not (Test-Path $markdownFile)) {
    Write-Host "Markdown file not found: $markdownFile" -ForegroundColor Red
    exit 1
}

Write-Host "Processing markdown file: $markdownFile" -ForegroundColor Cyan

# Create HTML content
$htmlContent = @"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Complete User Flow Guide - Secure Email System</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
            background-color: #ffffff;
        }
        
        h1, h2, h3, h4, h5, h6 {
            color: #2c3e50;
            margin-top: 1.5em;
            margin-bottom: 0.5em;
        }
        
        h1 {
            font-size: 2.5em;
            border-bottom: 3px solid #3498db;
            padding-bottom: 15px;
            margin-bottom: 30px;
            text-align: center;
        }
        
        h2 {
            font-size: 2em;
            border-bottom: 2px solid #ecf0f1;
            padding-bottom: 10px;
            margin-top: 40px;
        }
        
        h3 {
            font-size: 1.5em;
            color: #34495e;
            margin-top: 30px;
        }
        
        code {
            background-color: #f8f9fa;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #e74c3c;
        }
        
        pre {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #3498db;
            margin: 20px 0;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            line-height: 1.4;
            overflow-x: auto;
        }
        
        blockquote {
            border-left: 4px solid #3498db;
            margin: 20px 0;
            padding: 15px 20px;
            background-color: #f8f9fa;
            color: #7f8c8d;
            font-style: italic;
            border-radius: 0 5px 5px 0;
        }
        
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
        
        ul, ol {
            padding-left: 25px;
            margin: 15px 0;
        }
        
        li {
            margin-bottom: 8px;
            line-height: 1.5;
        }
        
        a {
            color: #3498db;
            text-decoration: none;
        }
        
        a:hover {
            text-decoration: underline;
        }
        
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 2px solid #ecf0f1;
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ecf0f1;
            text-align: center;
            color: #7f8c8d;
            font-size: 0.9em;
        }
        
        @media print {
            body {
                font-size: 12pt;
                line-height: 1.4;
                max-width: none;
                margin: 0;
                padding: 20px;
            }
            
            h1 { font-size: 18pt; }
            h2 { font-size: 16pt; }
            h3 { font-size: 14pt; }
            h4 { font-size: 12pt; }
            
            pre {
                white-space: pre-wrap;
                word-wrap: break-word;
                font-size: 10pt;
            }
            
            table {
                font-size: 10pt;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Complete User Flow Guide</h1>
        <div style="color: #7f8c8d; font-size: 1.1em;">Secure Email System</div>
        <div style="color: #7f8c8d; font-size: 1.1em;">Comprehensive User Experience Documentation</div>
    </div>
"@

# Read the markdown content
$markdownContent = Get-Content $markdownFile -Raw -Encoding UTF8

# Convert markdown to HTML (basic conversion)
$htmlContent = $htmlContent + $markdownContent

# Add footer
$htmlContent = $htmlContent + @"
    <div class="footer">
        <p><strong>Secure Email System - Complete User Flow Guide</strong></p>
        <p>Version 1.0 | Last Updated: August 2025 | System Version: Secure Email MVP v1.0</p>
        <p>This document provides a comprehensive guide to the Secure Email system user flows.</p>
        <p>For technical implementation details, please refer to the developer documentation and API specifications.</p>
    </div>
</body>
</html>
"@

# Save the HTML file
$htmlContent | Out-File -FilePath $htmlFile -Encoding UTF8

Write-Host "HTML file generated successfully: $htmlFile" -ForegroundColor Green

# Create conversion guide
$conversionGuide = @"
PDF Conversion Guide

Option 1: Browser Print to PDF (Recommended)
1. Open the HTML file in your web browser
2. Press Ctrl+P (or Cmd+P on Mac)
3. Select "Save as PDF" as the destination
4. Choose your preferred settings:
   - Page size: A4
   - Margins: Default or Minimum
   - Include background graphics: Yes
5. Click "Save" and choose your file location

Option 2: Online Converters
1. Visit an online HTML to PDF converter
2. Upload the HTML file: $htmlFile
3. Convert and download the PDF

File Locations
- HTML File: $htmlFile
- Output Directory: $outputDir
"@

$conversionGuide | Out-File -FilePath "$outputDir/pdf-conversion-guide.txt" -Encoding UTF8

Write-Host "Conversion guide created: $outputDir/pdf-conversion-guide.txt" -ForegroundColor Cyan

# Open the HTML file in default browser
Write-Host "Opening HTML file in browser..." -ForegroundColor Yellow
Start-Process $htmlFile

Write-Host "HTML Generation Complete!" -ForegroundColor Green
Write-Host "HTML File: $htmlFile" -ForegroundColor Cyan
Write-Host "To convert to PDF: Open the HTML file and use browser Print to PDF" -ForegroundColor Yellow
