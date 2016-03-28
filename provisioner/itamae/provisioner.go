/*
 * itamae.go
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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const (
	DefaultCommand    = "itamae"
	DefaultStagingDir = "/tmp/packer-itamae"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Command         string
	Vars            []string `mapstructure:"environment_vars"`
	ExecuteCommand  string   `mapstructure:"execute_command"`
	ExtraArguments  []string `mapstructure:"extra_arguments"`
	PreventSudo     bool     `mapstructure:"prevent_sudo"`
	StagingDir      string   `mapstructure:"staging_directory"`
	CleanStagingDir bool     `mapstructure:"clean_staging_directory"`
	SourceDir       string   `mapstructure:"source_directory"`
	LogLevel        string   `mapstructure:"log_level"`
	Shell           string   `mapstructure:"shell"`
	JsonPath        string   `mapstructure:"json_path"`
	YamlPath        string   `mapstructure:"yaml_path"`
	Recipes         []string `mapstructure:"recipes"`
	IgnoreExitCodes bool     `mapstructure:"ignore_exit_codes"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

type ExecuteTemplate struct {
	Command        string
	Vars           string
	StagingDir     string
	LogLevel       string
	Shell          string
	JsonPath       string
	YamlPath       string
	ExtraArguments string
	Recipes        string
	Sudo           bool
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

	if p.config.Command == "" {
		p.config.Command = DefaultCommand
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "cd {{.StagingDir}} && " +
			"{{.Vars}} {{if .Sudo}}sudo -E {{end}}" +
			"{{.Command}} local --color='false' " +
			"{{if ne .LogLevel \"\"}}--log-level='{{.LogLevel}}' {{end}}" +
			"{{if ne .Shell \"\"}}--shell='{{.Shell}}' {{end}}" +
			"{{if ne .JsonPath \"\"}}--node-json='{{.JsonPath}}' {{end}}" +
			"{{if ne .YamlPath \"\"}}--node-yaml='{{.YamlPath}}' {{end}}" +
			"{{if ne .ExtraArguments \"\"}}{{.ExtraArguments}} {{end}}" +
			"{{.Recipes}}"
	}

	if p.config.ExtraArguments == nil {
		p.config.ExtraArguments = make([]string, 0)
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = DefaultStagingDir
	}

	var errs *packer.MultiError

	if p.config.SourceDir != "" {
		if err := validateDirConfig(p.config.SourceDir, "source_directory"); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if p.config.Recipes == nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("A recipes must be specified."))
	} else {
		for idx, path := range p.config.Recipes {
			path = withDirPrefix(path, p.config.SourceDir)
			if err := validateFileConfig(path, fmt.Sprintf("recipes[%d]", idx)); err != nil {
				errs = packer.MultiErrorAppend(errs, err)
			}
		}
	}

	if p.config.JsonPath != "" {
		jsonPath := withDirPrefix(p.config.JsonPath, p.config.SourceDir)
		if err := validateFileConfig(jsonPath, "json_path"); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if p.config.YamlPath != "" {
		yamlPath := withDirPrefix(p.config.YamlPath, p.config.SourceDir)
		if err := validateFileConfig(yamlPath, "yaml_path"); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	for idx, kv := range p.config.Vars {
		vs := strings.SplitN(kv, "=", 2)
		if len(vs) != 2 || vs[0] == "" {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Environment variable not in format 'key=value': %s", kv))
		} else {
			vs[1] = strings.Replace(vs[1], "'", `'"'"'`, -1)
			p.config.Vars[idx] = fmt.Sprintf("%s='%s'", vs[0], vs[1])
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Itamae...")

	if p.config.SourceDir != "" {
		ui.Message("Uploading source directory to staging directory...")
		if err := p.uploadDir(ui, comm, p.config.StagingDir, p.config.SourceDir); err != nil {
			return fmt.Errorf("Error uploading source directory: %s", err)
		}
	} else {
		ui.Message("Creating staging directory...")
		if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error creating staging directory: %s", err)
		}

		ui.Message("Uploading recipes...")
		for _, src := range p.config.Recipes {
			dst := filepath.ToSlash(filepath.Join(p.config.StagingDir, src))
			if err := p.uploadFile(ui, comm, dst, src); err != nil {
				return fmt.Errorf("Error uploading recipe: %s", err)
			}
		}
	}

	if err := p.executeItamae(ui, comm); err != nil {
		return fmt.Errorf("Error executing Itamae: %s", err)
	}

	if p.config.CleanStagingDir {
		ui.Message("Removing staging directory...")
		if err := p.removeDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error removing staging directory: %s", err)
		}
	}
	return nil
}

func (p *Provisioner) Cancel() {
	os.Exit(0)
}

func withDirPrefix(path, prefix string) string {
	if prefix != "" {
		path = filepath.Join(prefix, path)
	}
	return filepath.ToSlash(path)
}

func validateDirConfig(path, config string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, path, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("%s: %s must point to a directory", config, path)
	}
	return nil
}

func validateFileConfig(path, config string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, path, err)
	} else if fi.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, path)
	}
	return nil
}

func (p *Provisioner) executeItamae(ui packer.Ui, comm packer.Communicator) error {
	envVars := make([]string, len(p.config.Vars)+2)
	envVars[0] = fmt.Sprintf("PACKER_BUILD_NAME='%s'", p.config.PackerBuildName)
	envVars[1] = fmt.Sprintf("PACKER_BUILDER_TYPE='%s'", p.config.PackerBuilderType)
	copy(envVars[2:], p.config.Vars)

	p.config.ctx.Data = &ExecuteTemplate{
		Command:        p.config.Command,
		Vars:           strings.Join(envVars, " "),
		StagingDir:     p.config.StagingDir,
		LogLevel:       p.config.LogLevel,
		Shell:          p.config.Shell,
		JsonPath:       p.config.JsonPath,
		YamlPath:       p.config.YamlPath,
		ExtraArguments: strings.Join(p.config.ExtraArguments, " "),
		Recipes:        strings.Join(p.config.Recipes, " "),
		Sudo:           !p.config.PreventSudo,
	}

	command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{
		Command: command,
	}

	ui.Message(fmt.Sprintf("Executing Itamae: %s", command))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 && !p.config.IgnoreExitCodes {
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}
	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("rm -rf '%s'", dir),
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}
	return nil
}

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return comm.Upload(dst, f, nil)
}

func (p *Provisioner) uploadDir(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	if ok := strings.HasSuffix(src, "/"); !ok {
		src += "/"
	}

	return comm.UploadDir(dst, src, nil)
}
