package graph

import "context"

// SubgraphResult holds the nodes and edges of a subgraph extraction.
type SubgraphResult struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Subgraph performs a BFS from startID up to the given depth,
// returning all reachable nodes and the edges between them.
func (s *Store) Subgraph(ctx context.Context, startID string, depth int) (*SubgraphResult, error) {
	if depth <= 0 {
		depth = 1
	}
	if depth > 3 {
		depth = 3
	}

	visited := make(map[string]bool)
	var allEdges []Edge
	edgeSeen := make(map[string]bool)

	// BFS frontier
	frontier := []string{startID}
	visited[startID] = true

	for d := 0; d < depth && len(frontier) > 0; d++ {
		var nextFrontier []string
		for _, nodeID := range frontier {
			edges, err := s.EdgesForNode(ctx, nodeID)
			if err != nil {
				return nil, err
			}
			for _, e := range edges {
				if !edgeSeen[e.ID] {
					edgeSeen[e.ID] = true
					allEdges = append(allEdges, e)
				}
				// Discover the neighbor
				neighbor := e.DstID
				if neighbor == nodeID {
					neighbor = e.SrcID
				}
				if !visited[neighbor] {
					visited[neighbor] = true
					nextFrontier = append(nextFrontier, neighbor)
				}
			}
		}
		frontier = nextFrontier
	}

	// Collect all visited nodes
	nodes := make([]Node, 0, len(visited))
	for id := range visited {
		n, err := s.GetNode(ctx, id)
		if err != nil {
			continue
		}
		nodes = append(nodes, *n)
	}
	if allEdges == nil {
		allEdges = []Edge{}
	}

	return &SubgraphResult{Nodes: nodes, Edges: allEdges}, nil
}
