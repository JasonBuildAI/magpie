# Claims

One line per claim, append-only, first timestamp wins. Do not open a PR from this branch.

- claim: <agent-id> | <ISO8601> | old PR: #<n or none> | defect: <one line> | files: <path,path> | issue: #<n>
- claim: a13aec | 2026-10-14T23:56:14+0800 | old PR: none | defect: a Codex subscription switched off or removed still has its allowance read and a reset about to run out spent | files: internal/provider/codex_expiring.go | issue: #2226
- claim: jx4d | 2026-10-15T00:10:12+08:00 | old PR: none | defect: a folder plugin written as /Users/me/plugin is read as an npm package on Windows, so a backup or sync carries it and the machine restoring it fails to install it | files: internal/plugin/store.go | issue: #2229
- claim: wren | 2026-10-15T00:18:31+08:00 | old PR: none (follows 7f968b95b) | defect: auto affinity drops a conversation as no-cache for a vendor that keeps the prompt cached without saying so (OpenAI's automatic prefix cache, Gemini), so the next turn is sent the whole prompt afresh elsewhere and paid for in full | files: internal/gateway/affinity.go,internal/gateway/gateway.go,internal/gateway/affinity_silent_cache_test.go | issue: #2230
- claim: ag7c2f | 2026-10-15T00:37:39+08:00 | old PR: none | defect: on Windows a port nothing listens on reads as an error, so the Gateway page's take-over ends the older magpie and still answers stuck | files: internal/proc/listener_windows.go,internal/proc/listener.go,internal/proc/listener_test.go | issue: #2231
- claim: mg7x3 | 2026-10-15T00:58:00+08:00 | old PR: none | defect: a Codex folder whose sessions folder is a junction (what Codex CLI makes on Windows when .codex is on another drive) reads as no sessions at all, so Sessions and Usage show none of that Codex's calls or tokens | files: internal/sessions/sessions.go,internal/sessions/codex_link_test.go | issue: #pending
