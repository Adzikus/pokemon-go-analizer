# Contributing

Thanks for considering a contribution!

## How to Contribute

1. **Fork** the repository
2. **Create a branch** for your change (`git checkout -b feat/my-feature`)
3. **Make your changes** with tests
4. **Run checks**: `make build && make test`
5. **Submit a PR** with a clear description

## Code Style

- Go 1.26+
- Standard formatting: `gofmt` (enforced in CI)
- Run `go vet ./src/pokemon/...` before committing
- Table-driven tests preferred
- No external dependencies beyond `github.com/rs/zerolog`

## Commit Messages

Use conventional commits:
```
feat: add new raid DPS algorithm
fix: correct CP cap calculation for PvP
docs: update README with new commands
test: add coverage for lucky target scoring
refactor: split analysis package
```

## Testing

```bash
# All tests
make test

# Single package
go test ./src/pokemon/pkg/analysis/...

# With coverage
go test -cover ./src/pokemon/...
```

## Architecture Notes

- **ODS is source of truth** — never CSV
- **Inline LO formula `cell-content-is-in-list("X")` does NOT work** — use cell range references
- **Candy = species family** (Bulbasaur line only), not type
- **Freeze panes**: LO uses `HorizontalSplitMode=2` + position
- **Sheet protection**: Disabled (prevents entry deletion)

## Adding New Analysis Sections

1. Add logic in `pkg/analysis/`
2. Extend `ReportData` in `commands/analyze.go`
3. Update `templates/report.md.tmpl`
4. Update `docs/05-analiza.md` with diagrams

## Updating External Data

```bash
# Regenerate gamemaster.json
pokemon update-gamemaster --force

# Regenerate meta rankings
pokemon update-meta --force
```

## Questions?

Open an issue or discussion.