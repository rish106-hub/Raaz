# Contributing to Raaz

> **Raaz** is a product built under [Arthakram](https://github.com/Arthakram) — the Product & Consulting Club at Rishihood University. Contributions follow a structured, GSoC-inspired process to maintain quality and fairness.

---

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Who Can Contribute](#who-can-contribute)
3. [Contribution Process (GSoC-Style)](#contribution-process-gsoc-style)
4. [Application & Selection](#application--selection)
5. [Project Ideas & Issues](#project-ideas--issues)
6. [Development Workflow](#development-workflow)
7. [Branch Naming Convention](#branch-naming-convention)
8. [Commit Message Standard](#commit-message-standard)
9. [Pull Request Rules](#pull-request-rules)
10. [Code Review Standards](#code-review-standards)
11. [Communication & Reporting](#communication--reporting)
12. [Mentors & Maintainers](#mentors--maintainers)

---

## Code of Conduct

All contributors must adhere to our [Code of Conduct](CODE_OF_CONDUCT.md). Violations will result in permanent ban from the project.

---

## Who Can Contribute

Raaz accepts contributions from:

- **Arthakram members** — direct contributors with Arthakram membership
- **External contributors** — anyone who completes the formal application process below

> ⚠️ **No unsolicited PRs.** All contributions require prior issue assignment and (for external contributors) a completed application. PRs without linked issues will be closed.

---

## Contribution Process (GSoC-Style)

Raaz follows a structured contribution model inspired by Google Summer of Code. Think of it as a mini-fellowship:

```
Explore Issues → Submit Application → Get Selected → Community Bonding → Coding Period → Review → Merge
```

### Phase 0: Explore (Always Open)
- Browse [open issues](https://github.com/rish106-hub/Raaz/issues) labelled `good first issue` or `help wanted`
- Read the codebase, ask questions in [Discussions](https://github.com/rish106-hub/Raaz/discussions)
- Fix typos or docs via small PRs (no application needed for `documentation` label issues)

### Phase 1: Application
- Find an issue or idea you want to work on
- Open a **Contributor Application** using the [application template](.github/ISSUE_TEMPLATE/contributor_application.md)
- Wait for maintainer review (SLA: 5 business days)

### Phase 2: Community Bonding
- Selected contributors are assigned the issue
- Schedule a 30-minute sync with your mentor
- Read all architecture docs and ask questions in the issue thread

### Phase 3: Coding Period
- Work on your feature/fix in a dedicated branch
- Open a Draft PR early (within 3 days of starting)
- Post weekly progress updates in your PR description

### Phase 4: Review & Merge
- Request review once all checks pass
- Address all review comments within 5 business days
- Maintainer merges after 2 approvals

---

## Application & Selection

Fill out the [Contributor Application issue template](.github/ISSUE_TEMPLATE/contributor_application.md) with:

| Field | Description |
|-------|-------------|
| **Your Background** | Brief intro, links to past work |
| **Issue/Feature** | Which issue or feature you want to tackle |
| **Proposal** | Your approach, timeline, and deliverables |
| **Availability** | Hours/week you can commit |
| **Questions** | Anything unclear about the codebase |

Selection criteria:
- Quality of proposal (40%)
- Demonstrated understanding of the codebase (30%)
- Past work / portfolio (20%)
- Communication quality (10%)

---

## Project Ideas & Issues

Issues are labelled by type and difficulty:

| Label | Meaning |
|-------|---------|
| `good first issue` | Ideal for first-time contributors |
| `help wanted` | Maintainers want community help |
| `gsoc-idea` | Larger scope, suitable for fellowship-style contribution |
| `bug` | Something is broken |
| `feature` | New capability |
| `documentation` | Docs, README, comments |
| `security` | Security-related — see SECURITY.md before touching |
| `blocked` | Waiting on another issue/PR |
| `in-progress` | Someone is actively working on this |

---

## Development Workflow

### Prerequisites
- Android Studio Hedgehog (2023.1.1) or newer
- Go 1.22+
- Docker & Docker Compose
- Node.js 20+
- PostgreSQL 15+

### Setup
```bash
git clone https://github.com/rish106-hub/Raaz.git
cd Raaz
cp .env.example .env
# Fill in your local env values
docker compose up -d
```

### Running Tests
```bash
# Backend (Go)
cd server && go test ./...

# Android lint
./gradlew lint

# Android unit tests
./gradlew testDebugUnitTest
```

---

## Branch Naming Convention

```
<type>/<short-description>
```

| Type | Use For |
|------|---------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation only |
| `chore/` | Build, CI, tooling |
| `security/` | Security patches |
| `refactor/` | Code restructure (no behaviour change) |
| `test/` | Tests only |

**Examples:**
```
feat/voice-mode-ui
fix/matching-engine-timeout
docs/api-reference
security/rotate-jwt-keys
```

---

## Commit Message Standard

We follow the **Conventional Commits** specification:

```
<type>(scope): <short description>

[optional body]

[optional footer: Closes #issue-number]
```

**Rules:**
- Use imperative mood: "add feature" not "added feature"
- Keep the subject line under 72 characters
- Reference the issue in the footer: `Closes #42`
- Never commit secrets, tokens, or PII

**Examples:**
```
feat(matching): add city-based fallback after 15-minute timeout

Closes #34

fix(auth): prevent JWT replay attacks via nonce validation

Closes #87

docs(readme): update setup instructions for M1 Mac
```

---

## Pull Request Rules

1. **One PR per issue** — never bundle unrelated changes
2. **Draft PR first** — open a Draft PR within 3 days of starting work
3. **All CI checks must pass** before requesting review
4. **Fill the PR template completely** — incomplete PRs will be closed
5. **Link the issue** — use `Closes #<issue-number>` in the PR body
6. **No force-pushes to `main`** — ever
7. **Squash commits** before merge (maintainer will squash)
8. **Two approvals required** to merge into `main`
9. **Self-reviews don't count** — you cannot approve your own PR
10. **Respond to review comments within 5 business days** or the PR will be closed

---

## Code Review Standards

As a reviewer:
- Be constructive, specific, and kind
- Use suggestion blocks (`` ```suggestion ``) for direct code fixes
- Batch your comments — use "Start a review" not individual comments
- Approve only when you are genuinely satisfied
- Label blocking comments `[BLOCKING]` and non-blocking ones `[NIT]`

As an author:
- Respond to every comment — even if you disagree, explain why
- Don't resolve others' comments — let the reviewer resolve
- Mark the PR as `ready for review` only when it's truly ready

---

## Communication & Reporting

| Channel | Purpose |
|---------|---------|
| [GitHub Issues](https://github.com/rish106-hub/Raaz/issues) | Bug reports, feature requests, contributor applications |
| [GitHub Discussions](https://github.com/rish106-hub/Raaz/discussions) | Architecture questions, design discussions, general Q&A |
| [Pull Requests](https://github.com/rish106-hub/Raaz/pulls) | Code review and contribution |
| Email: `arthakram@rishihood.edu.in` | Private / sensitive communications |

**Response SLAs:**
- Issue triage: 5 business days
- PR review: 7 business days
- Security reports: 48 hours (see [SECURITY.md](SECURITY.md))

---

## Mentors & Maintainers

| Role | GitHub | Responsibility |
|------|--------|---------------|
| Lead Maintainer | [@rish106-hub](https://github.com/rish106-hub) | Architecture, final merge approval |
| Org Sponsor | [Arthakram](https://github.com/Arthakram) | Product direction, Arthakram alignment |

---

## Recognition

All merged contributors are:
- Listed in `CONTRIBUTORS.md`
- Credited in release notes
- Eligible for an Arthakram contributor certificate

---

*Built under [Arthakram](https://github.com/Arthakram) · Rishihood University*
