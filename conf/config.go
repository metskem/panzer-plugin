package conf

import (
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"regexp"
)

const (
	DefaultHttpTimeout = 60
)

var (
	CurrentOrg                plugin_models.Organization
	CurrentSpace              plugin_models.Space
	CurrentUser               string
	SkipSSLValidation         bool
	AccessToken               string
	ApiEndpoint               string
	FlagLimit                 = 500
	FlagFilterEventTargetName string
	FlagFilterEventTargetType string
	FlagFilterEventTypes      string
	FlagFilterEventActor      string
	FlagFilterEventOrgName    string
	FlagFilterEventSpaceName  string
	FlagSwitchToSpace         bool
	FlagRoute                 string
	FlagAppName               string
	FlagHideHeaders           bool
	AppNameRegex              regexp.Regexp
)
