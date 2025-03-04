package event

import (
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/integrii/flaggy"
	"github.com/metskem/panzer-plugin/conf"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	ListEventsHelpText = "List recent audit events"
	timeFormat         = "2006-01-02T15:04:05"
)

var (
	ListEventsUsage  = "ev - List recent audit events, use \"cf ev -help\" for full help message"
	colNames         = []string{"timestamp", "event-type", "target-name", "target-type", "actor", "data"}
	beforeTime       time.Time
	afterTime        time.Time
	TypeAppCreate    = "audit.app.create"
	TypeProcessCrash = "audit.app.process.crash"
	TypeProcessReady = "audit.app.process.ready"
)

type DataProcessCrashEvent struct {
	Instance        string `json:"instance"`
	Index           int    `json:"index"`
	CellId          string `json:"cell_id"`
	Reason          string `json:"reason"`
	ExitDescription string `json:"exit_description"`
	CrashCount      int    `json:"crash_count"`
	CrashTimestamp  int64  `json:"crash_timestamp"`
}

type DataProcessReadyEvent struct {
	Instance string `json:"instance"`
	Index    int    `json:"index"`
	CellId   string `json:"cell_id"`
	Ready    bool   `json:"ready"`
}

type DataAppCreateEvent struct {
	Request struct {
		Lifecycle struct {
			Data struct {
				Buildpacks []string `json:"buildpacks"`
			} `json:"data"`
		} `json:"lifecycle"`
	} `json:"request"`
}

type AuditEventList []*resource.AuditEvent

func (list AuditEventList) Len() int {
	return len(list)
}

func (list AuditEventList) Less(i, j int) bool {
	return list[i].CreatedAt.Before(list[j].CreatedAt)
}

