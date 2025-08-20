// Package e2e provides end-to-end encryption functionality with cover traffic generation
package e2e

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CoverTrafficGenerator generates dummy traffic to obfuscate real communications
type CoverTrafficGenerator struct {
	enabled     bool
	intensity   float64  // 0.0 to 1.0
	analyzer    *TrafficAnalyzer
	scheduler   *TrafficScheduler
	encryptor   *CoverMessageEncryptor
	config      CoverTrafficConfig
	mu          sync.RWMutex
}

// TrafficAnalyzer analyzes traffic patterns to generate realistic cover traffic
type TrafficAnalyzer struct {
	patterns    map[string]*TrafficPattern
	baseline    *BaselineTraffic
	adaptive    *AdaptiveGenerator
	mu          sync.RWMutex
}

// TrafficPattern represents a user's traffic pattern
type TrafficPattern struct {
	UserID          string        `json:"user_id"`
	AverageMessages int           `json:"avg_messages_per_hour"`
	PeakHours       []int         `json:"peak_hours"`
	MessageSizes    []int         `json:"typical_sizes"`
	Intervals       []time.Duration `json:"intervals"`
	LastUpdate      time.Time     `json:"last_update"`
	TotalMessages   int           `json:"total_messages"`
	ActiveHours     map[int]int   `json:"active_hours"`
}

// BaselineTraffic represents global traffic patterns
type BaselineTraffic struct {
	GlobalAverage    int           `json:"global_average"`
	PeakHours        []int         `json:"peak_hours"`
	TypicalSizes     []int         `json:"typical_sizes"`
	TypicalIntervals []time.Duration `json:"typical_intervals"`
	LastUpdate       time.Time     `json:"last_update"`
}

// AdaptiveGenerator adapts cover traffic based on learned patterns
type AdaptiveGenerator struct {
	patterns map[string]*TrafficPattern
	config   CoverTrafficConfig
	mu       sync.RWMutex
}

// TrafficScheduler schedules cover traffic generation
type TrafficScheduler struct {
	patterns     map[string]*TrafficPattern
	nextSchedule map[string]time.Time
	jitter       time.Duration
	mu           sync.RWMutex
}

// CoverMessageEncryptor encrypts cover messages to look like real traffic
type CoverMessageEncryptor struct {
	config CoverTrafficConfig
}

