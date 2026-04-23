package graph

import (
	"context"
	"database/sql"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	db, err := OpenMemoryDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStore(db)
}

func TestCreateAndGetNode(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindPerson, Name: "Eric", Summary: "CTO of ControlTheory"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatalf("CreateNode: %v", err)
	}
	if n.ID == "" {
		t.Fatal("expected ID to be set")
	}
	if n.Strength != 1.0 {
		t.Errorf("strength = %v, want 1.0", n.Strength)
	}

	got, err := s.GetNode(ctx, n.ID)
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if got.Name != "Eric" {
		t.Errorf("name = %q, want %q", got.Name, "Eric")
	}
	if got.Kind != KindPerson {
		t.Errorf("kind = %q, want %q", got.Kind, KindPerson)
	}
	if got.Summary != "CTO of ControlTheory" {
		t.Errorf("summary = %q, want %q", got.Summary, "CTO of ControlTheory")
	}
}

func TestGetNodeByName(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindPerson, Name: "Eric"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	// Case-insensitive lookup
	got, err := s.GetNodeByName(ctx, KindPerson, "eric")
	if err != nil {
		t.Fatalf("GetNodeByName: %v", err)
	}
	if got.ID != n.ID {
		t.Errorf("got ID %q, want %q", got.ID, n.ID)
	}
}

func TestUpdateNode(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "test fact", Summary: "original"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	newSummary := "updated summary"
	if err := s.UpdateNode(ctx, n.ID, NodePatch{Summary: &newSummary}); err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}

	got, err := s.GetNode(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != "updated summary" {
		t.Errorf("summary = %q, want %q", got.Summary, "updated summary")
	}
	if got.Name != "test fact" {
		t.Errorf("name changed unexpectedly: %q", got.Name)
	}
}

