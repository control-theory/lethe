package handlers

// RememberInput is the input for the remember tool.
type RememberInput struct {
	Content   string   `json:"content"             jsonschema:"The fact, preference, or entity to remember. Be specific and self-contained."`
	Kind      string   `json:"kind,omitempty"      jsonschema:"The type of memory: person, place, project, organization, concept, event, artifact, preference, fact, or note. If omitted the server will infer from content."`
	Name      string   `json:"name,omitempty"      jsonschema:"A short canonical name for this memory. Used for deduplication."`
	Tags      []string `json:"tags,omitempty"      jsonschema:"Optional tags for categorization."`
	RelatedTo []string `json:"related_to,omitempty" jsonschema:"IDs or names of existing memories to link to."`
}

// RememberOutput is returned by the remember tool.
type RememberOutput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Action  string `json:"action"`
	Summary string `json:"summary,omitempty"`
}

// RecallInput is the input for the recall tool.
type RecallInput struct {
	Query        string `json:"query"                jsonschema:"Search query. Use specific keywords rather than broad questions."`
	Kind         string `json:"kind,omitempty"        jsonschema:"Filter by memory type."`
	Limit        int    `json:"limit,omitempty"       jsonschema:"Max results to return. Default 10, max 50."`
	IncludeFaded bool   `json:"include_faded,omitempty" jsonschema:"Include memories that have faded below the active threshold."`
}

// RecallOutput is returned by the recall tool.
type RecallOutput struct {
	Memories []NodeSummary `json:"memories"`
	Total    int           `json:"total"`
}

// NodeSummary is a compact representation of a node for tool responses.
type NodeSummary struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	Summary  string   `json:"summary,omitempty"`
	Strength float64  `json:"strength"`
	Tags     []string `json:"tags,omitempty"`
	Pinned   bool     `json:"pinned,omitempty"`
}

// ForgetInput is the input for the forget tool.
type ForgetInput struct {
	ID   string `json:"id"             jsonschema:"The ID of the memory to forget."`
	Hard bool   `json:"hard,omitempty" jsonschema:"If true, permanently delete. If false (default), soft-forget."`
}

// ForgetOutput is returned by the forget tool.
type ForgetOutput struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

// ListRecentInput is the input for the list_recent tool.
type ListRecentInput struct {
	Limit int    `json:"limit,omitempty" jsonschema:"Max results. Default 10, max 50."`
	Kind  string `json:"kind,omitempty"  jsonschema:"Filter by memory type."`
}

// StatsInput is the input for the stats tool.
type StatsInput struct{}

// StatsOutput is returned by the stats tool.
type StatsOutput struct {
	TotalNodes  int            `json:"total_nodes"`
	TotalEdges  int            `json:"total_edges"`
	ByKind      map[string]int `json:"by_kind"`
	PinnedCount int            `json:"pinned_count"`
	FadingCount int            `json:"fading_count"`
	RecentCount int            `json:"recent_count"`
}
