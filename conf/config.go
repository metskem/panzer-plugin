package conf

import (
	pluginmodels "code.cloudfoundry.org/cli/plugin/models"
	"context"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"regexp"
)

var (
	CfClient                  *client.Client
	CfCtx                     = context.Background()
	CurrentOrg                pluginmodels.Organization
	CurrentSpace              pluginmodels.Space
	CurrentUser               string
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
