# Contributing to Nopher

Thank you for your interest in contributing to Nopher!

## For AI Agents

If you're an AI agent (like Claude) working on this project, please read [AGENTS.md](AGENTS.md) first. It contains detailed instructions on:
- How to work with the memory/ documentation
- Code quality standards and expectations
- Phase-based implementation approach
- When and how to update documentation

## For Human Contributors

### Getting Started

1. **Fork and clone** the repository
2. **Read the documentation** in `memory/` to understand the architecture
3. **Check the issues** for tasks labeled "good first issue"
4. **Set up your development environment**

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/nopher.git
cd nopher

# Install dependencies
go mod download

# Run tests to verify setup
make test

# Run linters
make lint
```

### Code Standards

#### Quality Guidelines

- **File size**: Keep files under 500 lines (target 100-300 lines)
- **Function size**: Keep functions under 50 lines (target 10-30 lines)
- **DRY principle**: Don't repeat yourself - extract common logic
- **Clear naming**: Use descriptive names for files, functions, and variables
- **Single responsibility**: Each file/function should do one thing well

See `AGENTS.md` for detailed code quality examples.

#### Go Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (run `make fmt`)
- Write tests for new functionality
- Keep test coverage above 80%
- Add comments for exported functions/types

### Workflow

1. **Create a branch** for your work:
   ```bash
   git checkout -b feat/your-feature-name
   ```

2. **Make your changes** following the code standards

3. **Write tests** for new functionality

4. **Run checks locally**:
   ```bash
   make check  # Runs lint + test
   ```

5. **Commit your changes** using conventional commits:
   ```bash
   git commit -m "feat(gopher): add thread navigation"
   ```

6. **Push and create a pull request**

### Commit Message Format

We use conventional commits for automated changelog generation:

- `feat(scope):` - New feature
- `fix(scope):` - Bug fix
- `perf(scope):` - Performance improvement
- `docs(scope):` - Documentation changes
- `test(scope):` - Test additions/changes
- `chore(scope):` - Build/tooling changes
- `refactor(scope):` - Code refactoring

Examples:
```
feat(gemini): add TLS certificate generation
fix(gopher): correct selector escaping
perf(cache): optimize gophermap rendering
docs(readme): update installation instructions
```

### Pull Request Process

1. **Ensure all checks pass** (tests, lints, builds)
2. **Update documentation** if you changed functionality
3. **Reference any related issues** in the PR description
4. **Wait for review** from maintainers
5. **Address feedback** if requested

### What to Contribute

#### Good First Issues

Look for issues labeled:
- `good first issue` - Great for newcomers
- `help wanted` - Maintainers need assistance
- `documentation` - Docs improvements

#### Priority Areas

Based on current phase (see `memory/PHASES.md`):
- Configuration system (Phase 1)
- Storage layer implementation (Phase 2)
- Protocol server implementations (Phases 7-9)
- Testing and documentation (Phase 15)

#### Not Accepted

- Changes that violate the design in `memory/` (discuss first!)
- Breaking changes without discussion
- Adding dependencies without justification
- Code that doesn't meet quality standards

### Updating Documentation

If your changes affect design decisions:

1. **Read the current memory/ docs** relevant to your change
2. **Update memory/ files** to reflect the new design
3. **Keep the style consistent** with existing docs
4. **Update PHASES.md** if deliverables change

See `AGENTS.md` "Working with Memory" section for details.

### Testing

#### Running Tests

```bash
# Run all tests
make test

# Run with coverage report
HTML_COVERAGE=true make test

# Run specific package tests
go test ./internal/config/...

# Run with race detector
go test -race ./...
```

#### Writing Tests

- Use table-driven tests where appropriate
- Test happy paths and error cases
- Use descriptive test names
- Keep tests simple and focused

Example:
```go
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {
            name:  "valid config",
            input: "testdata/valid.yaml",
            want:  &Config{...},
        },
        {
            name:    "missing file",
            input:   "testdata/missing.yaml",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Getting Help

- **Questions?** Open a discussion on GitHub
- **Bug reports?** Open an issue with reproduction steps
- **Design discussions?** Reference memory/ docs in your proposal

### Code of Conduct

- Be respectful and professional
- Focus on the code, not the person
- Accept constructive feedback gracefully
- Help others learn and grow

### License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Nopher! 🎉
