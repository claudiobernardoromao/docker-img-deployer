// The MIT License (MIT)
//
// Copyright (c) 2016 Apprenda Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/context"

	t "github.com/claudiobernardoromao/docker-img-deployer/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const version = "1.0.0"
const instanceJSONFileName = "instance.json"
const dockerMarkerFileName = "apprenda-docker.properties"
const defaultBindsDirShared = "/apprenda/docker-binds"
const defaultBindsDirPermissions = 0777

// ACP Custom Property names
const propDockerImageName = "DockerImageName"
const propDockerImageTag = "DockerImageTag"
const propDockerCmd = "DockerCmd"
const propDockerEntrypoint = "DockerEntrypoint"
const propDockerBindHost = "DockerBindHost"
const propDockerBindLocal = "DockerBindLocal"
const propDockerBindShared = "DockerBindShared"
const propDockerBindSharedRootDir = "DockerBindSharedRootDir"
const propDockerBindDirPermissions = "DockerBindDirPermissions"
const propDockerBindHostApprovedDirs = "DockerBindHostApprovedDirs"
const propDockerNetwork = "DockerNetwork"
const propDockerNetworkScope = "DockerNetworkScope"
const propDockerHealthCheck = "DockerHealthCheck"
const propDockerHealthCheckPath = "DockerHealthCheckPath"
const propDockerHealthCheckScheme = "DockerHealthCheckScheme"
const propDockerHealthCheckTimeoutSecs = "DockerHealthCheckTimeoutSecs"
const propDockerReadinessCheck = "DockerReadinessCheck"
const propDockerReadinessCheckPath = "DockerReadinessCheckPath"
const propDockerReadinessCheckScheme = "DockerReadinessCheckScheme"
const propDockerReadinessCheckTimeoutSecs = "DockerReadinessCheckTimeoutSecs"
const propDockerForcePull = "DockerForcePull"
const propDockerImageRemove = "DockerImageRemove"
const propDockerRemoveImage = "DockerRemoveImage"

var ctx = context.Background()

// Apprenda Image-basd Docker deployer
//
// Designed to be invoked by the Linux Services Bootstrapper
// It assumes the presence of an instance.json file in the directory above PWD
func main() {

	logTo("init.out")
	log.Println("Initializing Docker Deployer version: ", version)

	flag.Parse()

	i, err := getInstance()
	if err != nil {
		log.Fatalln(err)
	}

	switch flag.Arg(0) {
	case "deploy":
		f := logTo("deployWorkload.out")
		defer f.Close()
		err = containerCreate(i)
		if err != nil {
			log.Fatalln(err)
		}
	case "start":
		f := logTo("startWorkload.out")
		defer f.Close()
		err = containerStart(i)
		if err != nil {
			log.Fatalln(err)
		}
	case "stop":
		f := logTo("stopWorkload.out")
		defer f.Close()
		err = containerStop(i)
		if err != nil {
			log.Fatalln(err)
		}
	case "undeploy":
		f := logTo("undeployWorkload.out")
		defer f.Close()
		err = containerRemove(i)
		if err != nil {
			log.Fatalln(err)
		}
	default:
		fmt.Println("Usage: instance [deploy|start|stop|undeploy]")
	}

}

func logTo(fileName string) *os.File {
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(f)
	return f
}

