package e2e

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// NodeDirectory manages discovery and verification of mixnet nodes
type NodeDirectory struct {
	config     MixnetConfig
	db         *sql.DB
	nodes      map[string]*MixnetNode
	mutex      sync.RWMutex
	httpClient *http.Client
}

// PathSelector implements intelligent path selection algorithms
type PathSelector struct {
	config    MixnetConfig
	directory *NodeDirectory
}

// MixnetCoverTrafficGenerator generates realistic dummy traffic for mixnet
type MixnetCoverTrafficGenerator struct {
	config   MixnetConfig
	patterns []MixnetTrafficPattern
}

// MixnetTrafficPattern defines a cover traffic generation pattern for mixnet
type MixnetTrafficPattern struct {
	Name        string        `json:"name"`
	Frequency   time.Duration `json:"frequency"`
	MessageSize int           `json:"message_size"`
	Burst       bool          `json:"burst"`
	Variation   float64       `json:"variation"`
}

// NodeHealth tracks the health and performance of mixnet nodes
type NodeHealth struct {
	NodeID           string        `json:"node_id"`
	IsOnline         bool          `json:"is_online"`
	ResponseTime     time.Duration `json:"response_time"`
	SuccessRate      float64       `json:"success_rate"`
	LastCheck        time.Time     `json:"last_check"`
	ErrorCount       int           `json:"error_count"`
	ConsecutiveFails int           `json:"consecutive_fails"`
}

// NewNodeDirectory creates a new node directory
func NewNodeDirectory(config MixnetConfig) (*NodeDirectory, error) {
	// Initialize database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to create node directory database: %w", err)
	}

	// Create tables
	if err := createNodeDirectoryTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create node directory tables: %w", err)
	}

	directory := &NodeDirectory{
		config: config,
		db:     db,
		nodes:  make(map[string]*MixnetNode),
		httpClient: &http.Client{
			Timeout: config.Timeouts.NodeDiscovery,
		},
	}

	return directory, nil
}

// RefreshNodes refreshes the list of available mixnet nodes
func (nd *NodeDirectory) RefreshNodes(ctx context.Context) error {
	if nd.config.NodeDirectoryURL == "" {
		// Generate mock nodes for testing
		return nd.generateMockNodes()
	}

	// Fetch nodes from directory service
	req, err := http.NewRequestWithContext(ctx, "GET", nd.config.NodeDirectoryURL+"/nodes", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := nd.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("directory service returned status %d", resp.StatusCode)
	}

	var nodes []*MixnetNode
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return fmt.Errorf("failed to decode nodes: %w", err)
	}

	// Update local node list
	nd.mutex.Lock()
	defer nd.mutex.Unlock()

	for _, node := range nodes {
		if nd.validateNode(node) {
			nd.nodes[node.ID] = node
			nd.updateNodeInDB(node)
		}
	}

	return nil
}

// GetAvailableNodes returns all available and trusted nodes
func (nd *NodeDirectory) GetAvailableNodes() []*MixnetNode {
	nd.mutex.RLock()
	defer nd.mutex.RUnlock()

	var available []*MixnetNode
	for _, node := range nd.nodes {
		if node.Reputation >= nd.config.TrustThreshold {
			available = append(available, node)
		}
	}

	return available
}

// GetNodeByID returns a specific node by ID
func (nd *NodeDirectory) GetNodeByID(nodeID string) (*MixnetNode, bool) {
	nd.mutex.RLock()
	defer nd.mutex.RUnlock()

	node, exists := nd.nodes[nodeID]
	return node, exists
}

// UpdateNodeHealth updates the health status of a node
func (nd *NodeDirectory) UpdateNodeHealth(nodeID string, health *NodeHealth) error {
	nd.mutex.Lock()
	defer nd.mutex.Unlock()

	node, exists := nd.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	// Update reputation based on health
	if health.IsOnline && health.SuccessRate > 0.9 {
		node.Reputation = min(1.0, node.Reputation+0.01)
	} else if health.ConsecutiveFails > 5 {
		node.Reputation = max(0.0, node.Reputation-0.05)
	}

	node.LastSeen = time.Now()
	node.Latency = health.ResponseTime

	return nd.updateNodeInDB(node)
}

