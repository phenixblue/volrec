/*
Copyright 2021 The WebRoot.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"flag"

	"github.com/go-logr/logr"
)

var (
	// VolrecConfig holds the controller configuration
	VolrecConfig ControllerConfig
)

// ControllerConfig represents configuration for the controller
type ControllerConfig struct {
	ReclaimPolicyLabel string
	OwnerLabel         string
	OwnerSet           bool
	NsLabel            string
	NsSet              bool
}

// InitConfig initializes the controller configuration
func InitConfig(setupLog logr.Logger) {

	// Initialize the config to be used everywhere
	VolrecConfig.ReclaimPolicyLabel = flag.Lookup("reclaim-label").Value.(flag.Getter).Get().(string)
	VolrecConfig.OwnerLabel = flag.Lookup("owner-label").Value.(flag.Getter).Get().(string)
	VolrecConfig.OwnerSet = flag.Lookup("set-owner").Value.(flag.Getter).Get().(bool)
	VolrecConfig.NsLabel = flag.Lookup("ns-label").Value.(flag.Getter).Get().(string)
	VolrecConfig.NsSet = flag.Lookup("set-ns").Value.(flag.Getter).Get().(bool)

}
