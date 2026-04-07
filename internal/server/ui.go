package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Dispatch — Stockyard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;--leather:#a0845c;--leather-light:#c4a87a;--cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;--gold:#d4a843;--green:#5ba86e;--red:#c0392b;--font-serif:'Libre Baskerville',Georgia,serif;--font-mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}a{color:var(--rust-light);text-decoration:none}a:hover{color:var(--gold)}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}.hdr-left{display:flex;align-items:center;gap:1rem}.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream);letter-spacing:1px}.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;letter-spacing:1px;text-transform:uppercase;border:1px solid;color:var(--green);border-color:var(--green)}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:1rem;margin-bottom:2rem}.card{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.2rem}.card-val{font-family:var(--font-mono);font-size:1.6rem;font-weight:700;color:var(--cream);display:block}.card-lbl{font-family:var(--font-mono);font-size:.58rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.2rem}
.section{margin-bottom:2rem}.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.75rem}th{background:var(--bg3);padding:.4rem .8rem;text-align:left;color:var(--leather-light);font-weight:400;letter-spacing:1px;font-size:.62rem;text-transform:uppercase}td{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim)}tr:hover td{background:var(--bg2)}.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.btn{font-family:var(--font-mono);font-size:.7rem;padding:.3rem .8rem;border:1px solid var(--leather);background:transparent;color:var(--cream);cursor:pointer;transition:all .2s}.btn:hover{border-color:var(--rust-light);color:var(--rust-light)}.btn-rust{border-color:var(--rust);color:var(--rust-light)}.btn-rust:hover{background:var(--rust);color:var(--cream)}.btn-sm{font-size:.62rem;padding:.2rem .5rem}
.pill{display:inline-block;font-family:var(--font-mono);font-size:.58rem;padding:.1rem .4rem;border-radius:2px;text-transform:uppercase}.pill-draft{background:var(--bg3);color:var(--cream-muted)}.pill-sent{background:#1a3a2a;color:var(--green)}.pill-sending{background:#2a2a1a;color:var(--gold)}
.lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:1px;text-transform:uppercase;color:var(--leather)}input,textarea{font-family:var(--font-mono);font-size:.78rem;background:var(--bg3);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .7rem;outline:none}input:focus,textarea:focus{border-color:var(--leather)}.row{display:flex;gap:.8rem;align-items:flex-end;flex-wrap:wrap;margin-bottom:1rem}.field{display:flex;flex-direction:column;gap:.3rem}
.tabs{display:flex;gap:0;margin-bottom:1.5rem;border-bottom:1px solid var(--bg3)}.tab{font-family:var(--font-mono);font-size:.72rem;padding:.6rem 1.2rem;color:var(--cream-muted);cursor:pointer;border-bottom:2px solid transparent;letter-spacing:1px;text-transform:uppercase}.tab:hover{color:var(--cream-dim)}.tab.active{color:var(--rust-light);border-bottom-color:var(--rust-light)}.tab-content{display:none}.tab-content.active{display:block}
pre{background:var(--bg3);padding:.8rem 1rem;font-family:var(--font-mono);font-size:.72rem;color:var(--cream-dim);overflow-x:auto}
</style></head><body>
<div class="hdr"><div class="hdr-left">
<svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
<span class="hdr-brand">Stockyard</span><span class="hdr-title">Dispatch</span></div>
<div style="display:flex;gap:.8rem;align-items:center"><span class="badge">Free</span><a href="/api/status" class="lbl" style="color:var(--leather)">API</a></div></div>
<div class="main"><div id="upgrade-banner" style="display:none;background:#241e18;border:1px solid #8b3d1a;border-left:3px solid #c45d2c;padding:.6rem 1rem;font-size:.78rem;color:#bfb5a3;margin-bottom:.8rem"><strong style="color:#f0e6d3">Free tier</strong> — 10 items max. <a href="https://stockyard.dev/dispatch/" target="_blank" style="color:#e8753a">Upgrade to Pro →</a></div>
<div class="cards">
  <div class="card"><span class="card-val" id="s-lists">—</span><span class="card-lbl">Lists</span></div>
  <div class="card"><span class="card-val" id="s-subs">—</span><span class="card-lbl">Subscribers</span></div>
  <div class="card"><span class="card-val" id="s-camps">—</span><span class="card-lbl">Campaigns</span></div>
  <div class="card"><span class="card-val" id="s-sent">—</span><span class="card-lbl">Sent</span></div>
</div>
<div class="tabs">
  <div class="tab active" onclick="switchTab('lists')">Lists</div>
  <div class="tab" onclick="switchTab('campaigns')">Campaigns</div>
  <div class="tab" onclick="switchTab('usage')">Usage</div>
</div>
<div id="tab-lists" class="tab-content active">
  <div class="section">
    <div class="section-title">Create List</div>
    <div class="row">
      <div class="field"><span class="lbl">Name</span><input id="c-name" placeholder="Newsletter" style="width:200px"></div>
      <button class="btn btn-rust" onclick="createList()">Create</button>
    </div><div id="c-result"></div>
  </div>
  <div class="section">
    <div class="section-title">Lists</div>
    <table><thead><tr><th>Name</th><th>Subscribers</th><th>Embed</th><th></th></tr></thead>
    <tbody id="lists-body"></tbody></table>
  </div>
