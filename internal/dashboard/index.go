package dashboard

import "net/http"

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Lethe — Memory Dashboard</title>
<script src="https://cdnjs.cloudflare.com/ajax/libs/cytoscape/3.30.4/cytoscape.min.js"></script>
<style>
  :root { --bg: #0d1117; --surface: #161b22; --border: #30363d; --text: #e6edf3; --muted: #8b949e; --accent: #58a6ff; --green: #3fb950; --red: #f85149; --orange: #d29922; }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif; background: var(--bg); color: var(--text); }
  a { color: var(--accent); text-decoration: none; }
  .container { max-width: 1400px; margin: 0 auto; padding: 16px; }
  header { display: flex; align-items: center; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--border); margin-bottom: 16px; }
  header h1 { font-size: 20px; font-weight: 600; }
  header h1 span { color: var(--muted); font-weight: 400; }
  .stats { display: flex; gap: 16px; font-size: 13px; color: var(--muted); }
  .stats .stat { background: var(--surface); padding: 4px 10px; border-radius: 6px; border: 1px solid var(--border); }
  .stats .stat b { color: var(--text); }
  .tabs { display: flex; gap: 0; border-bottom: 1px solid var(--border); margin-bottom: 16px; }
  .tab { padding: 8px 16px; cursor: pointer; color: var(--muted); font-size: 14px; border-bottom: 2px solid transparent; }
  .tab.active { color: var(--text); border-bottom-color: var(--accent); }
  .tab:hover { color: var(--text); }
  .panel { display: none; }
  .panel.active { display: block; }
  .controls { display: flex; gap: 8px; margin-bottom: 12px; align-items: center; }
  input, select { background: var(--surface); border: 1px solid var(--border); color: var(--text); padding: 6px 10px; border-radius: 6px; font-size: 13px; }
  input:focus, select:focus { outline: none; border-color: var(--accent); }
  input[type="text"] { width: 300px; }
  button { background: var(--surface); border: 1px solid var(--border); color: var(--text); padding: 6px 12px; border-radius: 6px; cursor: pointer; font-size: 13px; }
  button:hover { background: var(--border); }
  button.danger { border-color: var(--red); color: var(--red); }
  button.danger:hover { background: rgba(248,81,73,0.1); }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { text-align: left; padding: 8px; border-bottom: 1px solid var(--border); color: var(--muted); font-weight: 500; }
  td { padding: 8px; border-bottom: 1px solid var(--border); }
  tr:hover td { background: var(--surface); }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 11px; border: 1px solid var(--border); }
  .badge.pinned { border-color: var(--green); color: var(--green); }
  .badge.fading { border-color: var(--orange); color: var(--orange); }
  .strength-bar { width: 60px; height: 6px; background: var(--border); border-radius: 3px; display: inline-block; vertical-align: middle; }
  .strength-bar .fill { height: 100%; border-radius: 3px; background: var(--accent); }
  #graph-container { width: 100%; height: 600px; background: var(--surface); border: 1px solid var(--border); border-radius: 8px; }
  .detail { position: fixed; right: 0; top: 0; width: 400px; height: 100vh; background: var(--surface); border-left: 1px solid var(--border); padding: 16px; overflow-y: auto; z-index: 10; display: none; }
  .detail.open { display: block; }
  .detail h2 { font-size: 16px; margin-bottom: 12px; }
  .detail .close { float: right; cursor: pointer; color: var(--muted); font-size: 18px; }
  .detail label { display: block; font-size: 12px; color: var(--muted); margin-top: 10px; }
  .detail .val { font-size: 14px; margin-top: 2px; }
  .timeline { list-style: none; }
  .timeline li { padding: 8px 0; border-bottom: 1px solid var(--border); font-size: 13px; }
  .timeline .op { font-weight: 600; color: var(--accent); }
  .timeline .time { color: var(--muted); font-size: 12px; }
</style>
</head>
<body>
<div class="container">
  <header>
    <h1>Lethe <span>memory dashboard</span></h1>
    <div class="stats" id="stats-bar"></div>
  </header>

  <div class="tabs">
    <div class="tab active" data-tab="list">List</div>
    <div class="tab" data-tab="graph">Graph</div>
    <div class="tab" data-tab="timeline">Timeline</div>
    <div class="tab" data-tab="fading">Fading</div>
  </div>

  <div id="panel-list" class="panel active">
    <div class="controls">
      <input type="text" id="search-input" placeholder="Search memories...">
      <select id="kind-filter"><option value="">All types</option></select>
      <label style="font-size:13px;color:var(--muted);display:flex;align-items:center;gap:4px;">
        <input type="checkbox" id="show-faded"> Show faded
      </label>
    </div>
    <table>
      <thead><tr><th>Name</th><th>Type</th><th>Summary</th><th>Strength</th><th>Status</th></tr></thead>
      <tbody id="nodes-tbody"></tbody>
    </table>
  </div>

  <div id="panel-graph" class="panel">
    <div id="graph-container"></div>
  </div>

  <div id="panel-timeline" class="panel">
    <ul class="timeline" id="timeline-list"></ul>
  </div>

  <div id="panel-fading" class="panel">
    <p style="color:var(--muted);margin-bottom:12px;font-size:13px;">Memories fading below the active threshold. Pin or recall them to keep them alive.</p>
    <table>
      <thead><tr><th>Name</th><th>Type</th><th>Strength</th><th>Actions</th></tr></thead>
      <tbody id="fading-tbody"></tbody>
    </table>
  </div>