func getInstance() (*t.Instance, error) {

	f, err := os.Open(filepath.Join("..", instanceJSONFileName))
	defer f.Close()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var i t.Instance
	err = json.Unmarshal(b, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func containerCreate(i *t.Instance) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	repo := i.GetPropFirstValue(propDockerImageName)
	if repo == "" {
		return errors.New("ABORT: DockerImageName Custom Property for the component must be populated with a valid Registry name")
	}
	tag := i.GetPropFirstValue(propDockerImageTag)
	if tag == "" {
		tag = "latest"
	}
	ref := repo + ":" + tag
	
	// Claudio
	fmt.Println("redefinindo o nome da imagem")
	ref :="nginx"
	fmt.Println("Chamando o image pull antes de executar as alterações no objeto cli")
	err = imagePull(cli, ref)

	forcePull := strings.ToLower(i.GetPropFirstValue(propDockerForcePull))
	if forcePull == "yes" {
		log.Println("Forcing an image pull")
		err = imagePull(cli, ref)
		if err != nil {
			return err
		}
	}

	ports, portBindings, err := parseInstancePorts(i)
	if err != nil {
		return err
	}

	env, err := i.GetEnv()
	if err != nil {
		return err
	}

	config := &container.Config{
		Image:        ref,
		ExposedPorts: ports,
		Env:          env,
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		Tty:          false,
	}

	cmd := i.GetPropFirstValue(propDockerCmd)
	if cmd != "" {
		config.Cmd = strings.Fields(cmd)
	}

	entrypoint := i.GetPropFirstValue(propDockerEntrypoint)
	if entrypoint != "" {
		config.Entrypoint = strings.Fields(entrypoint)
	}

	binds, err := processBinds(i)
	if err != nil {
		return err
	}

	resources := container.Resources{}
	if i.Resource.ResourcePolicy.MemoryLimit > 0 {
		resources.Memory = i.Resource.ResourcePolicy.MemoryLimit * 1024 * 1024
	}
	if i.Resource.ResourcePolicy.CPULimit > 0 {
		resources.CPUShares = i.Resource.ResourcePolicy.CPULimit
	}

	networkName := i.GetPropFirstValue(propDockerNetwork)
	networkingConfig := &network.NetworkingConfig{}

	networkScope := strings.ToLower(i.GetPropFirstValue(propDockerNetworkScope))
	if networkScope != "" {
		networkName = getScopedNetworkName(i)

		networkExists, err := networkExists(cli, networkName)
		if err != nil {
			return err
		}

		if !networkExists {
			err = createNetwork(cli, networkName)
			if err != nil {
				return err
			}
		}

		netAlias := i.Workload.BundleName
		if networkScope != "app" {
			netAlias = strings.Join([]string{i.Workload.ApplicationAlias, i.Workload.VersionAlias, i.Workload.BundleName}, "-")
		}
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkName: {
					Aliases:   []string{netAlias},
					NetworkID: networkName,
				},
			},
		}
	}

	hostConfig := &container.HostConfig{
		Binds:        binds,
		Resources:    resources,
		PortBindings: portBindings,
		NetworkMode:  container.NetworkMode(networkName),
	}

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, i.ContainerName())
	if err != nil {
		if client.IsErrImageNotFound(err) {
			log.Println("Image not found locally, trying to pull it")
			fmt.Printf("chamando agora com o objeto preenchido")
			err = imagePull(cli, ref)
			if err != nil {
				return err
			}
			// and try again
			_, err = cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, i.ContainerName())
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	log.Println("Container created from", ref)
	return nil
}

func imagePull(cli *client.Client, ref string) error {
	log.Printf("Pulling %q from the registry...\n", ref)
	fmt.Println("entrando na funcao para baixar a imagem")
	fmt.Println(ref)
	resp, err := cli.ImagePull(context.Background(), ref, types.ImagePullOptions{})
	if err != nil {
		fmt.Println("problemas para baixar a imagem")
		fmt.Println(err)
		return err
	}
	defer resp.Close()
	if _, err = io.Copy(ioutil.Discard, resp); err != nil {
		return err
	}
	log.Println("Image pull complete")
	return nil
}

func getScopedNetworkName(i *t.Instance) (networkName string) {
	networkNameProp := strings.ToLower(i.GetPropFirstValue(propDockerNetwork))

	var nameParts []string
	networkScope := strings.ToLower(i.GetPropFirstValue(propDockerNetworkScope))
	switch networkScope {
	case "app":
		nameParts = []string{"app", i.TenantAlias(), i.Workload.ApplicationAlias, i.Workload.VersionAlias}
	case "tenant":
		nameParts = []string{"tenant", i.TenantAlias(), networkNameProp}
	case "global":
		nameParts = []string{"global-", networkNameProp}
	default:
		nameParts = []string{networkNameProp}
	}

	networkName = strings.Join(nameParts, "-")
	log.Printf("Network name is '%s'\n", networkName)
	return networkName
}

