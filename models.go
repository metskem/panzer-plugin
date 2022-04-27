package main

import "time"

// PanzerPlugin is the struct implementing the interface defined by the core CLI. It can be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type PanzerPlugin struct{}

type AppsListResponse struct {
	Pagination struct {
		TotalResults int `json:"total_results"`
		TotalPages   int `json:"total_pages"`
		First        struct {
			Href string `json:"href"`
		} `json:"first"`
		Last struct {
			Href string `json:"href"`
		} `json:"last"`
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
		Previous interface{} `json:"previous"`
	} `json:"pagination"`
	Resources []AppsListResource `json:"resources"`
}

type AppsListResource struct {
	GUID      string    `json:"guid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	State     string    `json:"state"`
	Lifecycle struct {
		Type string `json:"type"`
		Data struct {
			Buildpacks []string `json:"buildpacks"`
			Stack      string   `json:"stack"`
		} `json:"data"`
	} `json:"lifecycle"`
	//Relationships struct {
	//	Space struct {
	//		Data struct {
	//			GUID string `json:"guid"`
	//		} `json:"data"`
	//	} `json:"space"`
	//} `json:"relationships"`
	//Metadata struct {
	//	Labels struct {
	//	} `json:"labels"`
	//	Annotations struct {
	//	} `json:"annotations"`
	//} `json:"metadata"`
	//Links struct {
	//	Self struct {
	//		Href string `json:"href"`
	//	} `json:"self"`
	//	EnvironmentVariables struct {
	//		Href string `json:"href"`
	//	} `json:"environment_variables"`
	//	Space struct {
	//		Href string `json:"href"`
	//	} `json:"space"`
	//	Processes struct {
	//		Href string `json:"href"`
	//	} `json:"processes"`
	//	Packages struct {
	//		Href string `json:"href"`
	//	} `json:"packages"`
	//	CurrentDroplet struct {
	//		Href string `json:"href"`
	//	} `json:"current_droplet"`
	//	Droplets struct {
	//		Href string `json:"href"`
	//	} `json:"droplets"`
	//	Tasks struct {
	//		Href string `json:"href"`
	//	} `json:"tasks"`
	//	Start struct {
	//		Href   string `json:"href"`
	//		Method string `json:"method"`
	//	} `json:"start"`
	//	Stop struct {
	//		Href   string `json:"href"`
	//		Method string `json:"method"`
	//	} `json:"stop"`
	//	Revisions struct {
	//		Href string `json:"href"`
	//	} `json:"revisions"`
	//	DeployedRevisions struct {
	//		Href string `json:"href"`
	//	} `json:"deployed_revisions"`
	//	Features struct {
	//		Href string `json:"href"`
	//	} `json:"features"`
	//} `json:"links"`
}

type ProcessesListResponse struct {
	Pagination struct {
		TotalResults int `json:"total_results"`
		TotalPages   int `json:"total_pages"`
		First        struct {
			Href string `json:"href"`
		} `json:"first"`
		Last struct {
			Href string `json:"href"`
		} `json:"last"`
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
		Previous interface{} `json:"previous"`
	} `json:"pagination"`
	Resources []Process `json:"resources"`
}

type Process struct {
	GUID        string    `json:"guid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Type        string    `json:"type"`
	Command     string    `json:"command"`
	Instances   int       `json:"instances"`
	MemoryInMb  int       `json:"memory_in_mb"`
	DiskInMb    int       `json:"disk_in_mb"`
	HealthCheck struct {
		Type string `json:"type"`
		Data struct {
			Timeout           interface{} `json:"timeout"`
			InvocationTimeout interface{} `json:"invocation_timeout"`
		} `json:"data"`
	} `json:"health_check"`
	Relationships struct {
		App struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"app"`
		//	Revision struct {
		//		Data struct {
		//			GUID string `json:"guid"`
		//		} `json:"data"`
		//	} `json:"revision"`
	} `json:"relationships"`
	//Metadata struct {
	//	Labels struct {
	//	} `json:"labels"`
	//	Annotations struct {
	//	} `json:"annotations"`
	//} `json:"metadata"`
	Links struct {
		//	Self struct {
		//		Href string `json:"href"`
		//	} `json:"self"`
		//	Scale struct {
		//		Href   string `json:"href"`
		//		Method string `json:"method"`
		//	} `json:"scale"`
		//	App struct {
		//		Href string `json:"href"`
		//	} `json:"app"`
		//	Space struct {
		//		Href string `json:"href"`
		//	} `json:"space"`
		Stats struct {
			Href string `json:"href"`
		} `json:"stats"`
	} `json:"links"`
}

type ProcessStatsResponse struct {
	Resources []ProcessStats `json:"resources"`
}

type ProcessStats struct {
	Type             string      `json:"type"`
	Index            int         `json:"index"`
	State            string      `json:"state"`
	Host             string      `json:"host"`
	Uptime           int         `json:"uptime"`
	MemQuota         int         `json:"mem_quota"`
	DiskQuota        int         `json:"disk_quota"`
	FdsQuota         int         `json:"fds_quota"`
	IsolationSegment interface{} `json:"isolation_segment"`
	Details          interface{} `json:"details"`
	InstancePorts    []struct {
		External             int `json:"external"`
		Internal             int `json:"internal"`
		ExternalTLSProxyPort int `json:"external_tls_proxy_port"`
		InternalTLSProxyPort int `json:"internal_tls_proxy_port"`
	} `json:"instance_ports"`
	Usage struct {
		Time time.Time `json:"time"`
		CPU  float64   `json:"cpu"`
		Mem  int       `json:"mem"`
		Disk int       `json:"disk"`
	} `json:"usage"`
}
