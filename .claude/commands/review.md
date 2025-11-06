---
description: Run CodeRabbit CLI review for comprehensive code analysis and improvement suggestions
argument-hint: [--plain|--prompt-only]
allowed-tools: Bash(coderabbit:*)
---

Run CodeRabbit CLI review on the current codebase to get comprehensive code analysis.

**Execute CodeRabbit review:**
```bash
!coderabbit review ${ARGUMENTS:---plain}
```

After receiving the CodeRabbit analysis, please:

1. **Summarize Key Findings**
   - Highlight the most critical issues discovered
   - Group findings by severity (critical, high, medium, low)
   - Identify patterns or recurring problems

2. **Analyze Impact**
   - Assess which issues affect code quality most
   - Identify security vulnerabilities
   - Note performance concerns
   - Flag maintainability issues

3. **Provide Actionable Recommendations**
   - Prioritize fixes by impact and effort
   - Suggest specific code improvements
   - Reference best practices where applicable
   - Propose refactoring opportunities

4. **Create Implementation Plan**
   - Order fixes from highest to lowest priority
   - Estimate complexity of each fix
   - Suggest which issues can be batched together
   - Identify any that need immediate attention

5. **Apply Improvements** (if approved)
   - Implement suggested fixes systematically
   - Run tests after each change
   - Commit logical groups of related fixes

**Note:** If no arguments are provided, defaults to `--plain` for detailed feedback. Use `--prompt-only` for token-efficient minimal output.
