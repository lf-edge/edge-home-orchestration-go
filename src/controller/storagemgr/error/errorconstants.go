/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
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

// Package util provides string constants referred to in data storage for error scenario

package error

const (
	DSInitializeError   = "Device service intialize started"
	DSReadNotSupported  = "StorageDriver.HandleReadCommands; read commands not supported"
	DSWriteNotSupported = "StorageDriver.HandleWriteCommands; write commands not supported"
	DSForceStop         = "StorageDriver.Stop called: force=%v"
	DSSuccess           = "Device has been successfully added!!!!!!"
	DSRouteAddError     = "unable to add required route:"
	DSRouteSuccess      = "Route %s added."
	DSResourceError     = "Incoming reading ignored. Resource '%s' not found"
	DSNoContentType     = "No Content-Type"
	DSContentTypeError  = "Wrong Content-Type"
	DSReadError         = "Incoming reading ignored. Unable to read request body: %s"
	DSCommandValueError = "Incoming reading ignored. Unable to create Command Value for Device=%s Command=%s: %s"
	DSNoRequestBody     = "no request body provided"
	DSParserError       = "parse reading fail. Reading %v is out of the value type(%v)'s range"
	DSResultFailure     = "return result fail, none supported value type: %v"
)
