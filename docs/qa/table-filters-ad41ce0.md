# QA Report: Schema-declared table filters (`ad41ce0`)

**Verdict: PASS** (with minor residual UX notes)

## What changed

Commit `ad41ce0` (“Add schema-declared table filters”) adds:

- Schema annotation `FilterableColumns` (User: `email`, `is_staff`, `is_active`)
- Generated list handlers that read Datastar `filter.*` signals and apply Ent predicates (`ContainsFold` for strings, EQ for bools/ints)
- List UI filter form driven by Datastar + `data-query-string__history` (named query params, Clear, lazy Datastar table load)
- `vent.BoolFilter` (`true` / `false` / `all`) so unset bools do not 400

## Why this covers the change

Exercised the Users list end-to-end (codegen + templates + plugins + handlers): string filter, bool filters, empty result, Clear, URL sync, refresh hydration, and invalid Datastar payload error handling.

## Steps performed

1. Fast-forwarded QA branch to `origin/main` @ `ad41ce0`; ran `go test .` and `go test ./templates/gui/`
2. Started `go run ./examples/basic/cmd/server`; logged in as `admin@vent.com` / `test_user`
3. Seeded `viewer@vent.com` (non-staff/active) and `inactive@vent.com` (staff/inactive)
4. Scripted Datastar GETs against `/admin/users/` for filter combinations + malformed JSON
5. Manual UI: email filter → no-match → Clear → IsStaff=No → IsActive=No → refresh with query params

## Results

| Case | Result | Evidence |
|------|--------|----------|
| Happy path: email `viewer` | PASS — only `viewer@vent.com`; URL `filter.email=viewer`; Clear shown | video + `qa_filters_step2_email_viewer.webp` |
| Edge: email `zzz-nomatch` | PASS — “No data”; Clear shown | video + `qa_filters_step3_no_match.webp` |
| Clear | PASS — all 3 users restored | video + `qa_filters_step4_cleared.webp` |
| Bool IsStaff=No | PASS — only `viewer@vent.com` | video + `qa_filters_step5_staff_no.webp` |
| Bool IsActive=No | PASS — only `inactive@vent.com` | video + `qa_filters_step6_active_no.webp` |
| Refresh hydration | PASS — `filter.is_active=false` kept; same row | video + `qa_filters_step7_refresh_hydrate.webp` |
| ContainsFold (`VENT`) | PASS — all three emails | API log |
| Empty bool strings | PASS — HTTP 200, all rows | API log |
| Malformed `datastar` | PASS — HTTP 400 `invalid filter` | API log |
| Unit tests | PASS | `go test .`, `go test ./templates/gui/` |

Artifacts: `qa_table_filters_users_e2e.mp4`, step screenshots, `qa_filters_api_evidence.txt`

## Residual risk / untested

- **Bool select blanking:** when only a string filter is set, URL may contain `filter.is_staff=` / `filter.is_active=` (empty), so selects can render blank instead of “All” until reset/refresh normalizes them. Filtering still works.
- **Other schemas:** PermissionGroups/Permissions `name` filters not UI-tested (same generated pattern).
- **Int filters:** no filterable int column on User in this example.
- **Browser Back/Forward:** popstate handler present; only full refresh hydration was verified in the UI pass.
- Full suite not run (per focused-QA instructions).
