package main

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/integrii/flaggy"
	"github/metskem/panzer-plugin/conf"
	"github/metskem/panzer-plugin/model"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

var colNames = []string{"hostname", "domain", "org", "space", "bound apps"}

/** listRoutes - The main function to produce the response to list routes. */
func listRoutes() {
	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.ShowVersionWithVersionFlag = false
	// Add flags
	flaggy.Bool(&conf.FlagSwitchToSpace, "t", "target", "cf target the space where the route is found")
	flaggy.String(&conf.FlagRoute, "r", "route", "the route to lookup (specify only hostname, without the domain name)")
	// Parse the flags
	flaggy.Parse()

	if conf.FlagRoute == "" {
		fmt.Println("Please use the -r flag to specify the route name")
		os.Exit(1)
	}

	fmt.Printf("Getting routes for hostname %s as %s...\n\n", terminal.EntityNameColor(conf.FlagRoute), terminal.EntityNameColor(conf.CurrentUser))
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation}}
	httpClient = http.Client{Transport: transport, Timeout: time.Duration(conf.DefaultHttpTimeout) * time.Second}
	requestHeader = map[string][]string{"Content-Type": {"application/json"}, "Authorization": {conf.AccessToken}}
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/routes?per_page=100&hosts=%s", conf.ApiEndpoint, conf.FlagRoute))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	routesListResponse := model.RoutesListResponse{}
	err = json.Unmarshal(body, &routesListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse routes response: %s", err)))
	}
	if len(routesListResponse.Resources) == 0 {
		fmt.Printf("no routes found for hostname %s\n", conf.FlagRoute)
	} else {
		table := terminal.NewTable(colNames)
		var orgName, spaceName string
		for _, route := range routesListResponse.Resources {
			var colValues [5]string
			colValues[0] = conf.FlagRoute
			colValues[1] = getAPIResource(route.Relationships.Domain.Data.GUID, "domains").(model.Domain).Name
			space := getAPIResource(route.Relationships.Space.Data.GUID, "spaces").(model.Space)
			colValues[2] = getAPIResource(space.Relationships.Organization.Data.GUID, "organizations").(model.Org).Name
			colValues[3] = space.Name
			table.Add(colValues[:]...)
			orgName = colValues[2]
			spaceName = space.Name
			var destList string
			for _, dest := range route.Destinations {
				appName := getAPIResource(dest.App.GUID, "apps").(model.App).Name
				destList = fmt.Sprintf("%s%s ", destList, appName)
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

func getAPIResource(guid string, apiResource string) interface{} {
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/%s/%s", conf.ApiEndpoint, apiResource, guid))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	switch apiResource {
	case "organizations":
		org := model.Org{}
		if err = json.Unmarshal(body, &org); err == nil {
			return org
		}
	case "spaces":
		space := model.Space{}
		if err = json.Unmarshal(body, &space); err == nil {
			return space
		}
	case "domains":
		domain := model.Domain{}
		if err = json.Unmarshal(body, &domain); err == nil {
			return domain
		}
	case "apps":
		app := model.App{}
		if err = json.Unmarshal(body, &app); err == nil {
			return app
		}
	}
	fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get response: %s", err)))
	return nil
}
