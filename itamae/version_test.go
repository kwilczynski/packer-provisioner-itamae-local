/*
 * version_test.go
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
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestProvisioner_Version(t *testing.T) {
	var err error
	var p Provisioner

	buffer := &bytes.Buffer{}

	log.SetOutput(buffer)
	defer func() {
		buffer.Reset()
		buffer = nil
		log.SetOutput(ioutil.Discard)
	}()

	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := "[INFO] Provisioner Itamae v0.1.0"
	if ok := strings.Contains(buffer.String(), expected); !ok {
		t.Errorf("incorrect version, given: \"%v\", want \"%v\"",
			buffer.String(), expected)
	}

	buffer.Reset()

	Revision = "some-git-revision-12345"
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected = "[INFO] Provisioner Itamae v0.1.0 (some-git-revision-12345)"
	if ok := strings.Contains(buffer.String(), expected); !ok {
		t.Errorf("incorrect version, given: \"%v\", want \"%v\"",
			buffer.String(), expected)
	}
}
