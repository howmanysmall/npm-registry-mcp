package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ErrInvalidFlexInt is returned when FlexInt unmarshaling fails.
var ErrInvalidFlexInt = errors.New("cannot unmarshal into FlexInt")

// ErrInvalidFlexEngines is returned when FlexEngines unmarshaling fails.
var ErrInvalidFlexEngines = errors.New("cannot unmarshal into FlexEngines")

// ErrInvalidFlexLicense is returned when FlexLicense unmarshaling fails.
var ErrInvalidFlexLicense = errors.New("cannot unmarshal into FlexLicense")

// FlexInt is an integer that can be unmarshaled from either a JSON number or string.
// This handles cases where APIs return numbers as strings.
type FlexInt int

// UnmarshalJSON implements custom unmarshaling to handle both int and string values.
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexInt(i)
		return nil
	}

	// Try unmarshaling as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}

		parsed, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("cannot parse string %q as int: %w", s, err)
		}

		*f = FlexInt(parsed)

		return nil
	}

	return fmt.Errorf("%w: %s", ErrInvalidFlexInt, data)
}

// FlexEngines handles the engines field which can be either a map[string]string
// or an array of strings (for very old packages like lodash v0.1.0).
type FlexEngines map[string]string

// UnmarshalJSON implements custom unmarshaling to handle both map and array formats.
func (f *FlexEngines) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as map first (most common case)
	var m map[string]string
	if err := json.Unmarshal(data, &m); err == nil {
		*f = FlexEngines(m)
		return nil
	}

	// Try unmarshaling as array (old packages like lodash v0.1.0)
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		result := make(map[string]string, len(arr))
		for _, engine := range arr {
			result[engine] = "*"
		}

		*f = FlexEngines(result)

		return nil
	}

	return fmt.Errorf("%w: %s", ErrInvalidFlexEngines, data)
}

// FlexLicense handles the license field which can be either a string
// or an object (e.g., {"type": "MIT", "url": "..."}).
type FlexLicense string

// UnmarshalJSON implements custom unmarshaling to handle both string and object formats.
func (l *FlexLicense) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*l = ""
		return nil
	}

	// Try unmarshaling as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*l = FlexLicense(s)
		return nil
	}

	// Try unmarshaling as object (e.g., {"type": "MIT", "url": "..."})
	var obj struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		*l = FlexLicense(obj.Type)
		return nil
	}

	return fmt.Errorf("%w: %s", ErrInvalidFlexLicense, data)
}

// SearchResponse is the response from registry.npmjs.org/-/v1/search
type SearchResponse struct {
	Objects []SearchObject `json:"objects"`
	Total   int            `json:"total"`
	Time    string         `json:"time"`
}

// SearchObject is a single search result
type SearchObject struct {
	Package     SearchPackage `json:"package"`
	Score       Score         `json:"score"`
	SearchScore float64       `json:"searchScore"`
	Downloads   Downloads     `json:"downloads"`
	Dependents  FlexInt       `json:"dependents,omitempty"`
}

// SearchPackage contains package metadata from search results
type SearchPackage struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Keywords    []string     `json:"keywords"`
	License     FlexLicense  `json:"license"`
	Publisher   Publisher    `json:"publisher"`
	Maintainers []Maintainer `json:"maintainers"`
	Links       Links        `json:"links"`
	Date        string       `json:"date"`
}

// Score contains quality metrics
type Score struct {
	Final  float64     `json:"final"`
	Detail ScoreDetail `json:"detail"`
}

// ScoreDetail breaks down the score
type ScoreDetail struct {
	Quality     float64 `json:"quality"`
	Popularity  float64 `json:"popularity"`
	Maintenance float64 `json:"maintenance"`
}

// Downloads contains download counts
type Downloads struct {
	Weekly  int `json:"weekly"`
	Monthly int `json:"monthly"`
}

// Publisher is the package publisher
type Publisher struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Maintainer is a package maintainer
type Maintainer struct {
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
}

// Links contains package URLs
type Links struct {
	NPM        string `json:"npm"`
	Homepage   string `json:"homepage"`
	Repository string `json:"repository"`
	Bugs       string `json:"bugs"`
}

// PackageResponse is the response from registry.npmjs.org/{package}
type PackageResponse struct {
	ID          string                    `json:"_id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	DistTags    map[string]string         `json:"dist-tags"`
	Versions    map[string]PackageVersion `json:"versions"`
	Time        map[string]string         `json:"time"`
	Maintainers []Maintainer              `json:"maintainers"`
	Homepage    string                    `json:"homepage"`
	Keywords    []string                  `json:"keywords"`
	Repository  *Repository               `json:"repository"`
	License     FlexLicense               `json:"license"`
	Readme      string                    `json:"readme"`
}

// PackageVersion contains version-specific metadata
type PackageVersion struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Main            string            `json:"main"`
	Types           string            `json:"types,omitempty"`
	Typings         string            `json:"typings,omitempty"`
	License         FlexLicense       `json:"license"`
	Homepage        string            `json:"homepage"`
	Repository      *Repository       `json:"repository"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	PeerDeps        map[string]string `json:"peerDependencies"`
	Engines         FlexEngines       `json:"engines"`
	Dist            Dist              `json:"dist"`
}

