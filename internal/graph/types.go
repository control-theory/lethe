package graph

// NodeKind enumerates the typed node kinds.
type NodeKind string

const (
	KindPerson       NodeKind = "person"
	KindPlace        NodeKind = "place"
	KindProject      NodeKind = "project"
	KindOrganization NodeKind = "organization"
	KindConcept      NodeKind = "concept"
	KindEvent        NodeKind = "event"
	KindArtifact     NodeKind = "artifact"
	KindPreference   NodeKind = "preference"
	KindFact         NodeKind = "fact"
	KindNote         NodeKind = "note"
	KindCustom       NodeKind = "custom"
)

var validKinds = map[NodeKind]bool{
	KindPerson: true, KindPlace: true, KindProject: true,
	KindOrganization: true, KindConcept: true, KindEvent: true,
	KindArtifact: true, KindPreference: true, KindFact: true,
	KindNote: true, KindCustom: true,
}

// ValidNodeKind returns true if s is a recognized node kind.
func ValidNodeKind(s string) bool {
	return validKinds[NodeKind(s)]
}

// RelationType enumerates typed edge relations.
type RelationType string

const (
	RelKnows          RelationType = "knows"
	RelWorksOn        RelationType = "works_on"
	RelWorksAt        RelationType = "works_at"
	RelLocatedIn      RelationType = "located_in"
	RelRelatedTo      RelationType = "related_to"
	RelPartOf         RelationType = "part_of"
	RelCaused         RelationType = "caused"
	RelMentioned      RelationType = "mentioned"
	RelPrefers        RelationType = "prefers"
	RelOpposes        RelationType = "opposes"
	RelParticipatedIn RelationType = "participated_in"
	RelCreated        RelationType = "created"
	RelDependsOn      RelationType = "depends_on"
	RelCustom         RelationType = "custom"
)

// Node is the in-memory representation of a graph node.
type Node struct {
	ID             string         `json:"id"`
	Kind           NodeKind       `json:"kind"`
	CustomType     string         `json:"custom_type,omitempty"`
	Name           string         `json:"name"`
	Summary        string         `json:"summary,omitempty"`
	Body           string         `json:"body,omitempty"`
	Tags           []string       `json:"tags"`
	Attributes     map[string]any `json:"attributes"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at"`
	LastAccessedAt int64          `json:"last_accessed_at"`
	AccessCount    int            `json:"access_count"`
	Strength       float64        `json:"strength"`
	Pinned         bool           `json:"pinned"`
}

// Edge is the in-memory representation of a graph edge.
type Edge struct {
	ID             string         `json:"id"`
	SrcID          string         `json:"src_id"`
	DstID          string         `json:"dst_id"`
	Relation       RelationType   `json:"relation"`
	CustomRelation string         `json:"custom_relation,omitempty"`
	Weight         float64        `json:"weight"`
	Attributes     map[string]any `json:"attributes"`
	CreatedAt      int64          `json:"created_at"`
	LastAccessedAt int64          `json:"last_accessed_at"`
	AccessCount    int            `json:"access_count"`
}

// MemoryEvent represents an audit log entry.
type MemoryEvent struct {
	ID       int64  `json:"id"`
	TS       int64  `json:"ts"`
	Op       string `json:"op"`
	TargetID string `json:"target_id,omitempty"`
	Actor    string `json:"actor"`
	Payload  string `json:"payload,omitempty"`
}

// NodePatch represents a partial update to a node.
type NodePatch struct {
	Name       *string         `json:"name,omitempty"`
	Summary    *string         `json:"summary,omitempty"`
	Body       *string         `json:"body,omitempty"`
	Tags       *[]string       `json:"tags,omitempty"`
	Attributes *map[string]any `json:"attributes,omitempty"`
}

// SearchOpts configures a search query.
type SearchOpts struct {
	Kind         *NodeKind
	Limit        int
	IncludeFaded bool
}

// GraphStats holds aggregate statistics about the graph.
type GraphStats struct {
	TotalNodes  int            `json:"total_nodes"`
	TotalEdges  int            `json:"total_edges"`
	ByKind      map[string]int `json:"by_kind"`
	PinnedCount int            `json:"pinned_count"`
	FadingCount int            `json:"fading_count"`
	RecentCount int            `json:"recent_count"`
}

// DedupResult is returned by duplicate checking.
type DedupResult struct {
	IsDuplicate bool   `json:"is_duplicate"`
	ExistingID  string `json:"existing_id,omitempty"`
	Confidence  string `json:"confidence"` // "exact", "likely", "none"
}