func TestSoftForget(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "forgettable"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	if err := s.SoftForget(ctx, n.ID); err != nil {
		t.Fatalf("SoftForget: %v", err)
	}

	got, err := s.GetNode(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Strength != 0 {
		t.Errorf("strength = %v, want 0", got.Strength)
	}

	// Should be excluded from default ListRecent
	recent, err := s.ListRecent(ctx, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range recent {
		if r.ID == n.ID {
			t.Error("soft-forgotten node should not appear in ListRecent")
		}
	}
}

func TestDeleteNode(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "to delete"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteNode(ctx, n.ID); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}

	_, err := s.GetNode(ctx, n.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestDeleteNodeCascadesEdges(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n1 := &Node{Kind: KindPerson, Name: "Alice"}
	n2 := &Node{Kind: KindProject, Name: "Project X"}
	if err := s.CreateNode(ctx, n1); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateNode(ctx, n2); err != nil {
		t.Fatal(err)
	}

	e := &Edge{SrcID: n1.ID, DstID: n2.ID, Relation: RelWorksOn}
	if err := s.CreateEdge(ctx, e); err != nil {
		t.Fatal(err)
	}

	// Delete n1, edge should cascade
	if err := s.DeleteNode(ctx, n1.ID); err != nil {
		t.Fatal(err)
	}

	edges, err := s.EdgesForNode(ctx, n2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 0 {
		t.Errorf("expected 0 edges after cascade, got %d", len(edges))
	}
}

func TestListRecent(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	for _, name := range []string{"a", "b", "c"} {
		n := &Node{Kind: KindFact, Name: name}
		if err := s.CreateNode(ctx, n); err != nil {
			t.Fatal(err)
		}
	}

	recent, err := s.ListRecent(ctx, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 2 {
		t.Errorf("got %d results, want 2", len(recent))
	}

	// Filter by kind
	kind := KindPerson
	recent, err = s.ListRecent(ctx, 10, &kind)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 0 {
		t.Errorf("got %d person results, want 0", len(recent))
	}
}

func TestSearchFTS(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "helicopter pilot license", Summary: "Eric has a helicopter pilot license"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}
	n2 := &Node{Kind: KindPreference, Name: "prefers Go", Summary: "Eric prefers Go over Python"}
	if err := s.CreateNode(ctx, n2); err != nil {
		t.Fatal(err)
	}

	results, err := s.Search(ctx, "helicopter", SearchOpts{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Name != "helicopter pilot license" {
		t.Errorf("got %q, want %q", results[0].Name, "helicopter pilot license")
	}
}

func TestSearchExcludesFaded(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "faded memory", Summary: "this should be hidden"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}
	if err := s.SoftForget(ctx, n.ID); err != nil {
		t.Fatal(err)
	}

	results, err := s.Search(ctx, "faded memory", SearchOpts{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("faded node should not appear in default search, got %d results", len(results))
	}

	// With IncludeFaded
	results, err = s.Search(ctx, "faded memory", SearchOpts{Limit: 10, IncludeFaded: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("faded node should appear with IncludeFaded, got %d results", len(results))
	}
}

func TestDedupExactMatch(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindPerson, Name: "Eric"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	result, err := s.CheckDuplicate(ctx, KindPerson, "Eric", "")
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsDuplicate {
		t.Error("expected duplicate")
	}
	if result.ExistingID != n.ID {
		t.Errorf("existing ID = %q, want %q", result.ExistingID, n.ID)
	}
	if result.Confidence != "exact" {
		t.Errorf("confidence = %q, want %q", result.Confidence, "exact")
	}
}

func TestDedupNoMatch(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	result, err := s.CheckDuplicate(ctx, KindPerson, "Alice", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.IsDuplicate {
		t.Error("expected no duplicate")
	}
}

func TestReinforce(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "reinforceable"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	if err := s.Reinforce(ctx, n.ID, 1.0, 10.0); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetNode(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Strength != 2.0 {
		t.Errorf("strength = %v, want 2.0", got.Strength)
	}
	if got.AccessCount != 2 {
		t.Errorf("access_count = %d, want 2", got.AccessCount)
	}
}

func TestReinforceCap(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "capped", Strength: 9.5}
	// Override default strength
	n.Strength = 9.5
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	if err := s.Reinforce(ctx, n.ID, 1.0, 10.0); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetNode(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Strength != 10.0 {
		t.Errorf("strength = %v, want 10.0 (capped)", got.Strength)
	}
}

func TestEdgeCRUD(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n1 := &Node{Kind: KindPerson, Name: "Alice"}
	n2 := &Node{Kind: KindProject, Name: "Project X"}
	if err := s.CreateNode(ctx, n1); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateNode(ctx, n2); err != nil {
		t.Fatal(err)
	}

	e := &Edge{SrcID: n1.ID, DstID: n2.ID, Relation: RelWorksOn}
	if err := s.CreateEdge(ctx, e); err != nil {
		t.Fatal(err)
	}
	if e.ID == "" {
		t.Fatal("edge ID not set")
	}

	got, err := s.GetEdge(ctx, e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Relation != RelWorksOn {
		t.Errorf("relation = %q, want %q", got.Relation, RelWorksOn)
	}

	edges, err := s.EdgesForNode(ctx, n1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 1 {
		t.Errorf("got %d edges, want 1", len(edges))
	}

	if err := s.DeleteEdge(ctx, e.ID); err != nil {
		t.Fatal(err)
	}
	_, err = s.GetEdge(ctx, e.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestStats(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n1 := &Node{Kind: KindPerson, Name: "Alice"}
	n2 := &Node{Kind: KindPerson, Name: "Bob", Pinned: true}
	n3 := &Node{Kind: KindProject, Name: "Proj"}
	for _, n := range []*Node{n1, n2, n3} {
		if err := s.CreateNode(ctx, n); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := s.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalNodes != 3 {
		t.Errorf("total_nodes = %d, want 3", stats.TotalNodes)
	}
	if stats.ByKind["person"] != 2 {
		t.Errorf("person count = %d, want 2", stats.ByKind["person"])
	}
	if stats.ByKind["project"] != 1 {
		t.Errorf("project count = %d, want 1", stats.ByKind["project"])
	}
	if stats.RecentCount != 3 {
		t.Errorf("recent = %d, want 3", stats.RecentCount)
	}
}

func TestSetPinned(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "pinnable"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	// Pin it
	if err := s.SetPinned(ctx, n.ID, true); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetNode(ctx, n.ID)
	if !got.Pinned {
		t.Error("expected pinned=true")
	}

	// Unpin it
	if err := s.SetPinned(ctx, n.ID, false); err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetNode(ctx, n.ID)
	if got.Pinned {
		t.Error("expected pinned=false")
	}
}

func TestMaterializeStrength(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	// Create a node with artificially old last_accessed_at
	n := &Node{Kind: KindFact, Name: "old memory"}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	// Manually backdate the node's last_accessed_at by 30 days
	thirtyDaysMs := int64(30 * 24 * 60 * 60 * 1000)
	_, err := s.db.ExecContext(ctx,
		`UPDATE nodes SET last_accessed_at = last_accessed_at - ? WHERE id = ?`,
		thirtyDaysMs, n.ID)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := s.MaterializeStrength(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Errorf("updated = %d, want 1", updated)
	}

	// Read directly from DB (bypass decay computation) to check materialized value
	var strength float64
	s.db.QueryRowContext(ctx, `SELECT strength FROM nodes WHERE id = ?`, n.ID).Scan(&strength)
	if strength >= 0.6 || strength <= 0.4 {
		t.Errorf("materialized strength = %v, want ~0.5", strength)
	}
}

func TestPinnedNodeDoesNotDecay(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n := &Node{Kind: KindFact, Name: "core fact", Pinned: true}
	if err := s.CreateNode(ctx, n); err != nil {
		t.Fatal(err)
	}

	// Backdate last_accessed_at
	thirtyDaysMs := int64(30 * 24 * 60 * 60 * 1000)
	_, err := s.db.ExecContext(ctx,
		`UPDATE nodes SET last_accessed_at = last_accessed_at - ? WHERE id = ?`,
		thirtyDaysMs, n.ID)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.GetNode(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Pinned node should still have strength 1.0
	if got.Strength != 1.0 {
		t.Errorf("pinned node strength = %v, want 1.0", got.Strength)
	}

	// MaterializeStrength should not update pinned nodes
	updated, err := s.MaterializeStrength(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 0 {
		t.Errorf("pinned node should not be updated, got %d updates", updated)
	}
}

func TestSubgraph(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	// Create a small graph: A -> B -> C
	a := &Node{Kind: KindPerson, Name: "A"}
	b := &Node{Kind: KindPerson, Name: "B"}
	c := &Node{Kind: KindPerson, Name: "C"}
	for _, n := range []*Node{a, b, c} {
		if err := s.CreateNode(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CreateEdge(ctx, &Edge{SrcID: a.ID, DstID: b.ID, Relation: RelKnows}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateEdge(ctx, &Edge{SrcID: b.ID, DstID: c.ID, Relation: RelKnows}); err != nil {
		t.Fatal(err)
	}

	// Depth 1 from A should get A and B
	sg, err := s.Subgraph(ctx, a.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(sg.Nodes) != 2 {
		t.Errorf("depth=1: got %d nodes, want 2", len(sg.Nodes))
	}
	if len(sg.Edges) != 1 {
		t.Errorf("depth=1: got %d edges, want 1", len(sg.Edges))
	}

	// Depth 2 from A should get A, B, and C
	sg, err = s.Subgraph(ctx, a.ID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(sg.Nodes) != 3 {
		t.Errorf("depth=2: got %d nodes, want 3", len(sg.Nodes))
	}
	if len(sg.Edges) != 2 {
		t.Errorf("depth=2: got %d edges, want 2", len(sg.Edges))
	}
}

func TestListPinned(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	n1 := &Node{Kind: KindFact, Name: "pinned1", Pinned: true}
	n2 := &Node{Kind: KindFact, Name: "not pinned"}
	n3 := &Node{Kind: KindFact, Name: "pinned2", Pinned: true}
	for _, n := range []*Node{n1, n2, n3} {
		if err := s.CreateNode(ctx, n); err != nil {
			t.Fatal(err)
		}
	}

	pinned, err := s.ListPinned(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(pinned) != 2 {
		t.Errorf("got %d pinned, want 2", len(pinned))
	}
}
