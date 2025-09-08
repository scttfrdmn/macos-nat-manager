# Claude Code Guidelines for macOS NAT Manager

This document contains specific guidelines and requirements for AI assistants (particularly Claude Code) working on the macOS NAT Manager project.

## 🎯 Core Project Requirements

### Code Quality Standards - **MANDATORY**

**Go Report Card Grade A+ Requirement:**
- The project MUST maintain a Go Report Card grade of **A+ (≥90%)** at all times
- **NO exceptions, NO workarounds, NO compromises**
- Every commit must pass all Go Report Card equivalent checks:
  - `gofmt` - 100% code formatting compliance
  - `go vet` - Zero static analysis issues
  - `golint` - Zero style violations (warnings may be acceptable if they don't affect grade)
  - `gocyclo` - All functions MUST have cyclomatic complexity ≤15
  - `misspell` - Zero spelling errors
  - `ineffassign` - Zero ineffectual assignments

**Pre-commit Hook Enforcement:**
- The repository includes a mandatory pre-commit hook that blocks commits below grade A
- This hook MUST be used and MUST NOT be bypassed
- If the hook fails, the issues MUST be fixed before committing
- Never suggest using `--no-verify` to bypass quality checks

**Testing Requirements:**
- All tests MUST pass before any commit
- New functionality MUST include appropriate tests
- Test coverage should be maintained and improved where possible

## 🛠️ Development Guidelines

### Code Modifications

When making changes to the codebase:

1. **Always run quality checks first:** Use the pre-commit hook or run individual tools
2. **Refactor complex functions:** If cyclomatic complexity exceeds 15, refactor into smaller functions
3. **Handle all errors:** Use `_ = cmd.Run()` pattern for commands where errors are intentionally ignored
4. **Follow Go conventions:** Use proper naming, documentation, and package organization
5. **Maintain consistency:** Follow existing patterns and architectural decisions

### Architecture Principles

- **Separation of concerns:** Keep CLI, TUI, and NAT logic separate
- **Interface-based design:** Use interfaces for testability
- **Configuration management:** Use structured configuration with validation
- **Error handling:** Provide clear, actionable error messages
- **Clean shutdown:** Always implement proper cleanup and resource management

### Commands to Run

Always verify quality before committing:

```bash
# Run pre-commit hook manually
.git/hooks/pre-commit

# Individual quality checks
go fmt ./...
go vet ./...
go test ./...
$(go env GOPATH)/bin/gocyclo -over 15 .
$(go env GOPATH)/bin/ineffassign ./...
$(go env GOPATH)/bin/misspell .
```

## 🚫 Prohibited Actions

**Never:**
- Commit code that fails Go Report Card standards
- Bypass the pre-commit hook with `--no-verify`
- Leave functions with cyclomatic complexity >15
- Ignore static analysis warnings from `go vet`
- Create ineffectual assignments
- Leave spelling errors in code or comments
- Compromise on code quality for "quick fixes"

**Quality is non-negotiable.**

## 🔧 Tools Installation

Required tools for quality checking:

```bash
go install golang.org/x/lint/golint@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest  
go install github.com/client9/misspell/cmd/misspell@latest
go install github.com/gordonklaus/ineffassign@latest
```

## 📋 Checklist for Claude Code

Before making any commit, verify:

- [ ] All Go files are properly formatted (`gofmt`)
- [ ] No static analysis issues (`go vet`)
- [ ] No linting issues that affect the grade (`golint`)
- [ ] All functions have complexity ≤15 (`gocyclo`)
- [ ] No spelling errors (`misspell`)
- [ ] No ineffectual assignments (`ineffassign`)
- [ ] All tests pass (`go test ./...`)
- [ ] Pre-commit hook passes (Grade A+ achieved)

## 🎯 Success Metrics

- **Go Report Card Grade:** A+ (90-100%)
- **Build Status:** ✅ All builds successful
- **Test Coverage:** Maintained or improved
- **Complexity:** All functions ≤15 cyclomatic complexity
- **Lint Status:** Zero blocking issues

---

**Remember: Quality is not optional. The A+ grade requirement is absolute and must be maintained at all costs.**