# Husky Git Hooks

This project uses [Husky](https://typicode.github.io/husky/) to manage Git hooks that help maintain code quality and enforce development workflows.

## Available Hooks

### Pre-commit Hook (`.husky/pre-commit`)
Runs before each commit to ensure code quality:
- **Linting**: Runs `yarn lint` to check code style
- **Type Checking**: Runs `yarn typecheck` to verify TypeScript types
- **Testing**: Runs `yarn test:ci` to execute unit tests

If any of these checks fail, the commit will be blocked.

### Commit Message Hook (`.husky/commit-msg`)
Validates commit messages to follow [Conventional Commits](https://www.conventionalcommits.org/) format:
- **Format**: `<type>[optional scope]: <description>`
- **Types**: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert

**Examples of valid commit messages:**
```
feat: add new query builder feature
fix(auth): resolve connection timeout issue
docs: update installation instructions
chore: update dependencies
```

### Pre-push Hook (`.husky/pre-push`)
Runs before pushing to remote repository:
- **Changeset Check**: Warns if no changesets are found for user-facing changes
- **Interactive**: Allows you to continue if changes are internal/documentation only

## Installation

Husky hooks are automatically set up when the project is initialized. The hooks are stored in the `.husky/` directory and are ready to use after cloning the repository.

No additional installation steps are required - the hooks will work automatically once you have the repository cloned and dependencies installed.

## Managing Hooks

### Adding New Hooks
Create new hook files directly in the `.husky/` directory:

```bash
# Create a new hook file
echo '#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

# Your command here
yarn your-command' > .husky/your-hook-name

# Make it executable
chmod +x .husky/your-hook-name
```

### Modifying Existing Hooks
Edit the hook files directly in the `.husky/` directory:
- `.husky/pre-commit`
- `.husky/commit-msg`
- `.husky/pre-push`

### Disabling Hooks Temporarily
```bash
# Skip pre-commit hooks
git commit --no-verify -m "commit message"

# Skip pre-push hooks
git push --no-verify
```

## Hook Details

### Pre-commit Quality Checks
The pre-commit hook ensures:
1. **Code Style**: ESLint rules are followed
2. **Type Safety**: TypeScript compilation succeeds
3. **Test Coverage**: All tests pass

### Commit Message Validation
The commit-msg hook enforces:
- Conventional commit format
- Proper commit types
- Descriptive commit messages
- Maximum length limits

### Pre-push Changeset Validation
The pre-push hook:
- Checks for pending changesets
- Warns about missing changesets for user-facing changes
- Allows override for internal changes

## Integration with Changesets

The hooks work seamlessly with the changeset workflow:
1. **Make changes** to the codebase
2. **Pre-commit hook** ensures quality
3. **Commit-msg hook** enforces conventional commits
4. **Create changeset** for user-facing changes: `yarn changeset`
5. **Pre-push hook** reminds about changesets
6. **Push changes** to remote

## Troubleshooting

### Hook Not Running
If hooks aren't running, check that:
1. The hook files exist in `.husky/` directory
2. The hook files are executable (`chmod +x .husky/hook-name`)
3. You're in a Git repository
4. The `.husky/_/husky.sh` file exists

### Hook Failing
Check the specific error message:
- **Lint errors**: Fix with `yarn lint:fix`
- **Type errors**: Fix TypeScript issues
- **Test failures**: Fix failing tests
- **Commit format**: Follow conventional commit format

### Bypassing Hooks (Use Sparingly)
```bash
# Skip all hooks for emergency commits
git commit --no-verify -m "emergency fix"
git push --no-verify
```

## Best Practices

1. **Don't bypass hooks** unless absolutely necessary
2. **Fix issues** rather than skipping validation
3. **Use conventional commits** for better changelog generation
4. **Create changesets** for all user-facing changes
5. **Keep hooks fast** - they run on every commit/push

## Configuration

### Customizing Pre-commit Checks
Edit `.husky/pre-commit` to add/remove checks:
```bash
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

# Add your custom checks here
yarn lint
yarn typecheck
yarn test:ci
# yarn custom-check
```

### Customizing Commit Message Rules
Edit `.husky/commit-msg` to modify validation rules:
```bash
# Modify the regex pattern for different commit formats
commit_regex='^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .{1,50}'
```

## Learn More

- [Husky Documentation](https://typicode.github.io/husky/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Changesets Documentation](https://github.com/changesets/changesets)