# Sylfie Rebranding Plan

## Overview

This document outlines the plan for rebranding the project from "Biscuit" to "Sylfie", establishing PSX (Python Syntax eXtended) as the syntax name.

## Brand Identity

### Project Name: Sylfie
- **Pronunciation**: /ˈsɪlfi/ (SIL-fee)
- **Rationale**: Evokes "sylph" (an ethereal being), suggesting lightness and elegance in web UI creation
- **Tagline**: "Elegant Python UI Components"

### Syntax Name: PSX
- **Full Name**: Python Sylfie eXtension
- **File Extension**: `.psx` (already in use)
- **Analogy**: Sylfie:PSX :: React:JSX

## Rebranding Tasks

### Phase 1: Documentation Updates

1. **README.md**
   - Update project name from "Biscuit" to "Sylfie"
   - Update description to emphasize PSX syntax
   - Update all command examples to use `sylfie` CLI

2. **CLAUDE.md**
   - Update project overview section
   - Replace all "Biscuit" references with "Sylfie"
   - Clarify PSX as the syntax name

3. **Documentation Files** (`docs/`)
   - Update all markdown files
   - Update `index.md` with new branding
   - Update CLI documentation (`cli.md`)
   - Update architecture documentation

### Phase 2: Code Updates

1. **Binary Name**
   - Rename `biscuit` binary to `sylfie`
   - Update Makefile targets
   - Update build scripts

2. **Package References**
   - Update Go module name (if applicable)
   - Update import paths
   - Update package documentation

3. **CLI Commands**
   - Update command names: `biscuit` → `sylfie`
   - Update help text and descriptions
   - Maintain backward compatibility (optional)

4. **Test Files**
   - Update test descriptions
   - Update comments referencing "Biscuit"

### Phase 3: Repository Updates

1. **Repository Name** (if changing)
   - Consider renaming to `sylfie` or `sylfie-lang`
   - Update GitHub repository settings
   - Set up redirects from old repository

2. **Release Notes**
   - Prepare announcement for the rebrand
   - Explain the PSX syntax naming
   - Highlight that functionality remains the same

## Migration Guide for Users

### For Existing Users

```bash
# Old command
biscuit compile file.psx

# New command
sylfie compile file.psx
```

### Installation Update

```bash
# Remove old binary
rm $(which biscuit)

# Install new binary
go install github.com/fjvillamarin/sylfie/cmd/sylfie@latest
```

## Communication Strategy

### Announcement Template

```
🎉 Exciting News: Biscuit is now Sylfie!

We're rebranding to better reflect our vision for elegant Python UI components.

What's changing:
- Project name: Biscuit → Sylfie
- CLI command: biscuit → sylfie
- Our syntax is now officially called PSX (Python Syntax eXtended)

What's NOT changing:
- All your .psx files work exactly the same
- The syntax and features remain unchanged
- Full backward compatibility

Why the change?
- Clearer identity: Sylfie (project) uses PSX (syntax)
- Better alignment with industry standards (React/JSX pattern)
- More memorable and unique brand

Get started with Sylfie today!
```

## Implementation Timeline

### Week 1
- [ ] Update all documentation
- [ ] Create migration guide
- [ ] Prepare announcement

### Week 2
- [ ] Update codebase references
- [ ] Update build system
- [ ] Test all changes

### Week 3
- [ ] Create new releases
- [ ] Update repository (if renaming)
- [ ] Public announcement

## Technical Considerations

### Backward Compatibility

1. **Transition Period**
   - Consider supporting both `biscuit` and `sylfie` commands temporarily
   - Add deprecation warnings for `biscuit` command
   - Remove old command in future major version

2. **Import Paths**
   - If changing Go module name, consider using replace directives
   - Document import path changes clearly

### SEO and Discovery

1. **Keywords to Maintain**
   - "Python JSX alternative"
   - "PSX syntax"
   - "Python UI components"
   - "Python web framework"

2. **Documentation Updates**
   - Keep references to "formerly known as Biscuit"
   - Ensure search engines can find the project under both names

## Success Metrics

1. **User Adoption**
   - Monitor downloads of new `sylfie` binary
   - Track GitHub stars/forks after rebrand

2. **Community Feedback**
   - Gather user reactions to the rebrand
   - Address any confusion or concerns

3. **Documentation Clarity**
   - Ensure users understand PSX vs Sylfie distinction
   - Clear migration path for existing users

## FAQ

**Q: Why rebrand from Biscuit to Sylfie?**
A: To create a clearer distinction between the project (Sylfie) and the syntax (PSX), following industry patterns like React/JSX.

**Q: Will my existing .psx files still work?**
A: Yes! The syntax remains exactly the same. Only the project name and CLI command are changing.

**Q: Do I need to update my code?**
A: No code changes required. Just update your CLI commands from `biscuit` to `sylfie`.

**Q: What does Sylfie mean?**
A: Sylfie evokes "sylph" - an ethereal, graceful being - reflecting our goal of elegant, lightweight UI components in Python.

## Conclusion

The rebrand from Biscuit to Sylfie, with PSX as the official syntax name, will provide clearer identity and better alignment with industry standards while maintaining full compatibility for existing users.