# Todo

Prioritized list of commit-sized improvements.

## Priority 1 — Crash / Correctness Bugs

- [x] **Fix panic in `weather.go`**: Add bounds checks before accessing `weather.Daily.WeatherCode[1]`, temperature arrays at `[0]`/`[1]`, and `Hourly.PrecipitationProbability[24:]`.
- [x] **Fix panic in `chart.go`**: Guard `MaxValue()` against an empty slice before calling `slices.Max()`.
- [x] **Fix panic in `picture.go`**: Check `len(labels) > 0` before `rand.Int() % len(labels)`.
- [x] **Return errors from `Render()`**: Template parse/execute failures in `render.go` should be returned as errors, not just logged, so callers can detect a broken render.

## Priority 2 — Reliability

- [x] **Add HTTP timeouts**: Pass a `*http.Client` with a sensible timeout (e.g. 15s) to `http.Get` calls in `calendar.go` and `picture.go`.
- [x] **Surface failed cards to the user**: Log which card titles failed to load so the user knows what's missing, rather than silently dropping them.
- [x] **Handle `DevRender` listen error**: Return or log errors from `http.ListenAndServe` instead of swallowing them.
- [x] **Handle unknown weather codes**: Log a warning when a weather code isn't in `conditionToIcon`; use a default icon rather than an empty string.

## Priority 3 — Config & Validation

- [ ] **Implement `config.validate()`**: Validate non-empty calendar URLs, lat/lng in range, non-negative thresholds. Currently a no-op TODO stub.
- [ ] **Make viewport size configurable**: Add a `render.width`/`render.height` config field instead of hardcoding `1280x720` in `render.go`.
- [ ] **Make OpenAI model configurable**: Add a `generated.model` config field instead of hardcoding `GPT4o` in `generated.go`.
- [ ] **Compile attendee regexp at startup**: Move `regexp.MustCompile` out of the card loader closure in `calendar.go` — it's currently recompiled on every card load.

## Priority 4 — Code Quality

- [ ] **Fix error wrapping**: Replace `fmt.Errorf("...: %s", err)` with `fmt.Errorf("...: %w", err)` throughout so callers can use `errors.Is/As`.
- [ ] **Remove unused `Card.Footer` field**: `Footer` is declared in `card.go` and rendered in the template but never populated; remove or implement it.
- [ ] **Fix `DevRender` template path**: Use the embedded FS (like `Render()` does) instead of a relative file path `"internal/screen.go.html"` that breaks when the working directory is wrong.
- [ ] **Extract shared once-fetch pattern**: Weather, air quality, and calendar all repeat the same `sync.Once` + closure lazy-fetch pattern; extract a generic helper to reduce duplication.

## Priority 5 — Tests

- [ ] **Add unit tests for `chart.go`**: Pure functions `MaxValue()`, `Valid()`, `TopValue()`, and `Hours()` can be tested with no mocking.
- [ ] **Add unit tests for `config.go`**: Test `ReadConfig` with valid and invalid YAML, and `validate()` once it's implemented.
- [ ] **Add unit tests for `header.go`**: Test the French day/month formatting logic with a fixed `time.Time`.
- [ ] **Add integration test for `Render()`**: Use `fake.go` data sources with `httptest` to verify a PNG is produced without panicking.

## Priority 6 — Docs / Setup

- [ ] **Add a documented example `config.yaml`**: Expand `configs/config.yaml` with comments explaining every field, valid ranges, and example values.
- [ ] **Document locale/language hardcoding in README**: Note that card text is in French and explain where to change it if needed.