</div>

<div class="detail" id="detail-pane">
  <span class="close" onclick="closeDetail()">&times;</span>
  <div id="detail-content"></div>
</div>

<script>
const API = '';
let cy;

// Tabs
document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
    tab.classList.add('active');
    document.getElementById('panel-' + tab.dataset.tab).classList.add('active');
    if (tab.dataset.tab === 'graph') initGraph();
    if (tab.dataset.tab === 'timeline') loadTimeline();
    if (tab.dataset.tab === 'fading') loadFading();
  });
});

// Stats
async function loadStats() {
  const r = await fetch(API + '/api/stats');
  const s = await r.json();
  document.getElementById('stats-bar').innerHTML =
    '<div class="stat"><b>' + s.total_nodes + '</b> memories</div>' +
    '<div class="stat"><b>' + s.total_edges + '</b> connections</div>' +
    '<div class="stat"><b>' + s.pinned_count + '</b> pinned</div>' +
    '<div class="stat"><b>' + s.fading_count + '</b> fading</div>';
}

// List
async function loadNodes() {
  const q = document.getElementById('search-input').value;
  const kind = document.getElementById('kind-filter').value;
  const faded = document.getElementById('show-faded').checked;
  let url = API + '/api/nodes?limit=100';
  if (q) url += '&q=' + encodeURIComponent(q);
  if (kind) url += '&kind=' + kind;
  if (faded) url += '&include_faded=true';
  const r = await fetch(url);
  const nodes = await r.json();
  const tbody = document.getElementById('nodes-tbody');
  tbody.innerHTML = nodes.map(n => {
    const pct = Math.min(n.strength * 10, 100);
    let status = '';
    if (n.pinned) status = '<span class="badge pinned">pinned</span>';
    else if (n.strength < 0.1) status = '<span class="badge fading">fading</span>';
    return '<tr onclick="showDetail(\'' + n.id + '\')" style="cursor:pointer">' +
      '<td><b>' + esc(n.name) + '</b></td>' +
      '<td>' + esc(n.kind) + '</td>' +
      '<td style="color:var(--muted)">' + esc(n.summary || '') + '</td>' +
      '<td><div class="strength-bar"><div class="fill" style="width:' + pct + '%"></div></div> ' + n.strength.toFixed(2) + '</td>' +
      '<td>' + status + '</td></tr>';
  }).join('');
}

// Kind filter
const kinds = ['person','place','project','organization','concept','event','artifact','preference','fact','note'];
const sel = document.getElementById('kind-filter');
kinds.forEach(k => { const o = document.createElement('option'); o.value = k; o.textContent = k; sel.appendChild(o); });

document.getElementById('search-input').addEventListener('input', debounce(loadNodes, 300));
document.getElementById('kind-filter').addEventListener('change', loadNodes);
document.getElementById('show-faded').addEventListener('change', loadNodes);

// Detail pane
async function showDetail(id) {
  const r = await fetch(API + '/api/nodes/' + id);
  const d = await r.json();
  const n = d.node;
  const edges = d.edges || [];
  let html = '<h2>' + esc(n.name) + '</h2>';
  html += '<label>Type</label><div class="val">' + esc(n.kind) + '</div>';
  html += '<label>Summary</label><div class="val">' + esc(n.summary || '—') + '</div>';
  html += '<label>Body</label><div class="val" style="white-space:pre-wrap">' + esc(n.body || '—') + '</div>';
  html += '<label>Tags</label><div class="val">' + (n.tags && n.tags.length ? n.tags.map(t => '<span class="badge">' + esc(t) + '</span> ').join('') : '—') + '</div>';
  html += '<label>Strength</label><div class="val">' + n.strength.toFixed(4) + '</div>';
  html += '<label>Access count</label><div class="val">' + n.access_count + '</div>';
  html += '<label>Created</label><div class="val">' + new Date(n.created_at).toLocaleString() + '</div>';
  html += '<label>Last accessed</label><div class="val">' + new Date(n.last_accessed_at).toLocaleString() + '</div>';
  html += '<div style="margin-top:16px;display:flex;gap:8px;">';
  if (n.pinned) html += '<button onclick="togglePin(\'' + id + '\',false)">Unpin</button>';
  else html += '<button onclick="togglePin(\'' + id + '\',true)">Pin</button>';
  html += '<button class="danger" onclick="deleteNode(\'' + id + '\',false)">Soft forget</button>';
  html += '<button class="danger" onclick="deleteNode(\'' + id + '\',true)">Delete</button>';
  html += '</div>';
  if (edges.length) {
    html += '<label style="margin-top:16px">Connections (' + edges.length + ')</label>';
    edges.forEach(e => {
      const other = e.src_id === id ? e.dst_id : e.src_id;
      const dir = e.src_id === id ? '→' : '←';
      html += '<div class="val" style="cursor:pointer" onclick="showDetail(\'' + other + '\')">' + dir + ' ' + esc(String(e.relation)) + ' <a>' + other.substring(0,8) + '...</a></div>';
    });
  }
  document.getElementById('detail-content').innerHTML = html;
  document.getElementById('detail-pane').classList.add('open');
}
function closeDetail() { document.getElementById('detail-pane').classList.remove('open'); }

