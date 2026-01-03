package npm

import "time"

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
	Downloads   Downloads     `json:"downloads,omitempty"`
	Dependents  int           `json:"dependents,omitempty"`
}

// SearchPackage contains package metadata from search results
type SearchPackage struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Keywords    []string     `json:"keywords"`
	License     string       `json:"license"`
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
	License     string                    `json:"license"`
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
	License         string            `json:"license"`
	Homepage        string            `json:"homepage"`
	Repository      *Repository       `json:"repository"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	PeerDeps        map[string]string `json:"peerDependencies"`
	Engines         map[string]string `json:"engines"`
	Dist            Dist              `json:"dist"`
}

// Repository contains repository info
type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
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

// ParsedTime parses an NPM time string
func ParsedTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
