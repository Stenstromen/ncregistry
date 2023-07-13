package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Repository struct {
	Registry  string
	Name      string
	Tag       string
	CreatedAt time.Time
	Size      int64
}

// MenuItem struct
type MenuItem struct {
	Name     string
	Selected bool
}

var menuItems = []MenuItem{
	{"Add registry", false},
	{"Remove registry", false},
	{"Connect to registry", false},
	{"Exit", false},
}

var currentIndex = 0
var formItems = []string{"URL", "Username", "Password", "Save"}
var formIndex = 0
var formMode = false
var registryMode = false
var exitMode = false
var formInput = map[string]string{
	"URL":      "",
	"Username": "",
	"Password": "",
}

type Entry struct {
	URL      string
	Username string
	Password string
}

type Config struct {
	Entries []Entry
}

var currentRegistryIndex = 0
var config Config

func saveConfig() {
	newEntry := Entry{
		URL:      formInput["URL"],
		Username: formInput["Username"],
		Password: formInput["Password"],
	}
	config.Entries = append(config.Entries, newEntry)
	viper.Set("Entries", config.Entries)
	if err := viper.WriteConfig(); err != nil {
		log.Fatalf("Error writing config: %s", err)
	}
}

func printLineWithPadding(s tcell.Screen, str string, x int, y int, style tcell.Style, padding int) {
	startX := x + ((padding - len(str)) / 2)
	for _, r := range str {
		s.SetContent(startX, y, r, nil, style)
		startX++
	}
}

func printCenteredLineWithPadding(s tcell.Screen, str string, xStart, y int, style tcell.Style, padding int) {
	boxWidth := 2 * padding
	strStart := xStart + (boxWidth-len(str))/2 + 2
	for _, r := range str {
		s.SetContent(strStart, y, r, nil, style)
		strStart++
	}
}

func drawBorder(s tcell.Screen, x1, x2, y1, y2 int, style tcell.Style) {
	// Draw horizontal borders
	for x := x1; x <= x2; x++ {
		s.SetContent(x, y1, '-', nil, style)
		s.SetContent(x, y2, '-', nil, style)
	}
	// Draw vertical borders
	for y := y1; y <= y2; y++ {
		s.SetContent(x1, y, '|', nil, style)
		s.SetContent(x2, y, '|', nil, style)
	}
	// Draw corners
	s.SetContent(x1, y1, '+', nil, style)
	s.SetContent(x1, y2, '+', nil, style)
	s.SetContent(x2, y1, '+', nil, style)
	s.SetContent(x2, y2, '+', nil, style)
}

func readConfig() {
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}

