# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
