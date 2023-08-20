// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

// Behavior represents a dynamic file scan report.
type Behavior struct {
	Type             string      `json:"type,omitempty"`
	SHA256           string      `json:"sha256,omitempty"`
	Timestamp        int64       `json:"timestamp,omitempty"`
	Environment      interface{} `json:"env,omitempty"`
	APITrace         interface{} `json:"api_trace,omitempty"`
	SystemEvents     interface{} `json:"sys_events,omitempty"`
	ProcessTree      interface{} `json:"proc_tree,omitempty"`
	ScreenshotsCount int         `json:"screenshots_count,omitempty"`
	ScanConfig       ScanConfig  `json:"scan_cfg,omitempty"`
	SandboxLog       interface{} `json:"sandbox_log,omitempty"`
	AgentLog         interface{} `json:"agent_log,omitempty"`
	Status           int         `json:"status,omitempty"`
}

// ScanConfig represents the config used to detonate a file.
type ScanConfig struct {
	// Destination path where the sample will be located in the VM.
	DestPath string `json:"dest_path"`
	// Arguments used to run the sample.
	Arguments string `json:"args"`
	// Timeout in seconds for how long to keep the VM running.
	Timeout int `json:"timeout"`
	// Country to route traffic through.
	Country string `json:"country"`
	// Operating System used to run the sample.
	OS string `json:"os"`
}
