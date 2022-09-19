package cli

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubRepoEndpoint = "https://api.github.com/repos/nmaupu/gopicto/releases"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gopicto to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		execFile, err := os.Executable()
		if err != nil {
			log.Error().Err(err).Msg("An error occurred when getting executable path")
			os.Exit(1)
		}
		log.Info().Msg(getVersionMessage())
		log.Info().Str("executable", execFile).Msg("Binary info")
		filename := filepath.Base(execFile)
		if strings.HasPrefix(filename, "___") {
			log.Error().Msg("Cannot update because running from IDE")
			os.Exit(1)
		}

		releases, err := getGithubReleases()
		if err != nil {
			log.Error().Err(err).Msg("An error occurred getting releases from github")
			os.Exit(1)
		}

		if len(releases) == 0 {
			log.Error().Msgf("Unable to get any releases from %s", githubRepoEndpoint)
			os.Exit(1)
		}

		// Latest release is the first one
		release := releases[0]

		if release.TagName == ApplicationVersion {
			log.Info().Msg("This is already the latest version, no need to update !")
			os.Exit(0)
		}

		assetURL := ""
		for _, asset := range release.Assets {
			arch := runtime.GOARCH
			if runtime.GOARCH == "amd64" {
				arch = "x64"
			}
			if asset.Name == fmt.Sprintf("%s-%s-%s_%s", AppName, release.TagName, runtime.GOOS, arch) {
				assetURL = asset.BrowserDownloadURL
			}
		}

		log.Info().
			Str("version", release.TagName).
			Str("os", runtime.GOOS).
			Str("arch", runtime.GOARCH).
			Str("url", assetURL).
			Msg("Update info")

		// Downloading release
		tmpFile, err := downloadGithubAsset(assetURL)
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		// Overwriting execFile with file
		dstFile, err := os.OpenFile(
			fmt.Sprintf("%s/_%s", path.Dir(execFile), path.Base(execFile)),
			os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
			os.FileMode(0755))
		if err != nil {
			log.Error().Err(err).Msg("An error occurred trying to open file")
			os.Exit(1)
		}
		_, err = io.Copy(dstFile, tmpFile)
		if err != nil {
			log.Error().Err(err).Msg("An error occurred trying to write to file")
			os.Exit(1)
		}
		dstFile.Close()

		if err := os.Rename(dstFile.Name(), execFile); err != nil {
			log.Error().Err(err).Msg("An error occurred trying to write to file")
			os.Exit(1)
		}

		log.Info().Msgf("Update to %s done", release.TagName)
	},
}

type githubRelease struct {
	URL         string    `json:"url"`
	AssetsURL   string    `json:"assets_url"`
	UploadURL   string    `json:"upload_url"`
	HtmlURL     string    `json:"html_url"`
	ID          int       `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		URL                string `json:"url"`
		ID                 int    `json:"id"`
		NodeID             string `json:"node_id"`
		Name               string `json:"name"`
		ContentType        string `json:"content_type"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

func getGithubReleases() ([]githubRelease, error) {
	req, err := http.NewRequest(http.MethodGet, githubRepoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	ret := make([]githubRelease, 0)
	err = json.Unmarshal(responseData, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// downloadGithubAsset downloads a github asset and return the temp file
func downloadGithubAsset(url string) (*os.File, error) {
	tmpFile, err := ioutil.TempFile("", AppName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	_, err = io.Copy(tmpFile, response.Body)
	if err != nil {
		return nil, err
	}

	tmpFile.Seek(0, 0)
	return tmpFile, err
}
