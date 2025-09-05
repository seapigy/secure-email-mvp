#!/usr/bin/env bash
# Install git hooks for website freeze protection
# This script installs the pre-push hook that warns about website changes

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Installing Git Hooks for Website Freeze Protection${NC}"
echo ""

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}❌ Error: Not in a git repository${NC}"
    echo "Please run this script from the root of your git repository."
    exit 1
fi

# Check if .githooks directory exists
if [ ! -d ".githooks" ]; then
    echo -e "${YELLOW}⚠️  .githooks directory not found${NC}"
    echo "Creating .githooks directory..."
    mkdir -p .githooks
fi

# Check if pre-push hook exists
if [ ! -f ".githooks/pre-push" ]; then
    echo -e "${YELLOW}⚠️  Pre-push hook not found in .githooks/${NC}"
    echo "Please ensure .githooks/pre-push exists before running this script."
    exit 1
fi

# Make the hook executable
chmod +x .githooks/pre-push
echo -e "${GREEN}✅ Made .githooks/pre-push executable${NC}"

# Install the hook
if [ -f ".git/hooks/pre-push" ]; then
    echo -e "${YELLOW}⚠️  Pre-push hook already exists in .git/hooks/${NC}"
    echo "Backing up existing hook to .git/hooks/pre-push.backup"
    mv .git/hooks/pre-push .git/hooks/pre-push.backup
fi

# Copy the hook
cp .githooks/pre-push .git/hooks/pre-push
chmod +x .git/hooks/pre-push

echo -e "${GREEN}✅ Installed pre-push hook${NC}"
echo ""

echo -e "${BLUE}📋 Website Freeze Protection Installed${NC}"
echo ""
echo -e "${YELLOW}What this does:${NC}"
echo "- Warns you if you try to push changes to the website/ directory"
echo "- Blocks pushes unless you have proper approval markers"
echo "- Helps prevent accidental website changes"
echo ""
echo -e "${YELLOW}To push website changes:${NC}"
echo "1. Create a branch starting with 'website-edit/'"
echo "2. Include 'WEBSITE-EDIT-APPROVED: <JIRA/PR-ID>' in your commit message"
echo "3. Or get the 'website-edit-approved' label on your PR"
echo ""
echo -e "${YELLOW}For more information:${NC}"
echo "- Read WEBSITE_FREEZE_POLICY.md"
echo "- Check website/FROZEN.md"
echo "- Use .github/ISSUE_TEMPLATE/website-change-request.md for change requests"
echo ""
echo -e "${GREEN}🎉 Installation complete!${NC}"
echo ""
echo -e "${BLUE}Note:${NC} This hook only provides local warnings. The GitHub Actions"
echo "workflow will still enforce the freeze policy on the server side."
