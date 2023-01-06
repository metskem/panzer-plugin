package conf

import plugin_models "code.cloudfoundry.org/cli/plugin/models"

const (
	DefaultHttpTimeout = 60
)

var (
	CurrentOrg            plugin_models.Organization
	CurrentSpace          plugin_models.Space
	CurrentUser           string
	SkipSSLValidation     bool
	AccessToken           string
	ApiEndpoint           string
	FlagLimit             = 500
	FlagFilterEventAction string
	FlagFilterEventTarget string
	FlagFilterEventType   string
	FlagFilterEventActor  string
	FlagFilterOrgName     string
	FlagFilterSpaceName   string
	FlagSwitchToSpace     bool
	FlagRoute             string
	FlagAppName           string
)
