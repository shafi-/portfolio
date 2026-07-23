# Overengineering Cleanup Plan

## Priority: HIGH

### Phase 1: Delete Waste (Critical 🔴)

| File | Action | Lines Saved |
|------|--------|-------------|
| `internal/logging/logger_bench_test.go` | DELETE | 120 |
| `internal/cli/cli_test.go` | DELETE | 80 |
| `cmd/portfolio/main_test.go` | DELETE | 15 |
| `internal/config/config_test.go` | DELETE | 284 |
| `internal/logging/logger_test.go` | Reduce to 20 lines | 367 |

**Total: ~866 lines deleted**

**Rationale:**
- Benchmarks for I/O wrapper - pointless
- Tests verifying libraries work - not our job
- Logger wrapper tests - zap already tested

### Phase 2: CI Simplification

| File | Change |
|------|--------|
| `.github/workflows/ci.yml` | Remove coverage step, timeout 30s |
| `.github/workflows/security.yml` | Remove grep secrets (keep weekly scan) |

### Phase 3: Code Simplification (Moderate 🟡)

**Config:**
- Delete `validator.go` struct - use package functions
- Delete `defaults.go` Manager - call loader/validator directly
- Simplify `ConfigError` - remove unused Code field
- Merge `ValidationError` → `ConfigError`

**Logging:**
- Delete unused `Config.Components` map
- Remove `globalLogger` mutex - package vars adequate
- Delete `InitializeGlobalLogger` wrapper

**Models:**
- Delete `DatabaseInterface` - single implementation
- Delete `Result` struct - test-only dead code

### Phase 4: Test Reorganization

- Move acceptance criteria (AC05-AC13) from unit tests → E2E test suite
- Remove premature security checks (file permissions)
- Keep only: actual code behavior tests, not library tests

## Execution Order

1. **Phase 1** - Immediate line reduction
2. **Phase 2** - CI cost reduction
3. **Phase 3** - Code cleanup
4. **Phase 4** - Test structure (if E2E framework exists)

## Estimated Impact

- **Before:** ~1200 test lines, 40min CI potential
- **After:** ~200 test lines, <5min CI
- **Maintainability:** Focus on what matters
