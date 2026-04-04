# Contributing to Wallpaper CLI

Thank you for your interest in contributing! This document provides guidelines for contributing to this project.

## 🛡️ Repository Protection Rules

This repository has the following protections in place:

### Branch Protection (master/main)
- ✅ **Require pull request reviews** before merging
- ✅ **Require status checks** to pass (if CI is enabled)
- ✅ **Require conversation resolution** before merging
- ✅ **Require signed commits** (optional)
- ✅ **Require linear history** (no merge commits)
- ✅ **Include administrators** (rules apply to everyone)

### Code Review Requirements
- All changes must be made via **Pull Request**
- **1 approving review** required from code owner
- **CODEOWNERS** file defines mandatory reviewers

## 📝 How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/aurablacklight/wallpaper-cli/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Open a new issue with the "feature request" label
2. Describe the feature and its use case
3. Wait for maintainer feedback before implementing

### Pull Request Process

1. **Fork** the repository
2. **Create a branch** from `master`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes**:
   - Follow Go best practices
   - Add tests for new functionality
   - Update documentation as needed
4. **Commit** with clear messages:
   ```bash
   git commit -m "feat: add new sorting option --most-downloaded"
   ```
5. **Push** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
6. **Open a Pull Request** against `master`
   - Fill out the PR template
   - Link related issues
   - Request review from @aurablacklight

## 🎨 Code Standards

### Go Code Style
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `go fmt` before committing
- Run `go vet` to catch common mistakes
- Keep functions small and focused
- Add comments for exported functions

### Commit Message Format
We use conventional commits:
```
feat: add new feature
fix: fix bug in download manager
docs: update README
refactor: reorganize source adapters
test: add tests for deduplication
chore: update dependencies
```

### Testing
- Write tests for new functionality
- Ensure all tests pass before PR:
  ```bash
  make test
  ```
- Maintain or improve code coverage

## 🔒 Security

- Never commit secrets or API keys
- Report security vulnerabilities privately to the maintainer
- Be cautious when adding new dependencies

## 📋 PR Checklist

Before submitting:
- [ ] Code compiles (`go build`)
- [ ] Tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt`)
- [ ] No vet errors (`go vet`)
- [ ] Documentation updated (if needed)
- [ ] README updated (if needed)
- [ ] Commit messages are clear

## 🚫 What NOT to Do

- ❌ Push directly to `master`
- ❌ Merge your own PR without review
- ❌ Add large binary files to the repo
- ❌ Break backward compatibility without discussion
- ❌ Remove existing tests without justification

## 🆘 Getting Help

- Open an issue for questions
- Check existing documentation first
- Be respectful and patient

## 📜 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for helping make Wallpaper CLI better!** 🎉
