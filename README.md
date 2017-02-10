# Docker Registry-based Deployer for the Apprenda Cloud Platform

A Docker deployer for the Apprenda Cloud Platform that pulls ready-made images from a Docker Registry (e.g. hub.docker.com).

## How To Use

Configure as a Bootstrap Policy for Linux Application Components in the Apprenda SOC (see below for details). Trigger it by tagging Linux Services Component Types with a Custom Property (`DockerDeploy=Registry`).

To deploy a Docker workload, a placeholder Apprenda application archive is currently necessary. The only requirements are a `DeploymentManifest.xml` with a `linuxServices` component declaration, including corresponding HTTP port mappings, and a directory structure with (optional) content, in the form of `linuxServices/component_name/slug.txt`, where slug.txt is an empty text file (only necessary in the absence of other content).

The `DockerImageName` is the only **required** Custom Property and **must** be populated with the Docker image/repo name to be pulled from the Registry **before** promoting the app. This can be included in the `DeploymentManifest.xml` file or added after the application has been created in *Definition* stage.

#### Simple Example: Nginx Web Server

##### Apprenda Archive

```txt
- DeploymentManifest.xml
- linuxServices
  - nginx
    - slug.txt (placeholder text file)
```

##### Deployment Manifest

```xml
<?xml version="1.0"?>
<appManifest xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://schemas.apprenda.com/DeploymentManifest http://apprenda.com/schemas/platform/6.0/DeploymentManifest.xsd" xmlns="http://schemas.apprenda.com/DeploymentManifest">
  <presentation strategy="CommingledAppRoot" scalingType="Manual"/>
  <applicationServices stickySessions="False" />
  <linuxServices>
    <service name="nginx" throttlingPolicy="Small">
      <customProperties>
        <customProperty name="DockerDeploy">
          <values>
            <propertyValue value="Registry" />
          </values>
        </customProperty>
        <customProperty name="DockerImageName">
          <values>
            <propertyValue value="nginx" />
          </values>
        </customProperty>
      </customProperties>
      <ports>
        <dynamicPort httpMapped="true" portName="HTTP_80" />
      </ports>
    </service>
  </linuxServices>
</appManifest>
```

### Volume Mounting (Bind Mounts)

The Volume Mounting feature supports the mounting of host directories into the Docker container's filesystem (equivalent to using the `-v` command-line flag with the `docker` client). Because it's impossible to predict the node and file path destination for a workload, the host pasth cannot be specified directly by the developer. Instead, the Deployer uses a path-relative naming convention. There are three options to bind mount volumes: **Local**, **Shared** and **Host**, configured via Custom Properties to declare one or more container file paths to use for mounting.

The **Local** option bind mounts host paths under the `docker-binds` sub-directory of the instance's file space (usually something like `/apprenda/persistent-instance-state/instances/<INSTANCE_ID>/docker-binds/`), by creating the necessary sub-directories under that to match the specified internal container path.

For example, if the Custom Property is defined like this: `DockerBindLocal=/usr/share/nginx/html`, that container path will be bind mounted to `/apprenda/persistent-instance-state/instances/<INSTANCE_ID>/docker-binds/usr/share/nginx/html`. The Custom Property can contain multiple path values to be mounted in the same way.

It's important to note that the `docker-binds` directory, together with all other instance files, is deleted when that workload instance is undeployed.

The **Shared** option works in a similar way. The difference is that the volumes are mounted under a directory that is external to the instance's file space so the contents can survive specific instance deployments/undeployments.

The target directory is configurable with the `DockerBindSharedRootDir` Custom Property and defaults to `/apprenda/docker-binds/`. This folder is intended to be a mounted external filesystem (e.g. NFS, CIFS, etc.) shared by all target Linux nodes. Therefore, all instances of the same Docker-based application will have access to the same files.

A directory hierarchy is created under this parent using the Tenant, Application and Verison aliases and the specified container path. For example, if the Custom Property is defined like this: `DockerBindShared=/usr/share/nginx/html`, that path will be bind mounted to `/apprenda/docker-binds/<TENANT_ALIAS>/<APP_ALIAS>/<VERSION_ALIAS>/usr/share/nginx/html`. As with the **Local** option, the Custom Property can be populated with multiple values to mount multiple paths.

The **Host** option is different. It expects a regular Docker bind mount specification of the form `/absolute/host/path:/absolute/container/path` in each `DockerBindHost` Custom Property and, if authorized, will pass those straight through to the Docker Engine.

However, for this to work, an Apprenda operator must enable a "white list" of approved host paths to be used with this option. This is done by creating a Custom Property called `DockerBindHostApprovedDirs` and assigning it a list of approved absolute host paths separated by colons as the default value (e.g. `/var/run/docker:/var/lib/docker`).

