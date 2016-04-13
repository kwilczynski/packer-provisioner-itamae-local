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

package itamaelocal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

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
		t.Errorf("should be an error when key is invalid")
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
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		os.TempDir(),
	}

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipe file points to a directory")
	}

	config["recipes"] = []string{
		"fake.rb",
	}

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipe file does not exist")
	}

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	kind = reflect.ValueOf(p.config.Gems).Kind()
	if kind != reflect.Slice || len(p.config.Gems) != 2 {
		t.Errorf("incorrect gem, given {%v %d}, want {%v %d}",
			kind, len(p.config.Gems), reflect.Slice, 2)
	}

	if p.config.Command != DefaultCommand {
		t.Errorf("incorrect command, given \"%s\", want \"%s\"",
			p.config.Command, DefaultCommand)
	}

	kind = reflect.ValueOf(p.config.Vars).Kind()
	if kind != reflect.Slice || len(p.config.Vars) != 0 {
		t.Errorf("incorrect environment_vars, given {%v %d}, want {%v %d}",
			kind, len(p.config.Vars), reflect.Slice, 0)
	}

	kind = reflect.ValueOf(p.config.InstallCommand).Kind()
	if kind != reflect.String || p.config.InstallCommand == "" {
		t.Errorf("incorrect install_command, given {%v %d}, want {%v > 0}",
			kind, len(p.config.InstallCommand), reflect.String)
	}

	if p.config.SkipInstall {
		t.Errorf("incorrect skip_install, given: \"%v\", want \"%v\"",
			p.config.SkipInstall, false)
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

	kind = reflect.ValueOf(p.config.NodeJson).Kind()
	if kind != reflect.String || p.config.NodeJson != "" {
		t.Errorf("incorrect node_json, given {%v %d}, want {%v 0}",
			kind, len(p.config.NodeJson), reflect.String)
	}

	kind = reflect.ValueOf(p.config.NodeYaml).Kind()
	if kind != reflect.String || p.config.NodeYaml != "" {
		t.Errorf("incorrect node_yaml, given {%v %d}, want {%v 0}",
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
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
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
		t.Errorf("should be an error if a bad environment variable is present")
	}

	config["environment_vars"] = []string{
		"=bad",
	}

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if a bad environment variable is present")
	}

	config["environment_vars"] = []string{
		"EMPTY=",
		"UPPER=yes",
		"test1=variable",
		"test2=(abc=def)",
		"test3=baz=quux",
	}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := []string{
		"EMPTY=''",
		"UPPER='yes'",
		"test1='variable'",
		"test2='(abc=def)'",
		"test3='baz=quux'",
	}

	if ok := reflect.DeepEqual(p.config.Vars, expected); !ok {
		t.Errorf("value given %v, want %v", p.config.Vars, expected)
	}

	p = Provisioner{}
	delete(config, "environment_vars")

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
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["extra_arguments"] = "{{}}"
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if extra_arguments contains an illegal value")
	}

	p = Provisioner{}
	delete(config, "extra_arguments")

	arguments := []string{
		"--argument",
		"--option=value",
		"some-string",
		fmt.Sprintf("--date='%s'", time.Now()),
	}

	config["extra_arguments"] = arguments
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	if ok := reflect.DeepEqual(p.config.ExtraArguments, arguments); !ok {
		t.Errorf("value given %v, want %v", p.config.ExtraArguments, arguments)
	}
}

func TestProvisionerPrepare_StagingDirectory(t *testing.T) {
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
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["staging_directory"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := os.TempDir()
	if p.config.StagingDir != expected {
		t.Errorf("value given %v, want %v", p.config.StagingDir, expected)
	}
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
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		filepath.Base(recipeFile.Name()),
	}

	config["source_directory"] = recipeFile.Name()
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if source_directory points to a file")
	}

	config["source_directory"] = "/does/not/exist"
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if source_directory does not exist")
	}

	config["source_directory"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}
}

func TestProvisionerPrepare_Recipes(t *testing.T) {
	var err error
	var p Provisioner

	config := testConfig()

	path1, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	path2, err := ioutil.TempFile("", "role.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	directory, err := ioutil.TempDir("", "recipes")
	if err != nil {
		t.Fatalf("unable to create temporary directory: %s", err)
	}

	path3, err := ioutil.TempFile(directory, "test.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer func() {
		os.Remove(path1.Name())
		os.Remove(path2.Name())
		os.Remove(path3.Name())
		os.Remove(directory)
	}()

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		path1.Name(),
		path2.Name(),
		path3.Name(),
	}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := []string{
		path1.Name(),
		path2.Name(),
		path3.Name(),
	}

	if ok := reflect.DeepEqual(p.config.Recipes, expected); !ok {
		t.Errorf("value given %v, want %v", p.config.Recipes, expected)
	}
}

