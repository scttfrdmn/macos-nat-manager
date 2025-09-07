#!/bin/bash
#
# Pre-commit git hook for macOS NAT Manager
# This script runs various checks before allowing a commit
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  ${1}${NC}"
}

log_success() {
    echo -e "${GREEN}✅ ${1}${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  ${1}${NC}"
}

log_error() {
    echo -e "${RED}❌ ${1}${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 ${1}${NC}"
}

# Exit on any failure
exit_on_fail() {
    if [[ $? -ne 0 ]]; then
        log_error "Pre-commit hook failed!"
        log_info "Fix the issues above and try again."
        exit 1
    fi
}

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    log_error "Not in a git repository"
    exit 1
fi

log_info "Running pre-commit checks for macOS NAT Manager..."

# Check 1: Go formatting
log_step "Checking Go formatting..."
unformatted=$(gofmt -l . | grep -v vendor/ || true)
if [[ -n "$unformatted" ]]; then
    log_error "Some files are not properly formatted:"
    for file in $unformatted; do
        echo "  $file"
    done
    log_info "Run 'make fmt' to fix formatting issues"
    exit 1
else
    log_success "All Go files are properly formatted"
fi

# Check 2: Go mod tidy
log_step "Checking go.mod and go.sum..."
if command -v go &> /dev/null; then
    go mod tidy
    if ! git diff --quiet go.mod go.sum; then
        log_error "go.mod or go.sum needs to be updated"
        log_info "Run 'go mod tidy' and commit the changes"
        git checkout go.mod go.sum  # Reset changes
        exit 1
    else
        log_success "go.mod and go.sum are up to date"
    fi
else
    log_warning "Go not found, skipping mod check"
fi

# Check 3: Go vet
log_step "Running go vet..."
if command -v go &> /dev/null; then
    if go vet ./...; then
        log_success "go vet passed"
    else
        log_error "go vet found issues"
        exit 1
    fi
else
    log_warning "Go not found, skipping vet check"
fi

# Check 4: Tests
log_step "Running tests..."
if command -v go &> /dev/null; then
    if go test -short ./...; then
        log_success "Tests passed"
    else
        log_error "Tests failed"
        exit 1
    fi
else
    log_warning "Go not found, skipping tests"
fi

# Check 5: Build
log_step "Testing build..."
if command -v go &> /dev/null; then
    if make build > /dev/null 2>&1; then
        log_success "Build successful"
        # Clean up build artifact
        rm -f nat-manager
    else
        log_error "Build failed"
        exit 1
    fi
else
    log_warning "Go not found, skipping build check"
fi

# Check 6: Lint (if available)
log_step "Running linters..."
if command -v golint &> /dev/null; then
    lint_output=$(golint ./... | grep -v vendor/ || true)
    if [[ -n "$lint_output" ]]; then
        log_warning "Linter found some issues:"
        echo "$lint_output"
        # Don't fail on lint issues, just warn
    else
        log_success "Linter checks passed"
    fi
else
    log_info "golint not available, skipping lint check"
fi

# Check 7: Secrets/sensitive data
log_step "Checking for sensitive data..."
sensitive_files=$(git diff --cached --name-only | grep -E '\.(key|pem|p12|pfx|p8)$' || true)
if [[ -n "$sensitive_files" ]]; then
    log_error "Attempting to commit sensitive files:"
    for file in $sensitive_files; do
        echo "  $file"
    done
    log_info "Remove sensitive files or add to .gitignore"
    exit 1
fi

# Check for common secrets patterns
if git diff --cached | grep -qE '(password|secret|token|key|auth).*=.*[a-zA-Z0-9]{8,}'; then
    log_warning "Possible secrets detected in commit"
    log_info "Please review your changes for any sensitive information"
    # Don't fail, but warn the user
fi

log_success "Sensitive data check passed"

# Check 8: Documentation
log_step "Checking documentation..."
if git diff --cached --name-only | grep -qE '\.(go)$'; then
    # Check if any public functions lack documentation
    missing_docs=$(go doc ./... 2>/dev/null | grep -c "exported .* should have comment" || true)
    if [[ "$missing_docs" -gt 0 ]]; then
        log_warning "Some exported functions may lack documentation"
        log_info "Consider adding documentation for better code quality"
    fi
fi

# Check if README or other docs need updating
if git diff --cached --name-only | grep -qE '(cmd/|internal/).*\.go$'; then
    if ! git diff --cached --name-only | grep -qE 'README\.md|CHANGELOG\.md'; then
        log_info "Go files modified but no documentation updated"
        log_info "Consider updating README.md or CHANGELOG.md if needed"
    fi
fi

log_success "Documentation check completed"

# Check 9: Commit message format (if available)
if [[ -n "${GIT_COMMIT_MESSAGE:-}" ]]; then
    log_step "Checking commit message format..."
    
    # Check for conventional commit format
    if echo "$GIT_COMMIT_MESSAGE" | grep -qE '^(feat|fix|docs|style|refactor|test|chore|build|ci)(\(.+\))?: .{1,50}'; then
        log_success "Commit message follows conventional format"
    else
        log_warning "Commit message doesn't follow conventional format"
        log_info "Consider using: type(scope): description"
        log_info "Types: feat, fix, docs, style, refactor, test, chore, build, ci"
    fi
fi

# Check 10: File permissions
log_step "Checking file permissions..."
executable_files=$(git diff --cached --name-only | xargs -I {} find {} -perm +111 -type f 2>/dev/null | grep -v '\.sh$' | grep -v scripts/ || true)
if [[ -n "$executable_files" ]]; then
    log_warning "Non-script files with executable permissions:"
    for file in $executable_files; do
        echo "  $file"
    done
    log_info "Consider removing execute permissions if not needed"
fi

log_success "File permissions check completed"

# Final success message
log_success "All pre-commit checks passed! 🎉"
log_info "Commit is ready to proceed"

exit 0