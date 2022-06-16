/*******************************************************************************
 * Copyright 2021 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

package config

import (
	"gopkg.in/yaml.v3"
)

// Yaml contains the struct for building the DataStorage Device Resources configuration.
type Yaml struct {
	Name            string           `yaml:"name"`
	Manufacturer    string           `yaml:"manufacturer"`
	Model           string           `yaml:"model"`
	Labels          []string         `yaml:"labels,omitempty"`
	Description     string           `yaml:"description,omitempty"`
	DeviceResources []DeviceResource `yaml:"deviceResources,omitempty"`
}

// DeviceResource contains the resource information.
type DeviceResource struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Properties  Property `yaml:"properties"`
}

// Property has the value and units properties.
type Property struct {
	Value PropertyDetail `yaml:"value,flow"`
	Units PropertyDetail `yaml:"units,flow"`
}

// PropertyDetail contains the specific property(Value and Units) information.
// Type : bool, int8 - int64, uint8 - uint64, float32, float64, string, binary types are supported.
// ReadWrite : R, RW, or W
type PropertyDetail struct {
	Type      string `yaml:"type"`
	ReadWrite string `yaml:"readWrite"`
	MediaType string `yaml:"mediaType,omitempty"`
}

var (
	yamlInfo Yaml
)

// SetYaml configures the device resource information.
func SetYaml(name, manufac, model, desc string, labels []string, resources []DeviceResource) {
	yamlInfo = Yaml{
		Name:            name,
		Manufacturer:    manufac,
		Model:           model,
		Labels:          labels,
		Description:     desc,
		DeviceResources: resources}
}

// YamlMarshal returns bytes for DataStorage device resource configuration.
func YamlMarshal() (b []byte, err error) {
	return yaml.Marshal(yamlInfo)
}
