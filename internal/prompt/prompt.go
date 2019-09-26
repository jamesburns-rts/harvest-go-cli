package prompt

import (
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/manifoldco/promptui/list"
	"strings"
)

type (
	Value interface {
		String() string
		Set(string) error
	}

	Confirmation struct {
		Title string
		Value Value
	}
)

func ForSelection(title string, options interface{}) (int, error) {

	searcher := getSearcher(options)

	prompt := promptui.Select{
		Label:             title,
		Items:             options,
		Size:              10,
		Templates:         getSelectionTemplate(title, options),
		StartInSearchMode: searcher != nil,
		Searcher:          searcher,
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
			Active:   fmt.Sprintf("%s {{ .ClientName }}: {{ .Name }}", promptui.IconSelect),
			Inactive: "  {{ .ClientName | faint }}: {{ .Name | faint }}",
			Selected: fmt.Sprintf("%s %s: {{ .ClientName }}: {{ .Name }}", promptui.IconGood, title),
		}
	case []harvest.Task:
		return &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ .Name }}", promptui.IconSelect),
			Inactive: "  {{ .Name | faint }}",
			Selected: fmt.Sprintf("%s %s: {{ .Name }}", promptui.IconGood, title),
		}
	case []string:
		return &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ . }}", promptui.IconSelect),
			Inactive: "  {{ . | faint }}",
			Selected: fmt.Sprintf("%s %s: {{ . }}", promptui.IconGood, title),
		}
	default:
		return nil
	}
}

func getSearcher(options interface{}) list.Searcher {
	switch options.(type) {
	case []harvest.Project:
		return func(input string, index int) bool {
			project := options.([]harvest.Project)[index]
			return fuzzy.Match(strings.ToLower(input), strings.ToLower(project.Name))
		}
	case []harvest.Task:
		return func(input string, index int) bool {
			task := options.([]harvest.Task)[index]
			return fuzzy.Match(strings.ToLower(input), strings.ToLower(task.Name))
		}
	case []string:
		return func(input string, index int) bool {
			item := options.([]string)[index]
			return fuzzy.Match(strings.ToLower(input), strings.ToLower(item))
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
		if result, err := Confirm(c); err != nil {
			return err
		} else if err = c.Value.Set(result); err != nil {
			return err
		}
	}
	return nil
}

func Confirm(c Confirmation) (string, error) {
	value := strings.ReplaceAll(c.Value.String(), "\n", " \\")
	prompt := promptui.Prompt{
		Label:     c.Title,
		Validate:  c.Value.Set,
		Default:   value,
		AllowEdit: true,
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf("%s %s: ", promptui.IconGood, c.Title),
		},
	}

	result, err := prompt.Run()
	result = strings.ReplaceAll(result, " \\ ", "\n")
	result = strings.ReplaceAll(result, " \\", "\n")
	return result, err
}
