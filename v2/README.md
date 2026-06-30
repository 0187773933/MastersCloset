# Master's Closet — v2

A from-the-ground-up rewrite of the tracking server. Goals of v2:

- **No hardcoded paths.** All HTML/CSS/JS/images are `//go:embed`ed into the
  binary (`v2/server/{html,static,templates}`). The binary runs from any working
  directory; there is no `cp -r` build step. The only filesystem locations are
  the data dirs resolved in `v2/paths` (under `~/.config/mct`).
- **Live-editable config.** Configuration lives in BoltDB and is edited through a
  real settings panel at `/admin/settings`. v1 read config once at startup and
  could only "save" by rewriting `.js` source files by line number.
- **A clean user package.** `v2/user` splits pure data (`model.go`) from
  persistence (`store.go`) and has a single check-in calculation (`checkin.go`),
  one `Save`, one `Delete`, etc.

## Layout

```
v2/config      Config type + field metadata, Bolt-backed Manager (snapshot/edit/persist)
v2/user        model / store / checkin / balance / similar
v2/server      monolith Server, routes, auth, public + admin handlers, settings, embedded assets
v2/logger      leveled logger, daily file under ~/.config/mct/logs
v2/paths       data-dir resolution
v2/encryption  ChaCha + SecretBox helpers (ported from v1)
v2/printer     4x6 ticket PDF (embedded font/logo), CUPS/SumatraPDF send
```

## Build & run

```bash
go build -o mct .            # default build is v2
./mct config.json            # first run seeds config into the db, then ignores the file

# v1 fallback during the transition:
go build -tags legacy -o mct_legacy .
```

`LOG_LEVEL=debug` enables caller-aware debug logging.

### Config bootstrap

The JSON file passed as the first argument is parsed only for the bootstrap
fields needed to open the db (`bolt_db_path`, `bolt_db_encryption_key`,
`bleve_search_path`). On first run the full config is seeded into the db; after
that the db is authoritative and the file is ignored. `bolt_db_path` /
`bolt_db_encryption_key` / `bleve_search_path` are bootstrap-only and not
editable in the panel (changing them is a migration, not a setting).

The config blob is stored plaintext in the local (`0600`) db — user PII is
encrypted separately, and the same secrets previously lived in a plaintext
`config.json`.

### Per-machine data paths (v1-compatible auto-detection)

When `bolt_db_path` / `bleve_search_path` are **relative** (e.g. `mct.db`), v2
resolves them to `~/.config/mct/mct_<fingerprint>.db` (and `…_<fingerprint>.bleve`)
— the exact naming v1 used. The fingerprint (`v2/fingerprint`) reproduces v1's
`utils.FingerPrintPassive`, so an existing install's data is picked up
automatically with no config change. On a fresh machine the seed file/dir is
copied in on first run, just like v1's `FixDBAndSearchIndex`. Pass an **absolute**
path to override and pin a specific file.

## Routes (this pass)

| Route | Purpose |
| --- | --- |
| `GET /` | landing / user home / admin redirect |
| `GET /join`, `POST /user/api/new` | new-guest signup |
| `GET /user/login/fresh/:uuid` | QR hand-off: first check-in + login cookie |
| `GET /user/checkin`, `/user/checkin/display/:uuid`, `/user/checkin/silent/:uuid` | check-in flow |
| `GET/POST /admin/login`, `GET /admin/logout` | admin auth |
| `GET /admin` | dashboard |
| `GET /admin/settings`, `GET/POST /admin/settings/api` | **live config editor** |
| `GET /admin/users`, `/admin/user/new`, `/admin/checkin` | admin HTML pages |
| `GET /admin/user/all` · `/user/get/:uuid` · `/user/barcode/:barcode` · `/user/search/:name` · `/user/search/fuzzy/:query` · `/user/similar/:uuid` | user lookup/search |
| `POST /admin/user/new` · `/user/edit` · `DELETE /admin/user/:uuid` | create / edit / delete user |
| `GET /admin/user/handoff/:uuid/qr.png` | onboarding QR (server-rendered) |
| `GET /admin/user/checkin/test/:uuid` · `POST /admin/user/checkin/:uuid` (shopping) · `/checkin/force/:uuid` · `/refill/:uuid` | check-in |
| `GET /admin/checkins` (totals) · `/checkins/date/:date` · `GET/POST /admin/checkin/:ulid` · `DELETE /admin/checkin/:uuid/:ulid` | check-in history |
| `GET /admin/checkin/:ulid/ticket.pdf` | render/reprint a check-in's ticket |
| `POST /admin/print` | print a custom job |

The admin check-in is a **shopping transaction**: it decrements per-category
balances (refilling from the configured limit when dry), mints a virtual barcode
if the user has none, tallies guests, records a ULID'd check-in carrying the
shopping ticket, and **prints the 4×6 physical ticket** for the checkout station.

### Printing

`v2/printer` renders the ticket entirely in memory (font + logo are embedded via
`//go:embed`; the Code128 barcode and PDF are built in RAM). Only the final PDF
is briefly written to a temp file because the OS print command needs a path:
`lp` (CUPS) on macOS/Linux, `SumatraPDF.exe` on Windows — selected by the
`printer_name` / `printer_speed` config fields (editable in the settings panel).
A printing failure never fails the check-in itself; the visit is recorded and the
ticket can be re-printed from history. An empty `printer_name` renders without
sending (useful on dev boxes with no label printer).

## Deferred to later passes

**Email/SMS** (single + bulk), user **reports** (main / MailChimp / check-ins),
the audio→user transcription helper, and **remotesync v2**. v1 remains the
reference for these; they port onto the v2 `Store` / `Manager`.
