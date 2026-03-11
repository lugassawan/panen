#!/bin/sh
# Bump the version in all required files, commit, push, and create a PR.
# Complements release.sh: this creates the version bump PR; release.sh tags after merge.
set -eu

usage() {
    echo "Usage: $0 <version|--auto>"
    echo ""
    echo "Bump version in wails.json and Windows manifest, then create a PR."
    echo ""
    echo "Examples:"
    echo "  $0 1.2.0"
    echo "  $0 v1.2.0"
    echo "  $0 --auto"
    exit 1
}

die() {
    echo "Error: $1" >&2
    exit 1
}

# Detect next version by analyzing commits since the last tag.
# Classifies commits by conventional commit prefix (feat:/fix:/etc.)
# and determines the appropriate semver bump.
detect_next_version() {
    latest_tag=$(git tag -l 'v*' | sort -V | tail -1)
    if [ -z "$latest_tag" ]; then
        latest_tag="v0.0.0"
        echo "No existing tags found. Starting from v0.0.0."
        commit_range="HEAD"
    else
        commit_range="${latest_tag}..HEAD"
    fi

    commits=$(git log "$commit_range" --oneline 2>/dev/null || true)
    if [ -z "$commits" ]; then
        echo "No commits since $latest_tag. Nothing to bump."
        exit 0
    fi

    commit_count=$(echo "$commits" | wc -l | tr -d ' ')

    bump="patch"
    feat_commits=""
    fix_commits=""
    other_commits=""

    OLD_IFS="$IFS"
    IFS='
'
    for line in $commits; do
        subject=$(echo "$line" | sed 's/^[a-f0-9]* //')
        case "$subject" in
            feat:*)
                if [ "$bump" != "major" ]; then
                    bump="minor"
                fi
                feat_commits="${feat_commits}    - ${subject}
"
                ;;
            fix:*)
                fix_commits="${fix_commits}    - ${subject}
"
                ;;
            *)
                other_commits="${other_commits}    - ${subject}
"
                ;;
        esac
    done
    IFS="$OLD_IFS"

    # Parse current version
    major=$(echo "$latest_tag" | sed 's/^v//' | cut -d. -f1)
    minor=$(echo "$latest_tag" | sed 's/^v//' | cut -d. -f2)
    patch=$(echo "$latest_tag" | sed 's/^v//' | cut -d. -f3)

    case "$bump" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
    esac

    next_version="${major}.${minor}.${patch}"

    echo ""
    echo "Current version: $latest_tag"
    echo "Commits since $latest_tag: $commit_count"
    echo ""

    if [ -n "$feat_commits" ]; then
        echo "  Features (minor):"
        printf "%s" "$feat_commits"
        echo ""
    fi

    if [ -n "$fix_commits" ]; then
        echo "  Bug fixes (patch):"
        printf "%s" "$fix_commits"
        echo ""
    fi

    if [ -n "$other_commits" ]; then
        echo "  Other:"
        printf "%s" "$other_commits"
        echo ""
    fi

    echo "Next version: $next_version"
    echo ""

    printf "Continue with version bump? [y/N] "
    read -r confirm
    case "$confirm" in
        [yY]|[yY][eE][sS]) ;;
        *) echo "Aborted."; exit 0 ;;
    esac

    version="$next_version"
}

check_dependency() {
    command -v "$1" >/dev/null 2>&1 || die "$1 is required but not installed"
}

# Require exactly one argument
[ $# -eq 1 ] || usage

check_dependency jq
check_dependency gh
check_dependency sed

# Ensure working tree is clean (before any interactive prompts)
if [ -n "$(git status --porcelain)" ]; then
    die "working tree is dirty — commit or stash changes first"
fi

# Ensure on main and up to date
branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$branch" != "main" ]; then
    echo "Not on main (currently on $branch). Switching to main..."
    git checkout main
fi
git pull origin main

if [ "$1" = "--auto" ]; then
    detect_next_version
else
    version="$1"

    # Strip v prefix if present — we work with bare semver internally
    version="${version#v}"
fi

# Validate semver format (X.Y.Z)
echo "$version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$' \
    || die "invalid version format: $version (expected X.Y.Z)"

# Check it's not the current version
current_version=$(jq -r '.info.productVersion' wails.json)
if [ "$version" = "$current_version" ]; then
    die "version $version is already the current version in wails.json"
fi

# Create branch
bump_branch="chore/bump-version-${version}"
echo "Creating branch $bump_branch..."
git checkout -b "$bump_branch"

# Update wails.json
echo "Updating wails.json to $version..."
jq --arg v "$version" '.info.productVersion = $v' wails.json > wails.json.tmp \
    && mv wails.json.tmp wails.json

# Update Windows manifest (version format: X.Y.Z.0)
# Only replace the app assemblyIdentity version, not the Common-Controls dependency version.
# Anchor on the product name that precedes the version attribute.
manifest="build/windows/panen.exe.manifest"
echo "Updating $manifest to ${version}.0..."
sed -i.bak '/com.lugasseptiawan.panen/{
n
s/version="[0-9]*\.[0-9]*\.[0-9]*\.[0-9]*"/version="'"${version}"'.0"/
}' "$manifest" \
    && rm -f "${manifest}.bak"

# Commit
git add wails.json "$manifest"
git commit -m "chore: bump version to $version"

# Push
echo "Pushing $bump_branch..."
git push -u origin "$bump_branch"

# Create PR
echo "Creating pull request..."
pr_url=$(gh pr create \
    --title "chore: bump version to $version" \
    --body "$(cat <<EOF
## Issue
N/A

## Summary
- Bump version to $version in \`wails.json\` and Windows manifest

## Test Plan
- [ ] \`make release-check VERSION=$version\` passes
- [ ] \`wails.json\` shows \`productVersion: \"$version\"\`
- [ ] \`build/windows/panen.exe.manifest\` shows \`version=\"${version}.0\"\`

## Notes
Created by \`scripts/bump-version.sh\`.
EOF
)")

echo ""
echo "Pull request created: $pr_url"
echo ""
echo "Next steps:"
echo "  1. Review and merge the PR"
echo "  2. Run: scripts/release.sh $version"