// Repository contains repository info
type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// UnmarshalJSON implements custom unmarshaling to handle both string and object formats.
// The NPM registry sometimes returns repository as a string (e.g., "github:user/repo")
// instead of an object (e.g., {"type": "git", "url": "https://github.com/user/repo"}).
func (r *Repository) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as object first (the normal case)
	var repo struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	}

	if err := json.Unmarshal(data, &repo); err == nil {
		r.Type = repo.Type
		r.URL = repo.URL

		return nil
	}

	// If that fails, try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal repository: %w", err)
	}

	// Handle string format
	if s == "" {
		return nil
	}

	// If it starts with "github:", extract the repo path
	if len(s) > 7 && s[:7] == "github:" {
		r.Type = "git"
		r.URL = "https://github.com/" + s[7:]
	} else {
		r.Type = "git"
		r.URL = s
	}

	return nil
}

// Dist contains distribution info
type Dist struct {
	Tarball      string `json:"tarball"`
	Shasum       string `json:"shasum"`
	Integrity    string `json:"integrity"`
	FileCount    int    `json:"fileCount"`
	UnpackedSize int64  `json:"unpackedSize"`
}

// DownloadPoint is the response from api.npmjs.org/downloads/point/{period}/{package}
type DownloadPoint struct {
	Downloads int    `json:"downloads"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Package   string `json:"package"`
}

// DownloadRange is the response from api.npmjs.org/downloads/range/{period}/{package}
type DownloadRange struct {
	Downloads []DailyDownload `json:"downloads"`
	Start     string          `json:"start"`
	End       string          `json:"end"`
	Package   string          `json:"package"`
}

// DailyDownload is a single day's download count
type DailyDownload struct {
	Downloads int    `json:"downloads"`
	Day       string `json:"day"`
}

// AbbreviatedPackageResponse is the response from registry.npmjs.org/{package}
// with Accept: application/vnd.npm.install-v1+json
type AbbreviatedPackageResponse struct {
	Name     string                               `json:"name"`
	DistTags map[string]string                    `json:"dist-tags"`
	Versions map[string]AbbreviatedPackageVersion `json:"versions"`
	Modified string                               `json:"modified"`
}

// AbbreviatedPackageVersion is a smaller version of PackageVersion
type AbbreviatedPackageVersion struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies,omitempty"`
	DevDependencies      map[string]string `json:"devDependencies,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
	PeerDependencies     map[string]string `json:"peerDependencies,omitempty"`
	Engines              FlexEngines       `json:"engines,omitempty"`
	Dist                 Dist              `json:"dist"`
	Bin                  map[string]string `json:"bin,omitempty"`
}

// Advisory represents a security advisory from the NPM registry
type Advisory struct {
	ID                 int      `json:"id"`
	URL                string   `json:"url"`
	Title              string   `json:"title"`
	Severity           string   `json:"severity"`
	VulnerableVersions string   `json:"vulnerable_versions"`
	PatchedVersions    string   `json:"patched_versions"`
	Recommendation     string   `json:"recommendation"`
	CWEs               []string `json:"cwes"`
	CVSSVector         string   `json:"cvss_vector"`
	CVSSScore          float64  `json:"cvss_score"`
	ModuleName         string   `json:"module_name"`
	Cves               []string `json:"cves"`
	Access             string   `json:"access"`
	Created            string   `json:"created"`
	Updated            string   `json:"updated"`
	Findings           []any    `json:"findings"`
	References         string   `json:"references"`
	NpmAdvisoryID      any      `json:"npm_advisory_id"`
	GithubAdvisoryID   string   `json:"github_advisory_id"`
	CveID              string   `json:"cve_id"`
	Vendor             string   `json:"vendor"`
	Product            string   `json:"product"`
	Description        string   `json:"description"`
	Source             any      `json:"source"`
	Slug               string   `json:"slug"`
	Version            string   `json:"version"`
	FoundBy            any      `json:"found_by"`
	ReportedBy         any      `json:"reported_by"`
	Cwe                string   `json:"cwe"`
	Metadata           any      `json:"metadata"`
	AffectedVersions   string   `json:"affected_versions"`
	Recommendations    string   `json:"recommendations"`
}

// ParsedTime parses an NPM time string
func ParsedTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
