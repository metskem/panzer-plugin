package main

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var appData = make(map[string]AppsListResource)
var processData = make(map[string]Process)
var processStats = make(map[string]ProcessStatsResponse)

const colAppName = "Name"
const colState = "State"
const colMemory = "Memory"
const colDisk = "Disk"
const colType = "Type"
const colInstances = "# Inst"
const colIx = "Ix"
const colHost = "Host"
const colCpu = "Cpu"
const colMemUsed = "MemUsed"

var DefaultColumns = []string{colAppName, colState, colMemory, colDisk, colType, colInstances}

func listApps(cliConnection plugin.CliConnection, args []string) {
	if len(args) != 1 {
		fmt.Printf("Incorrect Usage: there should be no arguments to this command`\n\nNAME:\n   %s\n\nUSAGE:\n   %s\n", ListAppsHelpText, ListAppsUsage)
		os.Exit(1)
	}
	colNames := getColNames()

	//
	// get the /v3/apps data first
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/apps?order_by=name&space_guids=%s", apiEndpoint, currentSpace.Guid))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	fmt.Printf("Getting apps for org %s / space %s as %s\n\n", terminal.AdvisoryColor(currentOrg.Name), terminal.AdvisoryColor(currentSpace.Name), terminal.AdvisoryColor(currentUser))
	// TODO handle multi-page responses
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to list apps: %s", err)))
		os.Exit(1)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	appsListResponse := AppsListResponse{}
	err = json.Unmarshal(body, &appsListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}
	// convert the json response to a map of AppsListResource keyed by appguid
	for _, appsListResource := range appsListResponse.Resources {
		appData[appsListResource.GUID] = appsListResource
	}

	//
	// get the /v3/processes data next
	requestUrl, _ = url.Parse(fmt.Sprintf("%s/v3/processes?space_guids=%s", apiEndpoint, currentSpace.Guid))
	httpRequest = http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	fmt.Printf("Getting process info for org %s / space %s as %s\n\n", terminal.AdvisoryColor(currentOrg.Name), terminal.AdvisoryColor(currentSpace.Name), terminal.AdvisoryColor(currentUser))
	// TODO handle multi-page responses
	resp, err = httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to list apps: %s", err)))
		os.Exit(1)
	}
	body, _ = ioutil.ReadAll(resp.Body)
	processListResponse := ProcessesListResponse{}
	err = json.Unmarshal(body, &processListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}
	// convert the json response to a map of Process keyed by appguid
	for _, process := range processListResponse.Resources {
		processData[process.GUID] = process
	}

	//
	// optionally get the stats (per app instance stats)
	if processStatsRequired(colNames) {
		processStats = getAppProcessStats(appsListResponse)
	}

	table := terminal.NewTable(colNames)
	for _, app := range appData {
		var colValues []string
		for _, colName := range colNames {
			colValues = append(colValues, getColValue(app.GUID, colName))
		}
		table.Add(colValues[:]...)
	}
	_ = table.PrintTo(os.Stdout)
}

// processStatsRequired - If we want at least one instance level column, we need the app process stats
func processStatsRequired(colNames []string) bool {
	for _, colName := range colNames {
		if colName == colMemUsed || colName == colCpu || colName == colHost {
			return true
		}
	}
	return false
}

// getColNames - Find out what the desired columns are specified in envvar CF_COLS, or use the default set of columns
func getColNames() []string {
	if os.Getenv("CF_COLS") == "" {
		return DefaultColumns
	}
	return DefaultColumns // TODO, here handle the custom columns
}

func getColValue(appGuid string, colName string) string {
	var column string
	// per app instance columns
	if colName == colIx || colName == colHost || colName == colCpu || colName == colMemUsed {
		for ix, process := range processStats[appGuid].Resources {
			switch colName {
			case colIx:
				column = fmt.Sprintf("%s%d\n", column, ix)
			case colHost:
				column = fmt.Sprintf("%s%s\n", column, process.Host)
			case colCpu:
				column = fmt.Sprintf("%s%.3f\n", column, process.Usage.CPU)
			case colMemUsed:
				column = fmt.Sprintf("%s%v\n", column, process.Usage.Mem/1024/1024)
			}
		}
	} else {
		// other columns (not per app instance)
		switch colName {
		case colAppName:
			return appData[appGuid].Name
		case colState:
			return appData[appGuid].State
		case colMemory:
			return fmt.Sprintf("%d", processData[appGuid].MemoryInMb)
		case colDisk:
			return fmt.Sprintf("%d", processData[appGuid].DiskInMb)
		case colType:
			return processData[appGuid].Type
		case colInstances:
			return fmt.Sprintf("%d", processData[appGuid].Instances)
		}

		//var invocTmoutStr = "-"
		//var tmoutStr = "-"
		//if invocTmout := processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.InvocationTimeout; invocTmout != nil {
		//	invocTmoutStr = fmt.Sprintf("%v", processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.InvocationTimeout.(float64))
		//}
		//if tmout := processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.Timeout; tmout != nil {
		//	tmoutStr = fmt.Sprintf("%v", processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.Timeout.(float64))
		//}

	}
	return strings.TrimRight(column, "\n")
}

func getAppProcessStats(appsListResponse AppsListResponse) map[string]ProcessStatsResponse {
	processStats = make(map[string]ProcessStatsResponse)
	for _, app := range appsListResponse.Resources {
		requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/processes/%s/stats", apiEndpoint, app.GUID))
		httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
		resp, err := httpClient.Do(&httpRequest)
		if err != nil {
			fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
			os.Exit(1)
		}
		if err != nil {
			fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get app stats: %s", err)))
			os.Exit(1)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		processesStatsResponse := ProcessStatsResponse{}
		err = json.Unmarshal(body, &processesStatsResponse)
		if err != nil {
			fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
		}
		processStats[app.GUID] = processesStatsResponse
	}
	return processStats
}
