/*
 * provisioner_test.go
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

package itamae

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestProvisioner(t *testing.T) {
	p := &Provisioner{}

	func(v interface{}) {
		if _, ok := v.(packer.Provisioner); !ok {
			t.Fatalf("not a Provisioner type: %s", reflect.TypeOf(v).String())
		}
	}(p)
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var err error
	var p Provisioner

	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		"fake.rb",
	}

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if recipe file does not exist")
	}

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not an error, but got: %s", err)
	}

	if reflect.ValueOf(p.config.Vars).Kind() != reflect.Slice || len(p.config.Vars) > 0 {
		t.Errorf("incorrect environment_vars, given {%v %v}, want {%v %d}",
			reflect.ValueOf(p.config.Vars).Kind(), len(p.config.Vars), reflect.Slice, 0)
	}

	if p.config.Command != DefaultCommand {
		t.Errorf("incorrect command, given \"%s\", want \"%s\"",
			p.config.Command, DefaultCommand)
	}

	if reflect.ValueOf(p.config.ExtraArguments).Kind() != reflect.Slice || len(p.config.ExtraArguments) > 0 {
		t.Errorf("incorrect extra_arguments, given {%v %v}, want {%v %d}",
			reflect.ValueOf(p.config.ExtraArguments).Kind(),
			len(p.config.Vars), reflect.Slice, 0)
	}

	if p.config.StagingDir != DefaultStagingDir {
		t.Errorf("incorrect staging_directory, given \"%s\", want \"%s\"",
			p.config.StagingDir, DefaultStagingDir)
	}
}
