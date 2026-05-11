param(
  [string]$Version
)

# Check if version provided
if (-not $Version) {
  Write-Error "Version argument required. Usage: release.ps1 <version> (e.g., 1.2.3 or v1.2.3)"
  exit 1
}

# Check if on main branch
$currentBranch = git rev-parse --abbrev-ref HEAD
if ($currentBranch -ne "main") {
  Write-Error "Error: You are on branch '$currentBranch', not on main branch. Release must be from main branch."
  exit 1
}

# Normalize version: add 'v' prefix if not present
if (-not $Version.StartsWith("v")) {
  $Version = "v$Version"
}

# Validate format: must be vx.y.z or similar (v followed by at least one dot-separated number)
if ($Version -notmatch '^v\d+\.\d+\.\d+') {
  Write-Error "Error: Invalid version format '$Version'. Expected format: vx.y.z or x.y.z"
  exit 1
}

# Check if tag already exists
$tagExists = & git tag -l $Version
if ($tagExists) {
  Write-Error "Error: Tag '$Version' already exists."
  exit 1
}

# Create annotated tag for the release (standard Go package release mechanism)
try {
  git tag -a $Version -m "Release $Version"
  Write-Host "Successfully created tag: $Version"
}
catch {
  Write-Error "Error: Failed to create tag. $_"
  exit 1
}

# Push the tag to the remote repository
try {
  git push origin $Version
  Write-Host "Successfully pushed tag '$Version' to remote."
}
catch {
  Write-Error "Error: Failed to push tag to remote. $_"
  exit 1
}
