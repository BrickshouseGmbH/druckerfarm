package main

import "net/http"

// serveLoginPage is shown to unauthenticated visitors. It is standalone (no
// dependency on the main app bundle) and either logs in or, on first run,
// creates the first admin account.
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(loginHTML))
}

const loginHTML = `<!doctype html><html lang="de"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Druckerfarm — Anmeldung</title>
<style>
:root{--bg:#0f1216;--s1:#161b22;--s2:#1c2330;--border:#2a3342;--fg:#e6edf3;--muted:#8b97a7;--accent:#39d0d8;--red:#ff3355;}
*{box-sizing:border-box}
body{margin:0;font-family:Arial,Helvetica,sans-serif;background:var(--bg);color:var(--fg);display:flex;min-height:100vh;align-items:center;justify-content:center;}
.card{background:var(--s1);border:1px solid var(--border);border-radius:12px;padding:28px 26px;width:340px;max-width:92vw;box-shadow:0 12px 40px rgba(0,0,0,.4);}
.brand{display:flex;align-items:center;gap:10px;margin-bottom:6px;}
.brand img{height:34px;}
h1{font-size:17px;margin:10px 0 2px;}
p.sub{color:var(--muted);font-size:12px;margin:0 0 18px;}
label{display:block;font-size:12px;color:var(--muted);margin:12px 0 4px;}
input{width:100%;padding:9px 10px;border:1px solid var(--border);border-radius:7px;background:var(--s2);color:var(--fg);font-size:14px;}
button{width:100%;margin-top:18px;padding:10px;border:0;border-radius:7px;background:var(--accent);color:#001018;font-weight:700;font-size:14px;cursor:pointer;}
button:hover{filter:brightness(1.07);}
.err{color:var(--red);font-size:12px;margin-top:12px;min-height:16px;}
.hint{color:var(--muted);font-size:11px;margin-top:14px;line-height:1.5;}
</style></head><body>
<form class="card" id="f" onsubmit="return false;">
  <div class="brand"><img src="/logo.svg" alt="Druckerfarm"></div>
  <h1 id="title">Anmelden</h1>
  <p class="sub" id="sub">Bitte mit deinem Konto anmelden.</p>
  <label for="u">Benutzername</label>
  <input id="u" autocomplete="username" autofocus>
  <label for="p" id="plabel">Passwort</label>
  <input id="p" type="password" autocomplete="current-password">
  <button id="btn" type="submit">Anmelden</button>
  <div class="err" id="err"></div>
  <div class="hint" id="hint"></div>
</form>
<script>
async function init(){
  let need=false;
  try{ const d=await (await fetch('/api/needsetup',{cache:'no-store'})).json();
    need=!!d.needsetup; }catch(e){}
  if(need){
    // Accounts are created only at the server machine (local access has full
    // rights without login). A remote visitor cannot self-register.
    document.getElementById('sub').textContent='Für den Fernzugriff sind noch keine Konten angelegt.';
    document.getElementById('u').style.display='none';
    document.getElementById('p').style.display='none';
    document.getElementById('plabel').style.display='none';
    document.querySelector('label[for=u]').style.display='none';
    document.getElementById('btn').style.display='none';
    document.getElementById('hint').textContent='Bitte am Server-Rechner die Druckerfarm öffnen und unter Einstellungen › Benutzer & Zugang ein Konto anlegen. Danach kannst du dich hier anmelden.';
  }
}
async function submit(){
  const err=document.getElementById('err'); err.textContent='';
  const name=document.getElementById('u').value.trim();
  const password=document.getElementById('p').value;
  if(!name||!password){ err.textContent='Bitte Benutzername und Passwort eingeben.'; return; }
  try{
    const r=await fetch('/api/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name,password})});
    if(!r.ok){ let m='Anmeldung fehlgeschlagen.'; try{const j=await r.json(); if(j.error)m=j.error;}catch(e){} err.textContent=m; return; }
    location.href='/';
  }catch(e){ err.textContent=String(e); }
}
document.getElementById('f').addEventListener('submit',submit);
init();
</script></body></html>`