func networkExists(cli *client.Client, networkName string) (bool, error) {
	networkListOptions := types.NetworkListOptions{}
	networks, err := cli.NetworkList(context.Background(), networkListOptions)
	if err != nil {
		return false, err
	}

	for n := range networks {
		if networkName == networks[n].Name {
			log.Printf("Network '%s' already exists\n", networkName)
			return true, nil
		}
	}
	return false, nil
}

func createNetwork(cli *client.Client, networkName string) error {
	networkOptions := types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "overlay",
		Attachable:     true,
	}

	_, err := cli.NetworkCreate(context.Background(), networkName, networkOptions)
	if err != nil {
		return err
	}

	log.Printf("Successfully created network '%s'\n", networkName)
	return nil
}

func parseInstancePorts(i *t.Instance) (map[nat.Port]struct{}, map[nat.Port][]nat.PortBinding, error) {
	portSpecs := []string{}
	for _, portDef := range i.Process.Ports.Allocated {
		intPort := ""
		outPort := strconv.FormatInt(portDef.Port, 10)
		portRaw := portDef.Name
		if i := strings.LastIndex(portRaw, "_"); i != -1 {
			intPort = portRaw[i+1:]
		}
		portSpec := strings.Join([]string{outPort, intPort}, ":")
		portSpecs = append(portSpecs, portSpec)
	}
	return nat.ParsePortSpecs(portSpecs)
}

func processBinds(i *t.Instance) ([]string, error) {
	var archiveSrcDir string

	// Locate archive source in pre-zipped-repo (introduced in ACP 6.6.0) deploy structure
	if strings.HasPrefix(i.Platform.PlatformVersion, "6.5") {
		archiveSrcDir = filepath.Join(
			i.Host.RepositoryDir,
			i.TenantAlias(),
			i.Workload.ApplicationAlias,
			i.Workload.VersionAlias,
			"base/linuxServices",
			i.Workload.BundleName,
		)
	} else {
		tempDir := strings.Join([]string{i.Workload.InstanceID, "temp"}, "_")
		archiveSrcDir = filepath.Join(
			i.Host.Root,
			i.Workload.InstanceID,
			tempDir,
			"workload",
		)
	}

	dirPermInt, err := strconv.Atoi(i.GetPropFirstValue(propDockerBindDirPermissions))
	if err != nil {
		dirPermInt = defaultBindsDirPermissions
	}
	dirPerm := os.FileMode(dirPermInt)

	// Process local (instance) binds, if any
	lBinds := []string{}
	lPaths := i.GetProp(propDockerBindLocal)
	if len(lPaths) > 0 {
		lBindRoot := filepath.Join(i.Host.Root, i.Workload.InstanceID, "docker-binds")
		// Pre-create local bind directories with specified permissions
		err = preCreatePathDirs(lPaths, lBindRoot, dirPerm)
		if err != nil {
			return []string{}, err
		}
		// Process local binds, copying dirs from src archive if available
		lBinds, err = getBindsForPaths(lPaths, lBindRoot, archiveSrcDir)
		if err != nil {
			return []string{}, err
		}
	}

	// Process shared binds, if any
	sBinds := []string{}
	sPaths := i.GetProp(propDockerBindShared)
	if len(sPaths) > 0 {
		sBindRoot := i.GetPropFirstValue(propDockerBindSharedRootDir)
		if sBindRoot == "" {
			sBindRoot = defaultBindsDirShared
		}
		sBindRoot = filepath.Join(
			sBindRoot,
			i.TenantAlias(),
			i.Workload.ApplicationAlias,
			i.Workload.VersionAlias,
		)
		// Pre-create shared bind directories with specified permissions
		err = preCreatePathDirs(sPaths, sBindRoot, dirPerm)
		if err != nil {
			return []string{}, err
		}
		// Process shared binds, copying dirs from src archive if available
		sBinds, err = getBindsForPaths(sPaths, sBindRoot, archiveSrcDir)
		if err != nil {
			return []string{}, err
		}
	}

	// Process host binds, if any, validating against approved host dirs
	hPaths := i.GetProp(propDockerBindHost)
	hBinds := []string{}
	if len(hPaths) > 0 {
		hApprovedDirs := i.GetPropFirstValue(propDockerBindHostApprovedDirs)
		if hApprovedDirs != "" {
			hBinds, err = getBindsForHostPaths(hPaths, strings.Split(hApprovedDirs, ":"))
			if err != nil {
				return []string{}, err
			}
		} else {
			return []string{}, errors.New("ABORT: Host binding is not currently allowed")
		}
	}

	return append(append(lBinds, sBinds...), hBinds...), nil
}