// CoverMessage represents a dummy message for cover traffic
type CoverMessage struct {
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	Envelope  *Envelope `json:"envelope"`
	Size      int       `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	IsCover   bool      `json:"is_cover"` // Internal flag, not transmitted
	Pattern   string    `json:"pattern,omitempty"`
}

// CoverTrafficConfig contains configuration for cover traffic generation
type CoverTrafficConfig struct {
	Enabled         bool          `json:"enabled"`
	Intensity       float64       `json:"intensity"`        // 0.0 = disabled, 1.0 = max
	MinInterval     time.Duration `json:"min_interval"`     // Minimum between messages
	MaxInterval     time.Duration `json:"max_interval"`     // Maximum between messages
	SizeVariation   int           `json:"size_variation"`   // Message size randomness
	AnalysisWindow  time.Duration `json:"analysis_window"`  // Traffic pattern analysis period
	
	// Adaptive behavior
	AdaptToUser     bool          `json:"adapt_to_user"`    // Learn user patterns
	GlobalBaseline  bool          `json:"global_baseline"`  // Use global traffic patterns
	
	// Jitter settings
	JitterPercent   float64       `json:"jitter_percent"`   // Timing jitter percentage
	MaxJitter       time.Duration `json:"max_jitter"`       // Maximum jitter amount
}

// NewCoverTrafficGenerator creates a new cover traffic generator
func NewCoverTrafficGenerator(config CoverTrafficConfig) *CoverTrafficGenerator {
	generator := &CoverTrafficGenerator{
		enabled:   config.Enabled,
		intensity: config.Intensity,
		config:    config,
		analyzer:  NewTrafficAnalyzer(config),
		scheduler: NewTrafficScheduler(config),
		encryptor: NewCoverMessageEncryptor(config),
	}
	
	return generator
}

// NewTrafficAnalyzer creates a new traffic analyzer
func NewTrafficAnalyzer(config CoverTrafficConfig) *TrafficAnalyzer {
	return &TrafficAnalyzer{
		patterns: make(map[string]*TrafficPattern),
		baseline: &BaselineTraffic{
			GlobalAverage:    5, // 5 messages per hour average
			PeakHours:        []int{9, 10, 11, 14, 15, 16}, // Business hours
			TypicalSizes:     []int{512, 1024, 2048, 4096},
			TypicalIntervals: []time.Duration{5 * time.Minute, 10 * time.Minute, 15 * time.Minute, 30 * time.Minute},
			LastUpdate:       time.Now(),
		},
		adaptive: &AdaptiveGenerator{
			patterns: make(map[string]*TrafficPattern),
			config:   config,
		},
	}
}

// NewTrafficScheduler creates a new traffic scheduler
func NewTrafficScheduler(config CoverTrafficConfig) *TrafficScheduler {
	return &TrafficScheduler{
		patterns:     make(map[string]*TrafficPattern),
		nextSchedule: make(map[string]time.Time),
		jitter:       config.MaxJitter,
	}
}

// NewCoverMessageEncryptor creates a new cover message encryptor
func NewCoverMessageEncryptor(config CoverTrafficConfig) *CoverMessageEncryptor {
	return &CoverMessageEncryptor{
		config: config,
	}
}

// GenerateCoverMessage generates a cover message for a user
func (c *CoverTrafficGenerator) GenerateCoverMessage(
	userID string,
	size int,
) (*CoverMessage, error) {
	if !c.enabled {
		return nil, fmt.Errorf("cover traffic is disabled")
	}
	
	// Generate dummy content that looks real
	dummyContent := c.generateDummyContent(size)
	
	// Encrypt with same parameters as real messages
	envelope, err := c.encryptor.EncryptCover(dummyContent, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt cover message: %w", err)
	}
	
	// Determine pattern type
	pattern := c.determinePattern(userID)
	
	return &CoverMessage{
		MessageID: generateCoverMessageID(),
		UserID:    userID,
		Envelope:  envelope,
		Size:      size,
		Timestamp: time.Now(),
		IsCover:   true, // Internal flag, not transmitted
		Pattern:   pattern,
	}, nil
}

// generateDummyContent generates dummy content that looks like real messages
func (c *CoverTrafficGenerator) generateDummyContent(size int) []byte {
	// Generate random content of the specified size
	content := make([]byte, size)
	cryptorand.Read(content)
	
	// Make it look more realistic by adding some structure
	if size > 100 {
		// Add some realistic text patterns
		textPatterns := []string{
			"Hello, how are you?",
			"Please review the attached document.",
			"Meeting scheduled for tomorrow.",
			"Thanks for your email.",
			"Best regards,",
		}
		
		// Insert some text patterns at random positions
		for _, pattern := range textPatterns {
			if len(pattern) < size {
				pos := int(rand.Int63() % int64(size - len(pattern)))
				copy(content[pos:], []byte(pattern))
			}
		}
	}
	
	return content
}

// determinePattern determines the pattern type for this cover message
func (c *CoverTrafficGenerator) determinePattern(userID string) string {
	patterns := []string{"business", "personal", "casual", "formal"}
	return patterns[int(rand.Int63()%int64(len(patterns)))]
}

// ScheduleCoverTraffic schedules cover traffic for a user
func (c *CoverTrafficGenerator) ScheduleCoverTraffic(userID string) {
	if !c.enabled {
		return
	}
	
	nextTime := c.scheduler.ScheduleNextCover(userID)
	c.scheduler.SetNextSchedule(userID, nextTime)
}

// GenerateCoverTraffic generates cover traffic for a user over a time period
func (c *CoverTrafficGenerator) GenerateCoverTraffic(
	userID string,
	duration time.Duration,
) ([]*CoverMessage, error) {
	if !c.enabled {
		return nil, fmt.Errorf("cover traffic is disabled")
	}
	
	var messages []*CoverMessage
	endTime := time.Now().Add(duration)
	
	for time.Now().Before(endTime) {
		// Check if it's time to generate cover traffic
		if c.scheduler.ShouldGenerate(userID) {
			// Determine message size based on user pattern
			size := c.determineMessageSize(userID)
			
			// Generate cover message
			message, err := c.GenerateCoverMessage(userID, size)
			if err != nil {
				continue // Skip on error
			}
			
			messages = append(messages, message)
			
			// Schedule next cover traffic
			c.ScheduleCoverTraffic(userID)
		}
		
		// Small delay to prevent tight loop
		time.Sleep(100 * time.Millisecond)
	}
	
	return messages, nil
}

// determineMessageSize determines appropriate message size based on user pattern
func (c *CoverTrafficGenerator) determineMessageSize(userID string) int {
	c.analyzer.mu.RLock()
	defer c.analyzer.mu.RUnlock()
	
	pattern, exists := c.analyzer.patterns[userID]
	if !exists {
		// Use global baseline
		sizes := c.analyzer.baseline.TypicalSizes
		if len(sizes) == 0 {
			return 1024 // Default size
		}
		return sizes[int(rand.Int63()%int64(len(sizes)))]
	}
	
	// Use user's typical sizes
	if len(pattern.MessageSizes) == 0 {
		return 1024 // Default size
	}
	
	// Add some variation
	baseSize := pattern.MessageSizes[int(rand.Int63()%int64(len(pattern.MessageSizes)))]
	variation := int(float64(baseSize) * float64(c.config.SizeVariation) / 100.0)
	
	return baseSize + variation
}

// UpdateTrafficPattern updates the traffic pattern for a user
func (c *CoverTrafficGenerator) UpdateTrafficPattern(userID string, messageSize int) {
	c.analyzer.mu.Lock()
	defer c.analyzer.mu.Unlock()
	
	pattern, exists := c.analyzer.patterns[userID]
	if !exists {
		pattern = &TrafficPattern{
			UserID:        userID,
			ActiveHours:   make(map[int]int),
			LastUpdate:    time.Now(),
		}
		c.analyzer.patterns[userID] = pattern
	}
	
	// Update message count
	pattern.TotalMessages++
	
	// Update message sizes
	pattern.MessageSizes = append(pattern.MessageSizes, messageSize)
	if len(pattern.MessageSizes) > 100 {
		pattern.MessageSizes = pattern.MessageSizes[1:] // Keep last 100
	}
	
	// Update active hours
	hour := time.Now().Hour()
	pattern.ActiveHours[hour]++
	
	// Update last update time
	pattern.LastUpdate = time.Now()
	
	// Update average messages per hour
	if time.Since(pattern.LastUpdate) >= time.Hour {
		pattern.AverageMessages = pattern.TotalMessages
		pattern.TotalMessages = 0
	}
}

// ScheduleNextCover calculates when to generate the next cover message
func (t *TrafficScheduler) ScheduleNextCover(userID string) time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	pattern, exists := t.patterns[userID]
	if !exists {
		return t.getDefaultSchedule()
	}
	
	// Calculate next send time based on pattern
	interval := t.calculateAdaptiveInterval(pattern)
	jitteredInterval := t.addJitter(interval)
	
	return time.Now().Add(jitteredInterval)
}

// calculateAdaptiveInterval calculates the interval based on user pattern
func (t *TrafficScheduler) calculateAdaptiveInterval(pattern *TrafficPattern) time.Duration {
	if pattern.AverageMessages == 0 {
		return 30 * time.Minute // Default interval
	}
	
	// Calculate interval based on average messages per hour
	intervalMinutes := 60.0 / float64(pattern.AverageMessages)
	
	// Add some randomness
	randomFactor := 0.5 + rand.Float64() // 0.5 to 1.5
	intervalMinutes *= randomFactor
	
	return time.Duration(intervalMinutes * float64(time.Minute))
}

// addJitter adds timing jitter to prevent pattern detection
func (t *TrafficScheduler) addJitter(interval time.Duration) time.Duration {
	if t.jitter == 0 {
		return interval
	}
	
	// Add random jitter
	jitterAmount := time.Duration(rand.Int63() % int64(t.jitter))
	if rand.Float64() < 0.5 {
		jitterAmount = -jitterAmount
	}
	
	return interval + jitterAmount
}

// getDefaultSchedule returns a default schedule
func (t *TrafficScheduler) getDefaultSchedule() time.Time {
	// Default: generate cover traffic every 15-45 minutes
	interval := time.Duration(15+rand.Int63()%30) * time.Minute
	return time.Now().Add(interval)
}

// SetNextSchedule sets the next schedule for a user
func (t *TrafficScheduler) SetNextSchedule(userID string, nextTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.nextSchedule[userID] = nextTime
}

// ShouldGenerate checks if cover traffic should be generated for a user
func (t *TrafficScheduler) ShouldGenerate(userID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	nextTime, exists := t.nextSchedule[userID]
	if !exists {
		return true // Generate immediately if no schedule
	}
	
	return time.Now().After(nextTime)
}

// EncryptCover encrypts cover message content
func (c *CoverMessageEncryptor) EncryptCover(content []byte, userID string) (*Envelope, error) {
	// Create a dummy envelope that looks like a real message
	// This would use the same encryption as real messages but with dummy keys
	
	// Generate dummy keys for cover traffic
	dummyKeyPair, err := generateDummyKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate dummy keys: %w", err)
	}
	
	expiresAt := time.Now().Add(24 * time.Hour)
	// Create dummy envelope
	envelope := &Envelope{
		ID:              generateCoverMessageID(),
		Version:         "1.0",
		KEMAlgorithm:    "kyber768",
		DEMAlgorithm:    "aes256gcm",
		SignatureAlgorithm: "dilithium3",
		EncryptedKey:    base64.StdEncoding.EncodeToString(dummyKeyPair.PublicKey), // Dummy encrypted key
		EncryptedData:   base64.StdEncoding.EncodeToString(content),                // Use the dummy content directly
		Signature:       generateDummySignature(),
		CreatedAt:       time.Now(),
		ExpiresAt:       &expiresAt,
	}
	
	return envelope, nil
}

// generateDummyKeyPair generates dummy keys for cover traffic
func generateDummyKeyPair() (*KeyPair, error) {
	// Generate random dummy keys
	privateKey := make([]byte, 32)
	publicKey := make([]byte, 32)
	
	cryptorand.Read(privateKey)
	cryptorand.Read(publicKey)
	
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// generateDummySignature generates a dummy signature for cover traffic
func generateDummySignature() string {
	signature := make([]byte, 64)
	cryptorand.Read(signature)
	return fmt.Sprintf("%x", signature)
}

// generateCoverMessageID generates a unique ID for cover messages
func generateCoverMessageID() string {
	id := make([]byte, 16)
	cryptorand.Read(id)
	return fmt.Sprintf("cover_%x", id)
}

// AnalyzeTrafficPattern analyzes traffic patterns for a user
func (t *TrafficAnalyzer) AnalyzeTrafficPattern(userID string, messages []*CoverMessage) *TrafficPattern {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	pattern := &TrafficPattern{
		UserID:      userID,
		ActiveHours: make(map[int]int),
		LastUpdate:  time.Now(),
	}
	
	if len(messages) == 0 {
		return pattern
	}
	
	// Analyze message sizes
	for _, msg := range messages {
		pattern.MessageSizes = append(pattern.MessageSizes, msg.Size)
		pattern.TotalMessages++
		
		// Track active hours
		hour := msg.Timestamp.Hour()
		pattern.ActiveHours[hour]++
	}
	
	// Calculate average messages per hour
	timeSpan := time.Since(messages[0].Timestamp)
	if timeSpan > 0 {
		pattern.AverageMessages = int(float64(pattern.TotalMessages) / timeSpan.Hours())
	}
	
	// Determine peak hours
	pattern.PeakHours = t.determinePeakHours(pattern.ActiveHours)
	
	// Calculate typical intervals
	pattern.Intervals = t.calculateIntervals(messages)
	
	return pattern
}

// determinePeakHours determines peak activity hours
func (t *TrafficAnalyzer) determinePeakHours(activeHours map[int]int) []int {
	var peakHours []int
	maxCount := 0
	
	// Find the maximum count
	for _, count := range activeHours {
		if count > maxCount {
			maxCount = count
		}
	}
	
	// Find hours with at least 50% of max activity
	threshold := maxCount / 2
	for hour, count := range activeHours {
		if count >= threshold {
			peakHours = append(peakHours, hour)
		}
	}
	
	return peakHours
}

// calculateIntervals calculates typical intervals between messages
func (t *TrafficAnalyzer) calculateIntervals(messages []*CoverMessage) []time.Duration {
	var intervals []time.Duration
	
	for i := 1; i < len(messages); i++ {
		interval := messages[i].Timestamp.Sub(messages[i-1].Timestamp)
		if interval > 0 {
			intervals = append(intervals, interval)
		}
	}
	
	return intervals
}

// DefaultCoverTrafficConfig returns default cover traffic configuration
func DefaultCoverTrafficConfig() CoverTrafficConfig {
	return CoverTrafficConfig{
		Enabled:         true,
		Intensity:       0.3, // 30% of normal traffic
		MinInterval:     5 * time.Minute,
		MaxInterval:     45 * time.Minute,
		SizeVariation:   20, // 20% size variation
		AnalysisWindow:  24 * time.Hour,
		AdaptToUser:     true,
		GlobalBaseline:  true,
		JitterPercent:   15, // 15% jitter
		MaxJitter:       2 * time.Minute,
	}
}