func TestProvisionerPrepare_NodeJson(t *testing.T) {
	var err error
	var p Provisioner

	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	nodeFile, err := ioutil.TempFile("", "node.json")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer os.Remove(recipeFile.Name())
	defer os.Remove(nodeFile.Name())

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["node_json"] = os.TempDir()
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if node_json points to a directory")
	}

	p = Provisioner{}
	delete(config, "node_json")

	config["node_json"] = nodeFile.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}
}

func TestProvisionerPrepare_NodeYaml(t *testing.T) {
	var err error
	var p Provisioner

	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	nodeFile, err := ioutil.TempFile("", "node.yml")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer os.Remove(recipeFile.Name())
	defer os.Remove(nodeFile.Name())

	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is missing")
	}

	config["recipes"] = []string{}
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if recipes list is empty")
	}

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["node_yaml"] = os.TempDir()
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if node_yaml points to a directory")
	}

	p = Provisioner{}
	delete(config, "node_yaml")

	config["node_yaml"] = nodeFile.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}
}

func TestProvisionerProvision_Defaults(t *testing.T) {
	var err error
	var p Provisioner

	buffer := &bytes.Buffer{}

	ui := testUi(buffer)
	comm := testCommunicator()
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

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.ctx.Data = &InstallTemplate{
		Gems: strings.Join(p.config.Gems, " "),
		Sudo: !p.config.PreventSudo,
	}

	installCommand, err := interpolate.Render(p.config.InstallCommand, &p.config.ctx)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := "sudo -E gem install --quiet --no-document --no-suggestions itamae"
	if ok := strings.Contains(buffer.String(), expected); !ok {
		t.Errorf("incorrect install_command, given: \"%v\", want \"%v\"",
			installCommand, expected)
	}

	expected = fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' %s", recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_SkipInstall(t *testing.T) {
	var err error
	var p Provisioner

	buffer := &bytes.Buffer{}

	ui := testUi(buffer)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["install_command"] = "this should not be present"
	config["skip_install"] = true

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := "this should not be present"
	if ok := strings.Contains(buffer.String(), expected); ok {
		t.Errorf("should not include install_command, but got: %s", buffer)
	}
}

func TestProvisionerProvision_InstallCommand(t *testing.T) {
	var err error
	var p Provisioner

	buffer := &bytes.Buffer{}

	ui := testUi(buffer)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["install_command"] = "{{}}"
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if install_command contains an illegal value")
	}

	p = Provisioner{}
	delete(config, "install_command")

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	config["install_command"] = "gem install itamae -v 1.9.5"
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := "gem install itamae -v 1.9.5"
	if ok := strings.Contains(buffer.String(), expected); !ok {
		t.Errorf("incorrect install_command, given: \"%v\", want \"%v\"",
			p.config.InstallCommand, expected)
	}
}

