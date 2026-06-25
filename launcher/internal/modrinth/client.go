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

// GetProjectVersion fetches the version matching versionNumber for the given project slug or ID.
func GetProjectVersion(projectID, versionNumber string) (*Version, error) {
	u := fmt.Sprintf("%s/project/%s/version?%s", baseURL, projectID,
		url.Values{"version_number": {versionNumber}}.Encode())

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
		return nil, fmt.Errorf("version %q not found for project %s", versionNumber, projectID)
	}
	return &versions[0], nil
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