**SECURITY WARNING: THIS OPTION IS POTENTIALLY VERY DANGEROUS SINCE IT CAN EXPOSE SENSITIVE HOST DIRECTORIES TO GUEST APPLICATION CONTAINERS. THIS OPTION SHOULD ONLY BY ENABLED BY ADVANCED OPERATORS WITH FULL UNDERSTANDING OF THE CONSEQUENCES.**

#### Initializing Volumes With Archive Content

For each path specified to be bind mounted, the deployer will also look for a matching sub-directory inside the Deployment Archive's component folder. If found, the hierarchy and contents will be copied to the corresponding destination as explained above (**Local** or **Shared**, whichever applies) before bind mount occurs at container startup. This is a convenient way to insert files or whole directories into generic containers, that would otherwise require building new Docker images.

#### Volumes Example: Nginx Web Server with custom content

##### Apprenda Archive (Volumes)

```txt
- DeploymentManifest.xml
- linuxServices
  - nginx
    - user
      - share
        - nginx
          - html
            - index.html   <- custom index file
```

##### Deployment Manifest (Volumes)

```xml
<?xml version="1.0"?>
<appManifest xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://schemas.apprenda.com/DeploymentManifest http://apprenda.com/schemas/platform/6.0/DeploymentManifest.xsd" xmlns="http://schemas.apprenda.com/DeploymentManifest">
  <presentation strategy="CommingledAppRoot" scalingType="Manual"/>
  <applicationServices stickySessions="False" />
  <linuxServices>
    <service name="nginx" throttlingPolicy="Small">
      <customProperties>
        <customProperty name="DockerDeploy">
          <values>
            <propertyValue value="Registry" />
          </values>
        </customProperty>
        <customProperty name="DockerImageName">
          <values>
            <propertyValue value="nginx" />
          </values>
        </customProperty>
        <customProperty name="DockerBindLocal">
          <values>
            <propertyValue value="/usr/share/nginx/html" />
          </values>
        </customProperty>
      </customProperties>
      <ports>
        <dynamicPort httpMapped="true" portName="HTTP_80" />
      </ports>
    </service>
  </linuxServices>
</appManifest>
```

### Using Health Checks

Turning on Health Checking, by setting the `DockerReadinessCheck` Custom Property to `True`, ensures that HTTP traffic is not routed to a new deployment or instance of a workload until it is has been properly initialized (considered healty).

After a conainer is started it will not be immediately available for external requests, instead, the deployer will continuously send HTTP GET requests to a URL path, configurable using the `DockerReadinessCheckPath` Custom Property, until a valid HTTP Response Code is received (anything less than 300 is conisdered halthy). For containers exposing HTTPS endpoints, the URL scheme can also be configured, using the `DockerReadinessCheckScheme` Custom Property.

As soon as the first health check passes the deployer will consider the workload heathly and will complete the process to make it available and routable for external requests. If health checks continue to fail, the deployer will stop checking after a timeout period, configurable using the `DockerReadinessCheckTimeoutSecs` Custom Property, fail the deployment and return an error.

### Using Overlay Networking

Turn on *Overlay Networking* to place one or more Application Components in the same virtual network. Containers that share the same overlay network can find each other by **Network Alias** (see below) and can connect to each other directly. This is useful for containers that need horizontal clustering or for micro-services-based workloads that depend on non-HTTP, inter-component connectivity.

Set the `DockerNetworkScope` Custom Property of the component to one of its allowed values (`App`, `Tenant`, `Global`) to turn Overlay Networking on.

The **App** option will create an Overlay Network named after the Tenant, Application and Version that own the Component. Any other component(s) in the same Application can participate in the network by specifying the same `DockerNetworkScope` value. This is useful for Component interactions that depend on non-HTTP connectivity. For instance, connecting a middleware component to a back-end database or a message queue, without going (out and in again) through the main platform *Load Balancer*.

The **Tenant** option requires that the `DockerNetwork` Custom Property also be set. It will create an Overlay Network using the *name* specified in that second Property, in addition to the Tenant Alias. A tenant can create multiple **Tenant**-scoped Overlay Networks, and Components from different Applications also owned by that Tenant can be attached to any one of them. However, applications from **Other Tenants** will *not* be able to attach themselves to these networks.

The **Global** option also requires that the `DockerNetwork` Custom Property be set and creates an globally accessible Overlay Network, using the specified *name*. Components from differnt Tenants attached to the same Glocal network can communicate with each other. This may be useful in circumstances where applications from different Tenants may need access to non-HTTP (or alternative HTTP) enpoints of shared Components. However, if an HTTP connection is acceptable, using the front *Load Balancer* URL for an application is recommended instead.

**SECURITY WARNING: THIS OPTION EXPOSES COMPONENTS TO POTENTIALLY ROUGE APPLICATIONS FROM UNKNOWN TENANTS. BE SURE THE COMPONENT IS PROTECTED AT THE NETWORK LEVEL AND CAN WHISTAND NETWORK-BASED ATTACKS. THIS OPTION SHOULD ONLY BY USED BY ADVANCED USERS WITH FULL UNDERSTANDING OF THE CONSEQUENCES.**

