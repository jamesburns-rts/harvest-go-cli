package prompt

import (
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/manifoldco/promptui"
	"strings"
)

type (
	Confirmation struct {
		Title      string
		Value      *string
		Validation promptui.ValidateFunc
	}
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

func ForString(title string, validation func(s string) error) (string, error) {
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

func ConfirmAll(confirmations []Confirmation) (err error) {
	for _, c := range confirmations {
		if *c.Value, err = Confirm(c); err != nil {
			return err
		}
	}
	return nil
}

func Confirm(c Confirmation) (string, error) {
	value := strings.ReplaceAll(*c.Value, "\n", " \\")
	prompt := promptui.Prompt{
		Label:     c.Title,
		Validate:  c.Validation,
		Default:   value,
		AllowEdit: true,
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf("%s %s: ", promptui.IconGood, c.Title),
		},
	}

	result, err := prompt.Run()
	result = strings.ReplaceAll(result, " \\", "\n")
	return result, err
}
