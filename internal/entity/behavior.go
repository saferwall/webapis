// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package entity

// Behavior represents a dynamic file scan report.
type Behavior struct {
	Meta             *DocMetadata `json:"doc,omitempty"`
	Type             string       `json:"type,omitempty"`
	SHA256           string       `json:"sha256,omitempty"`
	Timestamp        int64        `json:"timestamp,omitempty"`
	Environment      interface{}  `json:"env,omitempty"`
	APITrace         interface{}  `json:"api_trace,omitempty"`
	Artifacts        interface{}  `json:"artifacts,omitempty"`
	SystemEvents     interface{}  `json:"sys_events,omitempty"`
	ProcessTree      interface{}  `json:"proc_tree,omitempty"`
	Capabilities     interface{}  `json:"capabilities,omitempty"`
	ScreenshotsCount int          `json:"screenshots_count,omitempty"`
	ScanConfig       interface{}  `json:"scan_cfg,omitempty"`
	SandboxLog       interface{}  `json:"sandbox_log,omitempty"`
	AgentLog         interface{}  `json:"agent_log,omitempty"`
	Status           int          `json:"status,omitempty"`
}
