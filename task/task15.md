Task 15 - Git Cleanup

Objective:
Remove generated binaries and logs from git tracking and update .gitignore.

---

### 1) Check tracked artifacts

```sh
git ls-files | grep -E "^(hms-control$|hms-control\\.exe|hms\\.log|tmp/)"
```

---

### 2) Update .gitignore (idempotent)

```sh
grep -qxF 'hms-control'     .gitignore || echo 'hms-control'     >> .gitignore
grep -qxF 'hms-control.exe' .gitignore || echo 'hms-control.exe' >> .gitignore
grep -qxF 'hms.log'         .gitignore || echo 'hms.log'         >> .gitignore
grep -qxF 'tmp/'            .gitignore || echo 'tmp/'            >> .gitignore
```

---

### 3) Remove from tracking (keep on disk)

```sh
git rm --cached hms-control hms-control.exe hms.log 2>/dev/null || true
git rm -r --cached tmp/ 2>/dev/null || true
git add .gitignore
git commit -m "chore: remove binaries and logs from tracking"
```

---

### 4) Acceptance Criteria

1) git status no longer shows these artifacts as tracked.
2) .gitignore contains required entries.

Expected Outcome:
- Repository is clean of generated artifacts.

Priority:
LOW
