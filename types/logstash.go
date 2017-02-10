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

// LogstashForwarderConfig holds configuration for the logstash-forwarder
type LogstashForwarderConfig struct {
	Network Network `json:"network"`
	Files   []File  `json:"files"`
}

// Network holds connection data for logstash
type Network struct {
	Servers []string `json:"servers"`
	SslCa   string   `json:"ssl ca"`
	Timeout int      `json:"timeout"`
}

// File holds information about files to be forwarded
type File struct {
	Fields Fields   `json:"fields"`
	Paths  []string `json:"paths"`
}

// Fields holds log field data
type Fields struct {
	InstanceID string `json:"instance_id"`
	ProviderID string `json:"provider_id"`
	Type       string `json:"type"`
	VersionID  string `json:"version_id"`
}
