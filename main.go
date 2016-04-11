// +build !windows

/*
 * main.go
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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilczynski/packer-provisioner-itamae-local/itamae"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "-help", "--help", "help":
			fmt.Printf(
				"Usage: %s [--version] [--help] <COMMAND>\n\n"+
					"Available commands are:\n"+
					"    version    Print the version and exit.\n"+
					"    help       Show this help screen.\n\n",
				filepath.Base(os.Args[0]))
		case "version":
			version := fmt.Sprintf("Provisioner Itamae v%s", itamaelocal.Version)
			if itamaelocal.Revision != "" {
				version += fmt.Sprintf(" (%s)", itamaelocal.Revision)
			}
			fmt.Println(version)
		case "-version", "--version":
			fmt.Printf("%s\n", itamaelocal.Version)
		}
		os.Exit(0)
	}

	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}

	server.RegisterProvisioner(&itamaelocal.Provisioner{})
	server.Serve()
}