func preCreatePathDirs(paths []string, rootPath string, perm os.FileMode) error {
	syscall.Umask(0)
	for _, path := range paths {
		localPath, _, err := getPaths(path, rootPath)
		if err != nil {
			return err
		}
		err = os.MkdirAll(localPath, perm)
		if err != nil {
			return err
		}
	}
	return nil
}

func getBindsForPaths(paths []string, rootPath, archiveSrcDir string) ([]string, error) {
	binds := []string{}
	for _, path := range paths {
		localPath, relPath, err := getPaths(path, rootPath)
		if err != nil {
			return []string{}, err
		}
		newBind := strings.Join([]string{localPath, path}, ":")
		binds = append(binds, newBind)

		srcDir := filepath.Join(archiveSrcDir, relPath)
		err = copyDirIfExists(srcDir, localPath)
		if err != nil {
			return []string{}, err
		}
	}
	return binds, nil
}

func getPaths(path, rootPath string) (localPath, relativePath string, err error) {
	if !strings.HasPrefix(path, "/") {
		err = errors.New("ABORT: All bind mounts must be absolute paths")
		return
	}
	relativePath = path[1:]
	if colIdx := strings.Index(relativePath, ":"); colIdx > -1 {
		relativePath = relativePath[0:colIdx]
	}
	localPath = filepath.Join(rootPath, relativePath)
	return
}

func copyDirIfExists(source, dest string) error {
	// check if the source exists and is a directory
	src, err := os.Stat(source)
	if err != nil {
		return nil
	}
	if !src.IsDir() {
		return errors.New("ABORT: Source is not a directory")
	}
	return copyDir(source, dest)
}

func copyFile(source, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err2 := os.Stat(source)
		if err2 != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}
	return
}

