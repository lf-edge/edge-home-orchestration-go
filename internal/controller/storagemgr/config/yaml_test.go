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
	"testing"
)

var (
	testName         = "datastorage"
	testManufacture  = "Home Edge"
	testModel        = "Home Edge"
	testLabel        = []string{"rest", "json", "numeric", "float", "int"}
	testDescription  = "REST Device"
	testPropertyJSON = Property{
		Value: PropertyDetail{
			Type:      "String",
			ReadWrite: "RW",
			MediaType: "application/json"},
		Units: PropertyDetail{
			Type:      "String",
			ReadWrite: "R"}}
	testPropertyInt = Property{
		Value: PropertyDetail{
			Type:      "Int64",
			ReadWrite: "RW",
			MediaType: "text/plain"},
		Units: PropertyDetail{
			Type:      "String",
			ReadWrite: "R"}}
	testPropertyFloat = Property{
		Value: PropertyDetail{
			Type:      "Float64",
			ReadWrite: "RW",
			MediaType: "text/plain"},
		Units: PropertyDetail{
			Type:      "String",
			ReadWrite: "R"}}
	testPropertyJpeg = Property{
		Value: PropertyDetail{
			Type:      "Binary",
			ReadWrite: "RW",
			MediaType: "image/jpeg"},
		Units: PropertyDetail{
			Type:      "String",
			ReadWrite: "R"}}
	testPropertyPng = Property{
		Value: PropertyDetail{
			Type:      "Binary",
			ReadWrite: "RW",
			MediaType: "image/png"},
		Units: PropertyDetail{
			Type:      "String",
			ReadWrite: "R"}}
	testPropertyString = Property{
		Value: PropertyDetail{
			Type:      "String",
			ReadWrite: "RW",
			MediaType: "text/plain"},
		Units: PropertyDetail{
			Type:      "String",
			ReadWrite: "R"}}
	testResource = []DeviceResource{
		{
			Name:        "json",
			Description: "json",
			Properties:  testPropertyJSON},
		{
			Name:       "int",
			Properties: testPropertyInt},
		{
			Name:        "float",
			Description: "float",
			Properties:  testPropertyFloat},
		{
			Name:        "jpeg",
			Description: "jpeg",
			Properties:  testPropertyJpeg},
		{
			Name:        "png",
			Description: "png",
			Properties:  testPropertyPng},
		{
			Name:        "string",
			Description: "string",
			Properties:  testPropertyString}}
)

func TestYaml(t *testing.T) {
	SetYaml(testName, testManufacture, testModel, testDescription, testLabel, testResource)

	b, err := YamlMarshal()
	if err != nil {
		t.Fatal("Unexpected Error")
	}

	log.Println(string(b))
}
