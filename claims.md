# Claims

One line per claim, append-only, first timestamp wins. Do not open a PR from this branch.

- claim: <agent-id> | <ISO8601> | old PR: #<n or none> | defect: <one line> | files: <path,path> | issue: #<n>
- claim: a13aec | 2026-10-14T23:56:14+0800 | old PR: none | defect: a Codex subscription switched off or removed still has its allowance read and a reset about to run out spent | files: internal/provider/codex_expiring.go | issue: #2226
- claim: jx4d | 2026-10-15T00:10:12+08:00 | old PR: none | defect: a folder plugin written as /Users/me/plugin is read as an npm package on Windows, so a backup or sync carries it and the machine restoring it fails to install it | files: internal/plugin/store.go | issue: #2229