async function togglePin(id, pin) {
  if (pin) await fetch(API + '/api/nodes/' + id + '/pin', {method:'POST'});
  else await fetch(API + '/api/nodes/' + id + '/pin', {method:'DELETE'});
  showDetail(id); loadNodes(); loadStats();
}
async function deleteNode(id, hard) {
  if (!confirm(hard ? 'Permanently delete this memory?' : 'Soft-forget this memory?')) return;
  await fetch(API + '/api/nodes/' + id + (hard ? '?hard=true' : ''), {method:'DELETE'});
  closeDetail(); loadNodes(); loadStats();
}

// Graph
async function initGraph() {
  const r = await fetch(API + '/api/graph?limit=200');
  const d = await r.json();
  const elements = [];
  (d.nodes || []).forEach(n => {
    elements.push({ data: { id: n.id, label: n.name, kind: n.kind, strength: n.strength, pinned: n.pinned } });
  });
  (d.edges || []).forEach(e => {
    elements.push({ data: { id: e.id, source: e.src_id, target: e.dst_id, relation: e.relation } });
  });
  if (cy) cy.destroy();
  cy = cytoscape({
    container: document.getElementById('graph-container'),
    elements: elements,
    style: [
      { selector: 'node', style: {
        'label': 'data(label)', 'font-size': '11px', 'color': '#e6edf3',
        'background-color': '#58a6ff', 'width': 'mapData(strength, 0, 10, 15, 50)',
        'height': 'mapData(strength, 0, 10, 15, 50)', 'text-valign': 'bottom', 'text-margin-y': 4
      }},
      { selector: 'node[pinned]', style: { 'border-width': 2, 'border-color': '#3fb950' } },
      { selector: 'edge', style: {
        'width': 1, 'line-color': '#30363d', 'target-arrow-color': '#30363d',
        'target-arrow-shape': 'triangle', 'curve-style': 'bezier',
        'label': 'data(relation)', 'font-size': '9px', 'color': '#8b949e', 'text-rotation': 'autorotate'
      }}
    ],
    layout: { name: 'cose', animate: false, nodeRepulsion: 8000, idealEdgeLength: 120 }
  });
  cy.on('tap', 'node', e => showDetail(e.target.id()));
}

// Timeline
async function loadTimeline() {
  const r = await fetch(API + '/api/events?limit=100');
  const events = await r.json();
  document.getElementById('timeline-list').innerHTML = events.map(e =>
    '<li><span class="op">' + esc(e.op) + '</span> ' +
    (e.target_id ? '<span style="color:var(--muted)">' + e.target_id.substring(0,8) + '...</span> ' : '') +
    'by ' + esc(e.actor) +
    '<div class="time">' + new Date(e.ts).toLocaleString() + '</div></li>'
  ).join('');
}

// Fading
async function loadFading() {
  const r = await fetch(API + '/api/nodes?include_faded=true&limit=100');
  const nodes = await r.json();
  const fading = nodes.filter(n => n.strength < 0.1 && !n.pinned);
  document.getElementById('fading-tbody').innerHTML = fading.map(n =>
    '<tr><td>' + esc(n.name) + '</td><td>' + esc(n.kind) + '</td>' +
    '<td>' + n.strength.toFixed(4) + '</td>' +
    '<td><button onclick="togglePin(\'' + n.id + '\',true)">Pin</button></td></tr>'
  ).join('') || '<tr><td colspan="4" style="color:var(--muted)">No fading memories</td></tr>';
}

function esc(s) { const d = document.createElement('div'); d.textContent = s; return d.innerHTML; }
function debounce(fn, ms) { let t; return (...a) => { clearTimeout(t); t = setTimeout(() => fn(...a), ms); }; }

// Init
loadStats();
loadNodes();
</script>
</body>
</html>`
