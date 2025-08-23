# =============================================================================
# Generate User Flow PDF Script
# =============================================================================
# Converts the complete user flow guide to PDF format
# =============================================================================

Write-Host "🚀 Generating User Flow PDF..." -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Gray

# Check if pandoc is installed
try {
    $pandocVersion = pandoc --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Pandoc found - using Pandoc for PDF generation" -ForegroundColor Green
        $usePandoc = $true
    } else {
        throw "Pandoc not found"
    }
} catch {
    Write-Host "⚠️ Pandoc not found - using alternative method" -ForegroundColor Yellow
    $usePandoc = $false
}

# Check if wkhtmltopdf is installed
try {
    $wkhtmltopdfVersion = wkhtmltopdf --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ wkhtmltopdf found - using wkhtmltopdf for PDF generation" -ForegroundColor Green
        $useWkhtmltopdf = $true
    } else {
        throw "wkhtmltopdf not found"
    }
} catch {
    Write-Host "⚠️ wkhtmltopdf not found - using alternative method" -ForegroundColor Yellow
    $useWkhtmltopdf = $false
}

# Check if we have any PDF generation tools
if (-not $usePandoc -and -not $useWkhtmltopdf) {
    Write-Host "❌ No PDF generation tools found" -ForegroundColor Red
    Write-Host "Please install one of the following:" -ForegroundColor Yellow
    Write-Host "  - Pandoc: https://pandoc.org/installing.html" -ForegroundColor Cyan
    Write-Host "  - wkhtmltopdf: https://wkhtmltopdf.org/downloads.html" -ForegroundColor Cyan
    Write-Host "  - Or use an online Markdown to PDF converter" -ForegroundColor Cyan
    exit 1
}

# Define paths
$markdownFile = "docs/complete-user-flow-guide.md"
$outputDir = "docs/pdf"
$pdfFile = "$outputDir/complete-user-flow-guide.pdf"
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

