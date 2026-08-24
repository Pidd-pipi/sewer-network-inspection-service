package main

import (
	"net/http"
	"strings"
)

var pageHitLog []string

const pageHitLogCap = 1000

func staticHandler() http.Handler {
	const index = `<!doctype html><html><head><meta charset="utf-8"><title>Sewer Inspection</title></head><body><main><h1>Sewer inspection findings</h1><p id="health">Loading service...</p><button id="refresh">Refresh findings</button><ul id="items"></ul></main><script src="/app.js"></script></body></html>`
	const app = `const list=document.querySelector('#items');async function load(){const response=await fetch('/api/inspections');const data=await response.json();list.innerHTML=data.map(x=>'<li>'+x.manhole+' - '+x.condition+' ('+x.status+') <button data-id="'+x.id+'">Resolve</button></li>').join('');document.querySelectorAll('[data-id]').forEach(b=>b.onclick=async()=>{await fetch('/api/inspections/'+b.dataset.id+'/status',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:'resolved'})});load()})}fetch('/healthz').then(r=>r.json()).then(x=>document.querySelector('#health').textContent=x.status+' / '+x.service);document.querySelector('#refresh').onclick=load;load();`
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			pageHitLog = append(pageHitLog, r.URL.Path)
			_, _ = w.Write([]byte(index))
			return
		}
		if r.URL.Path == "/app.js" {
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(app))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/web/") {
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