func copyDir(source, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := filepath.Join(source, obj.Name())
		destinationfilepointer := filepath.Join(dest, obj.Name())
		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func pathIsApproved(path string, approvedDirs []string) bool {
	for _, approved := range approvedDirs {
		if strings.HasPrefix(path, approved) {
			return true
		}
	}
	return false
}

func getBindsForHostPaths(paths, approvedDirs []string) ([]string, error) {
	binds := []string{}
	rejected := []string{}
	for _, path := range paths {
		if pathIsApproved(path, approvedDirs) {
			binds = append(binds, path)
		} else {
			rejected = append(rejected, path)
		}
	}
	if len(rejected) > 0 {
		return []string{}, fmt.Errorf("ABORT: The following host binds are not allowed: %s", strings.Join(rejected, ", "))
	}
	return binds, nil
}

func containerStart(i *t.Instance) error {
	err := os.Chdir(i.Token.Tokens["BASEPATH"])
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	err = cli.ContainerStart(ctx, i.ContainerName(), types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	log.Println("Container started")

	c, err := cli.ContainerInspect(ctx, i.ContainerName())
	if err != nil {
		return err
	}

	dockerVersion, err := cli.ServerVersion(ctx)
	if err != nil {
		return err
	}

	err = checkWorkloadReadiness(i)
	if err != nil {
		return err
	}

	err = createMonitorFiles(i, &c, dockerVersion.Version)
	if err != nil {
		return err
	}

	err = startLogForwarder(i, &c)
	if err != nil {
		return err
	}

	return nil
}

func checkWorkloadReadiness(i *t.Instance) error {
	checkReadiness := i.GetPropFirstValue(propDockerReadinessCheck)
	// Check for deprecated property name. To be removed in a future version.
	if checkReadiness == "" {
		checkReadiness = i.GetPropFirstValue(propDockerHealthCheck)
	}
	if checkReadiness == "Yes" {
		log.Println("Starting readiness checks")
		for _, portDef := range i.Process.Ports.Allocated {
			if portDef.PortType.Value == "Http" {
				outPort := strconv.FormatInt(portDef.Port, 10)
				transport := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
				httpClient := &http.Client{Transport: transport}
				scheme := i.GetPropFirstValue(propDockerReadinessCheckScheme)
				// Check for deprecated property name. To be removed in a future version.
				if scheme == "" {
					scheme = i.GetPropFirstValue(propDockerHealthCheckScheme)
				}
				if scheme != "http" && scheme != "https" {
					scheme = "http"
				}
				path := i.GetPropFirstValue(propDockerReadinessCheckPath)
				// Check for deprecated property name. To be removed in a future version.
				if path == "" {
					path = i.GetPropFirstValue(propDockerHealthCheckPath)
				}
				if !strings.HasPrefix(path, "/") {
					path = "/"
				}
				checkURL := url.URL{
					Scheme: scheme,
					Host:   "localhost:" + outPort,
					Path:   path,
				}
				timeout, err := strconv.Atoi(i.GetPropFirstValue(propDockerReadinessCheckTimeoutSecs))
				if err != nil {
					// Check for deprecated property name. To be removed in a future version.
					timeout, err = strconv.Atoi(i.GetPropFirstValue(propDockerHealthCheckTimeoutSecs))
					if err != nil {
						timeout = 300
					}
				}
				maxTime := time.Duration(timeout) * time.Second
				elapsed := time.Duration(0)
				t0 := time.Now()
				for i := 1; elapsed < maxTime; elapsed = time.Now().Sub(t0) {
					log.Println("Health check try #", i)
					resp, err := httpClient.Get(checkURL.String())
					if err != nil || resp.StatusCode >= 300 {
						if err != nil {
							log.Println(err)
						} else {
							log.Println("HTTP Rsponse Status Code: ", resp.StatusCode)
						}
						log.Println("Sleeping for 500 milliseconds...")
						time.Sleep(500 * time.Millisecond)
					} else {
						log.Println("Health check PASSED. HTTP Rsponse Status Code: ", resp.StatusCode)
						return nil
					}
				}
				return errors.New("ABORT: Health check timout reached")
			}
		}
	}
	return nil
}

func createMonitorFiles(i *t.Instance, c *types.ContainerJSON, dockerVersion string) error {
	pidFile := os.Getenv("APPRENDA_WORKLOAD_PIDFILE")
	if pidFile == "" {
		return errors.New("$APPRENDA_WORKLOAD_PIDFILE environment variable not defined")
	}

	cgroup := "/system.slice/docker-" + c.ID + ".scope"

	monitor := &t.Monitor{
		PidFilePath:     pidFile,
		Cgroup:          cgroup,
		LaunchLogPath:   filepath.Join(i.Token.Tokens["DEPLOYER_BASEDIR"], "startWorkload.out"),
		WorkloadLogPath: filepath.Join(i.Token.Tokens["BASEPATH"], "dockerStart.out"),
		ResourceConfig: t.ResourceConfig{
			StatsPollingInterval:    i.Resource.StatsPollingInterval,
			StatsPublishingInterval: i.Resource.StatsPublishingInterval,
			ResourcePolicy: t.ResourcePolicy{
				CPULimit:         i.Resource.ResourcePolicy.CPULimit,
				MemoryLimit:      i.Resource.ResourcePolicy.MemoryLimit,
				MemoryLimitBytes: i.Resource.ResourcePolicy.MemoryLimitBytes,
				Name:             i.Resource.ResourcePolicy.Name,
				VersionID:        i.Resource.ResourcePolicy.VersionID,
			},
		},
	}
	b, err := json.MarshalIndent(monitor, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.Create(pidFile)
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(strconv.Itoa(c.State.Pid))
	f.Close()
	log.Println("Created container PID file")

	f, err = os.Create(dockerMarkerFileName)
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(fmt.Sprintf("docker.version=%s\n", dockerVersion))
	f.WriteString(fmt.Sprintf("deployer.version=%s\n", version))
	f.Close()

	// Write a second marker file for ACP v6.6
	f, err = os.Create(filepath.Join(i.Token.Tokens["DEPLOYER_EVENTS_BASEDIR"], dockerMarkerFileName))
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(fmt.Sprintf("docker.version=%s\n", dockerVersion))
	f.WriteString(fmt.Sprintf("deployer.version=%s\n", version))
	f.Close()

	log.Println("Created Apprenda-Docker marker file")

	f, err = os.Create(filepath.Join(i.Token.Tokens["BASEPATH"], "monitor.json"))
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(string(b))
	log.Println("Created workload monitor.json file")
	return nil
}

func startLogForwarder(i *t.Instance, c *types.ContainerJSON) error {
	err := createLogForwarderConfig(i, c)
	if err != nil {
		return err
	}
	return startForwarderProcess(i)
}

func createLogForwarderConfig(i *t.Instance, c *types.ContainerJSON) error {
	forwarderConfig := &t.LogstashForwarderConfig{
		Network: t.Network{
			Servers: []string{"localhost:6782"},
			SslCa:   filepath.Join(i.Host.ProvidedPackageDir, "logstash-forwarder/etc/apprenda-logstash2.crt"),
			Timeout: 15,
		},
		Files: []t.File{
			{
				Paths: []string{c.LogPath},
				Fields: t.Fields{
					Type:       "v1 stdout/sderr",
					InstanceID: i.Workload.InstanceID,
					ProviderID: i.Workload.ProviderID,
					VersionID:  i.Workload.VersionID,
				},
			},
		},
	}
	b, err := json.MarshalIndent(forwarderConfig, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.Create(getLogConfigPath(i))
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(string(b))
	log.Println("Created logstash-forwarder-config.json file")
	return nil
}

func startForwarderProcess(i *t.Instance) error {
	var logstashStartScript = filepath.Join(i.Host.ProvidedPackageDir, "logstash-forwarder/bin/start-log-monitor.sh")
	err := exec.Command(logstashStartScript, getLogConfigPath(i)).Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Started logstash-forwarder")
	return nil
}

func getLogConfigPath(i *t.Instance) string {
	return filepath.Join(i.Token.Tokens["BASEPATH"], "logstash-forwarder-config.json")
}

func containerStop(i *t.Instance) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	timeout := 30 * time.Second
	err = cli.ContainerStop(ctx, i.ContainerName(), &timeout)
	if err != nil {
		return err
	}
	log.Println("Container stopped")

	err = stopLogForwarder(i)
	if err != nil {
		return err
	}
	log.Println("Stopped logstash-forwarder")

	return nil
}

func stopLogForwarder(i *t.Instance) error {
	f, err := os.Open(filepath.Join(i.Token.Tokens["BASEPATH"], "logstash_forwarder.pid"))
	defer f.Close()
	if err != nil {
		return err
	}
	var pid int
	_, err = fmt.Fscan(f, &pid)
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func containerRemove(i *t.Instance) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	options := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	err = cli.ContainerRemove(ctx, i.ContainerName(), options)
	if err != nil {
		return err
	}
	log.Println("Container removed")
	networkScope := strings.ToLower(i.GetPropFirstValue(propDockerNetworkScope))
	if networkScope != "" {
		networkName := getScopedNetworkName(i)
		err = cli.NetworkRemove(ctx, networkName)
		if err != nil {
			log.Println(err.Error())
		}
	}
	removeImage := i.GetPropFirstValue(propDockerRemoveImage)
	// Check for deprecated property name. To be removed in a future version.
	if removeImage == "" {
		removeImage = i.GetPropFirstValue(propDockerImageRemove)
	}
	if removeImage == "Yes" {
		err = imageRemove(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func imageRemove(i *t.Instance) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	options := types.ImageRemoveOptions{
		PruneChildren: true,
	}
	_, err = cli.ImageRemove(ctx, i.GetPropFirstValue(propDockerImageName), options)
	if err != nil {
		log.Println("Image not removed because other containers are still running")
	} else {
		log.Println("Image removed")
	}
	return nil
}
