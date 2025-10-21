# Contributing to lofigui

Thank you for contributing to lofigui!

## How to Contribute

### Reporting Bugs

Before creating a bug report, check existing issues. Include:
- Clear title and description
- Steps to reproduce
- Expected vs actual behavior
- Python version and OS

### Pull Requests

1. Fork the repository
2. Create a branch from `main`
3. Make your changes with tests
4. Update documentation
5. Submit a pull request

## Development Setup

```bash
git clone https://github.com/YOUR-USERNAME/lofigui.git
cd lofigui
uv sync --all-extras
```

## Running Tests

```bash
uv run pytest
uv run pytest --cov=lofigui
```

## Code Quality

```bash
uv run black lofigui tests
uv run flake8 lofigui tests --max-line-length=100
uv run mypy lofigui
```

## Coding Standards

- Follow PEP 8
- Use type hints
- Write docstrings (Google style)
- Add tests for new features
- Target 80%+ code coverage

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
