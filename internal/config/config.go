package config

type (
	HarvestProperties struct {
		AccessToken    string           `yaml:"accessToken"json:"accessToken"`
		AccountId      string           `yaml:"accountId"json:"accessToken"`
		ProjectAliases map[string]int64 `yaml:"projectAliases"json:"projectAliases"`
		TaskAliases    map[string]int64 `yaml:"taskAliases"json:"taskAliases"`
	}
)

var Harvest HarvestProperties
