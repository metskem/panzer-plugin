package event

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
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
	"sort"
	"strings"
	"time"
)

const (
	ListEventsHelpText = "List recent audit events"
	timeFormat         = "2006-01-02T15:04:05"
)

var (
	ListEventsUsage = "ev - List recent audit events, use \"cf ev -help\" for full help message"
	colNames        = []string{"timestamp", "event-type", "target-name", "target-type", "actor"}
)

type EventList []model.Event

func (list EventList) Len() int {
	return len(list)
}

func (list EventList) Less(i, j int) bool {
	return list[i].CreatedAt.Before(list[j].CreatedAt)
}

func (list EventList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// GetEvents - Perform an http request to get the audit events
func GetEvents(cliConnection plugin.CliConnection) {
	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.DefaultParser.ShowVersionWithVersionFlag = false
	// Add flags
	flaggy.Int(&conf.FlagLimit, "l", "limit", "Limit the output to max XXX events")
	flaggy.String(&conf.FlagFilterEventTypes, "e", "event-type", "Filter the output (server side), (comma separated list of) event type to exactly match the filter (i.e. audit.app.update,app.crash)")
	flaggy.String(&conf.FlagFilterEventTargetName, "n", "target-name", "Filter the output (client side), target name to fuzzy match the filter")
	flaggy.String(&conf.FlagFilterEventTargetType, "t", "target-type", "Filter the output (client side), target type to fuzzy match the filter (i.e. app service_binding route)")
	flaggy.String(&conf.FlagFilterEventActor, "a", "actor", "Filter the output (client side), actor name to fuzzy match the filter")
	flaggy.String(&conf.FlagFilterEventOrgName, "o", "org", "Filter the output (server side), org name to exactly match the filter")
	flaggy.String(&conf.FlagFilterEventSpaceName, "s", "space", "Filter the output (server side), space name to exactly match the filter")
	flaggy.Parse()
	if conf.FlagLimit > 5000 {
		fmt.Printf("Output limited to 5000 rows\n")
		conf.FlagLimit = 5000
	}

	var httpClient http.Client
	var requestHeader http.Header
	fmt.Printf("Getting events as %s...\n\n", terminal.EntityNameColor(conf.CurrentUser))
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation}}
	httpClient = http.Client{Transport: transport, Timeout: time.Duration(conf.DefaultHttpTimeout) * time.Second}
	requestHeader = map[string][]string{"Content-Type": {"application/json"}, "Authorization": {conf.AccessToken}}

	// handle the serverside filters. You can specify one or both of orgname and spacename.
	serverSideFilter := ""
	var orgGuid, spaceGuid string
	if conf.FlagFilterEventOrgName != "" {
		if conf.FlagFilterEventSpaceName != "" {
			orgGuid = getOrgGuid(conf.FlagFilterEventOrgName)
			spaceGuid = getSpaceGuid(orgGuid, conf.FlagFilterEventSpaceName)
			serverSideFilter = fmt.Sprintf("&space_guids=%s", spaceGuid)
		} else {
			orgGuid = getOrgGuid(conf.FlagFilterEventOrgName)
			serverSideFilter = fmt.Sprintf("&organization_guids=%s", orgGuid)
		}
	} else {
		if conf.FlagFilterEventSpaceName != "" {
			if currentOrg, err := cliConnection.GetCurrentOrg(); err != nil {
				fmt.Printf("failed to get current org: %s\n", err)
				os.Exit(1)
			} else {
				spaceGuid = getSpaceGuid(currentOrg.Guid, conf.FlagFilterEventSpaceName)
				serverSideFilter = fmt.Sprintf("&space_guids=%s", spaceGuid)
			}
		}
	}
	if conf.FlagFilterEventTypes != "" {
		serverSideFilter = fmt.Sprintf("%s&types=%s", serverSideFilter, conf.FlagFilterEventTypes)
	}

	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/audit_events?per_page=%d&order_by=-created_at%s", conf.ApiEndpoint, conf.FlagLimit, serverSideFilter))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	eventsListResponse := model.EventsListResponse{}
	err = json.Unmarshal(body, &eventsListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse audit_events response: %s", err)))
	}
	if len(eventsListResponse.Resources) == 0 {
		fmt.Println("no audit_events found")
	} else {
		table := terminal.NewTable(colNames)
		var eventList EventList
		eventList = eventsListResponse.Resources
		sort.Sort(eventList)
		for _, event := range eventList {
			if strings.Contains(event.Target.Name, conf.FlagFilterEventTargetName) && strings.Contains(event.Target.Type, conf.FlagFilterEventTargetType) && strings.Contains(event.Actor.Name, conf.FlagFilterEventActor) {
				var colValues [5]string
				colValues[0] = event.CreatedAt.Local().Format(timeFormat)
				colValues[1] = event.Type
				if event.Target.Name == "" {
					colValues[2] = "<N/A>"
				} else {
					colValues[2] = event.Target.Name
				}
				colValues[3] = event.Target.Type
				colValues[4] = fmt.Sprintf("%s: %s", event.Actor.Type, event.Actor.Name)
				table.Add(colValues[:]...)
			}
		}
		_ = table.PrintTo(os.Stdout)
	}
}

// getOrgGuid - Get the organization guid, given the organization name. Will os.Exit if it fails to find it.
func getOrgGuid(orgName string) string {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation}}
	httpClient := http.Client{Transport: transport, Timeout: time.Duration(conf.DefaultHttpTimeout) * time.Second}
	requestHeader := map[string][]string{"Content-Type": {"application/json"}, "Authorization": {conf.AccessToken}}
	//
	// get the /v3/apps data first
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/organizations?names=%s", conf.ApiEndpoint, orgName))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get org by name (%s): %s", orgName, err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	orgsListResponse := model.OrgsListResponse{}
	err = json.Unmarshal(body, &orgsListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}

	if len(orgsListResponse.Resources) == 0 {
		fmt.Printf("Org %s not found\n", orgName)
		os.Exit(1)
	}
	return orgsListResponse.Resources[0].GUID
}

// getSpaceGuid - Get the space guid, given the organization guid and space name. Will os.Exit if it fails to find it.
func getSpaceGuid(orgGuid, spaceName string) string {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation}}
	httpClient := http.Client{Transport: transport, Timeout: time.Duration(conf.DefaultHttpTimeout) * time.Second}
	requestHeader := map[string][]string{"Content-Type": {"application/json"}, "Authorization": {conf.AccessToken}}
	//
	// get the /v3/apps data first
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/spaces?names=%s&organization_guids=%s", conf.ApiEndpoint, spaceName, orgGuid))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get space by name (%s) and orgGuid (%s): %s", spaceName, orgGuid, err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	spacesListResponse := model.SpacesListResponse{}
	err = json.Unmarshal(body, &spacesListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}

	if len(spacesListResponse.Resources) == 0 {
		fmt.Printf("Space %s not found in current org\n", spaceName)
		os.Exit(1)
	}
	return spacesListResponse.Resources[0].GUID
}
