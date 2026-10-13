# Pokemon GO Collection Manager

CLI tool for managing Pokemon GO collections with comprehensive analysis, raid team building, PvP team builder, Lucky Trade targets, and development tracking — all from LibreOffice ODS spreadsheets.

## Features

- **Multi-account support** — Each player has their own `Pokemon-GO-<name>.ods` file
- **7-section Markdown reports** — Overview, Raid teams, PvP teams, Transfer candidates, Lucky Trade targets, Development ranking, Detailed analysis
- **Raid DPS algorithm** — Type-effective counters with actual equipped moves
- **PvP team builder** — Great/Ultra/Master League, 5 ready teams (3v3) with Lead/Switch/Closer roles
- **Lucky Trade targets** — Sourced from Pokedex (master list), filtered by meta relevance
- **Development ranking** — CP cap detection via `CelRozwoju`, excludes fully developed Pokemon
- **Safe ODS workflow** — Backup (max 5), verify, restore, regenerate with safety checks
- **External data** — PvPoke gamemaster + meta rankings (auto-updated, cached 30 days)

## Quick Start

### Prerequisites
- Go 1.22+
- LibreOffice (for ODS editing)

### Install from source
```bash
git clone https://github.com/Adzikus/pokemon-go
cd pokemon-go
make build
# Binary at bin/pokemon
```

### Create your first collection
```bash
# Generate template ODS from gamemaster
pokemon generate --konto Test

# Edit in LibreOffice: fill the "Kolekcja" sheet with your Pokemon
# Then update derived sheets:
pokemon update --konto Test
```

### Daily workflow
```bash
# 1. Backup before editing
pokemon backup --konto Test

# 2. Edit ODS in LibreOffice (add catches, update IVs, change Purpose)

# 3. Verify after editing
pokemon verify --konto Test

# 4. Regenerate derived sheets (Pokedex, Moves, Lists)
pokemon update --konto Test

# 5. Generate analysis report
pokemon analyze --konto Test
# Output: analizy/Analiza-YYYYMMDD-opencode-Test.md
```

### Update external data
```bash
pokemon update-gamemaster    # Fresh gamemaster.json from PvPoke
pokemon update-meta          # Fresh meta rankings from PvPoke
```

## Commands

| Command | Description |
|---------|-------------|
| `pokemon verify --konto <name>` | Verify ODS structure & data integrity |
| `pokemon backup --konto <name>` | Create verified backup (max 5 retained) |
| `pokemon restore --konto <name> -- <backup>` | Restore ODS from backup |
| `pokemon generate --konto <name>` | Generate new ODS from gamemaster |
| `pokemon update --konto <name>` | Regenerate derived sheets (Pokedex, Moves, Lists) |
| `pokemon analyze --konto <name>` | Generate full analysis report (MD + JSON) |
| `pokemon analyze --all` | Analyze all accounts (skips "Test") |
| `pokemon update-gamemaster [--force]` | Download fresh gamemaster.json |
| `pokemon update-meta [--force]` | Download fresh PvP meta rankings |

## Project Structure

```
pokemon-go/
├── bin/pokemon              # Compiled binary
├── data/
│   ├── gamemaster.json      # PvPoke game master (base stats, moves, CPM)
│   ├── meta_rankings.json   # PvP meta tiers (S/A/B/C per league)
│   └── config.json          # URLs, paths, scoring weights
├── Pokemon-GO-<name>.ods    # Collection per account (source of truth)
├── backups/                 # Timestamped ODS backups (max 5/account)
├── analizy/                 # Generated reports (MD + JSON)
├── templates/
│   └── report.md.tmpl       # Markdown report template
├── docs/                    # Architecture docs (01-07)
└── src/pokemon/             # Go source code
```

## ODS Spreadsheet Structure

Each account's `Pokemon-GO-<name>.ods` contains:

| Sheet | Purpose | Editable? |
|-------|---------|-----------|
| **Kolekcja** | Your Pokemon: UID, Species, IV, CP, Moves, Purpose, CelRozwoju, etc. | **YES** — main data entry |
| Pokedex | Master list (1598 species) with Lucky/Shiny/Caught flags | Auto-synced from Kolekcja |
| Pokemon | Species details (base stats, types, moves) | Read-only |
| Ruchy | Move database (DPS, energy, type) | Read-only |
| Listy | Dropdown values for validation | Read-only |

**Key columns in Kolekcja:**
- `Purpose` — PvP / Rajdy / Gym / Trade / Transfer / Kolekcja
- `CelRozwoju` (col 24) — Overrides Purpose for CP cap (e.g. "PvP GL 1450")
- `Priority` (col 22) — 1=High, 2=Med, 3=Low

## Analysis Report (7 Sections)

1. **Ogólne podsumowanie** — TOP 10 HUNDO/CP/Dynamax/Shadow/Shiny/Power-Up/Dev Score, Meta vs Collection
2. **Zestawy do Rajdów** — Tier list (S/A/B/C) per raid type, investment priorities
3. **Zestawy do PvP** — 5 ready teams (3v3) per league: Lead, Switch, Closer + playstyle guide
4. **Kandydaci do Transferu** — Safe to transfer, DO NOT transfer list
5. **Poszukiwani do statusu Lucky** — Pokedex-based, meta-filtered, score > 0, full list
6. **Szczegółowa analiza** — Legendaries, HUNDO, Shadow, PvP, Shiny (rating 1-10, recommendation)
7. **Pokemon-y do rozwoju** — Ranked by DevScore, filtered by CP cap, excludes fully developed

## Configuration

`data/config.json` controls:
- External URLs (PvPoke gamemaster, rankings)
- Paths (data, backups, analizy)
- Scoring weights (IV, Lucky, moves, CP)
- Tier thresholds (S/A+/A/B+/B/C/D per context)
- Raid DPS parameters (SE multiplier, shadow bonus, boss defense)
- PvP league caps (GL=1500, UL=2500, ML=unlimited)

## Building & Testing

```bash
make build      # Build binary to bin/pokemon
make test       # Run all tests (~200)
make build-test # Build + test
make clean      # Remove bin/pokemon
```

## Architecture

- **Go 1.26** — Single binary, no runtime deps
- **Modules**: `commands`, `config`, `model`, `ods`, `pkg/analysis`, `pkg/gamemaster`
- **Template**: `text/template` with custom functions (`pct`, `inc`, `tierOrDash`, etc.)
- **Tests**: Unit tests for all packages, table-driven

## License

MIT License — see [LICENSE](LICENSE)

## Data Sources

- **Game Master**: [PvPoke/gamemaster](https://github.com/pvpoke/pvpoke/tree/master/src/data)
- **Meta Rankings**: [PvPoke/rankings](https://github.com/pvpoke/pvpoke/tree/master/src/data/rankings)

---

*Built for Pokemon GO trainers who want data-driven decisions.*