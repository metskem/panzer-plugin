package main

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/integrii/flaggy"
	"github.com/metskem/panzer-plugin/conf"
	"os"
	"os/exec"
)

var colNames = []string{"hostname", "domain", "org", "space", "bound apps"}

/** listRoutes - The main function to produce the response to list routes. */
func listRoutes() {
	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.ShowVersionWithVersionFlag = false
	flaggy.Bool(&conf.FlagSwitchToSpace, "t", "target", "cf target the space where the route is found")
	flaggy.String(&conf.FlagRoute, "r", "route", "the route to lookup (specify only hostname, without the domain name)")
	flaggy.Parse()

	if conf.FlagRoute == "" {
		fmt.Println("Please use the -r flag to specify the route name")
		os.Exit(1)
	}

	fmt.Printf("Getting routes for hostname %s as %s...\n\n", terminal.EntityNameColor(conf.FlagRoute), terminal.EntityNameColor(conf.CurrentUser))
	routeListOptions := client.RouteListOptions{ListOptions: &client.ListOptions{}, Hosts: client.Filter{Values: []string{conf.FlagRoute}}}
	if routes, err := conf.CfClient.Routes.ListAll(conf.CfCtx, &routeListOptions); err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get routes: %s", err)))
	} else {
		if len(routes) == 0 {
			fmt.Printf("no routes found for hostname %s\n", conf.FlagRoute)
		} else {
			table := terminal.NewTable(colNames)
			var orgName, spaceName string
			for _, route := range routes {
				var colValues [5]string
				colValues[0] = conf.FlagRoute
				domain, _ := conf.CfClient.Domains.Get(conf.CfCtx, route.Relationships.Domain.Data.GUID)
				colValues[1] = domain.Name
				space, _ := conf.CfClient.Spaces.Get(conf.CfCtx, route.Relationships.Space.Data.GUID)
				org, _ := conf.CfClient.Organizations.Get(conf.CfCtx, space.Relationships.Organization.Data.GUID)
				colValues[2] = org.Name
				colValues[3] = space.Name
				table.Add(colValues[:]...)
				orgName = colValues[2]
				spaceName = colValues[3]
				var destList string
				for _, dest := range route.Destinations {
					app, _ := conf.CfClient.Applications.Get(conf.CfCtx, *dest.App.GUID)
					destList = fmt.Sprintf("%s%s ", destList, app.Name)
				}
				colValues[4] = destList
			}
			_ = table.PrintTo(os.Stdout)
			if conf.FlagSwitchToSpace {
				// You normally would use cliConnection.CliCommand, but that screws up my "NetworkPolicyV1Endpoint" in my cf config.json. So instead issue os command:
				cmd := exec.Command("cf", "target", "-o", orgName, "-s", spaceName)
				if err = cmd.Run(); err != nil {
					fmt.Printf("failed to set target to org %s and space %s: %s", orgName, spaceName, err)
				}
			}
		}
	}
}
