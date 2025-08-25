package ai_dlp

import (
	"fmt"
	"regexp"
	"strings"

	"secure-email-mvp/pkg/models"
)

// NLPClassifier implements the Classifier interface using NLP techniques
type NLPClassifier struct {
	categories   map[string]models.ContentCategory
	patterns     map[string][]*regexp.Regexp
	keywords     map[string][]string
	modelVersion string
}

// NewNLPClassifier creates a new NLP-based classifier
func NewNLPClassifier(modelVersion string) *NLPClassifier {
	classifier := &NLPClassifier{
		categories:   models.ContentCategories,
		patterns:     make(map[string][]*regexp.Regexp),
		keywords:     make(map[string][]string),
		modelVersion: modelVersion,
	}

	// Compile regex patterns for each category
	for categoryID, category := range classifier.categories {
		classifier.keywords[categoryID] = category.Keywords

		var patterns []*regexp.Regexp
		for _, pattern := range category.Patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				patterns = append(patterns, re)
			}
		}
		classifier.patterns[categoryID] = patterns
	}

	return classifier
}

// ClassifyContent performs AI-based content classification
func (c *NLPClassifier) ClassifyContent(content string) (*models.AIContentClassification, error) {
	content = strings.ToLower(content)

	// Analyze content for each category
	var bestCategory string
	var bestScore float64
	var detectedKeywords []string

	// Check each category
	for categoryID, category := range c.categories {
		score := c.calculateCategoryScore(content, categoryID)
		if score > bestScore {
			bestScore = score
			bestCategory = categoryID
		}

		// Collect keywords found in this category
		for _, keyword := range category.Keywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				detectedKeywords = append(detectedKeywords, keyword)
			}
		}
	}

	// If no significant matches found, classify as "none"
	if bestScore < 0.1 {
		bestCategory = "none"
		bestScore = 0.0
	}

	// Extract entities
	entities := c.extractEntitiesFromText(content)

	// Calculate confidence based on multiple factors
	confidence := c.calculateConfidence(bestScore, len(detectedKeywords), len(entities))

	// Determine severity based on category
	severity := c.getSeverityForCategory(bestCategory, bestScore)

	// Calculate risk score
	riskScore := c.calculateRiskScore(bestScore, bestCategory, len(detectedKeywords), len(entities))

	return &models.AIContentClassification{
		Category:     bestCategory,
		Confidence:   confidence,
		Severity:     severity,
		RiskScore:    riskScore,
		Keywords:     detectedKeywords,
		Entities:     entities,
		Context:      c.extractContext(content, bestCategory),
		ModelVersion: c.modelVersion,
	}, nil
}

// ExtractEntities extracts named entities from content
func (c *NLPClassifier) ExtractEntities(content string) ([]models.Entity, error) {
	return c.extractEntitiesFromText(content), nil
}

// CalculateRiskScore calculates the overall risk score for a classification
func (c *NLPClassifier) CalculateRiskScore(classification *models.AIContentClassification) float64 {
	if classification == nil {
		return 0.0
	}

	// Base risk score from classification
	baseScore := classification.RiskScore

	// Adjust based on number of entities found
	entityMultiplier := 1.0 + (float64(len(classification.Entities)) * 0.1)
	if entityMultiplier > 1.5 {
		entityMultiplier = 1.5
	}

	// Adjust based on number of keywords found
	keywordMultiplier := 1.0 + (float64(len(classification.Keywords)) * 0.05)
	if keywordMultiplier > 1.3 {
		keywordMultiplier = 1.3
	}

	// Calculate final risk score
	finalScore := baseScore * entityMultiplier * keywordMultiplier

	// Ensure score is between 0 and 1
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// calculateCategoryScore calculates how well content matches a specific category
func (c *NLPClassifier) calculateCategoryScore(content, categoryID string) float64 {
	category := c.categories[categoryID]
	score := 0.0

	// Check keyword matches
	keywordMatches := 0
	for _, keyword := range category.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			keywordMatches++
		}
	}

	// Calculate keyword score (0-0.4)
	keywordScore := float64(keywordMatches) / float64(len(category.Keywords)) * 0.4
	score += keywordScore

	// Check pattern matches
	patternMatches := 0
	if patterns, exists := c.patterns[categoryID]; exists {
		for _, pattern := range patterns {
			if pattern.MatchString(content) {
				patternMatches++
			}
		}
	}

	// Calculate pattern score (0-0.6)
	if len(c.patterns[categoryID]) > 0 {
		patternScore := float64(patternMatches) / float64(len(c.patterns[categoryID])) * 0.6
		score += patternScore
	}

	return score
}

