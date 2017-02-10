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

// Monitor represents an Apprenda monitoring file
type Monitor struct {
	PidFilePath     string         `json:"pidFilePath"`
	Cgroup          string         `json:"cgroup"`
	LaunchLogPath   string         `json:"launchLogPath"`
	WorkloadLogPath string         `json:"workloadLogPath"`
	ResourceConfig  ResourceConfig `json:"resourceConfig"`
}

// ResourceConfig represents resource conriguration
type ResourceConfig struct {
	StatsPollingInterval    int64          `json:"statsPollingInterval"`
	StatsPublishingInterval int64          `json:"statsPublishingInterval"`
	ResourcePolicy          ResourcePolicy `json:"resourcePolicy"`
}

// ResourcePolicy represents an Apprenda Resource Policy
type ResourcePolicy struct {
	CPULimit         int64  `json:"cpuLimit"`
	MemoryLimit      int64  `json:"memoryLimit"`
	MemoryLimitBytes int64  `json:"memoryLimitBytes"`
	Name             string `json:"name"`
	VersionID        string `json:"versionId"`
}
