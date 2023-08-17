/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)
var BRANCH = ""

func getFromInput(placeholder string) string {
	fmt.Printf("Enter %s: ", placeholder)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occur while reading input", err.Error())
		os.Exit(1)
	}
	return strings.TrimSuffix(input, "\n")
}

func runCli(cmd string) string {
	out, err := exec.Command("zsh", "-c", cmd).Output()
	if err != nil {
		fmt.Printf("Something wrong when run %s \n, %s", cmd, err.Error())
		os.Exit(1)
	}

	return string(out)
}

func gitBranchHandler(cmd *cobra.Command) {
	inputBranch, err := cmd.Flags().GetString("branch")

	if err == nil || inputBranch == "" {
		inputBranch = getFromInput("branch")
	}

	currentBranch := strings.Trim(runCli("git branch --show-current"), "\n")

	if inputBranch == currentBranch {
		BRANCH = currentBranch
		fmt.Println("Working on same branch. Skip create new branch")
	} else {
		fmt.Print("Checkout new branch....")
		BRANCH = inputBranch
		out := runCli("git checkout -b " + inputBranch)
		cmd.Flags()
		fmt.Println(out)
	}

}

type item struct {
	ID         string
	IsSelected bool
}

// selectItems() prompts user to select one or more items in the given slice
func selectItems(selectedPos int, allItems []*item) ([]*item, error) {
	// Always prepend a "Done" item to the slice if it doesn't
	// already exist.
	const doneID = "Done"
	if len(allItems) > 0 && allItems[0].ID != doneID {
		var items = []*item{
			{
				ID: doneID,
			},
		}
		allItems = append(items, allItems...)
	}

	// Define promptui template
	templates := &promptui.SelectTemplates{
		Label: `{{if .IsSelected}}
                    ✔
                {{end}} {{ .ID }} - label`,
		Active:   "→ {{if .IsSelected}}✔ {{end}}{{ .ID | cyan }}",
		Inactive: "{{if .IsSelected}}✔ {{end}}{{ .ID | cyan }}",
	}

	prompt := promptui.Select{
		Label:     "Item",
		Items:     allItems,
		Templates: templates,
		Size:      5,
		// Start the cursor at the currently selected index
		CursorPos:    selectedPos,
		HideSelected: true,
	}

	selectionIdx, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	chosenItem := allItems[selectionIdx]

	if chosenItem.ID != doneID {
		// If the user selected something other than "Done",
		// toggle selection on this item and run the function again.
		chosenItem.IsSelected = !chosenItem.IsSelected
		return selectItems(selectionIdx, allItems)
	}

	// If the user selected the "Done" item, return
	// all selected items.
	var selectedItems []*item
	for _, i := range allItems {
		if i.IsSelected {
			selectedItems = append(selectedItems, i)
		}
	}
	return selectedItems, nil
}

func gitAddHandler() {
	// get all files change
	fmt.Println("All files change: ")
	out := runCli("git status -s")
	fmt.Println(out)
	option := getFromInput("1. Add all | 2. Select manual")
	if option == "1" {
		fmt.Println("Add all file")
		runCli("git add .")
	} else {
		var allItems []*item
		listFiles := strings.Split(out, "\n")

		for i := 0; i < len(listFiles); i++ {
			if listFiles[i] != "" {
				allItems = append(allItems, &item{listFiles[i][3:], false})
			}
		}

		listFileResult, err := selectItems(1, allItems)
		if err != nil || len(listFileResult) == 0 {
			fmt.Printf("Have some error happen when get all file need to add")
			fmt.Println("Will run add all file")
			runCli("git add .")
		}

		for j := 0; j < len(listFileResult); j++ {
			if listFileResult[j].IsSelected {
				fmt.Printf("Adding file %s \n", listFileResult[j].ID)
				runCli("git add " + listFileResult[j].ID)
			}
		}

	}

}

func gitCommitHandler(cmd *cobra.Command) {
	msg, err := cmd.Flags().GetString("message")
	if err == nil || msg == "" {
		msg = getFromInput("commit message")
	}
	fmt.Printf("Commit file with message: %s \n", msg)
	runCli("git commit -m " + "\"" + msg + "\"")
}

func gitPushHandler(cmd *cobra.Command) {
	isVerify, err := cmd.Flags().GetBool("check")
	if err != nil {
		fmt.Println(err.Error())
	}
	
	if isVerify {
		fmt.Println("Push code")
		runCli("git push origin " + BRANCH)
	} else {
		fmt.Println("Push code and bypass check requirement")
		runCli("git push origin " + BRANCH + " --no-verify")
	}
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all cmd",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting....")
		gitBranchHandler(cmd)
		gitAddHandler()
		gitCommitHandler(cmd)
		gitPushHandler(cmd)
		fmt.Println("Done....")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().StringP("message", "m", "", "Message to commit")
	runCmd.Flags().StringP("branch", "b", "", "New branch")
	runCmd.Flags().BoolP("check", "c", false, "Need to check all requirement before push code")

}
