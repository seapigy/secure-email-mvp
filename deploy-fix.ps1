# Fix deployment workflow and push changes
Write-Host "Adding all changes..."
git add .

Write-Host "Committing deployment fix..."
git commit -m "fix: Update deploy workflow to trigger on feature branch

- Added feature/auth/core-signup to deploy workflow triggers
- Updated Netlify deployment condition to allow deployment from feature branch
- This will ensure the website goes live from our current branch"

Write-Host "Pushing changes..."
git push

Write-Host "Deployment fix complete!"
