package scanner

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/wellcom-rocks/updates-sucks/pkg/config"
	"github.com/wellcom-rocks/updates-sucks/pkg/version"
)

type GitScanner struct {
	verbose bool
}

func NewGitScanner(verbose bool) *GitScanner {
	return &GitScanner{verbose: verbose}
}

func (g *GitScanner) GetLatestVersion(repo *config.Repository) (string, error) {
	if repo.Type != "git" {
		return "", fmt.Errorf("unsupported repository type: %s", repo.Type)
	}

	// Prepare git command with authentication
	cmd := exec.Command("git", "ls-remote", "--tags", "--refs", repo.URL)

	// Configure authentication if needed
	if repo.Auth != nil && repo.Auth.EnvVariable != "" {
		token := os.Getenv(repo.Auth.EnvVariable)
		if token == "" {
			return "", fmt.Errorf("authentication token not found in environment variable %s", repo.Auth.EnvVariable)
		}

		// Configure git authentication based on auth type
		switch repo.Auth.Type {
		case "token":
			// For GitHub/GitLab tokens, modify the URL to include authentication
			authenticatedURL := g.addTokenToURL(repo.URL, token)
			cmd.Args[len(cmd.Args)-1] = authenticatedURL
		case "ssh":
			// For SSH authentication, the token should be an SSH key path
			cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", token))
		default:
			return "", fmt.Errorf("unsupported authentication type: %s", repo.Auth.Type)
		}
	}

	if g.verbose {
		fmt.Printf("Executing: git ls-remote --tags --refs %s\n", repo.URL)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute git ls-remote: %w", err)
	}

	// Parse tags from output
	tags := g.parseTags(string(output))
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in repository")
	}

	// Remove prefix if configured
	if repo.Versioning != nil && repo.Versioning.IgnorePrefix != "" {
		tags = g.removePrefix(tags, repo.Versioning.IgnorePrefix)
	}

	// Filter and sort tags based on versioning scheme
	scheme := "semver"
	if repo.Versioning != nil && repo.Versioning.Scheme != "" {
		scheme = repo.Versioning.Scheme
	}

	latestTag, err := g.findLatestVersion(tags, scheme)
	if err != nil {
		return "", fmt.Errorf("failed to find latest version: %w", err)
	}

	// Add prefix back if it was removed
	if repo.Versioning != nil && repo.Versioning.IgnorePrefix != "" {
		latestTag = repo.Versioning.IgnorePrefix + latestTag
	}

	return latestTag, nil
}

func (g *GitScanner) parseTags(output string) []string {
	var tags []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format: <commit-hash>\trefs/tags/<tag-name>
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue
		}

		ref := parts[1]
		if strings.HasPrefix(ref, "refs/tags/") {
			tag := strings.TrimPrefix(ref, "refs/tags/")
			tags = append(tags, tag)
		}
	}

	return tags
}

func (g *GitScanner) removePrefix(tags []string, prefix string) []string {
	var result []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			result = append(result, strings.TrimPrefix(tag, prefix))
		}
	}
	return result
}

func (g *GitScanner) findLatestVersion(tags []string, scheme string) (string, error) {
	switch scheme {
	case "semver":
		return g.findLatestSemVer(tags)
	case "calver":
		return g.findLatestCalVer(tags)
	case "string":
		return g.findLatestString(tags)
	default:
		return "", fmt.Errorf("unsupported versioning scheme: %s", scheme)
	}
}

func (g *GitScanner) findLatestSemVer(tags []string) (string, error) {
	// Import the version package functions to properly sort semver tags
	return findLatestVersionFromTags(tags, "semver")
}

func (g *GitScanner) findLatestCalVer(tags []string) (string, error) {
	return findLatestVersionFromTags(tags, "calver")
}

func (g *GitScanner) findLatestString(tags []string) (string, error) {
	return findLatestVersionFromTags(tags, "string")
}

func findLatestVersionFromTags(tags []string, scheme string) (string, error) {
	return version.GetLatestVersion(tags, scheme)
}

func (g *GitScanner) addTokenToURL(repoURL, token string) string {
	// Parse the URL
	u, err := url.Parse(repoURL)
	if err != nil {
		return repoURL // Return original if parsing fails
	}

	// Add token to URL based on the host
	if strings.Contains(u.Host, "github.com") {
		// GitHub: https://token@github.com/owner/repo.git
		u.User = url.User(token)
	} else if strings.Contains(u.Host, "gitlab.com") {
		// GitLab: https://oauth2:token@gitlab.com/owner/repo.git
		u.User = url.UserPassword("oauth2", token)
	} else {
		// Generic: https://token@host/path
		u.User = url.User(token)
	}

	return u.String()
}
