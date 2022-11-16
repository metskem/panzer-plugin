package main

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var appData = make(map[string]App)
var processListResponse = ProcessesListResponse{}
var processStats = make(map[string]ProcessStatsResponse)
var appnamePrefix = ""
var processMutex sync.Mutex
var concurrencyCounter int32
var concurrencyCounterP *int32

type processList []Process

func (list processList) Len() int {
	return len(list)
}

func (list processList) Less(i, j int) bool {
	return strings.ToLower(appData[list[i].Relationships.App.Data.GUID].Name) < strings.ToLower(appData[list[j].Relationships.App.Data.GUID].Name)
}

func (list processList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

const colAppName = "Name"
const colState = "State"
const colMemory = "Memory"
const colDisk = "Disk"
const colType = "Type"
const colInstances = "#Inst"
const colIx = "Ix"
const colHost = "Host"
const colCpu = "Cpu%"
const colMemUsed = "MemUsed"
const colLogUsed = "LogUsed"
const colCreated = "Created"
const colUpdated = "Updated"
const colBuildpacks = "Buildpacks"
const colHealthCheck = "HealthCheck"
const colHealthCheckInvocationTimeout = "InvocTmout"
const colHealthCheckTimeout = "Tmout"
const colGuid = "Guid"
const colProcState = "ProcState"
const colUptime = "Uptime"
const colInstancePorts = "InstancePorts"

var DefaultColumns = []string{colAppName, colState, colMemory, colDisk, colUpdated, colHealthCheck, colInstances, colHost, colProcState, colUptime, colCpu, colMemUsed}
var ValidColumns = []string{colAppName, colState, colMemory, colDisk, colType, colInstances, colHost, colCpu, colMemUsed, colLogUsed, colCreated, colUpdated, colBuildpacks, colHealthCheck, colHealthCheckInvocationTimeout, colHealthCheckTimeout, colGuid, colProcState, colUptime, colInstancePorts}
var InstanceLevelColumns = []string{colHost, colCpu, colMemUsed, colLogUsed, colProcState, colUptime, colInstancePorts}

/** listApps - The main function to produce the response. */
func listApps(args []string) {
	if len(args) < 1 || len(args) > 2 {
		fmt.Printf("Usage: \"cf aa [appname-prefix]\". (Use envvar CF_COLS to specify the output columns)`\n\nNAME:\n   %s\n\nUSAGE:\n   %s\n", ListAppsHelpText, ListAppsUsage)
		os.Exit(1)
	}
	fmt.Printf("Getting apps for org %s / space %s as %s...\n\n", terminal.EntityNameColor(currentOrg.Name), terminal.EntityNameColor(currentSpace.Name), terminal.EntityNameColor(currentUser))
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLValidation}}
	httpClient = http.Client{Transport: transport, Timeout: time.Duration(DefaultHttpTimeout) * time.Second}
	requestHeader = map[string][]string{"Content-Type": {"application/json"}, "Authorization": {accessToken}}
	colNames := getRequestedColNames()
	if len(args) == 2 {
		appnamePrefix = args[1]
	}

	//
	// get the /v3/apps data first
	requestUrl, _ := url.Parse(fmt.Sprintf("%s/v3/apps?per_page=1000&space_guids=%s", apiEndpoint, currentSpace.Guid))
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to list apps: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	appsListResponse := AppsListResponse{}
	err = json.Unmarshal(body, &appsListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}

	if len(appsListResponse.Resources) == 0 {
		fmt.Println("No apps found")
		os.Exit(0)
	}
	// convert the json response to a map of App keyed by appguid
	for _, appsListResource := range appsListResponse.Resources {
		if strings.HasPrefix(appsListResource.Name, appnamePrefix) {
			appData[appsListResource.GUID] = appsListResource
		}
	}

	//
	// get the /v3/processes data next
	requestUrl, _ = url.Parse(fmt.Sprintf("%s/v3/processes?per_page=1000&space_guids=%s", apiEndpoint, currentSpace.Guid))
	httpRequest = http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err = httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to list apps: %s", err)))
		os.Exit(1)
	}
	body, _ = io.ReadAll(resp.Body)
	processListResponse = ProcessesListResponse{}
	err = json.Unmarshal(body, &processListResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}
	var pList processList
	pList = processListResponse.Resources
	sort.Sort(pList)
	//
	// optionally get the stats (per instance stats)
	if processStatsRequired(colNames) {
		processStats = getProcessStats(processListResponse)
	}

	table := terminal.NewTable(colNames)
	for _, process := range processListResponse.Resources {
		if !(process.Type == "task" && process.Instances == 0) {
			if strings.HasPrefix(appData[process.Relationships.App.Data.GUID].Name, appnamePrefix) {
				var colValues []string
				for _, colName := range colNames {
					colValues = append(colValues, getColValue(process, colName))
				}
				table.Add(colValues[:]...)
			}
		}
	}
	_ = table.PrintTo(os.Stdout)

	fmt.Printf("\n  %s\n", terminal.StoppedColor(getTotals(colNames)))
}

