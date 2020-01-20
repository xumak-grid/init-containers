package main

// FileConfig represents the file object with the configuration to apply
type FileConfig struct {
	InitData      InitData       `json:"init_data"`
	Organizations []Organization `json:"organizations"`
	Repositories  []Repository   `json:"repositories"`
}

// InitData represents a configuration data for init installation in gogs
type InitData struct {
	Domain             string `json:"domain" validate:"required"`
	HTTPPort           string `json:"http_port" validate:"required,numeric"`
	APPUrl             string `json:"app_url" validate:"required,url"`
	AdminName          string `json:"admin_name" validate:"required,ne=admin"`
	AdminPasswd        string `json:"admin_passwd" validate:"required"`
	AdminConfirmPasswd string `json:"admin_confirm_passwd" validate:"required"`
	AdminEmail         string `json:"admin_email" validate:"required,email"`
	RepoRoot           string `json:"repo_root_path" validate:"required"`
	LogRoot            string `json:"log_root_path" validate:"required"`
}

// Organization represents a configuration for each organization in gogs
type Organization struct {
	Username    string `json:"username"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	WebSite     string `json:"website"`
	Location    string `json:"location"`
}

// Repository represents a configuration for each repository in gogs
type Repository struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Owner       string `json:"owner"`
	// pass true to create an initial commit with readme file
	AutoInit bool `json:"auto_init"`
	// desired readme template name to apply in the initial commit
	Readme string `json:"readme"`
	// ContentSetupType represents the type of content that will have the repository
	ContentSetupType string   `json:"content_setup_type"`
	DantaAEM         DantaAEM `json:"danta_aem_archetype"`
	EP               EP       `json:"ep_commerce"`
	BR               BR       `json:"bloomreach_archetype"`
}

// DantaAEM represents a configuration data to create a project with Danta AEM archetype
type DantaAEM struct {
	ArchetypeGroup    string `json:"archetype_group"`
	ArchetypeArtifact string `json:"archetype_artifact"`
	ArchetypeVersion  string `json:"archetype_version"`
	GroupID           string `json:"group_id"`
	ArtifactID        string `json:"artifact_id"`
	AppName           string `json:"app_name"`
	Package           string `json:"package"`
	AEMServer         string `json:"aem_server"`
	NexusURL          string `json:"nexus_url"`
	Interactive       string `json:"interactive"`
}

// EP represents a configuration data to create a project for EP commerce
type EP struct {
	SourceCodeURL    string `json:"source_code_url"`
	MavenRepURL      string `json:"maven_rep_url"`
	PlatformVersion  string `json:"platform_version"`
	ExtensionVersion string `json:"extension_version"`
}

// BR represents a configuration data to create a project with Bloomreach archetype
type BR struct {
	ArchetypeVersion string `json:"archetype_version"`
	GroupID          string `json:"group_id"`
	ArtifactID       string `json:"artifact_id"`
	Version          string `json:"version"`
	Package          string `json:"package"`
	ProjectName      string `json:"project_name"`
}