// Close closes the node directory
func (nd *NodeDirectory) Close() error {
	if nd.db != nil {
		return nd.db.Close()
	}
	return nil
}

// Helper methods for NodeDirectory

func (nd *NodeDirectory) validateNode(node *MixnetNode) bool {
	// Basic validation
	if node.ID == "" || node.Address == "" || len(node.PublicKey) == 0 {
		return false
	}

	// Check reputation threshold
	if node.Reputation < nd.config.TrustThreshold {
		return false
	}

	// Check if node was seen recently
	if time.Since(node.LastSeen) > 24*time.Hour {
		return false
	}

	return true
}

func (nd *NodeDirectory) generateMockNodes() error {
	// Generate mock nodes for testing
	regions := []string{"us-east", "us-west", "eu-west", "eu-central", "asia-pacific"}

	nd.mutex.Lock()
	defer nd.mutex.Unlock()

	for i := 0; i < 20; i++ {
		nodeID := fmt.Sprintf("node_%d", i)
		publicKey := make([]byte, 32)
		rand.Read(publicKey)

		node := &MixnetNode{
			ID:           nodeID,
			Address:      fmt.Sprintf("mixnode-%d.example.com:9001", i),
			PublicKey:    publicKey,
			Capabilities: []string{"onion-routing", "cover-traffic"},
			Reputation:   0.8 + (float64(i%3) * 0.1), // Vary reputation
			Latency:      time.Duration(50+i*10) * time.Millisecond,
			LastSeen:     time.Now(),
			Region:       regions[i%len(regions)],
			Metadata:     map[string]string{"version": "1.0", "operator": fmt.Sprintf("operator_%d", i%5)},
		}

		nd.nodes[nodeID] = node
		nd.updateNodeInDB(node)
	}

	return nil
}

