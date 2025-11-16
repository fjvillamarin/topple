Analyze CodeRabbit review comments from a GitHub PR and create detailed issue reports.

# Arguments

- `PR_NUMBER` (required) - The GitHub PR number to analyze
- `OUTPUT_DIR` (optional) - Output directory for review files (default: `tmp/review/`)

# Usage

```bash
/analyze-coderabbit 84
/analyze-coderabbit 84 docs/review/pr-84/
```

# Instructions

You are tasked with analyzing CodeRabbit AI review comments from a GitHub Pull Request and creating detailed, actionable issue reports.

## Step 1: Setup and Cleanup

1. Parse arguments:
   - Extract `PR_NUMBER` (required)
   - Extract `OUTPUT_DIR` (default: `tmp/review/`)

2. Clean output directory:
   ```bash
   rm -rf ${OUTPUT_DIR}/*
   mkdir -p ${OUTPUT_DIR}
   ```

## Step 2: Fetch CodeRabbit Comments

Use GitHub CLI to fetch review comments from the PR:

```bash
# Get PR review comments
gh api repos/fjvillamarin/topple/pulls/${PR_NUMBER}/comments --jq '.[] | select(.user.login == "coderabbitai[bot]") | {file: .path, line: .line, body: .body}'
```

Extract:
- File path
- Line number
- Comment body (contains severity, issue description, suggestions)

## Step 3: Parse and Deduplicate Issues

CodeRabbit may post duplicate comments. Deduplicate by:
- Unique combination of file + line + issue description
- Group related comments (same issue across multiple lines)

For each unique issue, extract:
- **Severity**: 🔴 Critical, 🟠 Major, 🟡 Minor (from comment tags)
- **Type**: Security, Bug, Error Handling, Test Quality, Documentation, Code Quality, Refactor
- **File/Location**: Path and line numbers
- **Original Comment**: Full CodeRabbit text
- **Suggested Fix**: Code diffs or recommendations from CodeRabbit

## Step 4: Analyze Each Issue

For each issue, create a detailed analysis:

### Template Structure:
```markdown
# Issue N: [Short Title]

**File**: `path/to/file.go:line`
**Severity**: [🔴 Critical | 🟠 Major | 🟡 Minor]
**Type**: [Category]

## Original CodeRabbit Comment

[Full comment text from CodeRabbit]

## Analysis

**Is this a real issue?** [✅ YES | ❌ NO | ❓ NEEDS INVESTIGATION]

[Detailed analysis covering:]
- Why this is/isn't a real issue
- Impact if not fixed
- Context from codebase knowledge
- Security implications (if applicable)
- Edge cases to consider

## Implementation Plan

[If it's a real issue, provide:]
1. Step-by-step fix instructions
2. Code examples (diffs)
3. Testing approach
4. Related files that may need updates

## Priority

**[HIGH | MEDIUM | LOW]** - [Justification]

[Explain:]
- When this should be fixed (before merge, cleanup phase, follow-up PR)
- Effort estimate
- Dependencies on other issues

## Recommendation

[Clear action item: Fix immediately | Fix before merge | Fix in cleanup | Defer to follow-up | Ignore]
```

## Step 5: Create Issue Files

For each unique issue:

1. Create file: `${OUTPUT_DIR}/NN-issue-slug.md`
   - `NN` = zero-padded issue number (01, 02, etc.)
   - `issue-slug` = kebab-case description

2. Write detailed analysis using template above

## Step 6: Create Summary File

Create `${OUTPUT_DIR}/00-SUMMARY.md`:

```markdown
# CodeRabbit PR #${PR_NUMBER} Review Summary

**Generated**: [Date]
**PR**: [PR title from GitHub]
**Reviewer**: CodeRabbit AI

## Overview

CodeRabbit identified **N unique issues** across PR #${PR_NUMBER}.

## Issues by Priority

### 🔴 Critical (N issues)
[List with links to individual files]

### 🟠 Major (N issues)
[List with links]

### 🟡 Minor (N issues)
[List with links]

## Recommended Action Plan

### Phase 1: Critical Issues (Before Merge)
[Must-fix items with effort estimates]

### Phase 2: Major Issues
[Should-fix items]

### Phase 3: Minor Issues
[Can-defer items]

## Issue Relationships

[Diagram or text showing related issues]

## Overall Assessment

**PR is [READY | NOT READY] to merge due to:**
[List blocking issues]

**Estimated effort to make merge-ready**: X hours

## Files Created

[List all MD files with brief descriptions]
```

## Step 7: Report Results

Output a summary:

```
✅ Analyzed CodeRabbit comments from PR #${PR_NUMBER}

📊 Summary:
- Total issues: N
- 🔴 Critical: N
- 🟠 Major: N
- 🟡 Minor: N

📁 Files created in ${OUTPUT_DIR}/:
- 00-SUMMARY.md (overview and action plan)
- 01-issue-name.md
- 02-issue-name.md
[... list all files]

🔍 Review the summary at: ${OUTPUT_DIR}/00-SUMMARY.md

⚠️  Critical/blocking issues:
[List critical items if any]

✅ Next steps: Review each issue file for detailed analysis and implementation plans.
```

## Important Notes

1. **Be Objective**: Not all CodeRabbit comments are real issues. Analyze each critically.

2. **Provide Context**: Use your knowledge of the codebase to assess whether issues are valid.

3. **Group Related Issues**: If multiple comments are about the same underlying problem, combine them.

4. **Actionable Recommendations**: Every issue should have clear next steps.

5. **Prioritize Correctly**:
   - 🔴 Critical: Security, data loss, breaking bugs
   - 🟠 Major: API design, error handling, functional bugs
   - 🟡 Minor: Style, test quality, documentation

6. **Estimate Effort**: Help the user understand time investment needed.

7. **Consider Dependencies**: Note when issues are related or should be fixed together.

## Error Handling

- If `gh` CLI is not available: Error and suggest installation
- If PR not found: Error with clear message
- If no CodeRabbit comments: Report "No CodeRabbit comments found"
- If API rate limit hit: Suggest waiting or using authentication

## Example Output Structure

```
tmp/review/
├── 00-SUMMARY.md
├── 01-security-path-traversal.md
├── 02-error-handling-inconsistent.md
├── 03-test-format-verbs.md
├── 04-dead-code-errors-field.md
└── 05-docs-outdated-paths.md
```
