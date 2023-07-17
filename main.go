package main

import (
	"fmt"

	"github.com/stenstromen/ncregistry/config"
	"github.com/stenstromen/ncregistry/prompts"
	"github.com/stenstromen/ncregistry/utils"
)

func main() {
	config.InitConfig()

	for {
		utils.ClearTerminal()

		result, err := prompts.Init()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Add registry":
			prompts.AddRegistry()
		case "Remove registry":
			prompts.RemoveRegistry()
		case "Connect to registry":
			prompts.ConnectToRegistry()
		case "Exit":
			return
		}

	}
}
