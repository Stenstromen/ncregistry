package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/spf13/viper"
)

// MenuItem struct
type MenuItem struct {
	Name     string
	Selected bool
}

var menuItems = []MenuItem{
	{"Add registry", false},
	{"Remove registry", false},
	{"Connect to registry", false},
}

var currentIndex = 0
var formItems = []string{"URL", "Username", "Password", "Add"}
var formIndex = 0
var formMode = false
var formInput = map[string]string{
	"URL":      "",
	"Username": "",
	"Password": "",
}

func saveConfig() {
	// Read in existing config
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		// Config file not found; ignore error
	} else if err != nil {
		// Config file was found but another error was produced
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	// Add new settings
	url := formInput["URL"]
	viper.Set(url+".username", formInput["Username"])
	viper.Set(url+".password", formInput["Password"])

	// Write the configuration
	viper.WriteConfigAs("config.yaml")
}

func main() {
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
		for i, item := range menuItems {
			style := style
			if i == currentIndex {
				style = selectedStyle
			}
			printLine(s, item.Name, 0, i, style)
		}
	}

	// Function to draw the form items
	drawForm := func() {
		for i, item := range formItems {
			style := style
			if i == formIndex {
				style = selectedStyle
			}
			if item != "Add" {
				input := formInput[item]
				if item == "Password" {
					input = strings.Repeat("*", len(input))
				}
				printLine(s, item+": "+input, 0, i, style)
			} else {
				printLine(s, item, 0, i, style)
			}
		}
	}

	if !formMode {
		drawMenu()
	} else {
		drawForm()
	}

	// Show the screen
	s.Show()

	// Wait for a key event
	for {
		event := s.PollEvent()
		switch event := event.(type) {
		case *tcell.EventKey:
			if !formMode {
				switch event.Key() {
				case tcell.KeyUp:
					currentIndex--
					if currentIndex < 0 {
						currentIndex = len(menuItems) - 1
					}
				case tcell.KeyDown:
					currentIndex++
					if currentIndex >= len(menuItems) {
						currentIndex = 0
					}
				case tcell.KeyEnter:
					switch currentIndex {
					case 0:
						// Enter form mode
						formMode = true
					case 1:
						// Code for "Remove registry" here
					case 2:
						// Code for "Connect to registry" here
					}
				case tcell.KeyEscape:
					return
				}
			} else {
				switch event.Key() {
				case tcell.KeyUp:
					formIndex--
					if formIndex < 0 {
						formIndex = len(formItems) - 1
					}
				case tcell.KeyDown:
					formIndex++
					if formIndex >= len(formItems) {
						formIndex = 0
					}
				case tcell.KeyBackspace2, tcell.KeyBackspace:
					if len(formInput[formItems[formIndex]]) > 0 {
						// Delete last character
						formInput[formItems[formIndex]] = formInput[formItems[formIndex]][:len(formInput[formItems[formIndex]])-1]
					}
				case tcell.KeyEnter:
					if formIndex == len(formItems)-1 {
						// "Add" button
						saveConfig()
						formMode = false
					} else {
						// Next field
						formIndex++
					}
				case tcell.KeyEscape:
					// Cancel form
					formMode = false
				case tcell.KeyRune:
					// Handle input
					formInput[formItems[formIndex]] += string(event.Rune())
				}
			}

			s.Clear()
			if !formMode {
				drawMenu()
			} else {
				drawForm()
			}
			s.Show()
		}
	}
}

func printLine(s tcell.Screen, str string, x int, y int, style tcell.Style) {
	for _, r := range str {
		s.SetContent(x, y, r, nil, style)
		x++
	}
}
