package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"github.com/stenstromen/ncregistry/config"
	"github.com/stenstromen/ncregistry/registry"
	"github.com/stenstromen/ncregistry/types"
	"github.com/stenstromen/ncregistry/utils"
)

func promptSelect(label string, items []string) (string, error) {
	prompt := promptui.Select{
		HideSelected: true,
		Label:        label,
		Items:        items,
		Templates: &promptui.SelectTemplates{
			Active:   `👉 {{ . | cyan | bold }}`,
			Inactive: `   {{ . | cyan }}`,
			Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
			Help:     `{{ "Use ↑/↓ to move and Enter to select" | bold }}`,
		},
	}
	_, result, err := prompt.Run()

	return result, err
}

func promptInput(label string, mask rune) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Mask:  mask,
		Templates: &promptui.PromptTemplates{
			Prompt:  `👉 {{ . | cyan | bold }} `,
			Valid:   `👉 {{ . | green | bold }} `,
			Invalid: `👉 {{ . | red | bold }} `,
			Success: `👉 {{ . | bold }} `,
		},
	}
	result, err := prompt.Run()

	return result, err
}

func handlePromptError(err error) {
	fmt.Printf("Prompt failed %v\n", err)
}

func addRegistry() {
	utils.ClearTerminal()
	url, err := promptInput("Registry URL", 0)
	if err != nil {
		handlePromptError(err)
		return
	}

	username, err := promptInput("Registry Username", 0)
	if err != nil {
		handlePromptError(err)
		return
	}

	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	password, err := promptInput("Registry Password", '*')
	if err != nil {
		handlePromptError(err)
		return
	}

	config.SaveConfig(types.Entry{URL: url, Username: username, Password: password})
}

func removeRegistry() {
	utils.ClearTerminal()
	var urls []string
	urls = append(urls, "../")
	for _, entry := range config.Config.Entries {
		urls = append(urls, strings.Split(entry.URL, "://")[1])
	}

	if len(urls) == 1 {
		fmt.Println("No registries found. Please add a registry first.")
		time.Sleep(2 * time.Second)
		return
	}

	prompt := promptui.Select{
		HideSelected: true,
		Label:        "Select Registry",
		Items:        urls,
		Templates: &promptui.SelectTemplates{
			Active:   `👉 {{ . | cyan | bold }}`,
			Inactive: `   {{ . | cyan }}`,
			Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
			Help:     `{{ "Remove registry" | bold }}`,
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if i == 0 {
		return
	}

	config.Config.Entries = append(config.Config.Entries[:i-1], config.Config.Entries[i:]...)
	viper.Set("Entries", config.Config.Entries)
	if err := viper.WriteConfig(); err != nil {
		log.Fatalf("Error writing config: %s", err)
	}
}

func main() {
	config.InitConfig()

	for {

		utils.ClearTerminal()

		result, err := promptSelect("Main Menu", []string{"Add registry", "Remove registry", "Connect to registry", "Exit"})
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Add registry":
			addRegistry()

		case "Remove registry":
			removeRegistry()

		case "Connect to registry":
			utils.ClearTerminal()
			var urls []string
			urls = append(urls, "../")
			urls[0] = "../"
			for _, entry := range config.Config.Entries {
				urls = append(urls, strings.Split(entry.URL, "://")[1])
			}

			if (len(urls)) == 1 {
				fmt.Println("No registries found. Please add a registry first.")
				time.Sleep(2 * time.Second)
				continue
			}
		Registrylist:
			prompt := promptui.Select{
				Label:        "Select Registry",
				HideSelected: true,
				Items:        urls,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
					Help:     `{{ "Connect to registry" | bold }}`,
				},
			}

			i, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			if result == "../" {
				continue
			}

			selectedRegistry := config.Config.Entries[i-1]

			fmt.Printf("Repositories for %s:\n", strings.Split(selectedRegistry.URL, "://")[1])
			repositories, err := registry.GetRepositories(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password)
			if err != nil {
				fmt.Println("Failed to fetch repositories:", err)
				return
			}

			repoItems := make([]string, len(repositories.Repositories)+1)
			repoItems[0] = "../"
			copy(repoItems[1:], repositories.Repositories)

		Repolist:
			prompt = promptui.Select{
				Label:        "Select Repository",
				HideHelp:     true,
				HideSelected: true,
				Items:        repoItems,
				Size:         30,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
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

			tags, err := registry.GetTags(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, result)
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
				manifest, err := registry.GetManifest(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, result, tag)
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

					blobResp, err := registry.GetBlob(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, result, digest)
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
			utils.ClearTerminal()
			prompt = promptui.Select{
				Label:        "Select Tag",
				HideHelp:     true,
				HideSelected: true,
				Items:        tagItems,
				Size:         100,
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
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
				Label:        "Select Action",
				HideHelp:     true,
				HideSelected: true,
				Items:        []string{"../", "Pull", "Delete"},
				Templates: &promptui.SelectTemplates{
					Active:   `👉 {{ . | cyan | bold }}`,
					Inactive: `   {{ . | cyan }}`,
					Selected: `{{ "✔" | green | bold }} {{ "Selected Option:" | bold }} {{ . | cyan }}`,
				},
			}

			_, result, err = prompt.Run()
			if err != nil || result == "../" {
				goto Taglist
			}

			switch result {
			case "Pull":
				fmt.Println("Pulling...")
				err = registry.DockerPull(strings.Split(selectedRegistry.URL, "://")[1], selectedRepository, selectedTag)
				if err != nil {
					fmt.Println("Failed to pull image:", err)
					return
				}
				goto Taglist

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
					manifest, err := registry.GetManifest(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, selectedRepository, selectedTag)
					digest := manifest.Config.Digest
					if err != nil {
						fmt.Println("Failed to fetch manifest:", err)
						return
					}

					fmt.Println("Deleting...")
					err = registry.DeleteManifest(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password, selectedRepository, digest)
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
