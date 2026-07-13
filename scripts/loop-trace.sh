#!/usr/bin/env bash
# loop-trace — sharded, size-capped observability for the autonomous loop.
#
# pi records full session transcripts (JSONL) per orchestrator turn and per subagent, but a
# single turn can be megabytes — unsafe to read whole into any agent's context. This tool
# distills sessions into SMALL per-slice/per-action digest pairs (human .md + machine .json,
# hard-capped) under .planning/auto-loop/trace/, with INDEX.md as the entry point. Progressive
# disclosure: digests are layer 1; the raw session (path recorded in every digest) is layer 2.
#
# Usage:
#   scripts/loop-trace.sh sessions            # list known sessions (run cwd + wt-* worktrees)
#   scripts/loop-trace.sh latest              # digest the newest session to stdout (no files)
#   scripts/loop-trace.sh distill [file]      # write digest pair + INDEX line (default: newest)
#   scripts/loop-trace.sh live                # follow the newest session, one line per event
#   scripts/loop-trace.sh full [file]         # COMPLETE human transcript (all thinking, all calls)
#   scripts/loop-trace.sh html [file]         # pi's native full HTML render (open in browser)
#   scripts/loop-trace.sh turn <n>            # show digests recorded for driver turn n
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${AUTO_LOOP_STATE_DIR:-$REPO_ROOT/.planning/auto-loop}"
[[ "$STATE_DIR" = /* ]] || STATE_DIR="$REPO_ROOT/$STATE_DIR"
TRACE_DIR="$STATE_DIR/trace"
CMD="${1:-latest}"; ARG="${2:-}"

session_dirs() {
  # project-local (post driver upgrade) first, then ~/.pi slugs for repo + sibling worktrees
  [[ -d "$STATE_DIR/sessions" ]] && echo "$STATE_DIR/sessions"
  local base="$HOME/.pi/agent/sessions"
  [[ -d "$base" ]] || return 0
  local slug; slug="$(echo "$REPO_ROOT" | tr '/' '-')"
  ls -d "$base"/*"${slug##*-}"* "$base"/*wt-*twenty* "$base"/*polymetrics* 2>/dev/null | sort -u
}

session_files() {
  local limit="${1:-200}"
  SESSION_DIRS="$(session_dirs | tr '\n' ':')" SESSION_LIMIT="$limit" python3 - <<'PY'
import os

dirs=[d for d in os.environ.get("SESSION_DIRS","").split(":") if d]
try:
    limit=max(1,int(os.environ.get("SESSION_LIMIT","200")))
except ValueError:
    limit=200
files=[]
seen=set()
for root in dirs:
    if not os.path.isdir(root):
        continue
    for cur, subdirs, names in os.walk(root):
        subdirs[:] = [d for d in subdirs if d not in {".git","node_modules","vendor"}]
        for name in names:
            if not name.endswith(".jsonl"):
                continue
            path=os.path.join(cur,name)
            if path in seen:
                continue
            seen.add(path)
            try:
                files.append((os.path.getmtime(path),path))
            except OSError:
                pass
files.sort(reverse=True)
for _, path in files[:limit]:
    print(path)
PY
}

newest_session() {
  session_files 1 | head -1
}

case "$CMD" in
  sessions)
    while IFS= read -r f; do
      [[ -n "$f" && -f "$f" ]] || continue
        python3 - "$f" <<'PY'
import json,sys,os,time
f=sys.argv[1]
n=0; last=None; start=None; cwd=""
for line in open(f):
    n+=1
    try: e=json.loads(line)
    except Exception: continue
    ts=e.get("timestamp")
    if e.get("type")=="session": cwd=e.get("cwd",""); start=ts
    if ts: last=ts
age=int(time.time()-os.path.getmtime(f))
state="active" if age<300 else ("recent" if age<3600 else "ended")
rel=os.path.relpath(f, os.getcwd())
print(f"{state:7} {rel[:72]:72}  events={n:<4} size={os.path.getsize(f)//1024}KB  last_event_age={age}s  cwd=…/{cwd.rsplit('/',1)[-1]}")
PY
    done < <(session_files 24) ;;

  latest|distill)
    f="${ARG:-$(newest_session)}"
    [[ -n "$f" && -f "$f" ]] || { echo "no session found" >&2; exit 1; }
    MODE="$CMD" OUT_DIR="$TRACE_DIR" RUN_JSON="$STATE_DIR/RUN.json" python3 - "$f" <<'PY'
import json,sys,os,time,re
f=sys.argv[1]; mode=os.environ["MODE"]; out_dir=os.environ["OUT_DIR"]; run_json=os.environ["RUN_JSON"]
MAX_MD=120; MAX_ITEMS=40
def ts_short(t): return (t or "")[11:19]
def parse_iso(t):
    from datetime import datetime
    return datetime.fromisoformat(t.replace("Z","+00:00")).timestamp()
events=[]
for line in open(f):
    try: events.append(json.loads(line))
    except Exception: pass
cwd=next((e.get("cwd","") for e in events if e.get("type")=="session"),"")
model=next((f"{e.get('provider')}/{e.get('modelId')}" for e in events if e.get("type")=="model_change"),"?")
rows=[]; commands=[]; files=set(); subagents=[]; anomalies=[]
prev_t=None; started=None; ended=None
for e in events:
    t=e.get("timestamp"); ended=t or ended; started=started or t
    if prev_t and t:
        gap=parse_iso(t)-parse_iso(prev_t)
        if gap>60: rows.append((ts_short(t),"gap",f"… {int(gap//60)}m{int(gap%60)}s idle/waiting"))
    prev_t=t or prev_t
    if e.get("type")!="message": continue
    m=e["message"]; role=m.get("role")
    for c in (m.get("content") or []):
        if not isinstance(c,dict): continue
        if c.get("type")=="toolCall":
            name=c.get("name","?"); a=c.get("arguments") or {}
            if name=="bash":
                cmd=(a.get("command") or "")[:160]
                commands.append({"t":ts_short(t),"cmd":cmd})
                rows.append((ts_short(t),"bash",cmd))
            elif name in ("edit","write"):
                p=a.get("path","?"); files.add(p); rows.append((ts_short(t),name,p))
            elif name=="subagent":
                agent=a.get("agent") or a.get("name") or "?"
                task=(a.get("task") or a.get("prompt") or "")[:120]
                subagents.append({"agent":agent,"objective":task})
                rows.append((ts_short(t),"SUBAGENT",f"{agent}: {task}"))
            elif name=="read": rows.append((ts_short(t),"read",a.get("path","?")))
            else: rows.append((ts_short(t),name,json.dumps(a)[:100]))
        elif c.get("type")=="thinking" and role=="assistant":
            th=(c.get("thinking") or "").strip().replace("\n"," ")
            if re.search(r"blocked|halt|fail|error|cannot|violat|scope",th,re.I):
                anomalies.append({"t":ts_short(t),"note":th[:160]})
dur=int(parse_iso(ended)-parse_iso(started)) if started and ended else 0
age=int(time.time()-os.path.getmtime(f))
stage="?"; slice_name="run"
try:
    rj=json.load(open(run_json)); stage=str(rj.get("stage"))
    for s in (rj.get("subissues") or []):
        if isinstance(s,dict) and s.get("stage") not in (None,"not_started","complete"):
            slice_name=s.get("slice") or s.get("name") or slice_name
except Exception: pass
action="turn"
joined=" ".join(r[2] for r in rows[-30:]).lower()
for k in ("review","verify","integrate","execute","plan","reconcile"):
    if k in joined: action=k; break
digest={"slice":slice_name,"action":action,"stage":stage,"model":model,"cwd":cwd,
        "started":started,"ended":ended,"duration_s":dur,"last_event_age_s":age,
        "session_file":f,"events":len(events),
        "commands":commands[-MAX_ITEMS:],"files_touched":sorted(files)[:MAX_ITEMS],
        "subagents":subagents,"anomalies":anomalies[-10:]}
md=[f"# {slice_name} / {action} — {os.path.basename(f)[:19]}",
    f"model {model} · {len(events)} events · {dur//60}m{dur%60}s · last event {age}s ago",
    f"raw session (layer 2): {f}",""]
shown=rows[-(MAX_MD-12):]
if len(rows)>len(shown): md.append(f"(… {len(rows)-len(shown)} earlier lines omitted — see raw session)")
for t,kind,detail in shown: md.append(f"- {t} **{kind}** {detail}")
if anomalies: md.append("\n## anomalies"); md+= [f"- {a['t']} {a['note']}" for a in anomalies[-8:]]
md_text="\n".join(md[:MAX_MD])
if mode=="latest":
    print(md_text)
else:
    base=f"turn-{time.strftime('%H%M')}-{action}"
    d=os.path.join(out_dir,slice_name); os.makedirs(d,exist_ok=True)
    open(os.path.join(d,base+".md"),"w").write(md_text+"\n")
    open(os.path.join(d,base+".json"),"w").write(json.dumps(digest,indent=1))
    idx=os.path.join(out_dir,"INDEX.md")
    line=f"- {time.strftime('%Y-%m-%dT%H:%M')}Z {slice_name}/{base} — {len(commands)} cmds, {len(subagents)} subagents, {len(anomalies)} anomalies, {dur//60}m\n"
    open(idx,"a").write(line)
    print(f"wrote {d}/{base}.md/.json (+INDEX)")
PY
    ;;

  live)
    echo "following ALL sessions in known dirs (orchestrator + subagents) — ^C to stop"
    SESSION_DIRS="$(session_dirs | tr '\n' ':')" python3 - <<'PY'
import json,os,re,time
dirs=[d for d in os.environ["SESSION_DIRS"].split(":") if d]
offsets={}
def tag(f):
    parent=os.path.basename(os.path.dirname(f))
    base=os.path.basename(f)
    return f"{parent}/{base[:18]}"
def session_files():
    files=[]; seen=set()
    for root in dirs:
        if not os.path.isdir(root): continue
        for cur, subdirs, names in os.walk(root):
            subdirs[:] = [d for d in subdirs if d not in {".git","node_modules","vendor"}]
            for name in names:
                if not name.endswith(".jsonl"): continue
                path=os.path.join(cur,name)
                if path in seen: continue
                seen.add(path)
                try: files.append((os.path.getmtime(path),path))
                except OSError: pass
    files.sort(reverse=True)
    return [p for _,p in files[:6]]
def redact(s):
    if not s: return s
    s=re.sub(r'(?i)(TWENTY_API_KEY|[A-Z0-9_]*(?:TOKEN|SECRET|PASSWORD|API_KEY))([\s:=]+)([^\s,"\']+)', r'\1\2[REDACTED]', s)
    s=re.sub(r'(?i)(bearer\s+)[A-Za-z0-9._~+/-]+=*', r'\1[REDACTED]', s)
    return s
def digest(line):
    try: e=json.loads(line)
    except Exception: return None
    t=(e.get("timestamp") or "")[11:19]
    if e.get("type")=="session": return f"{t} ── NEW SESSION cwd=…/{e.get('cwd','').rsplit('/',1)[-1]}"
    if e.get("type")!="message": return None
    m=e["message"]
    out=[]
    for c in (m.get("content") or []):
        if not isinstance(c,dict): continue
        k=c.get("type")
        if k=="toolCall":
            a=c.get("arguments") or {}
            out.append(f"{t} {c.get('name')}: {redact(a.get('command') or a.get('path') or a.get('agent') or json.dumps(a))[:110]}")
        elif k=="thinking":
            th=(c.get("thinking") or "").strip()
            if th:
                lim=100000 if os.environ.get("LIVE_FULL") else 90
                out.append(f"{t} 💭 {redact(th[:lim])}")
        elif k=="text" and m.get("role")=="toolResult":
            out.append(f"{t} ← result {len(c.get('text') or '')}B")
    return "\n".join(out) if out else None
try:
    while True:
        files=session_files()
        for f in files:
            size=os.path.getsize(f)
            off=offsets.get(f)
            if off is None:
                offsets[f]=size  # start at end for pre-existing files
                continue
            if size>off:
                with open(f) as fh:
                    fh.seek(off)
                    for line in fh:
                        d2=digest(line)
                        if d2:
                            for ln in d2.split("\n"): print(f"[{tag(f)}] {ln}",flush=True)
                offsets[f]=size
        time.sleep(2)
except KeyboardInterrupt: pass
PY
    ;;

  full)
    # Complete human-readable markdown: FULL thinking, full tool args, generous results.
    f="${ARG:-$(newest_session)}"
    [[ -n "$f" && -f "$f" ]] || { echo "no session found" >&2; exit 1; }
    python3 - "$f" <<'PY'
import json,sys
f=sys.argv[1]
print(f"# FULL transcript — {f}\n")
for line in open(f):
    try: e=json.loads(line)
    except Exception: continue
    t=(e.get("timestamp") or "")[11:19]
    ty=e.get("type")
    if ty=="session": print(f"**session start** cwd={e.get('cwd')}\n"); continue
    if ty=="model_change": print(f"**model** {e.get('provider')}/{e.get('modelId')}\n"); continue
    if ty!="message": continue
    m=e["message"]; role=m.get("role")
    for c in (m.get("content") or []):
        if not isinstance(c,dict): continue
        k=c.get("type")
        if k=="thinking":
            print(f"### {t} 💭 thinking\n\n{(c.get('thinking') or '').strip()}\n")
        elif k=="text":
            label="🛠 tool result" if role=="toolResult" else f"💬 {role}"
            txt=(c.get("text") or "")
            note="" if len(txt)<=4000 else f"\n\n*(truncated — {len(txt)} chars total; raw session has all of it)*"
            print(f"### {t} {label}\n\n```\n{txt[:4000]}\n```{note}\n")
        elif k=="toolCall":
            print(f"### {t} ▶ {c.get('name')}\n\n```json\n{json.dumps(c.get('arguments') or {},indent=1)[:4000]}\n```\n")
PY
    ;;

  html)
    # pi's native pretty rendering of the complete session (2MB-ish; open in a browser).
    f="${ARG:-$(newest_session)}"
    [[ -n "$f" && -f "$f" ]] || { echo "no session found" >&2; exit 1; }
    mkdir -p "$TRACE_DIR/full" && cd "$TRACE_DIR/full" && pi --export "$f" && echo "open with: open $TRACE_DIR/full/pi-session-*.html"
    ;;

  turn)
    [[ -n "$ARG" ]] || { echo "usage: loop-trace turn <n>" >&2; exit 1; }
    grep -rl "turn-.*" "$TRACE_DIR" 2>/dev/null | head -5
    grep -n "turn $ARG:" "$STATE_DIR/driver.log" | tail -5 ;;

  *) echo "usage: loop-trace.sh sessions|latest|distill [file]|full [file]|html [file]|live|turn <n>" >&2; exit 2 ;;
esac
