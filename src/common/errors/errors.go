/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
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

// Package errors defines error cases of edge-orchestration.
package errors

// InvalidParam will be used for return case of error
type InvalidParam struct {
	Message string
}

// Error sets an error message of invalid method
func (e InvalidParam) Error() string {
	return "invalid param : " + e.Message
}

// SystemError will be used for return case of error
type SystemError struct {
	Message string
}

// Error sets an error message of system error
func (e SystemError) Error() string {
	return "system error : " + e.Message
}

// NotSupport will be used for return case of error
type NotSupport struct {
	Message string
}

// Error sets an error message of not support error
func (e NotSupport) Error() string {
	return "not support error : " + e.Message
}

// NotFound will be used for return case of error
type NotFound struct {
	Message string
}

// Error sets an error message of not found error
func (e NotFound) Error() string {
	return "not found error : " + e.Message
}

// DBConnectionError will be used for return case of error
type DBConnectionError struct {
	Message string
}

// Error sets an error message of system error
func (e DBConnectionError) Error() string {
	return "db connection error : " + e.Message
}

// DBOperationError will be used for return case of error
type DBOperationError struct {
	Message string
}

// Error sets an error message of db operation error
func (e DBOperationError) Error() string {
	return "db operation error : " + e.Message
}

// InvalidJSON will be used for return case of error
type InvalidJSON struct {
	Message string
}

// Error sets an error message of system error
func (e InvalidJSON) Error() string {
	return "invalid json error : " + e.Message
}

// NetworkError is error related with network operation
type NetworkError struct {
	Message string
}

// Error sets an error message of network error
func (e NetworkError) Error() string {
	return "network error : " + e.Message
}
