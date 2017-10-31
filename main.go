package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer/plugin"
	"github.com/kwilczynski/packer-provisioner-itamae-local/itamae"
)

func main() {
	//
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
		case "-v", "-version", "--version":
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
