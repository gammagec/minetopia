package modrinth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const baseURL = "https://api.modrinth.com/v2"

var userAgent = "minetopia-launcher/1.0 (github.com/minetopia)"

type Version struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	VersionNumber string `json:"version_number"`
	Files         []File `json:"files"`
}

type File struct {
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Hashes   map[string]string `json:"hashes"`
	Primary  bool              `json:"primary"`
	Size     int64             `json:"size"`
}

// GetVersion fetches versions for a project filtered by MC version and loader.
// If versionNumber is non-empty, it additionally filters to that exact version.
// Returns the first (newest) matching version.
func GetVersion(projectID, mcVersion, loader, versionNumber string) (*Version, error) {
	q := url.Values{}
	q.Set("game_versions", `["`+mcVersion+`"]`)
	q.Set("loaders", `["`+loader+`"]`)
	if versionNumber != "" {
		q.Set("version_number", versionNumber)
	}

	u := fmt.Sprintf("%s/project/%s/version?%s", baseURL, projectID, q.Encode())
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("modrinth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth API returned %d for project %s", resp.StatusCode, projectID)
	}

	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if len(versions) == 0 {
		if versionNumber != "" {
			return nil, fmt.Errorf("version %q not found for project %s on MC %s/%s", versionNumber, projectID, mcVersion, loader)
		}
		return nil, fmt.Errorf("no compatible versions for project %s on MC %s/%s", projectID, mcVersion, loader)
	}

	ver := &versions[0]
	if versionNumber != "" && ver.VersionNumber != versionNumber {
		return nil, fmt.Errorf("pinned version %q not found for project %s; got %s", versionNumber, projectID, ver.VersionNumber)
	}
	return ver, nil
}

// PrimaryFile returns the primary file from the version, or the first file if none is marked primary.
func (v *Version) PrimaryFile() (*File, error) {
	for i := range v.Files {
		if v.Files[i].Primary {
			return &v.Files[i], nil
		}
	}
	if len(v.Files) > 0 {
		return &v.Files[0], nil
	}
	return nil, fmt.Errorf("no files in version %s", v.ID)
}