/** getTotals - Get all totals for the apps in the space, like total # of apps and total memory usage. */
func getTotals(colNames []string) string {
	var totalApps = 0
	var totalAppsStarted = 0
	var totalInstances = 0
	var totalMemory = 0
	var totalDisk = 0
	var totalMemoryUsed = 0
	var totalDiskUsed = 0
	var totalCpuUsed float64
	for _, process := range processListResponse.Resources {
		if strings.HasPrefix(appData[process.Relationships.App.Data.GUID].Name, appnamePrefix) {
			if !(process.Type == "task" && process.Instances == 0) {
				totalApps++
				if appData[process.Relationships.App.Data.GUID].State == "STARTED" {
					totalInstances = totalInstances + process.Instances
					totalAppsStarted++
					totalMemory = totalMemory + process.MemoryInMb*process.Instances
					totalDisk = totalDisk + process.DiskInMb*process.Instances
					for _, stat := range processStats[process.GUID].Resources {
						totalDiskUsed = totalDiskUsed + stat.Usage.Disk/1024/1024
						totalMemoryUsed = totalMemoryUsed + stat.Usage.Mem/1024/1024
						totalCpuUsed = totalCpuUsed + stat.Usage.CPU*100
					}
				}
			}
		}
	}
	if totalApps > 0 {
		memPerc := 0
		if totalMemory != 0 {
			memPerc = 100 * totalMemoryUsed / totalMemory
		}
		diskPerc := 0
		if totalDisk != 0 {
			diskPerc = 100 * totalDiskUsed / totalDisk
		}
		if processStatsRequired(colNames) {
			// we only have the "used" statistics if we requested at least one instance level column, if not we provide less statistics
			return fmt.Sprintf("%d apps (%d started), %d running instances, Memory(MB): requested:%d, used:%d (%2.0d%%), Cpu %4.0f%%, Disk(MB): requested:%d, used:%d (%2.0d%%)", totalApps, totalAppsStarted, totalInstances, totalMemory, totalMemoryUsed, memPerc, totalCpuUsed, totalDisk, totalDiskUsed, diskPerc)
		} else {
			return fmt.Sprintf("%d apps (%d started), %d running instances, Memory(MB): requested:%d, Cpu %4.0f%%, Disk(MB): requested:%d", totalApps, totalAppsStarted, totalInstances, totalMemory, totalCpuUsed, totalDisk)
		}
	} else {
		return ""
	}
}

/** processStatsRequired - If we want at least one instance level column, we need the app process stats (and we have to make a lot more http calls if the space has a lot of apps) */
func processStatsRequired(colNames []string) bool {
	var isProcessColumn bool = false
	for _, colName := range colNames {
		for _, processColumn := range InstanceLevelColumns {
			if colName == processColumn {
				isProcessColumn = true
			}
		}
	}
	return isProcessColumn
}

/** getRequestedColNames - Find out what the desired columns are specified in the envvar CF_COLS, or use the default set of columns */
func getRequestedColNames() []string {
	requestedColumns := os.Getenv("CF_COLS")
	if requestedColumns == "" {
		return DefaultColumns
	}
	if requestedColumns == "ALL" {
		return ValidColumns
	}
	customColNames := strings.Split(requestedColumns, ",")
	//
	// validate if invalid column names have been requested (if so, abort), and also check if instance level columns are present (if so, add Ix column)
	for _, customColName := range customColNames {
		isCustomColumnValid := false
		for _, validColumn := range ValidColumns {
			if customColName == validColumn {
				isCustomColumnValid = true
			}
		}
		if !isCustomColumnValid {
			fmt.Println(terminal.FailureColor(fmt.Sprintf("Invalid column in CF_COLS envvar : %s.", customColName)))
			fmt.Println(fmt.Sprintf("Valid column names are: %s", strings.Join(ValidColumns, ",")))
			os.Exit(1)
		}
	}
	for _, customColName := range customColNames {
		isInstanceLevelColumnRequested := false
		for _, instanceColumn := range InstanceLevelColumns {
			if customColName == instanceColumn {
				isInstanceLevelColumnRequested = true
			}
		}
		//
		// check if we have instance level columns, if so, we add an extra "ix" column to indicate which app index we have
		if isInstanceLevelColumnRequested {
			return strings.Split(fmt.Sprintf("%s,%s", colIx, requestedColumns), ",")
		}
	}
	return customColNames
}

