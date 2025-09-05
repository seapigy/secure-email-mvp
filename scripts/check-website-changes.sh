#!/usr/bin/env bash
# Helper script to check for website changes
# Usage: ./scripts/check-website-changes.sh [base-branch]
# Exit code 0: No website changes
# Exit code 1: Website changes detected

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default base branch
BASE_BRANCH=${1:-main}

echo -e "${BLUE}🔍 Checking for website changes against origin/${BASE_BRANCH}${NC}"

# Fetch the latest changes from origin
echo "Fetching latest changes from origin..."
git fetch origin "$BASE_BRANCH" >/dev/null 2>&1 || {
    echo -e "${RED}❌ Error: Could not fetch origin/${BASE_BRANCH}${NC}"
    echo "Make sure the remote 'origin' is configured and accessible."
    exit 1
}

# Check if the base branch exists
if ! git show-ref --verify --quiet "refs/remotes/origin/$BASE_BRANCH"; then
    echo -e "${RED}❌ Error: origin/${BASE_BRANCH} does not exist${NC}"
    echo "Available remote branches:"
    git branch -r | grep origin/ | sed 's/^/  /'
    exit 1
fi

# Get the list of changed files in website/ directory
CHANGED=$(git diff --name-only "origin/$BASE_BRANCH" -- website/ 2>/dev/null || true)

if [ -z "$CHANGED" ]; then
    echo -e "${GREEN}✅ No website changes detected vs origin/${BASE_BRANCH}${NC}"
    echo ""
    echo -e "${BLUE}Current website files:${NC}"
    find website/ -type f -name "*.tsx" -o -name "*.ts" -o -name "*.css" -o -name "*.md" | head -10 | sed 's/^/  /'
    if [ $(find website/ -type f -name "*.tsx" -o -name "*.ts" -o -name "*.css" -o -name "*.md" | wc -l) -gt 10 ]; then
        echo "  ... and more files"
    fi
    exit 0
else
    echo -e "${RED}❌ Website changes detected vs origin/${BASE_BRANCH}:${NC}"
    echo ""
    echo "$CHANGED" | while read -r file; do
        if [ -n "$file" ]; then
            echo -e "${YELLOW}  📝 $file${NC}"
        fi
    done
    echo ""
    echo -e "${RED}🚨 Website directory is frozen!${NC}"
    echo ""
    echo -e "${YELLOW}To make these changes:${NC}"
    echo "1. Create a GitHub Issue using .github/ISSUE_TEMPLATE/website-change-request.md"
    echo "2. Get approval from @seapigy/website-owners"
    echo "3. Get the 'website-edit-approved' label on your PR"
    echo "4. Ensure all tests pass before merging"
    echo ""
    echo -e "${BLUE}For more information:${NC}"
    echo "- WEBSITE_FREEZE_POLICY.md"
    echo "- website/FROZEN.md"
    echo ""
    exit 1
fi
