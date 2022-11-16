package main

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

var colNames = []string{"hostname", "domain", "org", "space", "bound apps"}

/** listRoutes - The main function to produce the response to list routes. */
func listRoutes(args []string, cliConnection plugin.CliConnection) {
	if len(args) < 2 || len(args) > 3 {
		fmt.Printf("Usage: \"cf lr [-t] <hostname>\"\n\nNAME:\n   %s\n\nUSAGE:\n   %s\n", ListRoutesHelpText, ListRoutesUsage)
		os.Exit(1)
	}
	hostname := args[1]
	setTarget := false
	if len(args) == 3 && args[1] == "-t" {
		hostname = args[2]
		setTarget = true
	}

	fmt.Printf("Getting routes for hostname %s as %s...\n\n", terminal.EntityNameColor(hostname), terminal.EntityNameColor(currentUser))
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLValidation}}
	httpClient = http.Client{Transport: transport, Timeout: time.Duration(DefaultHttpTimeout) * time.Second}
	requestHeader = map[string][]string{"Content-Type": {"application/json"}, "Authorization": {accessToken}}
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/routes?per_page=100&hosts=%s", apiEndpoint, hostname))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	routesListResponse := RoutesListResponse{}
	err = json.Unmarshal(body, &routesListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse routes response: %s", err)))
	}
	if len(routesListResponse.Resources) == 0 {
		fmt.Printf("no routes found for hostname %s\n", hostname)
	} else {
		table := terminal.NewTable(colNames)
		var orgName, spaceName string
		for _, route := range routesListResponse.Resources {
			var colValues [5]string
			colValues[0] = hostname
			colValues[1] = getAPIResource(route.Relationships.Domain.Data.GUID, "domains").(Domain).Name
			space := getAPIResource(route.Relationships.Space.Data.GUID, "spaces").(Space)
			colValues[2] = getAPIResource(space.Relationships.Organization.Data.GUID, "organizations").(Org).Name
			colValues[3] = space.Name
			table.Add(colValues[:]...)
			orgName = colValues[2]
			spaceName = space.Name
			var destList string
			for _, dest := range route.Destinations {
				appName := getAPIResource(dest.App.GUID, "apps").(App).Name
				destList = fmt.Sprintf("%s%s ", destList, appName)
			}
			colValues[4] = destList
		}
		_ = table.PrintTo(os.Stdout)
		if setTarget {
			if _, err = cliConnection.CliCommandWithoutTerminalOutput("target", "-o", orgName, "-s", spaceName); err != nil {
				fmt.Printf("failed to set target to org %s and space %s: %s", orgName, spaceName, err)
			}
		}
	}
}

func getAPIResource(guid string, apiResource string) interface{} {
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/%s/%s", apiEndpoint, apiResource, guid))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	switch apiResource {
	case "organizations":
		org := Org{}
		if err = json.Unmarshal(body, &org); err == nil {
			return org
		}
	case "spaces":
		space := Space{}
		if err = json.Unmarshal(body, &space); err == nil {
			return space
		}
	case "domains":
		domain := Domain{}
		if err = json.Unmarshal(body, &domain); err == nil {
			return domain
		}
	case "apps":
		app := App{}
		if err = json.Unmarshal(body, &app); err == nil {
			return app
		}
	}
	fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get response: %s", err)))
	return nil
}
