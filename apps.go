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
	"time"
)

var appData = make(map[string]AppsListResource)
var processData = make(map[string]Process)
var processStats = make(map[string]ProcessStatsResponse)

const colAppName = "Name"
const colState = "State"
const colMemory = "Memory"
const colDisk = "Disk"
const colType = "Type"
const colInstances = "Instances"
const colIx = "Ix"
const colHost = "Host"
const colCpu = "Cpu"
const colMemUsed = "MemUsed"
const colCreated = "Created"
const colUpdated = "Updated"
const colBuildpacks = "Buildpacks"
const colHealthCheck = "HealthCheck"
const colHealthCheckInvocationTimeout = "InvocTmout"
const colHealthCheckTimeout = "Tmout"
const colGuid = "Guid"

var DefaultColumns = []string{colAppName, colState, colMemory, colDisk, colType, colInstances}
var ValidColumns = []string{colAppName, colState, colMemory, colDisk, colType, colInstances, colHost, colCpu, colMemUsed, colCreated, colUpdated, colBuildpacks, colHealthCheck, colHealthCheckInvocationTimeout, colHealthCheckTimeout, colGuid}
var InstanceLevelColumns = []string{colHost, colCpu, colMemUsed}

func listApps(cliConnection plugin.CliConnection, args []string) {
	if len(args) != 1 {
		fmt.Printf("Incorrect Usage: there should be no arguments to this command`\n\nNAME:\n   %s\n\nUSAGE:\n   %s\n", ListAppsHelpText, ListAppsUsage)
		os.Exit(1)
	}
	colNames := getRequestedColNames()

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

	//
	// here we start building the table output
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

// getRequestedColNames - Find out what the desired columns are specified in envvar CF_COLS, or use the default set of columns
func getRequestedColNames() []string {
	requestedColumns := os.Getenv("CF_COLS")
	if requestedColumns == "" {
		return DefaultColumns
	}
	customColNames := strings.Split(requestedColumns, ",")
	//
	// validate if invalid column names have been requested (if so, abort), and also check if instance level columns are present (if so, add Ix column)
	for _, customColName := range customColNames {
		isCustomColumnValid := false
		isInstanceLevelColumnRequested := false
		for _, validColumn := range ValidColumns {
			if customColName == validColumn {
				isCustomColumnValid = true
			}
		}
		if !isCustomColumnValid {
			fmt.Println(terminal.FailureColor(fmt.Sprintf("Invalid column in CF_COLS envvar : %s", customColName)))
			os.Exit(1)
		}
		for _, instanceColumn := range InstanceLevelColumns {
			if customColName == instanceColumn {
				isInstanceLevelColumnRequested = true
			}
		}
		if !isCustomColumnValid {
			fmt.Println(terminal.FailureColor(fmt.Sprintf("Invalid column in CF_COLS envvar : %s", customColName)))
			os.Exit(1)
		}
		if isInstanceLevelColumnRequested {
			// prepend "ix,"
			return strings.Split(fmt.Sprintf("%s,%s", colIx, requestedColumns), ",")
		}
	}

	//
	// check if we have instance level columns, if so, we add an extra "ix" column to indicate which app index we have

	return customColNames
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
				column = fmt.Sprintf("%s%7d\n", column, process.Usage.Mem/1024/1024)
			}
		}
	} else {
		// other columns (not per app instance)
		switch colName {
		case colAppName:
			return appData[appGuid].Name
		case colGuid:
			return appData[appGuid].GUID
		case colState:
			return appData[appGuid].State
		case colMemory:
			return fmt.Sprintf("%6d", processData[appGuid].MemoryInMb)
		case colDisk:
			return fmt.Sprintf("%d", processData[appGuid].DiskInMb)
		case colType:
			return processData[appGuid].Type
		case colInstances:
			return fmt.Sprintf("%d", processData[appGuid].Instances)
		case colCreated:
			return appData[appGuid].CreatedAt.Format(time.RFC3339)
		case colUpdated:
			return appData[appGuid].UpdatedAt.Format(time.RFC3339)
		case colBuildpacks:
			return strings.Join(appData[appGuid].Lifecycle.Data.Buildpacks, ",")
		case colHealthCheck:
			return processData[appGuid].HealthCheck.Type
		case colHealthCheckInvocationTimeout:
			var invocTmoutStr = "-"
			if invocTmout := processData[appGuid].HealthCheck.Data.InvocationTimeout; invocTmout != nil {
				invocTmoutStr = fmt.Sprintf("%v", processData[appGuid].HealthCheck.Data.InvocationTimeout.(float64))
			}
			return invocTmoutStr
		case colHealthCheckTimeout:
			var tmoutStr = "-"
			if tmout := processData[appGuid].HealthCheck.Data.Timeout; tmout != nil {
				tmoutStr = fmt.Sprintf("%v", processData[appGuid].HealthCheck.Data.Timeout.(float64))
			}
			return tmoutStr
		}
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