# Method 1: Using Pandoc (preferred)
if ($usePandoc) {
    Write-Host "🔄 Converting to PDF using Pandoc..." -ForegroundColor Yellow
    
    try {
        # Create PDF with custom styling
        pandoc $markdownFile `
            --pdf-engine=xelatex `
            --variable geometry:margin=1in `
            --variable fontsize=11pt `
            --variable mainfont="DejaVu Sans" `
            --variable monofont="DejaVu Sans Mono" `
            --variable colorlinks=true `
            --variable linkcolor=blue `
            --variable urlcolor=blue `
            --variable toccolor=gray `
            --toc `
            --number-sections `
            --highlight-style=tango `
            -o $pdfFile
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ PDF generated successfully: $pdfFile" -ForegroundColor Green
        } else {
            throw "Pandoc conversion failed"
        }
    } catch {
        Write-Host "❌ Pandoc PDF generation failed: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "🔄 Trying HTML to PDF conversion..." -ForegroundColor Yellow
        
        # Fallback: Generate HTML first, then convert to PDF
        try {
            pandoc $markdownFile `
                --standalone `
                --self-contained `
                --css=scripts/pdf-styles.css `
                -o $htmlFile
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ HTML generated: $htmlFile" -ForegroundColor Green
                
                # Convert HTML to PDF using wkhtmltopdf if available
                if ($useWkhtmltopdf) {
                    wkhtmltopdf --page-size A4 --margin-top 0.75in --margin-right 0.75in --margin-bottom 0.75in --margin-left 0.75in --encoding UTF-8 $htmlFile $pdfFile
                    
                    if ($LASTEXITCODE -eq 0) {
                        Write-Host "✅ PDF generated successfully: $pdfFile" -ForegroundColor Green
                    } else {
                        throw "wkhtmltopdf conversion failed"
                    }
                } else {
                    Write-Host "⚠️ HTML file generated but no PDF converter available" -ForegroundColor Yellow
                    Write-Host "📄 HTML file: $htmlFile" -ForegroundColor Cyan
                }
            } else {
                throw "HTML generation failed"
            }
        } catch {
            Write-Host "❌ HTML generation failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

# Method 2: Using wkhtmltopdf directly (if Pandoc failed)
elseif ($useWkhtmltopdf) {
    Write-Host "🔄 Converting to PDF using wkhtmltopdf..." -ForegroundColor Yellow
    
    try {
        # First convert markdown to HTML
        pandoc $markdownFile `
            --standalone `
            --self-contained `
            --css=scripts/pdf-styles.css `
            -o $htmlFile
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ HTML generated: $htmlFile" -ForegroundColor Green
            
            # Convert HTML to PDF
            wkhtmltopdf --page-size A4 --margin-top 0.75in --margin-right 0.75in --margin-bottom 0.75in --margin-left 0.75in --encoding UTF-8 $htmlFile $pdfFile
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ PDF generated successfully: $pdfFile" -ForegroundColor Green
            } else {
                throw "wkhtmltopdf conversion failed"
            }
        } else {
            throw "HTML generation failed"
        }
    } catch {
        Write-Host "❌ PDF generation failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Create CSS file for better PDF styling
$cssContent = @"
/* PDF Styling for User Flow Guide */
body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    line-height: 1.6;
    color: #333;
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
}

h1, h2, h3, h4, h5, h6 {
    color: #2c3e50;
    margin-top: 1.5em;
    margin-bottom: 0.5em;
}

h1 {
    font-size: 2.5em;
    border-bottom: 3px solid #3498db;
    padding-bottom: 10px;
}

h2 {
    font-size: 2em;
    border-bottom: 2px solid #ecf0f1;
    padding-bottom: 8px;
}

h3 {
    font-size: 1.5em;
    color: #34495e;
}

h4 {
    font-size: 1.3em;
    color: #7f8c8d;
}

code {
    background-color: #f8f9fa;
    padding: 2px 4px;
    border-radius: 3px;
    font-family: 'Courier New', monospace;
    font-size: 0.9em;
}

pre {
    background-color: #f8f9fa;
    padding: 15px;
    border-radius: 5px;
    overflow-x: auto;
    border-left: 4px solid #3498db;
}

pre code {
    background-color: transparent;
    padding: 0;
}

blockquote {
    border-left: 4px solid #3498db;
    margin: 0;
    padding-left: 20px;
    color: #7f8c8d;
    font-style: italic;
}

table {
    border-collapse: collapse;
    width: 100%;
    margin: 20px 0;
}

th, td {
    border: 1px solid #ddd;
    padding: 12px;
    text-align: left;
}

th {
    background-color: #3498db;
    color: white;
    font-weight: bold;
}

tr:nth-child(even) {
    background-color: #f2f2f2;
}

ul, ol {
    padding-left: 20px;
}

li {
    margin-bottom: 5px;
}

a {
    color: #3498db;
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

/* Special styling for security features */
.security-feature {
    background-color: #e8f5e8;
    border: 1px solid #4caf50;
    border-radius: 5px;
    padding: 15px;
    margin: 10px 0;
}

.security-feature h4 {
    color: #2e7d32;
    margin-top: 0;
}

/* Workflow steps */
.workflow-step {
    background-color: #f0f8ff;
    border-left: 4px solid #2196f3;
    padding: 10px 15px;
    margin: 10px 0;
}

/* Code blocks */
.language-go, .language-typescript, .language-javascript {
    background-color: #f4f4f4;
    border: 1px solid #ddd;
    border-radius: 5px;
    padding: 15px;
    overflow-x: auto;
}

/* Emojis and icons */
.emoji {
    font-size: 1.2em;
    margin-right: 5px;
}

/* Print styles */
@media print {
    body {
        font-size: 12pt;
        line-height: 1.4;
    }
    
    h1 { font-size: 18pt; }
    h2 { font-size: 16pt; }
    h3 { font-size: 14pt; }
    h4 { font-size: 12pt; }
    
    .no-print {
        display: none;
    }
    
    pre {
        white-space: pre-wrap;
        word-wrap: break-word;
    }
}
"@

# Save CSS file
$cssFile = "scripts/pdf-styles.css"
$cssContent | Out-File -FilePath $cssFile -Encoding UTF8
Write-Host "✅ CSS styling file created: $cssFile" -ForegroundColor Green

# Check if PDF was generated successfully
if (Test-Path $pdfFile) {
    $fileSize = (Get-Item $pdfFile).Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    
    Write-Host "🎉 PDF Generation Complete!" -ForegroundColor Green
    Write-Host "==================================================================" -ForegroundColor Gray
    Write-Host "📄 PDF File: $pdfFile" -ForegroundColor Cyan
    Write-Host "📏 File Size: $fileSizeMB MB" -ForegroundColor Cyan
    Write-Host "📁 Output Directory: $outputDir" -ForegroundColor Cyan
    
    # Open the PDF file
    Write-Host "🔄 Opening PDF file..." -ForegroundColor Yellow
    Start-Process $pdfFile
    
} else {
    Write-Host "❌ PDF generation failed" -ForegroundColor Red
    Write-Host "📄 HTML file available: $htmlFile" -ForegroundColor Cyan
    Write-Host "💡 You can manually convert the HTML file to PDF using:" -ForegroundColor Yellow
    Write-Host "   - Browser: Print to PDF" -ForegroundColor Cyan
    Write-Host "   - Online converters" -ForegroundColor Cyan
    Write-Host "   - PDF software" -ForegroundColor Cyan
}

Write-Host "==================================================================" -ForegroundColor Gray
Write-Host "✅ User Flow PDF generation script completed" -ForegroundColor Green
