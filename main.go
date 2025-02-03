package main

import (
	"code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/util/configv3"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/metskem/panzer-plugin/conf"
	"github.com/metskem/panzer-plugin/event"
	"github.com/metskem/panzer-plugin/version"
	"os"
)

const (
	ListAppsHelpText   = "Lists basic information of apps in the current space"
	ListRoutesHelpText = "Find the routes with their domain/org/space"
)

var (
	ListAppsUsage   = fmt.Sprintf("aa [-a appname-filter] [-q] [-u], use \"cf aa -help\" for full help message - Use the envvar CF_COLS to specify the output columns, available columns are (comma separated): %s", ValidColumns)
	ListRoutesUsage = "lr [-t] <-r host-to-lookup>, use \"cf lr -help\" for full help message- Specify the host without the domain name, we will find all routes using this hostname, if option -t given we will also target the org/space"
)

// PanzerPlugin is the struct implementing the interface defined by the core CLI. It can be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type PanzerPlugin struct{}

// Run must be implemented by any plugin because it is part of the plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined by a plugin.
// The first parameter, plugin.CliConnection, is a struct that can be used to invoke cli commands. The second parameter, args, is a slice of strings.
// args[0] will be the name of the command, and will be followed by any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing user facing errors).
// The CLI will exit 0 if the plugin exits 0 and will exit 1 should the plugin exits nonzero.
func (c *PanzerPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	preCheck(cliConnection)
	cfHomeDir := os.Getenv("CF_HOME")
	if cfHomeDir == "" {
		cfHomeDir = os.Getenv("HOME")
	}
	if cfConfig, err := config.NewFromCFHomeDir(cfHomeDir); err != nil {
		fmt.Printf("failed to create new config: %s", err)
		os.Exit(1)
	} else {
		if conf.CfClient, err = client.New(cfConfig); err != nil {
			fmt.Printf("failed to create new cf client: %s\n", err)
			os.Exit(1)
		}
	}
	switch args[0] {
	case "aa":
		checkTarget(cliConnection)
		listApps(cliConnection)
	case "lr":
		listRoutes()
	case "ev":
		event.GetEvents(cliConnection)
	}
}

// GetMetadata returns a PluginMetadata struct. The first field, Name, determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name during uninstall otherwise the name will be treated as separate arguments.
// The second value is a slice of Command structs. Our slice only contains one Command Struct, but could contain any number of them.
// The first field Name defines the command `cf basic-plugin-command` once installed into the CLI.
// The second field, HelpText, is used by the core CLI to display help information to the user in the core commands `cf help`, `cf`, or `cf -h`.
func (c *PanzerPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:          "panzer",
		Version:       plugin.VersionType{Major: version.GetMajorVersion(), Minor: version.GetMinorVersion(), Build: version.GetPatchVersion()},
		MinCliVersion: plugin.VersionType{Major: 6, Minor: 7, Build: 0},
		Commands: []plugin.Command{
			{Name: "aa", HelpText: ListAppsHelpText, UsageDetails: plugin.Usage{Usage: ListAppsUsage}},
			{Name: "lr", HelpText: ListRoutesHelpText, UsageDetails: plugin.Usage{Usage: ListRoutesUsage}},
			{Name: "ev", HelpText: event.ListEventsHelpText, UsageDetails: plugin.Usage{Usage: event.ListEventsUsage}},
		},
	}
}

// checkTarget Checks if you currently have a targeted org and space.
func checkTarget(cliConnection plugin.CliConnection) {
	hasOrg, err := cliConnection.HasOrganization()
	if err != nil || !hasOrg {
		fmt.Println(terminal.FailureColor("please target your org/space first"))
		os.Exit(1)
	}
	org, _ := cliConnection.GetCurrentOrg()
	conf.CurrentOrg = org
	hasSpace, err := cliConnection.HasSpace()
	if err != nil || !hasSpace {
		fmt.Println(terminal.FailureColor("please target your space first"))
		os.Exit(1)
	}
	space, _ := cliConnection.GetCurrentSpace()
	conf.CurrentSpace = space
}

// preCheck Does all common validations, like being logged in.
func preCheck(cliConnection plugin.CliConnection) {
	v3Config, _ := configv3.LoadConfig()
	i18n.T = i18n.Init(v3Config)
	loggedIn, err := cliConnection.IsLoggedIn()
	if err != nil || !loggedIn {
		fmt.Println(terminal.NotLoggedInText())
		os.Exit(1)
	}
	conf.CurrentUser, _ = cliConnection.Username()
}

// Unlike most Go programs, the `Main()` function will not be used to run all the commands provided in your plugin.
// Main will be used to initialize the plugin process, as well as any dependencies you might require for your plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect metadata. The plugin will exit 0 and the Run([]string) method will not be invoked.
	plugin.Start(new(PanzerPlugin))
	// Plugin code should be written in the Run([]string) method, ensuring the plugin environment is bootstrapped.
}
