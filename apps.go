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

var appStats map[string]ProcessStatsResponse

const colState = "State"
const colMemory = "Memory"
const colDisk = "Disk"
const colType = "Type"
const colInstances = "# Inst"

var DefaultColumns = []string{colState, colMemory, colDisk, colType, colInstances}

func listApps(cliConnection plugin.CliConnection, args []string) {
	if len(args) != 1 {
		fmt.Printf("Incorrect Usage: there should be no arguments to this command`\n\nNAME:\n   %s\n\nUSAGE:\n   %s\n", ListAppsHelpText, ListAppsUsage)
		os.Exit(1)
	}
	cols := getColumns()
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
	processesListResponse := ProcessesListResponse{}
	err = json.Unmarshal(body, &processesListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}

	// optionally get the stats (per app instance stats)
	appStats = getAppStats(appsListResponse)

	table := terminal.NewTable([]string{"Name", "State", "Memory", "Disk", "Instas", "Type", "Created", "Updated", "Buildpacks", "Healthcheck", "Invoc Tmout", "Tmout", "Guid", "Ix", "Host", "Cpu", "Mem Used"})
	for _, app := range appsListResponse.Resources {
		var invocTmoutStr = "-"
		var tmoutStr = "-"
		if invocTmout := processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.InvocationTimeout; invocTmout != nil {
			invocTmoutStr = fmt.Sprintf("%v", processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.InvocationTimeout.(float64))
		}
		if tmout := processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.Timeout; tmout != nil {
			tmoutStr = fmt.Sprintf("%v", processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Data.Timeout.(float64))
		}
		table.Add(app.Name,
			strings.ToLower(app.State),
			fmt.Sprintf("%d", processForAppGuid(&processesListResponse, app.GUID).MemoryInMb),
			fmt.Sprintf("%d", processForAppGuid(&processesListResponse, app.GUID).DiskInMb),
			fmt.Sprintf("%d", processForAppGuid(&processesListResponse, app.GUID).Instances),
			processForAppGuid(&processesListResponse, app.GUID).Type,
			app.CreatedAt.Format(time.RFC3339),
			app.UpdatedAt.Format(time.RFC3339),
			strings.Join(app.Lifecycle.Data.Buildpacks, ","),
			fmt.Sprintf("%s", processForAppGuid(&processesListResponse, app.GUID).HealthCheck.Type),
			invocTmoutStr,
			tmoutStr,
			app.GUID,
			getColumnForApp(app.GUID, "ix"),
			getColumnForApp(app.GUID, "host"),
			getColumnForApp(app.GUID, "cpu"),
			getColumnForApp(app.GUID, "mem"),
		)
	}
	_ = table.PrintTo(os.Stdout)
}

func getColumns() []string {
	if os.Getenv("CF_COLS") == "" {
		return DefaultColumns
	}
}

func processForAppGuid(processes *ProcessesListResponse, appguid string) Process {
	var process Process
	for _, process = range processes.Resources {
		if process.Relationships.App.Data.GUID == appguid {
			return process
		}
	}
	return process
}

func getColumnForApp(appGuid string, colName string) string {
	var column string
	for ix, process := range appStats[appGuid].Resources {
		switch colName {
		case "ix":
			column = fmt.Sprintf("%s%d\n", column, ix)
		case "host":
			column = fmt.Sprintf("%s%s\n", column, process.Host)
		case "cpu":
			column = fmt.Sprintf("%s%.3f\n", column, process.Usage.CPU)
		case "mem":
			column = fmt.Sprintf("%s%v\n", column, process.Usage.Mem/1024/1024)
		}
	}
	return strings.TrimRight(column, "\n")
}

func getAppStats(appsListResponse AppsListResponse) map[string]ProcessStatsResponse {
	appStats = make(map[string]ProcessStatsResponse)
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
		appStats[app.GUID] = processesStatsResponse
	}
	return appStats
}