func (list AuditEventList) Swap(i, j int) {
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
	flaggy.Bool(&conf.FlagHideHeaders, "q", "hide-headers", "Hide the headers of the output (handy for automated processing), default is false")
	flaggy.String(&conf.FlagTimeBefore, "tb", "time-before", "Filter the output (server side), time before the given time (timeformat: YYYY-MM-DDThh:mm:ssZ)")
	flaggy.String(&conf.FlagTimeAfter, "ta", "time-after", "Filter the output (server side), time after the given time (timeformat: YYYY-MM-DDThh:mm:ssZ)")
	flaggy.Bool(&conf.FlagIncludeEventData, "d", "include-data", "Include the event data in the output (requires a lot of space), default is false")
	flaggy.Parse()
	if conf.FlagLimit > 5000 {
		fmt.Printf("Output limited to 5000 rows\n")
		conf.FlagLimit = 5000
	}
	if conf.FlagLimit == 0 {
		conf.FlagLimit = 500
	}

	if !conf.FlagHideHeaders {
		fmt.Printf("Getting events as %s...\n\n", terminal.EntityNameColor(conf.CurrentUser))
	}

	var timePatternRegex = regexp.MustCompile("^(\\d{4})-(\\d{2})-(\\d{2})T(\\d{2}):(\\d{2}):(\\d{2})Z$")
	if conf.FlagTimeBefore != "" {
		if timePatternRegex.MatchString(conf.FlagTimeBefore) {
			var err error
			if beforeTime, err = time.Parse(timeFormat+"Z", conf.FlagTimeBefore); err != nil {
				fmt.Printf("failed to parse time %s: %s\n", conf.FlagTimeBefore, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("invalid time format: %s\n", conf.FlagTimeBefore)
			os.Exit(1)
		}
	}
	if conf.FlagTimeAfter != "" {
		if timePatternRegex.MatchString(conf.FlagTimeAfter) {
			var err error
			if afterTime, err = time.Parse(timeFormat+"Z", conf.FlagTimeAfter); err != nil {
				fmt.Printf("failed to parse time %s: %s\n", conf.FlagTimeAfter, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("invalid time format: %s\n", conf.FlagTimeAfter)
			os.Exit(1)
		}
	}
	// handle the serverside filters. You can specify one or both of orgname and spacename.
	var orgGuid, spaceGuid string
	if conf.FlagFilterEventOrgName != "" {
		if conf.FlagFilterEventSpaceName != "" {
			orgGuid = getOrgGuid(conf.FlagFilterEventOrgName)
			spaceGuid = getSpaceGuid(orgGuid, conf.FlagFilterEventSpaceName)
		} else {
			orgGuid = getOrgGuid(conf.FlagFilterEventOrgName)
		}
	} else {
		if conf.FlagFilterEventSpaceName != "" {
			if currentOrg, err := cliConnection.GetCurrentOrg(); err != nil {
				fmt.Printf("failed to get current org: %s\n", err)
				os.Exit(1)
			} else {
				spaceGuid = getSpaceGuid(currentOrg.Guid, conf.FlagFilterEventSpaceName)
			}
		}
	}

	var types, orgGuids, spaceGuids client.Filter
	if conf.FlagFilterEventTypes != "" {
		types = client.Filter{Values: strings.Split(conf.FlagFilterEventTypes, ",")}
	}
	if conf.FlagFilterEventOrgName != "" {
		orgGuids = client.Filter{Values: []string{orgGuid}}
	}
	if conf.FlagFilterEventSpaceName != "" {
		spaceGuids = client.Filter{Values: []string{spaceGuid}}
	}

	var createdAfter, createdBefore client.TimestampFilter
	if conf.FlagTimeAfter != "" {
		createdAfter = client.TimestampFilter{Timestamp: []time.Time{afterTime}, Operator: client.FilterModifierGreaterThan}
	}
	if conf.FlagTimeBefore != "" {
		createdBefore = client.TimestampFilter{Timestamp: []time.Time{beforeTime}, Operator: client.FilterModifierLessThan}
	}
	timeStampFilterList := client.TimestampFilterList{createdAfter, createdBefore}

	auditListOptions := client.AuditEventListOptions{
		ListOptions:       &client.ListOptions{PerPage: conf.FlagLimit, Page: 1, OrderBy: "-created_at", CreateAts: timeStampFilterList},
		Types:             types,
		OrganizationGUIDs: orgGuids,
		SpaceGUIDs:        spaceGuids}

	if events, _, err := conf.CfClient.AuditEvents.List(conf.CfCtx, &auditListOptions); err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get audit events: %s", err)))
		os.Exit(1)
	} else {
		if len(events) == 0 {
			fmt.Println("no audit_events found")
		} else {
			table := terminal.NewTable(colNames)
			if conf.FlagHideHeaders {
				table.NoHeaders()
			}
			var eventList AuditEventList
			eventList = events
			sort.Sort(eventList)
			for _, event := range eventList {
				if strings.Contains(event.Target.Name, conf.FlagFilterEventTargetName) && strings.Contains(event.Target.Type, conf.FlagFilterEventTargetType) && strings.Contains(event.Actor.Name, conf.FlagFilterEventActor) {
					var colValues [6]string
					colValues[0] = event.CreatedAt.Local().Format(timeFormat)
					colValues[1] = event.Type
					if event.Target.Name == "" {
						colValues[2] = "<N/A>"
					} else {
						colValues[2] = event.Target.Name
					}
					colValues[3] = event.Target.Type
					colValues[4] = fmt.Sprintf("%s: %s", event.Actor.Type, event.Actor.Name)
					colValues[5] = "-"
					if conf.FlagIncludeEventData {
						if event.Type == TypeProcessCrash {
							var processCrashData DataProcessCrashEvent
							if err = json.Unmarshal(*event.Data, &processCrashData); err != nil {
								fmt.Printf("failed to unmarshal process crash data: %s\n", err)
							} else {
								colValues[5] = fmt.Sprintf("index: %d, cell_id: %s, crash_count: %d, exit_description: %s", processCrashData.Index, processCrashData.CellId, processCrashData.CrashCount, processCrashData.ExitDescription)
							}
						}
						if event.Type == TypeProcessReady {
							var processReadyData DataProcessReadyEvent
							if err = json.Unmarshal(*event.Data, &processReadyData); err != nil {
								fmt.Printf("failed to unmarshal process ready data: %s\n", err)
							} else {
								colValues[5] = fmt.Sprintf("index: %d, cell_id: %s", processReadyData.Index, processReadyData.CellId)
							}
						}
						if event.Type == TypeAppCreate {
							var appCreateData DataAppCreateEvent
							if err = json.Unmarshal(*event.Data, &appCreateData); err != nil {
								fmt.Printf("failed to unmarshal app create data: %s\n", err)
							} else {
								colValues[5] = fmt.Sprintf("buildpacks: %s", strings.Join(appCreateData.Request.Lifecycle.Data.Buildpacks, ","))
							}
						}
					}
					table.Add(colValues[:]...)
				}
			}
			_ = table.PrintTo(os.Stdout)
		}
	}
}

// getOrgGuid - Get the organization guid, given the organization name. Will os.Exit if it fails to find it.
func getOrgGuid(orgName string) string {
	org, err := conf.CfClient.Organizations.Single(conf.CfCtx, &client.OrganizationListOptions{ListOptions: &client.ListOptions{}, Names: client.Filter{Values: []string{orgName}}})
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get org by name (%s): %s", orgName, err)))
		os.Exit(1)
	}
	return org.GUID
}

// getSpaceGuid - Get the space guid, given the organization guid and space name. Will os.Exit if it fails to find it.
func getSpaceGuid(orgGuid, spaceName string) string {
	spaceListOptions := client.SpaceListOptions{ListOptions: &client.ListOptions{}, OrganizationGUIDs: client.Filter{Values: []string{orgGuid}}, Names: client.Filter{Values: []string{spaceName}}}
	space, err := conf.CfClient.Spaces.Single(conf.CfCtx, &spaceListOptions)
	if err != nil {
		fmt.Println(terminal.FailureColor(fmt.Sprintf("failed to get space by name (%s): %s", spaceName, err)))
		os.Exit(1)
	}
	return space.GUID
}