/** - getColValue - Get the value of the given column.*/
func getColValue(process Process, colName string) string {
	var column string
	// per app instance columns
	if isInstanceColumn(colName) {
		for statsIndex, stats := range processStats[process.GUID].Resources {
			if appData[process.Relationships.App.Data.GUID].State != "STOPPED" {
				switch colName {
				case colIx:
					column = fmt.Sprintf("%s%d\n", column, statsIndex)
				case colHost:
					column = fmt.Sprintf("%s%s\n", column, stats.Host)
				case colCpu:
					column = fmt.Sprintf("%s%5.1f\n", column, stats.Usage.CPU*100)
				case colMemUsed:
					// calculate and color the memory used percentage
					usedMem := stats.Usage.Mem / 1024 / 1024
					memPercent := 100 * usedMem / process.MemoryInMb
					memPercentColored := terminal.SuccessColor(fmt.Sprintf("%2s", strconv.Itoa(memPercent)))
					if memPercent < 25 {
						memPercentColored = terminal.AdvisoryColor(fmt.Sprintf("%2s", strconv.Itoa(memPercent)))
					}
					if memPercent > 90 {
						memPercentColored = terminal.FailureColor(fmt.Sprintf("%2s", strconv.Itoa(memPercent)))
					}
					column = fmt.Sprintf("%s%4d (%s%%)\n", column, usedMem, memPercentColored)
				case colLogUsed:
					// calculate and color the log used percentage
					usedLog := stats.Usage.LogRate
					if process.LogRateBPS == -1 || process.LogRateBPS == 0 { // unlimited or undefined log rate
						column = fmt.Sprintf("%s%6d\n", column, usedLog)
					} else {
						logPercent := 100 * usedLog / process.LogRateBPS
						logPercentColored := terminal.SuccessColor(fmt.Sprintf("%2s", strconv.Itoa(logPercent)))
						if logPercent > 80 {
							logPercentColored = terminal.FailureColor(fmt.Sprintf("%2s", strconv.Itoa(logPercent)))
						}
						column = fmt.Sprintf("%s%4d (%s%%)\n", column, usedLog, logPercentColored)
					}
				case colProcState:
					if appData[process.Relationships.App.Data.GUID].State == "STARTED" && (stats.State == "CRASHED" || stats.State == "DOWN") {
						column = fmt.Sprintf("%s%s\n", column, terminal.FailureColor(strings.ToLower(stats.State)))
					} else {
						if appData[process.Relationships.App.Data.GUID].State == "STOPPED" && stats.State == "DOWN" {
							column = fmt.Sprintf("%s%s\n", column, terminal.EntityNameColor(strings.ToLower(stats.State)))
						} else {
							if appData[process.Relationships.App.Data.GUID].State == "STARTED" && stats.State == "STARTING" {
								column = fmt.Sprintf("%s%s\n", column, terminal.EntityNameColor(strings.ToLower(stats.State)))
							} else {
								column = fmt.Sprintf("%s%s\n", column, terminal.SuccessColor(strings.ToLower(stats.State)))
							}
						}
					}
				case colUptime:
					column = fmt.Sprintf("%s%12s\n", column, getFormattedElapsedTime(stats.Uptime))
				case colInstancePorts:
					var instancePorts []string
					for _, port := range stats.InstancePorts {
						instancePorts = append(instancePorts, fmt.Sprintf("%d", port.Internal))
					}
					column = fmt.Sprintf("%s%s\n", column, strings.Join(instancePorts, ","))
				}
			}
		}
	} else {
		// other columns (per app, not per app instance)
		switch colName {
		case colAppName:
			return appData[process.Relationships.App.Data.GUID].Name
		case colGuid:
			return appData[process.Relationships.App.Data.GUID].GUID
		case colState:
			if appData[process.Relationships.App.Data.GUID].State == "STOPPED" {
				return terminal.StoppedColor(strings.ToLower(appData[process.Relationships.App.Data.GUID].State))
			} else {
				return terminal.SuccessColor(strings.ToLower(appData[process.Relationships.App.Data.GUID].State))
			}
		case colMemory:
			return fmt.Sprintf("%6d", process.MemoryInMb)
		case colDisk:
			return fmt.Sprintf("%6d", process.DiskInMb)
		case colType:
			return fmt.Sprintf("%4s", process.Type)
		case colInstances:
			return fmt.Sprintf("%5d", process.Instances)
		case colCreated:
			return appData[process.Relationships.App.Data.GUID].CreatedAt.Format(time.RFC3339)
		case colUpdated:
			return appData[process.Relationships.App.Data.GUID].UpdatedAt.Format(time.RFC3339)
		case colBuildpacks:
			return strings.Join(appData[process.Relationships.App.Data.GUID].Lifecycle.Data.Buildpacks, ",")
		case colHealthCheck:
			return fmt.Sprintf("%11s", process.HealthCheck.Type)
		case colHealthCheckInvocationTimeout:
			var invocTmoutStr = "-"
			if invocTmout := process.HealthCheck.Data.InvocationTimeout; invocTmout != nil {
				invocTmoutStr = fmt.Sprintf("%v", process.HealthCheck.Data.InvocationTimeout.(float64))
			}
			return invocTmoutStr
		case colHealthCheckTimeout:
			var tmoutStr = "-"
			if tmout := process.HealthCheck.Data.Timeout; tmout != nil {
				tmoutStr = fmt.Sprintf("%v", process.HealthCheck.Data.Timeout.(float64))
			}
			return tmoutStr
		}
	}
	return strings.TrimRight(column, "\n")
}

