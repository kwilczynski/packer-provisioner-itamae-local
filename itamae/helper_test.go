/*
 * helper_test.go
 *
 * Copyright 2016 Krzysztof Wilczynski
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package itamaelocal

import (
	"io"
	"io/ioutil"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	return make(map[string]interface{})
}

func testUI(writer io.Writer) *packer.MachineReadableUi {
	if writer == nil {
		writer = ioutil.Discard
	}
	return &packer.MachineReadableUi{
		Writer: writer,
	}
}

func testCommunicator() *packer.MockCommunicator {
	return &packer.MockCommunicator{}
}
