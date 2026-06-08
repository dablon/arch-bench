#!/usr/bin/env python3
import json, sys
CSS = "body{font-family:Arial;background:#0f1419;color:#fff;padding:20px}"
def make_row(svc):
    lang = "rust" if "rust" in svc["name"] else "go"
    return f"<tr><td>{svc[NAME]}</td><td class={lang}>{lang.upper()}</td><td>{svc[TPS]:.1f}</td><td>{svc[AVG]:.2f}ms</td></tr>"
def make_row(svc):
    lang = "rust" if "rust" in svc["name"] else "go"
    n = svc["name"]; t = round(svc["tps"], 1); a = round(svc["avg_ms"], 2)
    return "<tr><td>" + n + "</td><td>" + lang.upper() + "</td><td>" + str(t) + "</td><td>" + str(a) + "ms</td></tr>"
def generate_html(data):
    html = "<!DOCTYPE html><html><head><meta charset=UTF-8><title>arch-bench</title><style>" + CSS + "</style></head><body>"
    html += "<h1>arch-bench Benchmark Report</h1>"
    total = sum(s["ok"] for s in data["services"])
    best = max(s["tps"] for s in data["services"])
    html += "<div class=kpi><div class=kpi-v>" + str(total) + "</div><div class=kpi-l>Total Requests</div></div>"
    html += "<div class=kpi><div class=kpi-v style=color:#00ff88>" + str(round(best,1)) + "</div><div class=kpi-l>Best TPS</div></div>"
    html += "<table><thead><tr><th>Service</th><th>Lang</th><th>TPS</th><th>Avg</th><th>p50</th><th>p95</th><th>p99</th><th>Max</th></tr></thead><tbody>"
    for s in sorted(data["services"], key=lambda x: -x["tps"]):
        html += make_row(s)
    html += "</tbody></table></body></html>"
    return html
if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: generate_report.py <results.json>")
        sys.exit(1)
    with open(sys.argv[1]) as f:
        data = json.load(f)
    html = generate_html(data)
    with open("/home/nalcaraz/arch-bench/index.html", "w") as f:
        f.write(html)
    print("Report generated: /home/nalcaraz/arch-bench/index.html")
