# 🚀 CI/CD Pipeline Setup Guide

This document explains the GitHub Actions CI/CD pipeline implementation for Proyecto 1 Video Platform.

## 📁 Files Created

### GitHub Workflows

- `.github/workflows/pr-ci.yml` - **Pull Request CI/CD Pipeline**
- `.github/workflows/code-quality.yml` - **Advanced Code Quality Checks**
- `.github/workflows/main-ci.yml` - **Main Branch CI/CD**

### Configuration Files

- `.golangci.yml` - **Go Linting Configuration**
- `.github/PULL_REQUEST_TEMPLATE.md` - **PR Template**

## 🔄 Pipeline Overview

### 1. Pull Request Pipeline (`pr-ci.yml`)

**Triggers:** Pull requests to `main` or `develop` branches

**Jobs:**

- 🔍 **Code Quality & Linting** - Format checking, go vet, golangci-lint
- 🛡️ **Security Scanning** - Gosec, Nancy vulnerability scanner
- 🧪 **Unit Tests** - Complete test suite with coverage reporting
- 🏗️ **Build Applications** - Compile API and Worker binaries
- 🐳 **Docker Build** - Verify Docker image builds
- 🎯 **Integration Tests** - Optional (requires `integration-tests` label)
- 📋 **CI Status Summary** - Final status check

**Features:**

- ✅ **PostgreSQL service** for database tests
- ✅ **Coverage reporting** with PR comments
- ✅ **Artifact uploads** (binaries, coverage reports)
- ✅ **Security SARIF** uploads to GitHub Security tab
- ✅ **Codecov integration** for coverage tracking

### 2. Code Quality Pipeline (`code-quality.yml`)

**Triggers:** Pull requests and pushes to `main`

**Jobs:**

- 📊 **Dependency Analysis** - Vulnerability checks with govulncheck
- 📏 **Code Metrics** - Lines of code, complexity analysis
- 🔒 **Advanced Security** - CodeQL analysis, secret scanning

### 3. Main Branch Pipeline (`main-ci.yml`)

**Triggers:** Pushes to `main` branch

**Jobs:**

- 🧪 **Complete Test Suite** - Full tests with LocalStack
- 🏗️ **Build & Release** - Production binaries with checksums
- 🐳 **Docker Images** - Build and push to GitHub Container Registry
- 📊 **Quality Report** - Generate quality metrics

## ⚙️ Configuration

### Environment Variables

Set these secrets in your GitHub repository settings:

| Secret          | Description      | Required For                   |
| --------------- | ---------------- | ------------------------------ |
| `GITHUB_TOKEN`  | Auto-generated   | Docker registry, SARIF uploads |
| `CODECOV_TOKEN` | Codecov.io token | Coverage reporting (optional)  |

### Go Version

The pipeline uses **Go 1.23.12** as specified in your go.mod files.

### Linting Configuration

The `.golangci.yml` includes:

- **Enabled Linters:**

  - `errcheck` - Unchecked errors
  - `gosimple` - Code simplification
  - `govet` - Suspicious constructs
  - `staticcheck` - Advanced analysis
  - `gosec` - Security issues
  - `gofmt` - Formatting
  - `goimports` - Import formatting
  - And many more...

- **Quality Standards:**
  - Max line length: 120
  - Max cyclomatic complexity: 10
  - Max function length: 80 lines / 50 statements
  - Security issues: Medium severity

## 🚦 Pipeline Behavior

### Pull Request Workflow

1. **Draft PRs** - Skip all CI checks
2. **Ready PRs** - Run full pipeline
3. **Failed Jobs** - Block merge until fixed
4. **Coverage Comments** - Auto-posted on PRs
5. **Security Issues** - Uploaded to Security tab

### Branch Protection

Recommended branch protection rules for `main`:

```yaml
required_status_checks:
  strict: true
  contexts:
    - "🔍 Code Quality & Linting"
    - "🧪 Unit Tests"
    - "🏗️ Build Applications"
    - "🐳 Docker Build Verification"
enforce_admins: false
required_pull_request_reviews:
  required_approving_review_count: 1
  dismiss_stale_reviews: true
restrictions: null
```

