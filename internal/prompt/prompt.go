package prompt

import (
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/manifoldco/promptui"
)

func ForSelection(title string, options interface{}) int {
	prompt := promptui.Select{
		Label:     title,
		Items:     options,
		Templates: getSelectionTemplate(options),
	}

	i, _, err := prompt.Run()

	if err != nil {
		panic(fmt.Sprintf("Prompt failed %v\n", err))
	}
	return i
}

func getSelectionTemplate(options interface{}) *promptui.SelectTemplates {

	switch options.(type) {
	case []harvest.Project:
		return &promptui.SelectTemplates{
			Active:   "{{ .Name }}",
			Inactive: "{{ .Name | faint }}",
			Selected: promptui.IconGood + " {{ .Name }}",
		}
	case []harvest.Task:
		return &promptui.SelectTemplates{
			Active:   "{{ .Name }}",
			Inactive: "{{ .Name | faint }}",
			Selected: promptui.IconGood + " {{ .Name }}",
		}
	default:
		return nil
	}
}
