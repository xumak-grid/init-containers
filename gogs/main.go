package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	cms "github.com/xumak-grid/init-containers/pkg/commons"
	validator "gopkg.in/go-playground/validator.v9"
)

const (
	gogsConfigFileEnv = "GOGS_CONFIG_FILE"
	gogsHostEnv       = "GOGS_HOST"
)

// serviceReady checks if a service is ready
func serviceReady(url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	client := cms.GetClient(5)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	return false
}

// initSetup post initial configuration in gogs
func initSetup(host string, data InitData) error {
	var validate *validator.Validate
	validate = validator.New()
	err := validate.Struct(data)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("domain", data.Domain)
	values.Set("http_port", data.HTTPPort)
	values.Set("app_url", data.APPUrl)
	values.Set("admin_name", data.AdminName)
	values.Set("admin_passwd", data.AdminPasswd)
	values.Set("admin_confirm_passwd", data.AdminConfirmPasswd)
	values.Set("admin_email", data.AdminEmail)
	values.Set("repo_root_path", data.RepoRoot)
	values.Set("log_root_path", data.LogRoot)
	values.Set("db_type", "SQLite3")
	values.Set("ssl_mode", "disable")
	values.Set("db_path", "data/gogs.db")
	values.Set("app_name", "Gogs")
	values.Set("run_user", "git")
	// empty to disable ability to clone via ssh
	values.Set("ssh_port", "")
	values.Set("enable_federated_avatar", "on")
	values.Set("enable_captcha", "on")

	url := host + "/install"
	resp, err := cms.GetClient(5).PostForm(url, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error creating init config code: %v message: %v", resp.StatusCode, resp.Status)
	}

	return nil
}

// gogsPost creates a new resource making a POST request to gogs
func gogsPost(user, pass, host string, obj interface{}) error {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, host, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, pass)
	req.Header.Add("Content-Type", "application/json")
	client := cms.GetClient(5)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error creating resource code: %d message: %v", resp.StatusCode, resp.Status)
	}
	return nil
}

// addCode supports to add code to a repository
func addCode(rep Repository, data InitData, host string) error {
	log.Println("configuring git with username and email")
	err := configGit(data.AdminName, data.AdminEmail)
	CheckIfError(err)

	switch rep.ContentSetupType {
	// add code from danta aem demo repository
	case "danta-aem-demo":
		demoURL := "git@github.com:xumak-grid/demo.git"
		addCodeFromRepo(rep, data, demoURL, host)
	// add code from project generated using danta AEM archetype
	case "danta-aem-archetype":
		addCodeFromDantaAEM(rep, data, host)
	// add code for EP commerce project
	case "ep-commerce":
		addCodeEP(rep, data, host)
	case "bloomreach-archetype":
		addCodeBR(rep, data, host)
	default:
		return fmt.Errorf("error adding code: the %v is not a valid content type for a repository", rep.ContentSetupType)
	}
	return nil
}

// addCodeFromRepo gets initial code from a git reposotory,
// and make a push with that code to a new gogs repository
func addCodeFromRepo(rep Repository, gogs InitData, src, host string) {
	log.Println("adding code from existing repository")
	dir, err := ioutil.TempDir("", rep.Owner)
	CheckIfError(err)

	log.Printf("cloning %v repository\n", src)
	err = clone(src, dir, ".")
	CheckIfError(err)

	log.Println("removing .git directory in source repository")
	err = os.RemoveAll(filepath.Join(dir, ".git"))
	CheckIfError(err)

	// create repo
	createRepo(rep, gogs, host, dir)
}

// addCodeFromDantaAEM generates a project using the Danta AEM archetype
// and make a push with that code to a gogs repository
func addCodeFromDantaAEM(rep Repository, gogs InitData, host string) {
	log.Println("generating danta aem project")
	dir, err := ioutil.TempDir("", rep.DantaAEM.AppName)
	CheckIfError(err)

	cmd := newCMD("mvn",
		"archetype:generate",
		"-DarchetypeGroupId=io.tikaltechnologies.danta",
		"-DarchetypeArtifactId=danta-aem-archetype",
		"-DarchetypeVersion="+rep.DantaAEM.ArchetypeVersion,
		"-DgroupId="+rep.DantaAEM.GroupID,
		"-DartifactId="+rep.DantaAEM.ArtifactID,
		"-Dproject-app-name="+rep.DantaAEM.AppName,
		"-Dpackage="+rep.DantaAEM.Package,
		"-Dcq-server="+rep.DantaAEM.AEMServer,
		"-Dnexus-public-url="+rep.DantaAEM.NexusURL,
		"-DinteractiveMode=false")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Fatalf("error: %v ouput: %v \n", err.Error(), out.String())
	}
	// create repository
	dir = filepath.Join(dir, rep.DantaAEM.AppName)
	createRepo(rep, gogs, host, dir)
}

