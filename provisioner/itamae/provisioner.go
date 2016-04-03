/*
 * provisioner.go
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
	"runtime"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/provisioner"
	"github.com/mitchellh/packer/template/interpolate"
)

const (
	DefaultGem        = "itamae"
	DefaultCommand    = "itamae"
	DefaultStagingDir = "/tmp/packer-itamae"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Gem             string
	Command         string
	Vars            []string `mapstructure:"environment_vars"`
	InstallCommand  string   `mapstructure:"install_command"`
	SkipInstall     bool     `mapstructure:"skip_install"`
	ExecuteCommand  string   `mapstructure:"execute_command"`
	ExtraArguments  []string `mapstructure:"extra_arguments"`
	PreventSudo     bool     `mapstructure:"prevent_sudo"`
	StagingDir      string   `mapstructure:"staging_directory"`
	CleanStagingDir bool     `mapstructure:"clean_staging_directory"`
	SourceDir       string   `mapstructure:"source_directory"`
	LogLevel        string   `mapstructure:"log_level"`
	Shell           string   `mapstructure:"shell"`
	NodeJson        string   `mapstructure:"node_json"`
	NodeYaml        string   `mapstructure:"node_yaml"`
	Recipes         []string `mapstructure:"recipes"`
	IgnoreExitCodes bool     `mapstructure:"ignore_exit_codes"`

	ctx interpolate.Context
}

type Provisioner struct {
	config        Config
	guestCommands *provisioner.GuestCommands
}

type ExecuteTemplate struct {
	Command        string
	Vars           string
	StagingDir     string
	LogLevel       string
	Shell          string
	NodeJson       string
	NodeYaml       string
	ExtraArguments string
	Recipes        string
	Sudo           bool
}

type InstallTemplate struct {
	Gem  string
	Sudo bool
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"install_command",
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	p.guestCommands, err = provisioner.NewGuestCommands(p.guestOStype(), !p.config.PreventSudo)
	if err != nil {
		return err
	}

	if p.config.Gem == "" {
		p.config.Gem = DefaultGem
	}

	if p.config.Command == "" {
		p.config.Command = DefaultCommand
	}

	if p.config.Vars == nil {
		p.config.Vars = make([]string, 0)
	}

	if p.config.InstallCommand == "" {
		p.config.InstallCommand = "{{ if .Sudo}}sudo -E {{end}}" +
			"gem install --quiet --no-document --no-suggestions {{.Gem}}"
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "cd {{.StagingDir}} && " +
			"{{.Vars}} {{if .Sudo}}sudo -E {{end}}" +
			"{{.Command}} local --color='false' " +
			"{{if ne .LogLevel \"\"}}--log-level='{{.LogLevel}}' {{end}}" +
			"{{if ne .Shell \"\"}}--shell='{{.Shell}}' {{end}}" +
			"{{if ne .NodeJson \"\"}}--node-json='{{.NodeJson}}' {{end}}" +
			"{{if ne .NodeYaml \"\"}}--node-yaml='{{.NodeYaml}}' {{end}}" +
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
		if err := p.validateDirConfig(p.config.SourceDir, "source_directory"); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if p.config.Recipes == nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("A list of recipes must be specified."))
	} else if len(p.config.Recipes) == 0 {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("A list of recipes cannot be empty."))
	} else {
		for idx, path := range p.config.Recipes {
			if err := p.validateFileConfig(path, fmt.Sprintf("recipes[%d]", idx)); err != nil {
				errs = packer.MultiErrorAppend(errs, err)
			}
		}
	}

	if p.config.NodeJson != "" {
		if err := p.validateFileConfig(p.config.NodeJson, "node_json"); err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if p.config.NodeYaml != "" {
		if err := p.validateFileConfig(p.config.NodeYaml, "node_yaml"); err != nil {
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

	if !p.config.SkipInstall {
		if err := p.installItamae(ui, comm); err != nil {
			return fmt.Errorf("Error installing Itamae: %s", err)
		}
	}

	ui.Message("Creating staging directory...")
	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	if p.config.SourceDir != "" {
		ui.Message("Uploading source directory to staging directory...")
		if err := p.uploadDir(ui, comm, p.config.StagingDir, p.config.SourceDir); err != nil {
			return fmt.Errorf("Error uploading source directory: %s", err)
		}
	} else {
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

func (p *Provisioner) guestOStype() string {
	unixes := map[string]bool{
		"darwin":  true,
		"freebsd": true,
		"linux":   true,
		"openbsd": true,
	}

	if unixes[runtime.GOOS] {
		return "unix"
	}
	return runtime.GOOS
}

func (p *Provisioner) prefixPath(path, prefix string) string {
	if prefix != "" {
		path = filepath.Join(prefix, path)
	}
	return filepath.ToSlash(path)
}

func (p *Provisioner) validateDirConfig(path, config string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, path, err)
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s: %s must point to a directory", config, path)
	}
	return nil
}

func (p *Provisioner) validateFileConfig(path, config string) error {
	path = p.prefixPath(path, p.config.SourceDir)

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %s is invalid: %s", config, path, err)
	}

	if fi.IsDir() {
		return fmt.Errorf("%s: %s must point to a file", config, path)
	}
	return nil
}

func (p *Provisioner) installItamae(ui packer.Ui, comm packer.Communicator) error {
	ui.Message("Installing Itamae...")

	p.config.ctx.Data = &InstallTemplate{
		Gem:  p.config.Gem,
		Sudo: !p.config.PreventSudo,
	}

	command, err := interpolate.Render(p.config.InstallCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{
		Command: command,
	}

	ui.Message(fmt.Sprintf("Executing: %s", command))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) executeItamae(ui packer.Ui, comm packer.Communicator) error {
	ui.Message("Executing Itamae...")

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
		NodeJson:       p.config.NodeJson,
		NodeYaml:       p.config.NodeYaml,
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

	ui.Message(fmt.Sprintf("Executing: %s", command))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 && !p.config.IgnoreExitCodes {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	cmd := &packer.RemoteCmd{
		Command: p.guestCommands.CreateDir(dir),
	}

	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}

	cmd = &packer.RemoteCmd{
		Command: p.guestCommands.Chmod(dir, "0777"),
	}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	cmd := &packer.RemoteCmd{
		Command: p.guestCommands.RemoveDir(dir),
	}

	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more information.")
	}
	return nil
}

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	ui.Message(fmt.Sprintf("Uploading file: %s", src))
	return comm.Upload(dst, f, nil)
}

func (p *Provisioner) uploadDir(ui packer.Ui, comm packer.Communicator, dst, src string) error {
	ui.Message(fmt.Sprintf("Uploading directory: %s", src))
	if ok := strings.HasSuffix(src, "/"); !ok {
		src += "/"
	}
	return comm.UploadDir(dst, src, nil)
}