</div>
<div id="tab-campaigns" class="tab-content">
  <div class="section"><div class="section-title">Campaigns</div><div id="camps-list"></div></div>
</div>
<div id="tab-usage" class="tab-content">
  <div class="section"><div class="section-title">Quick Start</div>
    <pre>
# Create a list
curl -X POST http://localhost:8900/api/lists \
  -H "Content-Type: application/json" \
  -d '{"name":"Newsletter"}'

# Add subscriber
curl -X POST http://localhost:8900/api/lists/{id}/subscribers \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"Jane"}'

# Embed subscribe form on your site
&lt;form method="POST" action="http://localhost:8900/subscribe/{list_id}"&gt;
  &lt;input name="email" type="email" required&gt;
  &lt;button&gt;Subscribe&lt;/button&gt;
&lt;/form&gt;

# Create + send a campaign
curl -X POST http://localhost:8900/api/lists/{id}/campaigns \
  -H "Content-Type: application/json" \
  -d '{"subject":"Weekly Update","body_html":"&lt;h1&gt;Hello!&lt;/h1&gt;"}'

curl -X POST http://localhost:8900/api/campaigns/{id}/send
    </pre>
  </div>
</div>
</div>
<script>
let lists=[];
function switchTab(n){document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.textContent.toLowerCase()===n));document.querySelectorAll('.tab-content').forEach(t=>t.classList.toggle('active',t.id==='tab-'+n));if(n==='campaigns')loadCampaigns();}
async function refresh(){
  try{const s=await(await fetch('/api/status')).json();document.getElementById('s-lists').textContent=s.lists||0;document.getElementById('s-subs').textContent=fmt(s.subscribers||0);document.getElementById('s-camps').textContent=s.campaigns||0;document.getElementById('s-sent').textContent=fmt(s.emails_sent||0);}catch(e){}
  try{const d=await(await fetch('/api/lists')).json();lists=d.lists||[];const tb=document.getElementById('lists-body');
  if(!lists.length){tb.innerHTML='<tr><td colspan="4" class="empty">No lists yet.</td></tr>';return;}
  tb.innerHTML=lists.map(l=>'<tr><td style="color:var(--cream);font-weight:600">'+esc(l.name)+'<br><span style="font-size:.55rem;color:var(--cream-muted)">'+l.id+'</span></td><td>'+l.subscriber_count+'</td><td style="font-size:.6rem"><code>POST /subscribe/'+l.id+'</code></td><td><button class="btn btn-sm" onclick="deleteList(\''+l.id+'\')">Delete</button></td></tr>').join('');}catch(e){}
}
async function loadCampaigns(){
  let html='';for(const l of lists){const d=await(await fetch('/api/lists/'+l.id+'/campaigns')).json();const cs=d.campaigns||[];
  html+='<div class="section-title" style="margin-top:1rem">'+esc(l.name)+'</div>';
  if(!cs.length){html+='<div class="empty">No campaigns</div>';continue;}
  html+='<table><thead><tr><th>Subject</th><th>Status</th><th>Sent</th><th>Opens</th></tr></thead><tbody>';
  html+=cs.map(c=>'<tr><td>'+esc(c.subject)+'</td><td><span class="pill pill-'+c.status+'">'+c.status+'</span></td><td>'+c.sent_count+'</td><td>'+c.open_count+'</td></tr>').join('');
  html+='</tbody></table>';}
  document.getElementById('camps-list').innerHTML=html||'<div class="empty">No lists yet</div>';
}
async function createList(){const name=document.getElementById('c-name').value.trim();if(!name)return;const r=await fetch('/api/lists',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name})});const d=await r.json();if(r.ok){document.getElementById('c-result').innerHTML='<span style="color:var(--green)">Created</span>';document.getElementById('c-name').value='';refresh();}else{document.getElementById('c-result').innerHTML='<span style="color:var(--red)">'+esc(d.error)+'</span>';}}
async function deleteList(id){if(!confirm('Delete list?'))return;await fetch('/api/lists/'+id,{method:'DELETE'});refresh();}
function fmt(n){if(n>=1e6)return(n/1e6).toFixed(1)+'M';if(n>=1e3)return(n/1e3).toFixed(1)+'K';return n;}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
refresh();setInterval(refresh,8000);
fetch('/api/tier').then(r=>r.json()).then(j=>{if(j.tier==='free'){var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'}}).catch(()=>{var b=document.getElementById('upgrade-banner');if(b)b.style.display='block'});
</script><script>
(function(){
  fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
    if(!cfg||typeof cfg!=='object')return;
    if(cfg.dashboard_title){
      document.title=cfg.dashboard_title;
      var h1=document.querySelector('h1');
      if(h1){
        var inner=h1.innerHTML;
        var firstSpan=inner.match(/<span[^>]*>[^<]*<\/span>/);
        if(firstSpan){h1.innerHTML=firstSpan[0]+' '+cfg.dashboard_title}
        else{h1.textContent=cfg.dashboard_title}
      }
    }
  }).catch(function(){});
})();
</script>
</body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