// addCodeEP creates a project using the EP code sources
// and make a push with that code to a gogs repository
func addCodeEP(rep Repository, gogs InitData, host string) {
	log.Println("downloading EP commerce project")
	url := rep.EP.SourceCodeURL
	dir, err := ioutil.TempDir("", rep.Name)
	CheckIfError(err)

	// create file path
	file := filepath.Join(dir, "source")
	output, err := os.Create(file)
	CheckIfError(err)

	// download file
	response, err := http.Get(url)
	CheckIfError(err)
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	CheckIfError(err)
	log.Printf("%v bytes downloaded", n)

	log.Println("unzipping EP commerce source code")
	cmd := newCMD("unzip", file)
	cmd.Dir = dir
	err = cmd.Run()
	CheckIfError(err)

	log.Println("editing settings.xml file")
	path := filepath.Join(dir, "ep-commerce", "extensions", "maven", "settings.xml")
	err = cms.ReplaceStr(path, "PROJECT_REPOSITORY_GROUP_URL", rep.EP.MavenRepURL)
	if err != nil {
		log.Printf("error editing settings file: %s \n", err.Error())
	}

	log.Println("changing versions")
	dir = filepath.Join(dir, "ep-commerce")
	cmd = newCMD("./devops/scripts/set-ep-versions.sh", "-s", path, rep.EP.PlatformVersion, rep.EP.ExtensionVersion)
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		log.Printf("error setting versions: %s \n", err.Error())
	}

	log.Println("removing unused files")
	cmd = newCMD("rm", "commerce-manager/cm-modules/pom.xml.versionsBackup")
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		log.Printf("error removing unused files: %v \n", err.Error())
	}

	// create repository
	createRepo(rep, gogs, host, dir)
}

func addCodeBR(rep Repository, gogs InitData, host string) {
	log.Println("generating bloomreach project")
	dir, err := ioutil.TempDir("", rep.BR.ProjectName)
	CheckIfError(err)

	cmd := newCMD("mvn",
		"org.apache.maven.plugins:maven-archetype-plugin:2.4:generate",
		"-DarchetypeRepository=https://maven.onehippo.com/maven2",
		"-DarchetypeGroupId=org.onehippo.cms7",
		"-DarchetypeArtifactId=hippo-project-archetype",
		"-DarchetypeVersion=12.2.0",
		"-DgroupId="+rep.BR.GroupID,
		"-DartifactId="+rep.BR.ArtifactID,
		"-Dversion="+rep.BR.Version,
		"-Dpackage="+rep.BR.Package,
		"-DprojectName="+rep.BR.ProjectName,
		"-DinteractiveMode=false")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Fatalf("error: %v ouput: %v \n", err.Error(), out.String())
	}
	// create repository
	dir = filepath.Join(dir, rep.BR.ArtifactID)
	createRepo(rep, gogs, host, dir)
}

func createRepo(rep Repository, gogs InitData, host, dir string) {
	log.Println("initializing new repository")
	err := initRepo(dir)
	CheckIfError(err)

	log.Println("commit files")
	err = commitAll(dir, "Initial code")
	CheckIfError(err)

	log.Println("adding remote")
	repURL := fmt.Sprintf("%v/%v/%v", host, rep.Owner, rep.Name)
	err = addRemote(repURL, dir, "origin")
	CheckIfError(err)

	log.Printf("push files to %v repository\n", repURL)
	u, err := url.Parse(gogs.APPUrl)
	CheckIfError(err)

	u.Path = rep.Owner + "/" + rep.Name
	u.User = url.UserPassword(gogs.AdminName, gogs.AdminPasswd)
	url := u.String()
	err = push(dir, url)
	CheckIfError(err)
}