func getRepositories(url, username, password string) (Repository, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+"/v2/_catalog", nil)
	if err != nil {
		return Repository{}, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return Repository{}, err
	}
	defer resp.Body.Close()

	var repo Repository
	err = yaml.NewDecoder(resp.Body).Decode(&repo)
	return repo, err
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

	readConfig()

	// Create a new screen
	s, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	if err = s.Init(); err != nil {
		panic(err)
	}

	// Defer a cleanup function
	defer s.Fini()

	// Set the style
	style := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)

	// Set the selected style
	selectedStyle := tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack)

	// Function to draw the menu items
	drawMenu := func() {
		width, height := s.Size()
		startY := (height - len(menuItems)) / 2
		padding := 10
		maxLength := len(menuItems[0].Name) + padding
		xStart := (width - maxLength) / 2
		xEnd := xStart + maxLength
		drawBorder(s, xStart-1, xEnd+1, startY-1, startY+len(menuItems)+1, style)

		for i, item := range menuItems {
			currentStyle := style
			if i == currentIndex {
				currentStyle = selectedStyle
			}
			printCenteredLineWithPadding(s, item.Name, xStart, startY+i, currentStyle, padding)
		}
	}

	// Function to draw the form items
	drawForm := func() {
		width, height := s.Size()
		startY := (height - len(formItems)) / 2
		padding := 10
		xStart := (width - len(formItems[0]) - padding) / 2
		xEnd := xStart + len(formItems[0]) + padding
		drawBorder(s, xStart-1, xEnd+1, startY-1, startY+len(formItems)+2, style)

		for i, item := range formItems {
			currentStyle := style
			if i == formIndex {
				currentStyle = selectedStyle
			}
			if item == "Save" {
				printCenteredLineWithPadding(s, item, xStart, startY+i, currentStyle, padding)
			} else {
				input := formInput[item]
				if item == "Password" {
					input = strings.Repeat("*", len(input))
				}
				printLineWithPadding(s, item+": "+input, xStart, startY+i, currentStyle, padding)
			}
		}
	}

	// Function to draw the registry items
	drawRegistries := func() {
		width, height := s.Size()
		startY := (height - len(config.Entries)) / 2
		padding := 10
		maxLength := 0
		for _, entry := range config.Entries {
			if len(entry.URL) > maxLength {
				maxLength = len(entry.URL)
			}
		}
		maxLength += padding
		xStart := (width - maxLength) / 2
		xEnd := xStart + maxLength
		drawBorder(s, xStart-1, xEnd+1, startY-1, startY+len(config.Entries)+1, style)

		for i, entry := range config.Entries {
			currentStyle := style
			if i == currentRegistryIndex {
				currentStyle = selectedStyle
			}
			printCenteredLineWithPadding(s, entry.URL, xStart, startY+i, currentStyle, padding)
		}
	}

	drawMenu()
	s.Show()

	for {
		if exitMode {
			break
		}
		event := s.PollEvent()
		switch event := event.(type) {
		case *tcell.EventKey:
			switch event.Key() {
			case tcell.KeyUp:
				if formMode {
					formIndex--
					if formIndex < 0 {
						formIndex = len(formItems) - 1
					}
				} else if registryMode {
					currentRegistryIndex--
					if currentRegistryIndex < 0 {
						currentRegistryIndex = len(config.Entries) - 1
					}
				} else {
					currentIndex--
					if currentIndex < 0 {
						currentIndex = len(menuItems) - 1
					}
				}
			case tcell.KeyDown:
				if formMode {
					formIndex++
					if formIndex >= len(formItems) {
						formIndex = 0
					}
				} else if registryMode {
					currentRegistryIndex++
					if currentRegistryIndex >= len(config.Entries) {
						currentRegistryIndex = 0
					}
				} else {
					currentIndex++
					if currentIndex >= len(menuItems) {
						currentIndex = 0
					}
				}
			case tcell.KeyEnter:
				if formMode {
					if formItems[formIndex] == "Save" {
						// Submit the form
						saveConfig()
						formMode = false
						formInput = map[string]string{
							"URL":      "",
							"Username": "",
							"Password": "",
						}
					}
				} else if registryMode {
					// Get repositories for selected registry
					selectedRegistry := config.Entries[currentRegistryIndex]
					repositories, err := getRepositories(selectedRegistry.URL, selectedRegistry.Username, selectedRegistry.Password)
					if err != nil {
						fmt.Printf("Failed to fetch repositories: %v\n", err)
					} else {
						fmt.Printf("Repositories for %s:\n", selectedRegistry.URL)
						for _, repo := range repositories {
							fmt.Printf("- %s\n", repo)
						}
					}
					registryMode = false
				} else {
					if currentIndex == 0 {
						formMode = true
					} else if currentIndex == 1 {
						// remove registry
					} else if currentIndex == 2 {
						registryMode = true
					} else if currentIndex == 3 {
						exitMode = true
					}
				}
			case tcell.KeyEscape:
				formMode = false
				registryMode = false
			case tcell.KeyRune:
				if formMode && formIndex < len(formItems) {
					formInput[formItems[formIndex]] += string(event.Rune())
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if formMode && formIndex < len(formItems) {
					formInput[formItems[formIndex]] = strings.TrimSuffix(formInput[formItems[formIndex]], formInput[formItems[formIndex]][len(formInput[formItems[formIndex]])-1:])
				}
			}

			s.Clear()
			if !formMode && !registryMode && !exitMode {
				drawMenu()
			} else if formMode {
				drawForm()
			} else if registryMode {
				drawRegistries()
			}
			s.Show()
		}
	}
}
