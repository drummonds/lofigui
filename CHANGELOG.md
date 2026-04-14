# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.17.34] - 2026-04-14

 - Add App.RunWASM() for service worker WASM and switch example 01 to use it

## [0.17.33] - 2026-04-13

 - Add post:release task to commit rebuilt docs assets

## [0.17.32] - 2026-04-13

 - Split Python Jinja2 templates from Go html/template templates

## [0.17.31] - 2026-04-13

 - Merge main into task/WTteletype

## [0.17.30] - 2026-04-06

 - Simplify WASM examples to use RunWASM and unify demo templates

## [0.17.29] - 2026-04-05

 - Release prep

## [0.17.28] - 2026-04-04

 - New exxmaple 02

## [0.17.27] - 2026-04-02

 - Updating example 01

## [0.17.26] - 2026-03-30

 - Still cleaning example01

## [0.17.25] - 2026-03-29

 - Updating docs

## [0.17.24] - 2026-03-29

 - Simplify example 01 with Handle, lazy controller, auto-flush and graceful shutdown

## [0.17.23] - 2026-03-26

 - Add lofigui.Yield() for WASM-friendly cooperative scheduling (issue #4)

## [0.17.22] - 2026-03-12

 - Decouple Python App from FastAPI, add docs index

## [0.17.21] - 2026-03-12

 - Update CHANGELOG for v0.17.20

## [0.17.20] - 2026-03-12

 - Slimmming release

## [0.17.19] - 2026-03-12

 - Updating version

## [0.17.18] - 2026-03-12

 - Tidying docs and trying to get dual tag

## [0.17.17] - 2026-03-12

 - Just updating versions

## [0.17.16] - 2026-03-12

 - Updating readme and starting dual tagging

## [0.17.15] - 2026-03-07

 - Adding research

## [0.17.14] - 2026-03-07

 - Adding research paper

## [0.17.13] - 2026-03-06

 - Reising seaweed fs

## [0.17.12] - 2026-03-06

 - Adding example 11

## [0.17.11] - 2026-03-02

 - Adding example 10 docs

## [0.17.10] - 2026-03-02

 - Example 10 non blocking multi action

- Example 10: Water tank maintenance — background goroutines with progress, cancellation, and equipment lockout

## [0.17.6] - 2026-03-01

 - Updating documentation## [0.17.8] - 2026-03-01

 - Adding 04 and 09 and summary pages
## [0.17.9] - 2026-03-01

 - Adding better documentation

## [0.17.7] - 2026-03-01

 - Updating docs

## [0.17.5] - 2026-03-01

 - Adding HTMX versions

## [0.17.5] - 2026-02-28

- Example 09: Water tank with HTMX partial updates (no full-page polling)

## [0.17.4] - 2026-02-27

- working on WASM documentation of Water tank simulation

## [0.17.0] - 2026-02-26

- Go water tank simulator

## [0.16.1] - 2026-02-25

### Added
- Built-in Bulma 1.0.4 layout templates: LayoutSingle, LayoutNavbar, LayoutThreePanel (Go + Python)
- `NewControllerWithLayout()` convenience constructor
- Python `lofigui.LAYOUT_SINGLE`, `LAYOUT_NAVBAR`, `LAYOUT_THREE_PANEL` exports

## [0.15.1] - 2026-02-25

### Fixed
- Python: `reset()` now drains the queue (BUG 4 — stale items persisted across resets)
- Python: `PrintContext.__exit__` also drains queue on context manager exit
- Python: Fixed mutable default arguments `extra: dict = {}` in `App.state_dict()` and `App.template_response()`
- Python: Fixed example 05 to use `app_instance` API instead of removed `Controller` methods; replaced blocking `time.sleep` with `asyncio.sleep`

## [0.15.0] - 2026-02-25

### Changed
- **Breaking**: Go model function signature changed from `func(*App)` to `func(context.Context, *App)`
- `StartAction()` now returns a `context.Context` that is cancelled on `EndAction()` or when a new action starts
- `HandleRoot()` passes cancellable context to model goroutines
- Old goroutines are automatically cancelled when a new action starts (prevents stale goroutines)

## [0.14.0] - 2026-02-25

### Fixed
- Go: `StateDict()` deadlock — replaced nested `ControllerName()` call with inline lookup, added `defer` for unlock
- Go: `HandleDisplay()` now injects polling state (refresh meta tag, polling status) via `StateDict()` + `RenderTemplate()`
- Go: Extra context merge moved inside lock scope for safety

## [0.13.4] - 2026-02-15

- Project conformance: added ROADMAP.md, standard tasks (fmt, vet, check, clean), fixed CHANGELOG formatting
- Visual documentation: SVG captures of examples using url2svg, docs/UI_PATTERNS.md

## [0.13.2] - 2026-02-13

- Making more LLM compatible
 
## [0.13.0] - 2025-11-05

- Making auto favicon configurable 

## [0.12.0] - 2025-10-28

- Refactoring to make app a more centralised and opinionated controller with just 
model specific info in the controller.

## [0.11.0] - 2025-10-27

- Using taskfile in CI to use same commands

## [0.10.0] - 2025-10-27

- Adding tinygo example
- Adding polling state

## [0.9.0] - 2025-10-27

- Make changes to controller idempotent for both go and python
- adding tests and lint to precommit hook.

## [0.8.0] - 2025-10-27

- When change python controller, make app shut down previous action safely.

## [0.6.0] - 2025-10-23
 - Bug fixing and making sure if start up on after action action is taken.

## [0.5.0] - 2025-10-23
 - restructuring to put core controller function into lofigui
 - refactored example 01 to show this currently python only

## [0.4.0]

### Added
- Comprehensive test suite with pytest
- Type hints for all public functions and classes
- Docstrings for all public API functions
- HTML escaping by default in `print()` and `table()` functions with `escape` parameter
- Context manager support for `PrintContext` class
- Error handling and meaningful exceptions across all modules
- Buffer size warning system for `PrintContext`
- Development dependencies: pytest, pytest-cov, mypy, flake8
- CI/CD GitHub Actions workflow
- MIT LICENSE file
- Community files: CONTRIBUTING.md, CODE_OF_CONDUCT.md
- Improved .gitignore

### Changed
- Minimum Python version updated from 3.7 to 3.8 for better type hints support
- Updated all dependencies to latest versions
- Improved README with installation instructions and API documentation
- Fixed changelog filename from `changehistory,md` to `CHANGELOG.md`
- Enhanced example documentation

### Security
- Fixed XSS vulnerabilities by adding HTML escaping to all output functions
- Added explicit warnings about using raw HTML functions

## [0.2.3] - 2023-06-XX

### Changed
- Code reformatted with black
- Minor improvements and bug fixes

## [0.2.2] - 2023-06-XX

### Changed
- Package improvements

## [0.2.1] - 2023-06-XX

### Changed
- Minor updates

## [0.2.0] - 2023-06-XX

### Added
- Additional features and improvements

## [0.1.0] - 2023-06-08

### Added
- Initial release
- Basic print functionality
- Markdown and HTML rendering
- Table generation with Bulma CSS
- PrintContext for buffering
- Example applications (hello world, SVG graph)
- MVC architecture pattern
- FastAPI integration examples

[Unreleased]: https://github.com/drummonds/lofigui/compare/v0.2.3...HEAD
[0.2.3]: https://github.com/drummonds/lofigui/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/drummonds/lofigui/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/drummonds/lofigui/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/drummonds/lofigui/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/drummonds/lofigui/releases/tag/v0.1.0
