package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"
)

var Version = "dev"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update git-index to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", Version)

		// 1. Fetch latest release info
		fmt.Println("Checking for latest release...")
		resp, err := http.Get("https://api.github.com/repos/acmcelwee/git-index/releases/latest")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error fetching latest release:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Failed to get release info: HTTP %d\n", resp.StatusCode)
			os.Exit(1)
		}

		var release struct {
			TagName string `json:"tag_name"`
			Assets  []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			fmt.Fprintln(os.Stderr, "Error decoding release info:", err)
			os.Exit(1)
		}

		if Version != "dev" && release.TagName == Version {
			fmt.Println("You are already on the latest version!")
			return
		}

		fmt.Printf("Latest version: %s\n", release.TagName)

		// 2. Find correct asset
		assetSuffix := fmt.Sprintf("%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		var downloadURL string
		for _, asset := range release.Assets {
			if strings.HasSuffix(asset.Name, assetSuffix) {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			fmt.Fprintln(os.Stderr, "No precompiled binary found for your OS/Architecture.")
			os.Exit(1)
		}

		// 3. Download the tarball
		fmt.Printf("Downloading %s...\n", downloadURL)
		req, err := http.NewRequest("GET", downloadURL, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating request:", err)
			os.Exit(1)
		}
		// Some github releases return 403 or need a user agent
		req.Header.Set("User-Agent", "git-index-updater")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error downloading update:", err)
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Failed to download update: HTTP %d\n", res.StatusCode)
			os.Exit(1)
		}

		// 4. Extract the binary from the tarball
		gzr, err := gzip.NewReader(res.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading gzip stream:", err)
			os.Exit(1)
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)
		var binaryReader io.Reader
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading tar stream:", err)
				os.Exit(1)
			}

			// The binary is inside the tarball and named git-index-<os>-<arch>
			if !header.FileInfo().IsDir() && strings.HasPrefix(header.Name, "git-index-") && !strings.HasSuffix(header.Name, ".sh") {
				binaryReader = tr
				break
			}
		}

		if binaryReader == nil {
			fmt.Fprintln(os.Stderr, "Could not find the executable inside the release archive.")
			os.Exit(1)
		}

		// 5. Apply the update
		fmt.Println("Applying update...")
		err = selfupdate.Apply(binaryReader, selfupdate.Options{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "Update failed:", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated to %s!\n", release.TagName)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