/** isInstanceColumn - Return true if the given column name is an instance column (and requires us to call the /stats for all processes) */
func isInstanceColumn(name string) bool {
	if name == colIx {
		return true
	}
	for _, instanceColumn := range InstanceLevelColumns {
		if name == instanceColumn {
			return true
		}
	}
	return false
}

/** getProcessStats - Iterate over all processes and get the stats from them (concurrently) */
func getProcessStats(processListResponse ProcessesListResponse) map[string]ProcessStatsResponse {
	processStats = make(map[string]ProcessStatsResponse)
	concurrencyCounterP = &concurrencyCounter
	for _, process := range processListResponse.Resources {
		if strings.HasPrefix(appData[process.Relationships.App.Data.GUID].Name, appnamePrefix) {
			if !(process.Type == "task" && process.Instances == 0) {
				atomic.AddInt32(concurrencyCounterP, 1)
				// throttle a bit:
				time.Sleep(time.Millisecond * 25 * time.Duration(concurrencyCounter))
				go getProcessStat(process)
			}
		}
	}

	// wait for all routines to end:
	for {
		time.Sleep(time.Millisecond * 100)
		if concurrencyCounter == 0 {
			break
		}
	}
	return processStats
}

/** getProcessStat - Perform an http request to get the stats. This function is called concurrently. */
func getProcessStat(process Process) {
	defer atomic.AddInt32(concurrencyCounterP, -1)
	requestUrl, _ := url.Parse(process.Links.Stats.Href)
	httpRequest := http.Request{Method: http.MethodGet, URL: requestUrl, Header: requestHeader}
	resp, err := httpClient.Do(&httpRequest)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed response: %s", err)))
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	processesStatsResponse := ProcessStatsResponse{}
	err = json.Unmarshal(body, &processesStatsResponse)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to parse response: %s", err)))
	}
	processMutex.Lock()
	processStats[process.GUID] = processesStatsResponse
	processMutex.Unlock()
}

/** getFormattedElapsedTime - Transform the input (time in seconds) to a string with number of days, hours, mins and secs, like "1d01h54m10s" */
func getFormattedElapsedTime(timeInSecs int) string {
	days := timeInSecs / 86400
	secsLeft := timeInSecs % 86400
	hours := secsLeft / 3600
	secsLeft = secsLeft % 3600
	mins := secsLeft / 60
	secs := secsLeft % 60
	if days > 0 {
		return fmt.Sprintf("%dd%02dh%02dm%02ds", days, hours, mins, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, mins, secs)
	} else if mins > 0 {
		return fmt.Sprintf("%dm%02ds", mins, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}
