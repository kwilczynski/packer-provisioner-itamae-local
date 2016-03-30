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
	"path/filepath"
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

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["invalid-key"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should be an error when key is invalid")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var err error
	var p Provisioner
	var kind reflect.Kind

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
		os.TempDir(),
	}

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if recipe file points to a directory")
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
		t.Fatalf("should not error, but got: %s", err)
	}

	kind = reflect.ValueOf(p.config.Vars).Kind()
	if kind != reflect.Slice || len(p.config.Vars) != 0 {
		t.Errorf("incorrect environment_vars, given {%v %d}, want {%v %d}",
			kind, len(p.config.Vars), reflect.Slice, 0)
	}

	if p.config.Command != DefaultCommand {
		t.Errorf("incorrect command, given \"%s\", want \"%s\"",
			p.config.Command, DefaultCommand)
	}

	kind = reflect.ValueOf(p.config.ExecuteCommand).Kind()
	if kind != reflect.String || p.config.ExecuteCommand == "" {
		t.Errorf("incorrect execute_command, given {%v %d}, want {%v > 0}",
			kind, len(p.config.ExecuteCommand), reflect.String)
	}

	kind = reflect.ValueOf(p.config.ExtraArguments).Kind()
	if kind != reflect.Slice || len(p.config.ExtraArguments) != 0 {
		t.Errorf("incorrect extra_arguments, given {%v %d}, want {%v %d}",
			kind, len(p.config.Vars), reflect.Slice, 0)
	}

	if p.config.PreventSudo {
		t.Errorf("incorrect prevent_sudo, given: \"%v\", want \"%v\"",
			p.config.PreventSudo, false)
	}

	if p.config.StagingDir != DefaultStagingDir {
		t.Errorf("incorrect staging_directory, given \"%s\", want \"%s\"",
			p.config.StagingDir, DefaultStagingDir)
	}

	if p.config.CleanStagingDir {
		t.Errorf("incorrect clean_staging_directory, given: \"%v\", want \"%v\"",
			p.config.CleanStagingDir, false)
	}

	kind = reflect.ValueOf(p.config.SourceDir).Kind()
	if kind != reflect.String || p.config.SourceDir != "" {
		t.Errorf("incorrect source_directory, given {%v %d}, want {%v 0}",
			kind, len(p.config.SourceDir), reflect.String)
	}

	kind = reflect.ValueOf(p.config.LogLevel).Kind()
	if kind != reflect.String || p.config.LogLevel != "" {
		t.Errorf("incorrect log_level, given {%v %d}, want {%v 0}",
			kind, len(p.config.LogLevel), reflect.String)
	}

	kind = reflect.ValueOf(p.config.Shell).Kind()
	if kind != reflect.String || p.config.Shell != "" {
		t.Errorf("incorrect shell, given {%v %d}, want {%v 0}",
			kind, len(p.config.Shell), reflect.String)
	}

	kind = reflect.ValueOf(p.config.JsonPath).Kind()
	if kind != reflect.String || p.config.JsonPath != "" {
		t.Errorf("incorrect json_path, given {%v %d}, want {%v 0}",
			kind, len(p.config.JsonPath), reflect.String)
	}

	kind = reflect.ValueOf(p.config.YamlPath).Kind()
	if kind != reflect.String || p.config.YamlPath != "" {
		t.Errorf("incorrect yaml_path, given {%v %d}, want {%v 0}",
			kind, len(p.config.Shell), reflect.String)
	}

	p = Provisioner{}

	delete(config, "recipes")
	p.Prepare(config)

	kind = reflect.ValueOf(p.config.Recipes).Kind()
	if kind != reflect.Slice || len(p.config.Recipes) != 0 {
		t.Errorf("incorrect recipes, given {%v %d}, want {%v %d}",
			kind, len(p.config.Recipes), reflect.Slice, 0)
	}

	if p.config.IgnoreExitCodes {
		t.Errorf("incorrect ignore_exit_codes, given: \"%v\", want \"%v\"",
			p.config.IgnoreExitCodes, false)
	}
}

func TestProvisionerPrepare_EnvironmentVars(t *testing.T) {
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
		recipeFile.Name(),
	}

	config["environment_vars"] = []string{
		"badvariable",
		"good=variable",
	}

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if a bad environment variable is present")
	}

	config["environment_vars"] = []string{
		"=bad",
	}

	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if a bad environment variable is present")
	}

	config["environment_vars"] = []string{
		"test=variable",
		"WORKING=yes",
		"EMPTY=",
	}

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not error, but got: %s", err)
	}

	expected := []string{
		"test='variable'",
		"WORKING='yes'",
		"EMPTY=''",
	}

	if ok := reflect.DeepEqual(p.config.Vars, expected); !ok {
		t.Errorf("value given %v, want %v", p.config.Vars, expected)
	}

	p = Provisioner{}
	delete(config, "environment_vars")

	// config["environment_vars"] = []string{"keyone=valueone", "keytwo=value\ntwo"}

	config["environment_vars"] = []string{
		"one=two",
		"two=three\nfour",
		"four='five'",
		"five='six\nseven'",
	}

	p.Prepare(config)

	expected = []string{
		"one='two'",
		"two='three\nfour'",
		"four=''\"'\"'five'\"'\"''",
		"five=''\"'\"'six\nseven'\"'\"''",
	}

	if ok := reflect.DeepEqual(p.config.Vars, expected); !ok {
		t.Errorf("value given %v, want %v", p.config.Vars, expected)
	}
}

func TestProvisionerPrepare_ExtraArguments(t *testing.T) {
}

func TestProvisionerPrepare_StagingDirectory(t *testing.T) {
}

func TestProvisionerPrepare_SourceDirectory(t *testing.T) {
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
		filepath.Base(recipeFile.Name()),
	}

	config["source_directory"] = recipeFile.Name()
	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if source_directory points to a file")
	}

	config["source_directory"] = "/does/not/exist"
	err = p.Prepare(config)
	if err == nil {
		t.Fatalf("should be an error if source_directory does not exist")
	}

	config["source_directory"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not error, but got: %s", err)
	}
}

func TestProvisionerPrepare_Recipes(t *testing.T) {
}

func TestProvisionerPrepare_JsonPath(t *testing.T) {
}

func TestProvisionerPrepare_YamlPath(t *testing.T) {
}

func TestProvisionerProvision_Command(t *testing.T) {
}

func TestProvisionerProvision_EnvironmentVars(t *testing.T) {
}

func TestProvisionerProvision_StagingDirectory(t *testing.T) {
}

func TestProvisionerProvision_LogLevel(t *testing.T) {
}

func TestProvisionerProvision_Shell(t *testing.T) {
}

func TestProvisionerProvision_JsonPath(t *testing.T) {
}

func TestProvisionerProvision_YamlPath(t *testing.T) {
}

func TestProvisionerProvision_ExtraArguments(t *testing.T) {
}

func TestProvisionerProvision_Recipes(t *testing.T) {
}

func TestProvisionerProvision_Sudo(t *testing.T) {
}
