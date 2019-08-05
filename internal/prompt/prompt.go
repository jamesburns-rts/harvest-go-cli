package prompt

import (
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
		Templates: getSelectionTemplate(options),
	}

	i, _, err := prompt.Run()

	if err != nil {
		if err == promptui.ErrInterrupt {
			return 0, util.QuitError
		}
	}
	return i, nil
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

func ForWord(title string) (string, error) {
	return ForStringWithValidation(title, func(s string) error {
		if s == "" || strings.ContainsAny(s, " \t\n") {
			return errors.New("Must input word with no spaces")
		}
		return nil
	})
}

func ForString(title string) (string, error) {
	return ForStringWithValidation(title, func(s string) error {
		if s == "" {
			return errors.New("Must input value")
		}
		return nil
	})
}

func ForStringWithValidation(title string, validation func(s string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    title,
		Validate: validation,
	}

	result, err := prompt.Run()
	return result, err
}