func (nd *NodeDirectory) updateNodeInDB(node *MixnetNode) error {
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return err
	}

	_, err = nd.db.Exec(`
		INSERT OR REPLACE INTO mixnet_nodes 
		(id, address, public_key, reputation, last_seen, region, node_data) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		node.ID, node.Address, node.PublicKey, node.Reputation,
		node.LastSeen, node.Region, string(nodeJSON))

	return err
}

// NewPathSelector creates a new path selector
func NewPathSelector(config MixnetConfig) *PathSelector {
	return &PathSelector{
		config: config,
	}
}

// SetDirectory sets the node directory for the path selector
func (ps *PathSelector) SetDirectory(directory *NodeDirectory) {
	ps.directory = directory
}

// SelectPath selects an optimal path through the mixnet
func (ps *PathSelector) SelectPath(ctx context.Context, minHops int) ([]*MixnetNode, error) {
	if ps.directory == nil {
		return nil, fmt.Errorf("node directory not set")
	}

	availableNodes := ps.directory.GetAvailableNodes()
	if len(availableNodes) < minHops {
		return nil, fmt.Errorf("insufficient nodes available: need %d, have %d", minHops, len(availableNodes))
	}

	switch ps.config.PathSelectionMode {
	case "random":
		return ps.selectRandomPath(availableNodes, minHops)
	case "latency":
		return ps.selectLatencyOptimizedPath(availableNodes, minHops)
	case "security":
		return ps.selectSecurityOptimizedPath(availableNodes, minHops)
	default:
		return ps.selectRandomPath(availableNodes, minHops)
	}
}

func (ps *PathSelector) selectRandomPath(nodes []*MixnetNode, minHops int) ([]*MixnetNode, error) {
	if len(nodes) < minHops {
		return nil, fmt.Errorf("insufficient nodes for random path")
	}

	// Shuffle nodes
	shuffled := make([]*MixnetNode, len(nodes))
	copy(shuffled, nodes)

	for i := range shuffled {
		j := i + int(mathrand.Int31n(int32(len(shuffled)-i)))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:minHops], nil
}

func (ps *PathSelector) selectLatencyOptimizedPath(nodes []*MixnetNode, minHops int) ([]*MixnetNode, error) {
	// Sort by latency
	sorted := make([]*MixnetNode, len(nodes))
	copy(sorted, nodes)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Latency < sorted[j].Latency
	})

	if len(sorted) < minHops {
		return nil, fmt.Errorf("insufficient nodes for latency-optimized path")
	}

	return sorted[:minHops], nil
}

func (ps *PathSelector) selectSecurityOptimizedPath(nodes []*MixnetNode, minHops int) ([]*MixnetNode, error) {
	// Select nodes from different regions and operators for diversity
	regionMap := make(map[string][]*MixnetNode)
	operatorMap := make(map[string][]*MixnetNode)

	for _, node := range nodes {
		regionMap[node.Region] = append(regionMap[node.Region], node)
		if operator, exists := node.Metadata["operator"]; exists {
			operatorMap[operator] = append(operatorMap[operator], node)
		}
	}

	var selected []*MixnetNode
	usedRegions := make(map[string]bool)
	usedOperators := make(map[string]bool)

	// Try to select nodes from different regions and operators
	for len(selected) < minHops && len(nodes) > 0 {
		var bestNode *MixnetNode
		var bestScore float64

		for _, node := range nodes {
			if ps.isNodeSelected(node, selected) {
				continue
			}

			score := node.Reputation

			// Bonus for region diversity
			if !usedRegions[node.Region] {
				score += 0.2
			}

			// Bonus for operator diversity
			if operator, exists := node.Metadata["operator"]; exists && !usedOperators[operator] {
				score += 0.1
			}

			if bestNode == nil || score > bestScore {
				bestNode = node
				bestScore = score
			}
		}

		if bestNode == nil {
			break
		}

		selected = append(selected, bestNode)
		usedRegions[bestNode.Region] = true
		if operator, exists := bestNode.Metadata["operator"]; exists {
			usedOperators[operator] = true
		}
	}

	if len(selected) < minHops {
		return nil, fmt.Errorf("insufficient diverse nodes for security-optimized path")
	}

	return selected, nil
}

func (ps *PathSelector) isNodeSelected(node *MixnetNode, selected []*MixnetNode) bool {
	for _, s := range selected {
		if s.ID == node.ID {
			return true
		}
	}
	return false
}

// NewMixnetCoverTrafficGenerator creates a new mixnet cover traffic generator
func NewMixnetCoverTrafficGenerator(config MixnetConfig) *MixnetCoverTrafficGenerator {
	patterns := []MixnetTrafficPattern{
		{
			Name:        "steady_background",
			Frequency:   30 * time.Second,
			MessageSize: 1024,
			Burst:       false,
			Variation:   0.2,
		},
		{
			Name:        "periodic_burst",
			Frequency:   5 * time.Minute,
			MessageSize: 4096,
			Burst:       true,
			Variation:   0.5,
		},
		{
			Name:        "random_small",
			Frequency:   10 * time.Second,
			MessageSize: 256,
			Burst:       false,
			Variation:   0.8,
		},
	}

	return &MixnetCoverTrafficGenerator{
		config:   config,
		patterns: patterns,
	}
}

// GenerateTraffic generates cover traffic according to configured patterns
func (ctg *MixnetCoverTrafficGenerator) GenerateTraffic(ctx context.Context) error {
	// Select a random pattern
	pattern := ctg.patterns[int(mathrand.Int31n(int32(len(ctg.patterns))))]

	// Generate message with size variation
	baseSize := pattern.MessageSize
	variation := int(float64(baseSize) * pattern.Variation)
	messageSize := baseSize + int(mathrand.Int31n(int32(variation*2))) - variation

	// Create dummy message
	dummyMessage := make([]byte, messageSize)
	rand.Read(dummyMessage)

	// Log cover traffic generation (in production, route through mixnet)
	// This is a placeholder - in real implementation, route through mixnet
	_ = dummyMessage

	return nil
}

// Helper functions

func createNodeDirectoryTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS mixnet_nodes (
		id TEXT PRIMARY KEY,
		address TEXT NOT NULL,
		public_key BLOB NOT NULL,
		reputation REAL NOT NULL,
		last_seen DATETIME NOT NULL,
		region TEXT,
		node_data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_mixnet_nodes_reputation ON mixnet_nodes(reputation);
	CREATE INDEX IF NOT EXISTS idx_mixnet_nodes_region ON mixnet_nodes(region);
	CREATE INDEX IF NOT EXISTS idx_mixnet_nodes_last_seen ON mixnet_nodes(last_seen);
	`

	_, err := db.Exec(schema)
	return err
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