#### Component Names as Service Names (Network Alias)

Every Component that gets connected to an Overlay Network can be looked up in DNS by  its *Network Alias*, which effectively becomes its **Service Name** inside that Overlay Network. For **App**-scoped networks the Service Name is simply the Component Alias (a.k.a. the Bundle Name). For **Tenant**- and **Global**-scoped networks the Service Name is constructed by concatenating the Application Alias, the Version Alias and the Component Alias, separated by dashes (-). For example, an application with an alias of `app1`, version alias `v1` and a component named `frontend` will get a Service Name of `app1-v1-frontend`. Any workload attached to the same Overlay Network will be able to look up this component by using its Service Name.

If a component has been horizontally scaled, looking up its Service Name will return a list of Overlay Network IPs corresponding to all the instances that exist for that component. Notably, subsequent lookups for the same Service Name will result in a different ordering of the target IPs.

**IMPORTANT NOTE: The client component must know the IP port(s) of the target Service. There is currently no way to look up the exposed ports programatically or environmentally.**

## Configuring The Apprenda Cloud Platform

Create the Following new Custom Properties **scoped to Applications > Linux Executables**:

### Developer Accesible Custom Properties (Visible to Developers)

Property Name | Allowed Values | Default Value | Description
------------- | -------------- | ------------- | -----------
`DockerDeploy` | `No`, `Dockerfile`, `Registry` | `No` | Used to trigger the Bootstrapper, etc.
`DockerImageName` | *custom* | - | The nme of the image to pull from the registry
`DockerImageTag` | *custom* | - | The specific image tag to use when pulling
`DockerCmd` | *custom* | - | Override the command and arguments to invoke inside de container
`DockerEntrypoint` | *custom* | - | Override the entrypoint set inside the container
`DockerBindHost` | *custom*, *allow multiple* | - | Local host directory absolute path to mount
`DockerBindLocal` | *custom*, *allow multiple* | - | Instance-space sub-directory path to mount
`DockerBindShared` | *custom*, *allow multiple*  | - | Global-space sub-directory path to mount
`DockerNetwork` | *custom* | - | The network name to use for the container
`DockerNetworkScope` | `App`, `Tenant`, `Global` | - | Use overlay networking with this scope
`DockerReadinessCheck` | `Yes`, `No` | `No` | Whether health checks should be performed before routing traffic to instance
`DockerReadinessCheckPath` | *custom* | `/` | A URL path to check for HTTP response codes < 300
`DockerReadinessCheckScheme` | `http`, `https` | `http` | The scheme that should be used for checks
`DockerReadinessCheckTimeoutSecs` | *custom* | `300` | Abort deployment after this timeout in seconds

### Administrative Custom Properties (Not Visible to Developers)

Property Name | Allowed Values | Default Value | Description
------------- | -------------- | ------------- | -----------
`DockerForcePull` | `Yes`, `No` | `No` | Should a pull be forced for every deployment
`DockerRemoveImage` | `Yes`, `No` | `No` | Should the cached image be removed when no containers are left using it
`DockerBindSharedRootDir` | *custom*  | `/apprenda/docker-binds` | The Shared root path for binds
`DockerBindDirPermissions` | *custom* | `0777` | Force specific permissions on bind directory creation
`DockerBindHostApprovedDirs` | *custom* | - | Colon-separated white list of approved absolute paths for host bind mounting

## Hacking On The Code

This codebase uses Go Vendoring for predictable builds. You need to have a working installation of [Glide](https://glide.sh) to build this project.

1. `go get https://bitbucket.org/apprenda/docker-img-deployer`
2. `cd $GOPATH/src/bitbucket.org/docker-img-deployer`
3. `glide install -v`

And hack away...

### How To Cut A Release

1. Update the value of `const version =` in `main.go`
2. Run `./build.sh` to stamp the VERSION file and create a Linux executable in `apprenda/platform-events/bin`

## Creating A Bootstrap Policy Archive For Apprenda

This repo does not cotnain all necessary files to create a complete and valid Bootstrap Archive. You can create a complete BSP Archive by following these steps:

1. Download the "Preview" Docker deployer as a base from [this documentation page](http://docs.apprenda.com/workloads/docker)
2. Expand the Zip file
3. Replace the `platform-events` directory with the one under the `apprenda` directory (after running the `./build.sh` script, see "Hacking On The Code" above)
4. Compress back into a Zip file
5. In the SOC, Create a new Bootstrap Policy for Linux Application Components using the new Zip file, that runs when `DockerDeploy=Registry`
6. Optionally create a Deployment Policy to target a subset of Linux nodes

## Who do I talk to?

- [Isaac (Ike) Arias](mailto:ike@apprenda.com)
