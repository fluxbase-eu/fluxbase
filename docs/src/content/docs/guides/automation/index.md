---
title: Issue Automation
description: Automate GitHub issue handling with Claude AI
---

# Issue Automation

Fluxbase includes a comprehensive system for automating GitHub issue handling using Claude AI. This allows you to automatically analyze issues, implement fixes, and create pull requests with minimal human intervention.

## Overview

The automation system consists of three main components:

1. **Issue Templates** - Structured templates that capture the information Claude needs
2. **GitHub Actions Workflow** - Triggers Claude to fix issues labeled with `claude-fix`
3. **MCP GitHub Tools** - Programmatic access to GitHub from within Claude sessions

## Quick Start

### 1. Enable the Automation

First, ensure you have the following GitHub secrets configured in your repository:

- `ANTHROPIC_API_KEY` - Your Anthropic API key for Claude access

### 2. Create an Issue

Use one of the issue templates:

- **Bug Report** - For reporting bugs
- **Feature Request** - For suggesting new features
- **Claude Fix Request** - Specifically designed for automated fixes

### 3. Trigger the Fix

Add the `claude-fix` label to the issue. The automation will:

1. Create a new branch (`claude-fix/issue-{number}-{timestamp}`)
2. Analyze the issue and codebase
3. Implement the fix
4. Create a PR for review
5. Comment on the issue with the PR link

## When to Use Claude Fix

The `claude-fix` label works best for:

- Bug fixes with clear reproduction steps
- Small refactoring tasks
- Documentation updates
- Test additions
- Code style/lint fixes
- Type safety improvements

It's not recommended for:

- Large architectural changes
- Security-sensitive code
- Breaking API changes
- Features requiring product decisions

## Components

### Issue Templates

Located in `.github/ISSUE_TEMPLATE/`:

| Template | Purpose |
|----------|---------|
| `bug_report.yml` | Standard bug reports with severity and component |
| `feature_request.yml` | Feature suggestions with priority |
| `claude_fix.yml` | Structured format for automated fixes |

### GitHub Actions Workflow

The `.github/workflows/claude-fix.yml` workflow:

- **Triggers**: Issue labeled with `claude-fix`, or manual dispatch
- **Actions**: Creates branch, runs Claude Code, creates PR
- **Comments**: Updates issue with progress and results

### MCP GitHub Tools

Available through the MCP server:

| Tool | Scope | Description |
|------|-------|-------------|
| `list_github_issues` | `github:read` | List issues with filtering |
| `get_github_issue` | `github:read` | Get issue details |
| `create_github_issue` | `github:write` | Create new issues |
| `create_github_issue_comment` | `github:write` | Add comments |
| `update_github_issue_labels` | `github:write` | Manage labels |
| `trigger_claude_fix` | `github:write` | Trigger the fix workflow |

## Configuration

### Repository Secrets

Required secrets for the automation:

```yaml
# GitHub Actions secrets
ANTHROPIC_API_KEY: sk-ant-...  # Required for Claude API access
```

### Webhook Configuration

To receive issue events in your Fluxbase server:

```bash
# Configure the webhook endpoint
POST /api/v1/admin/branches/github/configs
{
  "repository": "owner/repo",
  "webhook_secret": "your-webhook-secret"
}
```

Then in GitHub, add a webhook pointing to:
```
https://your-fluxbase-server.com/api/v1/webhooks/github
```

With events:
- Issues
- Pull requests
- Push (optional)

## Best Practices

### Writing Good Fix Requests

For the best results with Claude Fix:

1. **Be specific** - Include file paths, function names, error messages
2. **Provide context** - Explain why the fix is needed
3. **Define acceptance criteria** - List what "done" looks like
4. **Include reproduction steps** - For bugs, show how to trigger the issue

### Example Fix Request

```markdown
## Task Description

Fix the nil pointer dereference in `internal/auth/service.go` in the
`ValidateToken` function when the token is expired.

**Files to Modify:**
- internal/auth/service.go
- internal/auth/service_test.go

**Acceptance Criteria:**
- [ ] No panic when token is expired
- [ ] Proper error returned for expired tokens
- [ ] Unit test added for this case
- [ ] All existing tests pass
```

## Monitoring

### Check Workflow Status

View workflow runs in GitHub Actions:

```
https://github.com/owner/repo/actions/workflows/claude-fix.yml
```

### Issue Comments

The automation adds comments to issues:

1. **Started** - When processing begins
2. **Complete** - With PR link when done
3. **Failed** - With error details if something goes wrong

## Troubleshooting

### Workflow Not Triggering

Check that:
- The `claude-fix` label exists in the repository
- The issue is in "open" state
- GitHub Actions are enabled
- The `ANTHROPIC_API_KEY` secret is set

### No Changes Made

Claude may not make changes if:
- The issue description is too vague
- The relevant code cannot be located
- The fix requires decisions beyond the issue scope

In this case, provide more specific details and re-run.

### PR Has Issues

If the generated PR doesn't fully solve the problem:
- Add comments on the PR with specific feedback
- Close the PR and create a new issue with more details
- Fix manually and reference the original issue
