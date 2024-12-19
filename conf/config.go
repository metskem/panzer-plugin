package conf

import (
	pluginmodels "code.cloudfoundry.org/cli/plugin/models"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"regexp"
)

const (
	DefaultHttpTimeout = 60
)

var (
	CfClient                  *client.Client
	CurrentOrg                pluginmodels.Organization
	CurrentSpace              pluginmodels.Space
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
	FlagShowQuotaUsage        bool
	AppNameRegex              regexp.Regexp
)