// extractEntitiesFromText extracts named entities from text
func (c *NLPClassifier) extractEntitiesFromText(content string) []models.Entity {
	var entities []models.Entity

	// Extract credit card numbers
	ccPattern := regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
	ccMatches := ccPattern.FindAllStringIndex(content, -1)
	for _, match := range ccMatches {
		entities = append(entities, models.Entity{
			Type:       "account",
			Value:      content[match[0]:match[1]],
			Confidence: 0.95,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	// Extract SSNs
	ssnPattern := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	ssnMatches := ssnPattern.FindAllStringIndex(content, -1)
	for _, match := range ssnMatches {
		entities = append(entities, models.Entity{
			Type:       "person",
			Value:      content[match[0]:match[1]],
			Confidence: 0.98,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	// Extract email addresses
	emailPattern := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	emailMatches := emailPattern.FindAllStringIndex(content, -1)
	for _, match := range emailMatches {
		entities = append(entities, models.Entity{
			Type:       "email",
			Value:      content[match[0]:match[1]],
			Confidence: 0.9,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	// Extract phone numbers
	phonePattern := regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
	phoneMatches := phonePattern.FindAllStringIndex(content, -1)
	for _, match := range phoneMatches {
		entities = append(entities, models.Entity{
			Type:       "phone",
			Value:      content[match[0]:match[1]],
			Confidence: 0.85,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	// Extract amounts (currency)
	amountPattern := regexp.MustCompile(`\$\d{1,3}(?:,\d{3})*(?:\.\d{2})?`)
	amountMatches := amountPattern.FindAllStringIndex(content, -1)
	for _, match := range amountMatches {
		entities = append(entities, models.Entity{
			Type:       "amount",
			Value:      content[match[0]:match[1]],
			Confidence: 0.8,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	// Extract dates
	datePattern := regexp.MustCompile(`\b\d{1,2}[/-]\d{1,2}[/-]\d{2,4}\b`)
	dateMatches := datePattern.FindAllStringIndex(content, -1)
	for _, match := range dateMatches {
		entities = append(entities, models.Entity{
			Type:       "date",
			Value:      content[match[0]:match[1]],
			Confidence: 0.7,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	return entities
}

// calculateConfidence calculates confidence score based on multiple factors
func (c *NLPClassifier) calculateConfidence(categoryScore float64, keywordCount, entityCount int) float64 {
	// Base confidence from category score
	confidence := categoryScore

	// Boost confidence based on keyword matches
	keywordBoost := float64(keywordCount) * 0.1
	if keywordBoost > 0.3 {
		keywordBoost = 0.3
	}
	confidence += keywordBoost

	// Boost confidence based on entity matches
	entityBoost := float64(entityCount) * 0.15
	if entityBoost > 0.4 {
		entityBoost = 0.4
	}
	confidence += entityBoost

	// Ensure confidence is between 0 and 1
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getSeverityForCategory determines severity based on category and score
func (c *NLPClassifier) getSeverityForCategory(category string, score float64) string {
	if category == "none" {
		return "none"
	}

	categoryInfo := c.categories[category]
	baseSeverity := categoryInfo.Severity

	// Adjust severity based on score
	if score > 0.8 {
		switch baseSeverity {
		case "medium":
			return "high"
		case "low":
			return "medium"
		default:
			return baseSeverity
		}
	} else if score < 0.3 {
		switch baseSeverity {
		case "critical":
			return "high"
		case "high":
			return "medium"
		case "medium":
			return "low"
		default:
			return baseSeverity
		}
	}

	return baseSeverity
}

// calculateRiskScore calculates risk score based on category and factors
func (c *NLPClassifier) calculateRiskScore(categoryScore float64, category string, keywordCount, entityCount int) float64 {
	// Get base risk weight for category
	baseRisk := 0.5
	if categoryInfo, exists := c.categories[category]; exists {
		baseRisk = categoryInfo.RiskWeight
	}

	// Calculate risk score
	riskScore := baseRisk * categoryScore

	// Adjust based on keyword density
	keywordDensity := float64(keywordCount) * 0.1
	if keywordDensity > 0.3 {
		keywordDensity = 0.3
	}
	riskScore += keywordDensity

	// Adjust based on entity density
	entityDensity := float64(entityCount) * 0.15
	if entityDensity > 0.4 {
		entityDensity = 0.4
	}
	riskScore += entityDensity

	// Ensure risk score is between 0 and 1
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore
}

// extractContext extracts contextual information about the classification
func (c *NLPClassifier) extractContext(content, category string) string {
	if category == "none" {
		return "No sensitive content detected"
	}

	categoryInfo := c.categories[category]

	// Count occurrences of category keywords
	keywordCount := 0
	for _, keyword := range categoryInfo.Keywords {
		if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
			keywordCount++
		}
	}

	// Generate context description
	switch {
	case keywordCount > 5:
		return fmt.Sprintf("High concentration of %s keywords detected", categoryInfo.Name)
	case keywordCount > 2:
		return fmt.Sprintf("Multiple %s keywords detected", categoryInfo.Name)
	case keywordCount > 0:
		return fmt.Sprintf("Some %s keywords detected", categoryInfo.Name)
	default:
		return fmt.Sprintf("Pattern-based %s detection", categoryInfo.Name)
	}
}
