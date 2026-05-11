Task 11 - SSE Live Provisioning Logs

Objective:
Stream provisioning logs to admins via Server-Sent Events (SSE).

---

### 1) Route

Add to main.go:
```go
admin.Get("/tenants/:id/logs/stream", h.StreamProvisionLogs)
```

---

### 2) Handler

Add StreamProvisionLogs in internal/handlers/admin.go:
```go
func (h *Handler) StreamProvisionLogs(c *fiber.Ctx) error {
    tenantID := c.Params("id")
    logPath  := fmt.Sprintf("./tmp/provision-%s.log", tenantID)

    c.Set("Content-Type",  "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection",    "keep-alive")
    c.Set("X-Accel-Buffering", "no")

    c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
        f, err := os.Open(logPath)
        if err != nil {
            fmt.Fprintf(w, "data: log file not found for tenant %s\n\n", tenantID)
            w.Flush()
            return
        }
        defer f.Close()

        reader := bufio.NewReader(f)
        for {
            select {
            case <-c.Context().Done():
                return
            default:
            }

            line, err := reader.ReadString('\n')
            if len(line) > 0 {
                fmt.Fprintf(w, "data: %s\n\n", strings.TrimRight(line, "\r\n"))
                if flushErr := w.Flush(); flushErr != nil {
                    return
                }
            }
            if err != nil {
                time.Sleep(500 * time.Millisecond)
            }
        }
    })

    return nil
}
```

Imports needed:
- bufio
- fmt
- os
- strings
- time

---

### 3) UI snippet for deployment detail page

Add to the templ file:
```html
<pre id="log-output"
style="height:300px;overflow-y:auto;background:#1e1e1e;color:#d4d4d4;padding:1rem;font-family:monospace;">
</pre>
<script>
(function() {
  const out = document.getElementById('log-output');
  const es = new EventSource('/admin/tenants/{{ .TenantID }}/logs/stream');
  es.onmessage = function(e) {
    out.textContent += e.data + '\n';
    out.scrollTop = out.scrollHeight;
  };
  es.onerror = function() {
    out.textContent += '\n[stream closed]\n';
    es.close();
  };
})();
</script>
```

---

### 4) Acceptance Criteria

1) Log stream loads and updates in real time.
2) Missing log file shows a clear message.
3) Stream closes gracefully.

Expected Outcome:
- Admins can view live provisioning logs.

Priority:
MEDIUM