func TestProvisionerProvision_ExecuteCommand(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["execute_command"] = "{{}}"
	err = p.Prepare(config)
	if err == nil {
		t.Errorf("should be an error if execute_command contains an illegal value")
	}

	p = Provisioner{}
	delete(config, "execute_command")

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	config["execute_command"] = "{{.Vars}} {{.Command}} local {{.Recipes}}"
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' itamae local %s", recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_EnvironmentVars(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	date := time.Now()

	varibales := []string{
		"name=value",
		fmt.Sprintf("DATE='%s'", date),
	}

	config["environment_vars"] = varibales
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	execptedVariables := []string{
		"name='value'",
		fmt.Sprintf("DATE=''\"'\"'%s'\"'\"''", date),
	}

	if ok := reflect.DeepEqual(p.config.Vars, execptedVariables); !ok {
		t.Errorf("value given %v, want %v", p.config.Vars, execptedVariables)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' %s "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' %s", strings.Join(execptedVariables, " "),
		recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_StagingDirectory(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	directory, err := ioutil.TempDir("", "staging")
	if err != nil {
		t.Fatalf("unable to create temporary directory: %s", err)
	}

	defer os.Remove(recipeFile.Name())
	defer os.Remove(directory)

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["staging_directory"] = directory
	config["clean_staging_directory"] = true

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	if p.config.StagingDir != directory {
		t.Errorf("value given %v, want %v", p.config.StagingDir, directory)
	}

	if !p.config.CleanStagingDir {
		t.Errorf("value given %v, want %v", p.config.CleanStagingDir, true)
	}

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p = Provisioner{}
	delete(config, "clean_staging_directory")

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd %s && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' %s", directory, recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_SourceDirectory(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	directory, err := ioutil.TempDir("", "source")
	if err != nil {
		t.Fatalf("unable to create temporary directory: %s", err)
	}

	recipeFile, err := ioutil.TempFile(directory, "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer os.Remove(recipeFile.Name())
	defer os.Remove(directory)

	config["recipes"] = []string{
		filepath.Base(recipeFile.Name()),
	}

	config["source_directory"] = directory
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	if p.config.SourceDir != directory {
		t.Errorf("value given %v, want %v", p.config.SourceDir, directory)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' %s", filepath.Base(recipeFile.Name()))

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_LogLevel(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["log_level"] = "debug"
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' --log-level='debug' %s",
		recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_Shell(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["shell"] = "/bin/bash"
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' --shell='/bin/bash' %s",
		recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_NodeJson(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	nodeFile, err := ioutil.TempFile("", "node.json")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer os.Remove(recipeFile.Name())
	defer os.Remove(nodeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["node_json"] = nodeFile.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' --node-json='%s' %s", nodeFile.Name(),
		recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_YamlPath(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	nodeFile, err := ioutil.TempFile("", "node.yml")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer os.Remove(recipeFile.Name())
	defer os.Remove(nodeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["node_yaml"] = nodeFile.Name()
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' --node-yaml='%s' %s", nodeFile.Name(),
		recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_ExtraArguments(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	arguments := []string{
		"--argument",
		"--option=value",
		"some-string",
		fmt.Sprintf("--date='%s'", time.Now()),
	}

	config["extra_arguments"] = arguments
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	if ok := reflect.DeepEqual(p.config.ExtraArguments, arguments); !ok {
		t.Errorf("value given %v, want %v", p.config.ExtraArguments, arguments)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := strings.Join(arguments, " ")
	if ok := strings.Contains(comm.StartCmd.Command, expected); !ok {
		t.Errorf("incorrect execute_command, given \"%v\" does not contain "+
			"the expected arguments: \"%v\"", comm.StartCmd.Command, expected)
	}

	expected = fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' %s %s", strings.Join(arguments, " "),
		recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_Recipes(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	path1, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	path2, err := ioutil.TempFile("", "role.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	directory, err := ioutil.TempDir("", "recipes")
	if err != nil {
		t.Fatalf("unable to create temporary directory: %s", err)
	}

	path3, err := ioutil.TempFile(directory, "test.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}

	defer func() {
		os.Remove(path1.Name())
		os.Remove(path2.Name())
		os.Remove(path3.Name())
		os.Remove(directory)
	}()

	config["recipes"] = []string{
		path1.Name(),
		path2.Name(),
		path3.Name(),
	}

	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	recipes := []string{
		path1.Name(),
		path2.Name(),
		path3.Name(),
	}

	if ok := reflect.DeepEqual(p.config.Recipes, recipes); !ok {
		t.Errorf("value given %v, want %v", p.config.Recipes, recipes)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"sudo -E itamae local --detailed-exitcode "+
		"--color='false' %s", strings.Join(recipes, " "))

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisionerProvision_IgnoreExitCodes(t *testing.T) {
	// XXX(kwilczynski): How to test this?
}

func TestProvisionerProvision_PreventSudo(t *testing.T) {
	var err error
	var p Provisioner

	ui := testUi(nil)
	comm := testCommunicator()
	config := testConfig()

	recipeFile, err := ioutil.TempFile("", "recipe.rb")
	if err != nil {
		t.Fatalf("unable to create temporary file: %s", err)
	}
	defer os.Remove(recipeFile.Name())

	config["recipes"] = []string{
		recipeFile.Name(),
	}

	config["prevent_sudo"] = true
	err = p.Prepare(config)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	p.config.PackerBuildName = "virtualbox"
	p.config.PackerBuilderType = "iso"

	err = p.Provision(ui, comm)
	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}

	expected := fmt.Sprintf("cd /tmp/packer-itamae && "+
		"PACKER_BUILD_NAME='virtualbox' "+
		"PACKER_BUILDER_TYPE='iso' "+
		"itamae local --detailed-exitcode "+
		"--color='false' %s", recipeFile.Name())

	if comm.StartCmd.Command != expected {
		t.Errorf("incorrect execute_command, given: \"%v\", want \"%v\"",
			comm.StartCmd.Command, expected)
	}
}

func TestProvisioner_Cancel(t *testing.T) {
	var p Provisioner

	if os.Getenv("TEST_CANCEL") == "1" {
		p.Cancel()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestProvisioner_Cancel")
	cmd.Env = append(os.Environ(), "TEST_CANCEL=1")

	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && e.Success() {
		return
	}

	if err != nil {
		t.Errorf("should not error, but got: %s", err)
	}
}
