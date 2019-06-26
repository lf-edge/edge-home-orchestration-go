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

package resourceutil

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
