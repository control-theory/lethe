package mcpserver

import (
	"github.com/control-theory/lethe/internal/mcp/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerTools(s *mcp.Server, h *handlers.Handlers) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "remember",
		Description: `Persist a fact, preference, or entity so it can be recalled in future conversations. ` +
			`Use this whenever the user explicitly asks you to remember something, OR when you notice a stable, ` +
			`reusable piece of information about the user, their projects, their people, or their preferences ` +
			`that would be useful later. Do NOT call this for ephemeral things like "I'm going to the store now." ` +
			`DO call it for things like "I'm vegetarian," "my daughter's name is Ellie," "I prefer pytest over unittest." ` +
			`The server detects duplicates—if you remember something that already exists, it will be reinforced rather than duplicated.`,
	}, h.Remember)

	mcp.AddTool(s, &mcp.Tool{
		Name: "recall",
		Description: `Search stored memories by keyword or phrase. Returns the most relevant memories ranked by ` +
			`text match and recency. Use specific keywords rather than broad questions—"helicopter license" is ` +
			`better than "what do I like." Memories that are actively used are reinforced and stay strong; unused ` +
			`memories naturally fade over time. Results include each memory's id, name, kind, summary, and current strength.`,
	}, h.Recall)

	mcp.AddTool(s, &mcp.Tool{
		Name: "forget",
		Description: `Remove a memory. By default this is a soft forget—the memory's strength drops to zero and it ` +
			`fades from results but remains in the database and can be recovered. Pass hard=true to permanently and ` +
			`irreversibly delete the memory and all its connections. Use soft forget for "I don't care about this ` +
			`anymore" and hard forget only when the user explicitly asks you to delete something.`,
	}, h.Forget)

	mcp.AddTool(s, &mcp.Tool{
		Name: "list_recent",
		Description: `List the most recently created or reinforced memories. Use this to see what has been ` +
			`remembered lately. Optionally filter by kind (person, project, preference, etc). Good for getting ` +
			`a quick overview of recent memory activity.`,
	}, h.ListRecent)

	mcp.AddTool(s, &mcp.Tool{
		Name: "stats",
		Description: `Get a summary of the memory system's current state: total memories and connections, ` +
			`counts by type, how many are pinned, how many are fading, and recent activity. Call this when the ` +
			`user asks "what do you know about me?" or when you want to understand the scope of stored knowledge ` +
			`before doing a more targeted recall.`,
	}, h.Stats)

	mcp.AddTool(s, &mcp.Tool{
		Name: "pin",
		Description: `Pin a memory so it never decays. Use this for core facts about the user that should ` +
			`always be available—their name, role, key preferences, important people. Pinned memories always ` +
			`appear at full strength in search results and in the active context resource.`,
	}, h.Pin)

	mcp.AddTool(s, &mcp.Tool{
		Name: "unpin",
		Description: `Unpin a memory so it decays normally again. Use this when a previously-pinned fact ` +
			`is no longer important enough to keep permanently active.`,
	}, h.Unpin)

	mcp.AddTool(s, &mcp.Tool{
		Name: "link",
		Description: `Create a named relation between two memories. Use this to connect people to projects, ` +
			`concepts to each other, or any two memories that have a meaningful relationship. Available relation ` +
			`types: knows, works_on, works_at, located_in, related_to, part_of, caused, mentioned, prefers, ` +
			`opposes, participated_in, created, depends_on. Default is "related_to."`,
	}, h.Link)

	mcp.AddTool(s, &mcp.Tool{
		Name: "unlink",
		Description: `Remove a relation between two memories by its edge ID. Use get_context to find edge IDs.`,
	}, h.Unlink)

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_context",
		Description: `Fetch the subgraph around a memory—the memory itself plus all memories connected to it ` +
			`within 1 or 2 hops. Use this to pull rich context before answering a question about a topic. ` +
			`Returns nodes and edges as structured data. More expensive than recall for a single keyword, but ` +
			`gives you the full picture of how a topic connects to other knowledge.`,
	}, h.GetContext)

	mcp.AddTool(s, &mcp.Tool{
		Name: "update_memory",
		Description: `Edit an existing memory's name, summary, body, or tags. Use this when the user corrects ` +
			`a stored fact or when you need to update a memory with new information. Pass only the fields you ` +
			`want to change; omitted fields are left untouched.`,
	}, h.UpdateMemory)
}
