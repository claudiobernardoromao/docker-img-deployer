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

package types

import "strings"

// Instance represents an Apprenda workload instance
type Instance struct {
	ComponentType string `json:"componentType"`

	Platform struct {
		RedisServers       []string `json:"redisServers"`
		ZkConnString       string   `json:"zkConnString"`
		ClaimEncryptionKey string   `json:"claimEncryptionKey"`
		ClaimEncryptionIv  string   `json:"claimEncryptionIv"`
		CloudURL           string   `json:"cloudUrl"`
		PlatformVersion    string   `json:"platformVersion"`
	} `json:"platform"`

	Host struct {
		HostName           string `json:"hostName"`
		Root               string `json:"root"`
		Core               string `json:"core"`
		ProvidedPackageDir string `json:"providedPackageDir"`
		RepositoryDir      string `json:"repositoryDir"`
		Fqdn               string `json:"fqdn"`
	} `json:"host"`

	Process struct {
		WorkloadUserAccount          string     `json:"workloadUserAccount"`
		AutoCreateUserAccount        bool       `json:"autoCreateUserAccount"`
		MaxOpenFiles                 int64      `json:"maxOpenFiles"`
		EnvironmentVariables         [][]string `json:"environmentVariables"`
		EnvironmentVariableBlacklist []string   `json:"environmentVariableBlacklist"`

		LogWatch struct {
			MaxLogLineCount  int64 `json:"maxLogLineCount"`
			MaxLogLineLength int64 `json:"maxLogLineLength"`
			WatchInterval    int64 `json:"watchInterval"`
		} `json:"logWatch"`

		Ports struct {
			MaxDynamic int64 `json:"maxDynamic"`
			MinDynamic int64 `json:"minDynamic"`
			Allocated  []struct {
				Name     string `json:"name"`
				Port     int64  `json:"port"`
				PortType struct {
					EnumClass string `json:"enumClass"`
					Value     string `json:"value"`
				} `json:"portType"`
			} `json:"allocated"`
		} `json:"ports"`

		RuntimeOpts struct {
			Allowed string   `json:"allowed"`
			Flags   []string `json:"flags"`
		} `json:"runtimeOpts"`
	} `json:"process"`

	Workload struct {
		ApplicationAlias   string `json:"applicationAlias"`
		ApplicationID      string `json:"applicationId"`
		ArtifactID         string `json:"artifactId"`
		BundleName         string `json:"bundleName"`
		ContainerProfileID string `json:"containerProfileId"`
		Flags              string `json:"flags"`
		InstanceID         string `json:"instanceId"`
		ProviderID         string `json:"providerId"`
		Source             string `json:"source"`
		TransactionID      string `json:"transactionId"`
		VersionAlias       string `json:"versionAlias"`
		VersionID          string `json:"versionId"`
		VersionTenantID    string `json:"versionTenantId"`

		CustomProps []struct {
			Name   string   `json:"name"`
			Values []string `json:"values"`
		} `json:"customProps"`

		Logging struct {
			ApprendaLoggerLevel      string `json:"apprendaLoggerLevel"`
			Log4jPattern             string `json:"log4jPattern"`
			LogAggregationMethod     string `json:"logAggregationMethod"`
			LogFormatRefreshInterval int64  `json:"logFormatRefreshInterval"`
			LogLevelsFiltered        string `json:"logLevelsFiltered"`
			LoggerNamesFiltered      string `json:"loggerNamesFiltered"`
			MaxBatchSize             int64  `json:"maxBatchSize"`
			MaxBatchTimeMs           int64  `json:"maxBatchTimeMs"`
		} `json:"logging"`

		Stage struct {
			EnumClass string `json:"enumClass"`
			Value     string `json:"value"`
		} `json:"stage"`

		TokenReplacement struct {
			ReplaceDefaultFilePatterns bool `json:"replaceDefaultFilePatterns"`
			DefaultFilePatterns        []struct {
				Pattern string `json:"pattern"`
			} `json:"defaultFilePatterns"`
			Replacements []struct {
				IncludePlatformTokens bool `json:"includePlatformTokens"`
				ExcludePatterns       []struct {
					Pattern string `json:"pattern"`
				} `json:"excludePatterns"`
				IncludePatterns []struct {
					Pattern string `json:"pattern"`
				} `json:"includePatterns"`
			} `json:"replacements"`
		} `json:"tokenReplacement"`
	} `json:"workload"`

	Resource struct {
		StatsPollingInterval    int64 `json:"statsPollingInterval"`
		StatsPublishingInterval int64 `json:"statsPublishingInterval"`

		ResourcePolicy struct {
			CPULimit         int64  `json:"cpuLimit"`
			MemoryLimit      int64  `json:"memoryLimit"`
			MemoryLimitBytes int64  `json:"memoryLimitBytes"`
			Name             string `json:"name"`
			VersionID        string `json:"versionId"`
		} `json:"resourcePolicy"`
	} `json:"resource"`

	WebDeploy struct {
		CustomURLEnabled          bool `json:"customUrlEnabled"`
		DistributedSessionEnabled bool `json:"distributedSessionEnabled"`
		KeyStoreConfig            struct {
			Location   string `json:"location"`
			Password   string `json:"password"`
			Type       string `json:"type"`
			EntryAlias string `json:"entryAlias"`
		} `json:"keyStoreConfig"`
		URLAlias                    string `json:"urlAlias"`
		WarnIfCertExpiresWithinDays int64  `json:"warnIfCertExpiresWithinDays"`

		DeploymentStrategy struct {
			EnumClass string `json:"enumClass"`
			Value     string `json:"value"`
		} `json:"deploymentStrategy"`

		ServiceLevel struct {
			Flags int64 `json:"flags"`
		} `json:"serviceLevel"`
	} `json:"webDeploy"`

	Token struct {
		Tokens map[string]string `json:"tokens"`
	} `json:"token"`
}

// ContainerName constructs a name for the conainer that will hold the instance
func (i *Instance) ContainerName() string {
	nameParts := []string{"apprenda", i.Workload.ApplicationAlias, i.Workload.VersionAlias, i.Workload.InstanceID}
	return strings.Join(nameParts, "-")
}

// TenantAlias returns the Tenant alias, extracted from the Workload Source property
func (i *Instance) TenantAlias() string {
	return i.Workload.Source[1 : strings.Index(i.Workload.Source[1:], "/")+1]
}

// GetProp returns the value of a property if it exists or and empty string otherwise
func (i *Instance) GetProp(key string) []string {
	for _, prop := range i.Workload.CustomProps {
		if prop.Name == key {
			if len(prop.Values) > 0 {
				return prop.Values
			}
		}
	}
	return []string{}
}

// GetPropFirstValue returns the value of a property if it exists or and empty string otherwise
func (i *Instance) GetPropFirstValue(key string) string {
	for _, prop := range i.Workload.CustomProps {
		if prop.Name == key {
			if len(prop.Values) > 0 {
				return prop.Values[0]
			}
		}
	}
	return ""
}

// GetEnv reurns the Environment Variables and Platform Tokens
func (i *Instance) GetEnv() ([]string, error) {
	env := []string{}
	for key, val := range i.Token.Tokens {
		env = append(env, strings.Join([]string{key, val}, "="))
	}
	for _, envVar := range i.Process.EnvironmentVariables {
		env = append(env, strings.Join(envVar, "="))
	}
	return env, nil
}