// clone repository in a given directory
func clone(url, dir, name string) error {
	cmd := newCMD("git", "clone", url, name)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// commitAll creates a commit adding all the changed files in a repository
func commitAll(dir, msg string) error {
	// adds all files to staging area
	cmd := newCMD("git", "add", ".")
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return err
	}

	// commit files
	cmd = newCMD("git", "commit", "-m", msg)
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// push makes a push in a repository
func push(dir, url string) error {
	cmd := newCMD("git", "push", url)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// initRepo initializes a new repository
func initRepo(dir string) error {
	cmd := newCMD("git", "init")
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// addRemote adds a new remote url in a repository
func addRemote(url, dir, remote string) error {
	cmd := newCMD("git", "remote", "add", remote, url)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// configGit configs user name and email, only for the current repository
func configGit(user, email string) error {
	// config name
	cmd := newCMD("git", "config", "--global", "user.name", user)
	err := cmd.Run()
	if err != nil {
		return err
	}

	// config email
	cmd = newCMD("git", "config", "--global", "user.email", email)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// newCMD returns a new cmd and redirects the stderr to stdout
func newCMD(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stdout
	return cmd
}

func main() {
	// environment variables
	configFile := cms.GetEnv(gogsConfigFileEnv, "examples/configFile.json")
	host := cms.GetEnv(gogsHostEnv, "http://localhost:8181")

	// read config file
	data := FileConfig{}
	err := cms.DecodeFromFile(configFile, &data)
	if err != nil {
		log.Fatalf("reading config file %v", err.Error())
	}

	// checks and waits for gogs
	log.Printf("check and wait for gogs on host: %v\n", host)
	timeout := time.After(1 * time.Minute)
	check := true
	for check {
		select {
		case <-time.After(3 * time.Second):
			if serviceReady(host) {
				check = false
				break
			}
		case <-timeout:
			log.Fatalf("timeout reached, host: %v", host)
		}
		log.Println("host not ready, 3s")
	}

	// read user
	user := data.InitData.AdminName
	pass := data.InitData.AdminPasswd

	// post gogs setup
	log.Println("initializing gogs")
	err = initSetup(host, data.InitData)
	if err != nil {
		log.Fatalf("error in gogs setup %s", err.Error())
	}

	// gogs healthcheck
	log.Printf("healthcheck for gogs on host %v\n", host)
	url := host + "/healthcheck"
	timeout = time.After(1 * time.Minute)
	check = true
	for check {
		select {
		case <-time.After(3 * time.Second):
			if serviceReady(url) {
				check = false
				break
			}
		case <-timeout:
			log.Fatalf("timeout reached, host: %v", host)
		}
		log.Println("host config not ready, 3s")
	}
	log.Println("initial configuration done!")

	// post organizations
	url = fmt.Sprintf("%v/api/v1/admin/users/%v/orgs", host, user)
	for _, org := range data.Organizations {
		log.Printf("creating %s organization", org.Username)
		err = gogsPost(user, pass, url, org)
		if err != nil {
			log.Fatalf("error creating %s organization %s\n", org.Username, err.Error())
		}
	}

	// post repositories
	for _, rep := range data.Repositories {
		log.Printf("creating %s repository", rep.Name)
		// if there is no value for the owner, it will use the admin user as the repository owner
		owner := rep.Owner
		if rep.Owner == "" {
			owner = user
		}
		rep.AutoInit = false
		rep.Readme = "Default"
		// if there is no value for add code to the repository it will be initialized
		if rep.ContentSetupType == "" || rep.ContentSetupType == "empty" {
			rep.AutoInit = true
		}
		url = fmt.Sprintf("%v/api/v1/admin/users/%v/repos", host, owner)
		err = gogsPost(user, pass, url, rep)
		if err != nil {
			log.Fatalf("error creating %s repository %s\n", rep.Name, err.Error())
		}
		// add code to the repository
		if rep.ContentSetupType != "" && rep.ContentSetupType != "empty" {
			err = addCode(rep, data.InitData, host)
			if err != nil {
				log.Fatalf("error adding code to %v repository %s\n", rep.Name, err.Error())
			}
		}
	}
	log.Println("the job is done!")
}

// CheckIfError is used to call a log.Fatalf if an error is not nil
func CheckIfError(err error) {
	if err != nil {
		log.Fatalf("error %s\n", err.Error())
	}
}
