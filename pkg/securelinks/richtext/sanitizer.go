package richtext

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"time"

	"secure-email-mvp/pkg/models"
)

// Sanitizer provides HTML sanitization and feature detection
type Sanitizer struct {
	allowedTags    map[string]bool
	blockedTags    map[string]bool
	blockedAttrs   map[string]bool
	urlWhitelist   []string
	maxContentSize int
}

// NewSanitizer creates a new HTML sanitizer with secure defaults
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		allowedTags: map[string]bool{
			"p": true, "br": true, "strong": true, "b": true, "em": true, "i": true,
			"u": true, "ul": true, "ol": true, "li": true, "blockquote": true,
			"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
			"code": true, "pre": true, "table": true, "thead": true, "tbody": true,
			"tr": true, "td": true, "th": true, "a": true, "img": true, "span": true,
			"div": true, "hr": true,
		},
		blockedTags: map[string]bool{
			"script": true, "style": true, "iframe": true, "object": true,
			"embed": true, "form": true, "input": true, "button": true,
			"select": true, "textarea": true, "meta": true, "link": true,
			"base": true, "title": true, "head": true, "body": true,
		},
		blockedAttrs: map[string]bool{
			"onclick": true, "onload": true, "onerror": true, "onmouseover": true,
			"onmouseout": true, "onfocus": true, "onblur": true, "onchange": true,
			"onsubmit": true, "onreset": true, "onselect": true, "onunload": true,
			"onkeydown": true, "onkeyup": true, "onkeypress": true,
		},
		urlWhitelist: []string{
			"http://", "https://", "mailto:", "tel:", "ftp://",
		},
		maxContentSize: 1024 * 1024, // 1MB max
	}
}

// SanitizeHTML sanitizes HTML content and returns sanitized version with features
func (s *Sanitizer) SanitizeHTML(content string) (*models.RichTextContent, error) {
	if len(content) > s.maxContentSize {
		return nil, fmt.Errorf("content too large: %d bytes (max: %d)", len(content), s.maxContentSize)
	}

	// Basic HTML sanitization using regex
	sanitizedHTML := s.sanitizeHTMLBasic(content)

	// Extract features
	features := s.extractFeaturesBasic(sanitizedHTML)

	// Generate content hash
	contentHash := s.generateContentHash(sanitizedHTML)

	// Convert features to JSON
	featuresJSON, err := features.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize features: %w", err)
	}

	return &models.RichTextContent{
		ContentID:        s.generateContentID(),
		RawContent:       &content,
		SanitizedContent: sanitizedHTML,
		ContentHash:      contentHash,
		FeaturesUsed:     &featuresJSON,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

// sanitizeHTMLBasic performs basic HTML sanitization using regex
func (s *Sanitizer) sanitizeHTMLBasic(content string) string {
	// Remove blocked tags and their content
	for tag := range s.blockedTags {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>.*?</%s>`, tag, tag))
		content = pattern.ReplaceAllString(content, "")
	}

	// Remove blocked attributes
	for attr := range s.blockedAttrs {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)\s+%s\s*=\s*["'][^"']*["']`, attr))
		content = pattern.ReplaceAllString(content, "")
	}

	// Remove dangerous URLs
	content = s.sanitizeURLs(content)

	// Escape any remaining dangerous content
	content = html.EscapeString(content)

	return content
}

// sanitizeURLs removes dangerous URLs
func (s *Sanitizer) sanitizeURLs(content string) string {
	// Remove javascript: URLs
	pattern := regexp.MustCompile(`(?i)href\s*=\s*["']javascript:[^"']*["']`)
	content = pattern.ReplaceAllString(content, `href="#"`)

	// Remove data: URLs
	pattern = regexp.MustCompile(`(?i)src\s*=\s*["']data:[^"']*["']`)
	content = pattern.ReplaceAllString(content, `src=""`)

	return content
}

// extractFeaturesBasic extracts features using regex patterns
func (s *Sanitizer) extractFeaturesBasic(content string) *models.RichTextFeatures {
	features := &models.RichTextFeatures{}

	// Check for bold text
	if regexp.MustCompile(`(?i)<(strong|b)[^>]*>`).MatchString(content) {
		features.Bold = true
	}

	// Check for italic text
	if regexp.MustCompile(`(?i)<(em|i)[^>]*>`).MatchString(content) {
		features.Italic = true
	}

	// Check for underlined text
	if regexp.MustCompile(`(?i)<u[^>]*>`).MatchString(content) {
		features.Underline = true
	}

	// Check for lists
	if regexp.MustCompile(`(?i)<(ul|ol)[^>]*>`).MatchString(content) {
		features.Lists = true
	}

	// Check for images
	if regexp.MustCompile(`(?i)<img[^>]*>`).MatchString(content) {
		features.Images = true
	}

	// Check for tables
	if regexp.MustCompile(`(?i)<table[^>]*>`).MatchString(content) {
		features.Tables = true
	}

	// Check for code blocks
	if regexp.MustCompile(`(?i)<(code|pre)[^>]*>`).MatchString(content) {
		features.CodeBlocks = true
	}

	// Check for quotes
	if regexp.MustCompile(`(?i)<blockquote[^>]*>`).MatchString(content) {
		features.Quotes = true
	}

	// Check for headings
	if regexp.MustCompile(`(?i)<h[1-6][^>]*>`).MatchString(content) {
		features.Headings = true
	}

	// Extract links
	linkPattern := regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*)["']`)
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			features.Links = append(features.Links, match[1])
		}
	}

	// Check for colors and fonts in style attributes
	if regexp.MustCompile(`(?i)color\s*:`).MatchString(content) {
		features.Colors = true
	}
	if regexp.MustCompile(`(?i)font-size\s*:`).MatchString(content) {
		features.FontSizes = true
	}

	return features
}

// generateContentHash creates a SHA-256 hash of the content
func (s *Sanitizer) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// generateContentID creates a unique content ID
func (s *Sanitizer) generateContentID() string {
	return fmt.Sprintf("content_%s", hex.EncodeToString([]byte(time.Now().Format("20060102150405"))))
}

// ValidateContentSize checks if content is within size limits
func (s *Sanitizer) ValidateContentSize(content string) error {
	if len(content) > s.maxContentSize {
		return fmt.Errorf("content too large: %d bytes (max: %d)", len(content), s.maxContentSize)
	}
	return nil
}

// GetContentStats returns statistics about the content
func (s *Sanitizer) GetContentStats(content string) map[string]interface{} {
	stats := map[string]interface{}{
		"size_bytes": len(content),
		"tags":       make(map[string]int),
		"links":      []string{},
		"images":     0,
	}

	// Count tags
	for tag := range s.allowedTags {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>`, tag))
		matches := pattern.FindAllString(content, -1)
		stats["tags"].(map[string]int)[tag] = len(matches)
	}

	// Count images
	imgPattern := regexp.MustCompile(`(?i)<img[^>]*>`)
	stats["images"] = len(imgPattern.FindAllString(content, -1))

	// Extract links
	linkPattern := regexp.MustCompile(`(?i)href\s*=\s*["']([^"']*)["']`)
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			stats["links"] = append(stats["links"].([]string), match[1])
		}
	}

	return stats
}
