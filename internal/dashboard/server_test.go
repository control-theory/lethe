package dashboard

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/control-theory/lethe/internal/graph"
)

func testDashboard(t *testing.T) (*Server, string) {
	t.Helper()
	db, err := graph.OpenMemoryDB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	store := graph.NewStore(db)
	dash := New(store)
	port, err := dash.Start(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { dash.Stop(nil) })

	base := "http://127.0.0.1:" + itoa(port)
	return dash, base
}

func itoa(i int) string {
	return json.Number(json.Number(string(rune('0'+i/10000%10)) +
		string(rune('0'+i/1000%10)) +
		string(rune('0'+i/100%10)) +
		string(rune('0'+i/10%10)) +
		string(rune('0'+i%10)))).String()
}

func TestDashboardIndex(t *testing.T) {
	_, base := testDashboard(t)
	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("content-type = %q", ct)
	}
}

func TestDashboardStatsAPI(t *testing.T) {
	_, base := testDashboard(t)
	resp, err := http.Get(base + "/api/stats")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var stats graph.GraphStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}
	if stats.TotalNodes != 0 {
		t.Errorf("total_nodes = %d, want 0", stats.TotalNodes)
	}
}

func TestDashboardNodesAPI(t *testing.T) {
	_, base := testDashboard(t)
	resp, err := http.Get(base + "/api/nodes")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var nodes []graph.Node
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 0 {
		t.Errorf("got %d nodes, want 0", len(nodes))
	}
}

func TestDashboardLocalhostOnly(t *testing.T) {
	_, base := testDashboard(t)
	// Request with a non-localhost Host header
	req, _ := http.NewRequest("GET", base+"/api/stats", nil)
	req.Host = "evil.com"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Errorf("non-localhost request: status = %d, want 403", resp.StatusCode)
	}
}
