package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"github.com/stenstromen/ncregistry/types"
	"github.com/stenstromen/ncregistry/utils"
)

var config types.Config

func saveConfig(newEntry types.Entry) {
	config.Entries = append(config.Entries, newEntry)
	viper.Set("Entries", config.Entries)
	if err := viper.WriteConfig(); err != nil {
		log.Fatalf("Error writing config: %s", err)
	}
}

func getRepositories(url, username, password string) (types.RepositoryResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/_catalog", nil)
	if err != nil {
		return types.RepositoryResponse{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return types.RepositoryResponse{}, err
	}
	defer resp.Body.Close()

	var repoResp types.RepositoryResponse
	err = json.NewDecoder(resp.Body).Decode(&repoResp)
	return repoResp, err
}

func getTags(url, username, password, repository string) (types.TagResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/tags/list", nil)
	if err != nil {
		return types.TagResponse{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return types.TagResponse{}, err
	}
	defer resp.Body.Close()

	var tagResp types.TagResponse
	err = json.NewDecoder(resp.Body).Decode(&tagResp)
	return tagResp, err
}

func getManifest(url, username, password, repository, tag string) (*types.ManifestResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/manifests/"+tag, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var manifestResp types.ManifestResponse
	err = json.NewDecoder(resp.Body).Decode(&manifestResp)
	if err != nil {
		return nil, err
	}

	return &manifestResp, nil
}

func getBlob(url, username, password, repository, digest string) (types.BlobResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/"+repository+"/blobs/"+digest, nil)
	if err != nil {
		return types.BlobResponse{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return types.BlobResponse{}, err
	}
	defer resp.Body.Close()

	var blobResp types.BlobResponse
	err = json.NewDecoder(resp.Body).Decode(&blobResp)
	return blobResp, err
}

func deleteManifest(url, username, password, repository, digest string) error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url+"/v2/"+repository+"/manifests/"+digest, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
		return fmt.Errorf("unexpected response from server: %s, body: %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func main() {
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		file, err := os.Create("config.yaml")
		if err != nil {
			log.Fatalf("Failed to create file: %s", err)
		}
		file.Close()
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SafeWriteConfig()
		} else {
			fmt.Printf("Error reading config file: %s", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	for {

		clear := exec.Command("clear")
		clear.Stdout = os.Stdout
		clear.Run()

		prompt := promptui.Select{
			Label: "Main Menu",
			Items: []string{"Add registry", "Remove registry", "Connect to registry", "Exit"},
			Templates: &promptui.SelectTemplates{
				Active:   `👉 {{ . | cyan | bold }}`,
				Inactive: `   {{ . | cyan }}`,
				Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
				Help:     `{{ "Use ↑/↓ to move and Enter to select" | bold }}`,
			},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Add registry":
			prompt := promptui.Prompt{
				Label: "Registry URL",
				Templates: &promptui.PromptTemplates{
					Prompt:  `👉 {{ . | cyan | bold }} `,
					Valid:   `👉 {{ . | green | bold }} `,
					Invalid: `👉 {{ . | red | bold }} `,
					Success: `👉 {{ . | bold }} `,
				},
			}

			url, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			prompt = promptui.Prompt{
				Label: "Registry Username",
				Templates: &promptui.PromptTemplates{
					Prompt:  `👉 {{ . | cyan | bold }} `,
					Valid:   `👉 {{ . | green | bold }} `,
					Invalid: `👉 {{ . | red | bold }} `,
					Success: `👉 {{ . | bold }} `,
				},
			}

			username, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			// Prepend "https://" if not present
			if !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}

			prompt = promptui.Prompt{
				Label: "Registry Password",
				Mask:  '*',
				Templates: &promptui.PromptTemplates{
					Prompt:  `👉 {{ . | cyan | bold }} `,
					Valid:   `👉 {{ . | green | bold }} `,
					Invalid: `👉 {{ . | red | bold }} `,
					Success: `👉 {{ . | bold }} `,
				},
			}

			password, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			saveConfig(types.Entry{URL: url, Username: username, Password: password})

		case "Remove registry":
			var urls []string
			for _, entry := range config.Entries {
				urls = append(urls, strings.Split(entry.URL, "://")[1])
			}

			if (len(urls)) == 0 {
				fmt.Println("No registries found. Please add a registry first.")
				time.Sleep(2 * time.Second)
				continue
			}

			prompt := promptui.Select{
				Label: "Select Registry",
				Items: urls,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
					Help:     `{{ "Press ESC to go back" | bold }}`,
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			config.Entries = append(config.Entries[:i], config.Entries[i+1:]...)
			viper.Set("Entries", config.Entries)
			if err := viper.WriteConfig(); err != nil {
				log.Fatalf("Error writing config: %s", err)
			}

		case "Connect to registry":
			var urls []string
			for _, entry := range config.Entries {
				urls = append(urls, strings.Split(entry.URL, "://")[1])
			}

			if (len(urls)) == 0 {
				fmt.Println("No registries found. Please add a registry first.")
				time.Sleep(2 * time.Second)
				continue
			}
		Registrylist:
			prompt := promptui.Select{
				Label: "Select Registry",
				Items: urls,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
					Help:     `{{ "Press ESC to go back" | bold }}`,
				},
			}

			i, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			selectedRegistry := config.Entries[i]

			fmt.Printf("Repositories for %s:\n", strings.Split(selectedRegistry.URL, "://")[1])
			repositories, err := getRepositories(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password)
			if err != nil {
				fmt.Println("Failed to fetch repositories:", err)
				return
			}

			repoItems := make([]string, len(repositories.Repositories)+1)
			repoItems[0] = "../"
			copy(repoItems[1:], repositories.Repositories)

		Repolist:
			prompt = promptui.Select{
				Label: "Select Repository",
				Items: repoItems,
				Size:  30,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
					Help:     `{{ "Press ESC to go back" | bold }}`,
				},
			}

			_, result, err = prompt.Run()

			if result == "../" {
				goto Registrylist
			}

			selectedRepository := result

			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			tags, err := getTags(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, result)
			if err != nil {
				fmt.Println("Failed to fetch tags:", err)
				return
			}

			if (len(tags.Tags)) == 0 {
				fmt.Println("No tags found for this repository.")
				time.Sleep(2 * time.Second)
				continue
			}

			tagInfos := make([]types.TagInfo, len(tags.Tags))

			for i, tag := range tags.Tags {
				manifest, err := getManifest(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, result, tag)
				digest := manifest.Config.Digest
				if err != nil {
					// Check if error is because manifest doesn't exist
					if strings.Contains(err.Error(), "MANIFEST_UNKNOWN") {
						tagInfos[i] = types.TagInfo{
							Name: tag + " (empty)",
							Date: "N/A",
							Size: 0,
						}
					} else {
						fmt.Println("Failed to fetch manifest:", err)
						return
					}
				} else {
					// Calculate total size
					totalSize := manifest.Config.Size
					for _, layer := range manifest.Layers {
						totalSize += layer.Size
					}

					blobResp, err := getBlob(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, result, digest)
					if err != nil {
						fmt.Println("Failed to fetch blob:", err)
						return
					}

					daysAgo, err := utils.ConvertToDaysAgo(blobResp.Created)
					if err != nil {
						fmt.Println("Failed to convert timestamp:", err)
						return
					}

					tagInfos[i] = types.TagInfo{
						Name: tag,
						Date: daysAgo,
						Size: totalSize, // Add size here
					}
				}
			}

			tagItems := make([]string, len(tagInfos)+1)
			tagItems[0] = "../"
			for i, info := range tagInfos {
				tagItems[i+1] = fmt.Sprintf("%s (Created %s) %s", info.Name, info.Date, utils.FormatBytes(info.Size))
			}
		Taglist:
			prompt = promptui.Select{
				Label: "Select Tag",
				Items: tagItems,
				Size:  100,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
					Help:     `{{ "Press ESC to go back" | bold }}`,
				},
			}
			_, result, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			if result == "../" {
				goto Repolist
			}

			selectedTag := result[:strings.Index(result, " (")]

			prompt = promptui.Select{
				Label: "Select Action",
				Items: []string{"Pull", "Delete", "Exit"},
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
					Help:     `{{ "Press ESC to go back" | bold }}`,
				},
			}

			_, result, err = prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			switch result {
			case "Pull":
				fmt.Println("Pulling...")
				time.Sleep(2 * time.Second)

			case "Delete":
				prompt = promptui.Select{
					Label: "Are you sure?",
					Items: []string{"Yes", "No"},
					Templates: &promptui.SelectTemplates{
						Active:   `👉 {{ . | cyan | bold }}`,
						Inactive: `   {{ . | cyan }}`,
						Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
						Help:     `{{ "Press ESC to go back" | bold }}`,
					},
				}

				_, result, err = prompt.Run()
				if err != nil {
					fmt.Printf("Prompt failed %v\n", err)
					return
				}

				if result == "Yes" {
					manifest, err := getManifest(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, selectedRepository, selectedTag)
					digest := manifest.Config.Digest
					if err != nil {
						fmt.Println("Failed to fetch manifest:", err)
						return
					}

					fmt.Println("Deleting...")
					err = deleteManifest(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, selectedRepository, digest)
					if err != nil {
						fmt.Println("Failed to delete manifest:", err)
						return
					}

					fmt.Println("Delete successful")
					time.Sleep(2 * time.Second)
					goto Taglist
				}
			}

		case "Exit":
			return
		}
	}
}