## 🔧 Local Development

### Running Checks Locally

```bash
# Install golangci-lint
brew install golangci-lint

# Run linting
golangci-lint run ./api/... --config .golangci.yml
golangci-lint run ./worker/... --config .golangci.yml

# Run tests with coverage
make test-coverage

# Build applications
make build

# Format code
make lint
```

### Pre-commit Hooks (Recommended)

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go-fmt
        entry: gofmt
        language: system
        args: [-w, -s]
        files: \.go$

      - id: go-vet
        name: go-vet
        entry: bash -c 'cd api && go vet ./... && cd ../worker && go vet ./...'
        language: system
        files: \.go$
        pass_filenames: false

      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        files: \.go$
        args: [--config=.golangci.yml]
```

## 📊 Monitoring & Metrics

### Coverage Requirements

- **Minimum Coverage:** 80% for new code
- **Coverage Reports:** Generated as HTML artifacts
- **Coverage Trends:** Tracked in Codecov

### Build Artifacts

**Pull Requests:**

- `coverage-reports/` - HTML coverage reports
- `binaries/` - Compiled executables

**Main Branch:**

- `proyecto1-binaries/` - Production binaries with checksums
- `quality-summary/` - Quality metrics report

### Docker Images

Main branch builds push to:

- `ghcr.io/your-username/proyecto-1-cloud-development-uniandes-api:latest`
- `ghcr.io/your-username/proyecto-1-cloud-development-uniandes-worker:latest`

## 🔥 Troubleshooting

### Common Issues

#### 1. Linting Failures

```bash
# Fix formatting
make lint

# Check specific issues
golangci-lint run --config .golangci.yml ./...
```

#### 2. Test Failures

```bash
# Run tests locally
make test

# Run specific test
make test-specific TEST=TestName DIR=api
```

#### 3. Build Failures

```bash
# Ensure dependencies are up to date
cd api && go mod tidy
cd worker && go mod tidy

# Build locally
make build
```

#### 4. Docker Build Issues

```bash
# Test Docker builds locally
docker build -t test-api ./api
docker build -t test-worker ./worker
```

### Pipeline Debugging

- **View detailed logs** in GitHub Actions tab
- **Download artifacts** to inspect build outputs
- **Check Security tab** for security scan results
- **Review PR comments** for coverage information

## 🎯 Best Practices

### For Developers

1. **Run checks locally** before pushing
2. **Write meaningful tests** with good coverage
3. **Keep PRs small** and focused
4. **Use conventional commits** for clarity
5. **Address security issues** immediately

### For Pull Requests

1. **Fill out PR template** completely
2. **Add integration-tests label** if needed
3. **Ensure all checks pass** before requesting review
4. **Respond to coverage feedback** promptly
5. **Test manually** in addition to automated tests

### For Releases

1. **Main branch** always contains production-ready code
2. **Docker images** are automatically built and tagged
3. **Binaries** include checksums for verification
4. **Quality metrics** are tracked over time

## 🚀 Future Enhancements

### Planned Features

- [ ] **Deployment automation** to staging/production
- [ ] **Performance testing** integration
- [ ] **End-to-end testing** with Playwright
- [ ] **Dependency updates** automation with Dependabot
- [ ] **Release automation** with semantic versioning

### Monitoring Integration

- [ ] **Prometheus metrics** collection
- [ ] **Grafana dashboards** for CI/CD metrics
- [ ] **Slack notifications** for build status
- [ ] **SonarQube** integration for code quality

## 📚 References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Configuration](https://golangci-lint.run/usage/configuration/)
- [Codecov Documentation](https://docs.codecov.com/)
- [Docker Actions](https://github.com/docker/build-push-action)

---

**Last Updated:** Created with CI/CD pipeline implementation  
**Version:** 1.0.0  
**Maintainer:** Proyecto 1 Development Team
