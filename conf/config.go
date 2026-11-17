package conf

import (
	"context"
	"regexp"

	pluginmodels "code.cloudfoundry.org/cli/v8/plugin/models"
	"github.com/cloudfoundry/go-cfclient/v3/client"
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
	FlagTimeBefore            string
	FlagTimeAfter             string
	FlagIncludeEventData      bool
	AppNameRegex              regexp.Regexp
)
