package main

// ArtifactoryConfig represents the global configuration to apply in nexus
// this configuration comes from the k8s secret
type ArtifactoryConfig struct {
	Groups  []ArtifactoryGroup  `json:"groups,omitempty"`
	Hosteds []ArtifactoryHosted `json:"hosteds,omitempty"`
	Proxies []ArtifactoryProxy  `json:"proxies,omitempty"`
}

// ArtifactoryUser represents a user in the server
type ArtifactoryUser struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewPassword string `json:"newpassword"`
}

// ArtifactoryGroup represents a group repository
type ArtifactoryGroup struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// ArtifactoryHosted represents a hosted repository
type ArtifactoryHosted struct {
	Name string `json:"name"`
	// VersionPolicy the options are: RELEASE SNAPSHOT MIXED
	VersionPolicy string `json:"versionPolicy"`
	// LayoutPolicy the options are: STRICT PERMISSIVE
	LayoutPolicy string `json:"layoutPolicy"`
}

// ArtifactoryProxy represents a proxy repository
type ArtifactoryProxy struct {
	Name string `json:"name"`
	// VersionPolicy the options are: RELEASE SNAPSHOT MIXED
	VersionPolicy string `json:"versionPolicy"`
	// LayoutPolicy the options are: STRICT PERMISSIVE
	LayoutPolicy string `json:"layoutPolicy"`
	// RemoteURL is remote url to proxied
	RemoteURL string `json:"remoteUrl"`
	// RequiredAuth set to true if the proxy required authentication
	RequiredAuth bool `json:"requiredAuth"`
	// Authentication is required if RequiredAuth is set to true
	Authentication *ArtifactoryAuth `json:"authentication"`
}

// ArtifactoryAuth is the auth for artifactory proxy repository
type ArtifactoryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func hostedDataConfig(h ArtifactoryHosted) DataConfig {
	return DataConfig{
		Name:   h.Name,
		Online: true,
		Recipe: "maven2-hosted",
		Attributes: Attributes{
			Maven: &Maven{
				VersionPolicy: h.VersionPolicy,
				LayoutPolicy:  h.LayoutPolicy,
			},
			Storage: Storage{
				BlobStoreName:               "default",
				StrictContentTypeValidation: true,
				WritePolicy:                 "ALLOW",
			},
		},
	}
}

func proxyDataConfig(p ArtifactoryProxy) DataConfig {
	auth := &Authentication{}
	if p.RequiredAuth {
		auth.Type = "username"
		auth.UserName = p.Authentication.Username
		auth.Password = p.Authentication.Password
		auth.NtlmDomain = ""
		auth.NtlmHost = ""
	} else {
		// nil is ignored in the json
		auth = nil
	}
	return DataConfig{
		Name:        p.Name,
		Online:      true,
		AuthEnabled: p.RequiredAuth,
		Recipe:      "maven2-proxy",
		Attributes: Attributes{
			Maven: &Maven{
				VersionPolicy: p.VersionPolicy,
				LayoutPolicy:  p.LayoutPolicy,
			},
			Proxy: &Proxy{
				RemoteURL:      p.RemoteURL,
				ContentMaxAge:  -1,
				MetadataMaxAge: 1440,
			},
			HTTPClient: &HTTPClient{
				Blocked:        false,
				AutoBlock:      true,
				Authentication: auth,
			},
			Storage: Storage{
				BlobStoreName:               "default",
				StrictContentTypeValidation: true,
			},
			NegativeCache: &NegativeCache{
				Enabled:    true,
				TimeToLive: 1440,
			},
		},
	}
}

func groupDataConfig(g ArtifactoryGroup) DataConfig {
	return DataConfig{
		Name:   g.Name,
		Online: true,
		Recipe: "maven2-group",
		Attributes: Attributes{
			Storage: Storage{
				BlobStoreName:               "default",
				StrictContentTypeValidation: true,
			},
			Group: &Group{
				MemberNames: g.Members,
			},
		},
	}
}

func nexusConfig(data DataConfig) NexusConfig {
	return NexusConfig{
		Action: "coreui_Repository",
		Method: "create",
		Data: []DataConfig{
			data,
		},
		Type: "rpc",
		TID:  27,
	}
}
