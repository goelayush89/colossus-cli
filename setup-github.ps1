# Setup GitHub Repository for Colossus CLI
# This script helps set up the GitHub repository with proper configuration

param(
    [Parameter(Mandatory=$true)]
    [string]$GitHubUsername,
    
    [string]$RepoName = "colossus-cli",
    [switch]$CreateRepo,
    [switch]$SetupPages
)

$InfoColor = "Green"
$WarnColor = "Yellow"
$ErrorColor = "Red"

function Write-Info($msg) { Write-Host $msg -ForegroundColor $InfoColor }
function Write-Warn($msg) { Write-Host $msg -ForegroundColor $WarnColor }
function Write-Error($msg) { Write-Host $msg -ForegroundColor $ErrorColor }

Write-Info "🚀 Setting up GitHub Repository for Colossus CLI"
Write-Info "================================================"

# Check if we're in a Git repository
if (!(Test-Path ".git")) {
    Write-Info "📁 Initializing Git repository..."
    git init
    git add .
    git commit -m "🎉 Initial commit: Colossus CLI - Ollama alternative in Go"
}

# Update download page with correct GitHub URLs
Write-Info "🔗 Updating download URLs..."
$downloadPage = "docs/index.html"
if (Test-Path $downloadPage) {
    $content = Get-Content $downloadPage -Raw
    $content = $content -replace "https://github.com/user/colossus-cli", "https://github.com/$GitHubUsername/$RepoName"
    Set-Content $downloadPage $content
    Write-Info "   ✅ Updated GitHub URLs in download page"
}

# Set up remote origin
$repoUrl = "https://github.com/$GitHubUsername/$RepoName.git"
Write-Info "🌐 Setting up remote origin: $repoUrl"

try {
    git remote remove origin 2>$null
    git remote add origin $repoUrl
    Write-Info "   ✅ Remote origin set"
} catch {
    Write-Warn "   ⚠️  Could not set remote origin. You may need to create the repository first."
}

# Create initial tag
Write-Info "🏷️  Creating initial tag..."
git tag v1.0.0 2>$null
Write-Info "   ✅ Tagged v1.0.0"

Write-Info ""
Write-Info "📋 Next Steps:"
Write-Info "=============="
Write-Info "1. Create GitHub repository: https://github.com/new"
Write-Info "   - Repository name: $RepoName"
Write-Info "   - Description: 'A powerful Go-based alternative to Ollama for running LLMs locally'"
Write-Info "   - Make it public for GitHub Pages"
Write-Info ""
Write-Info "2. Push code to GitHub:"
Write-Info "   git push -u origin main"
Write-Info "   git push origin v1.0.0"
Write-Info ""
Write-Info "3. Enable GitHub Pages:"
Write-Info "   - Go to repository Settings → Pages"
Write-Info "   - Source: Deploy from a branch"
Write-Info "   - Branch: gh-pages"
Write-Info "   - Folder: / (root)"
Write-Info ""
Write-Info "4. Create first release:"
Write-Info "   - Go to repository → Releases → Create a new release"
Write-Info "   - Tag: v1.0.0"
Write-Info "   - Title: 'Colossus CLI v1.0.0 - Initial Release'"
Write-Info "   - The GitHub Action will automatically build and attach binaries"
Write-Info ""
Write-Info "🌐 Your download page will be available at:"
Write-Info "   https://$GitHubUsername.github.io/$RepoName"
Write-Info ""
Write-Info "✅ Repository setup complete!"

if ($CreateRepo) {
    Write-Info ""
    Write-Info "🔧 To create the repository via GitHub CLI (if installed):"
    Write-Info "   gh repo create $RepoName --public --description 'A powerful Go-based alternative to Ollama for running LLMs locally'"
    Write-Info "   git push -u origin main"
    Write-Info "   git push origin v1.0.0"
}
