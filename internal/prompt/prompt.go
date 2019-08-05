package prompt

import (
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"strings"
)

func ForSelection(title string, options interface{}) (int, error) {
	prompt := promptui.Select{
		Label:     title,
		Items:     options,
		Templates: getSelectionTemplate(title, options),
	}

	i, _, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrInterrupt {
			return 0, util.QuitError
		}
	}
	return i, nil
}

func getSelectionTemplate(title string, options interface{}) *promptui.SelectTemplates {

	switch options.(type) {
	case []harvest.Project:
		return &promptui.SelectTemplates{
			Active:   "{{ .Name }}",
			Inactive: "{{ .Name | faint }}",
			Selected: fmt.Sprintf("%s %s: {{ .Name }}", promptui.IconGood, title),
		}
	case []harvest.Task:
		return &promptui.SelectTemplates{
			Active:   "{{ .Name }}",
			Inactive: "{{ .Name | faint }}",
			Selected: fmt.Sprintf("%s %s: {{ .Name }}", promptui.IconGood, title),
		}
	default:
		return nil
	}
}

func ForWord(title string) (string, error) {
	return ForStringWithValidation(title, func(s string) error {
		if s == "" || strings.ContainsAny(s, " \t\n") {
			return errors.New("Must input word with no spaces")
		}
		return nil
	})
}

func ForString(title string) (string, error) {
	return ForStringWithValidation(title, nil)
}

func ForStringWithValidation(title string, validation func(s string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    title,
		Validate: validation,
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf("%s %s: ", promptui.IconGood, title),
		},
	}

	result, err := prompt.Run()
	return result, err
}
