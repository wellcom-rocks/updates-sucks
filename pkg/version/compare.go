package version

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	Original string
	Major    int
	Minor    int
	Patch    int
	PreRelease string
	Build    string
}

type CompareResult int

const (
	Equal CompareResult = iota
	Greater
	Less
)

func ParseSemVer(version string) (*Version, error) {
	// Remove leading 'v' if present
	v := strings.TrimPrefix(version, "v")
	
	// Regex for semantic versioning
	semverRegex := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	
	matches := semverRegex.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("invalid semver format: %s", version)
	}
	
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	
	return &Version{
		Original:   version,
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: matches[4],
		Build:      matches[5],
	}, nil
}

func CompareSemVer(current, latest string) (CompareResult, error) {
	currentVer, err := ParseSemVer(current)
	if err != nil {
		return Equal, fmt.Errorf("failed to parse current version: %w", err)
	}
	
	latestVer, err := ParseSemVer(latest)
	if err != nil {
		return Equal, fmt.Errorf("failed to parse latest version: %w", err)
	}
	
	return currentVer.Compare(latestVer), nil
}

func (v *Version) Compare(other *Version) CompareResult {
	// Compare major version
	if v.Major > other.Major {
		return Greater
	} else if v.Major < other.Major {
		return Less
	}
	
	// Compare minor version
	if v.Minor > other.Minor {
		return Greater
	} else if v.Minor < other.Minor {
		return Less
	}
	
	// Compare patch version
	if v.Patch > other.Patch {
		return Greater
	} else if v.Patch < other.Patch {
		return Less
	}
	
	// Compare pre-release versions
	if v.PreRelease == "" && other.PreRelease != "" {
		return Greater // Release version is greater than pre-release
	} else if v.PreRelease != "" && other.PreRelease == "" {
		return Less // Pre-release is less than release
	} else if v.PreRelease != "" && other.PreRelease != "" {
		// Both have pre-release, compare lexicographically
		if v.PreRelease > other.PreRelease {
			return Greater
		} else if v.PreRelease < other.PreRelease {
			return Less
		}
	}
	
	return Equal
}

func CompareCalVer(current, latest string) (CompareResult, error) {
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")
	
	if len(currentParts) != 3 || len(latestParts) != 3 {
		return Equal, fmt.Errorf("invalid calver format")
	}
	
	for i := 0; i < 3; i++ {
		currentNum, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return Equal, fmt.Errorf("invalid calver number: %s", currentParts[i])
		}
		
		latestNum, err := strconv.Atoi(latestParts[i])
		if err != nil {
			return Equal, fmt.Errorf("invalid calver number: %s", latestParts[i])
		}
		
		if currentNum > latestNum {
			return Greater, nil
		} else if currentNum < latestNum {
			return Less, nil
		}
	}
	
	return Equal, nil
}

func CompareString(current, latest string) (CompareResult, error) {
	if current == latest {
		return Equal, nil
	} else if current > latest {
		return Greater, nil
	} else {
		return Less, nil
	}
}

func FilterValidSemVer(tags []string) []string {
	var validTags []string
	for _, tag := range tags {
		if _, err := ParseSemVer(tag); err == nil {
			validTags = append(validTags, tag)
		}
	}
	return validTags
}

func FilterValidCalVer(tags []string) []string {
	calverRegex := regexp.MustCompile(`^(\d{4})\.(\d{1,2})\.(\d+)$`)
	var validTags []string
	for _, tag := range tags {
		if calverRegex.MatchString(tag) {
			validTags = append(validTags, tag)
		}
	}
	return validTags
}

func SortSemVer(tags []string) []string {
	var versions []*Version
	for _, tag := range tags {
		if v, err := ParseSemVer(tag); err == nil {
			versions = append(versions, v)
		}
	}
	
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Compare(versions[j]) == Less
	})
	
	var sorted []string
	for _, v := range versions {
		sorted = append(sorted, v.Original)
	}
	
	return sorted
}

func SortCalVer(tags []string) []string {
	validTags := FilterValidCalVer(tags)
	sort.Strings(validTags)
	return validTags
}

func GetLatestVersion(tags []string, scheme string) (string, error) {
	switch scheme {
	case "semver":
		validTags := FilterValidSemVer(tags)
		if len(validTags) == 0 {
			return "", fmt.Errorf("no valid semver tags found")
		}
		sorted := SortSemVer(validTags)
		return sorted[len(sorted)-1], nil
		
	case "calver":
		validTags := FilterValidCalVer(tags)
		if len(validTags) == 0 {
			return "", fmt.Errorf("no valid calver tags found")
		}
		sorted := SortCalVer(validTags)
		return sorted[len(sorted)-1], nil
		
	case "string":
		if len(tags) == 0 {
			return "", fmt.Errorf("no tags found")
		}
		sort.Strings(tags)
		return tags[len(tags)-1], nil
		
	default:
		return "", fmt.Errorf("unsupported versioning scheme: %s", scheme)
	}
}