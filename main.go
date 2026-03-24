/*
Copyright © 2026 Julian Easterling

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// provision-test is a CLI tool for testing Ansible provision playbooks against
// a Vagrant-based virtual machine.
//
// It automates the lifecycle of a Vagrant VM (create, destroy, SSH) and then
// runs an Ansible provisioning playbook against it over SSH. The tool reads a
// local hosts.ini inventory file, creates a temporary hosts.test.ini with
// Vagrant-specific connection overrides (IP, SSH key, user), invokes
// [ansible-host] to provision the specified hostname, and cleans up the
// temporary inventory file afterward.
//
// Usage:
//
//	provision-test <hostname>                Destroy + recreate the VM, then provision <hostname>
//	provision-test <hostname> norecreate     Skip recreation; just bring the VM up and provision
//	provision-test ssh                       Open an interactive SSH session to the VM
//	provision-test destroy                   Destroy the VM without provisioning
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/textformat"
)

// main is the entry point for the provision-test CLI. It inspects the
// command-line arguments to determine the requested action:
//
//   - With no arguments, it prints usage guidance.
//   - "ssh" opens an interactive SSH session to the Vagrant VM via [vagrant ssh].
//   - "destroy" tears down the Vagrant VM via [vagrant destroy --force].
//   - A bare hostname destroys and recreates the VM, then provisions it.
//   - A hostname followed by "norecreate" skips the destroy step, brings the
//     VM up, and provisions it. Any other second argument prints an error.
func main() {
	switch len(os.Args) {
	case 1:
		fmt.Println(textformat.Fatal("Please provide host name to test or action (ssh or destroy)"))
		fmt.Println(textformat.Warn("To skip recreating, add 'norecreate' after hostname"))

		os.Exit(1)
	case 2:
		if strings.EqualFold(os.Args[1], "ssh") {
			if err := execute.ExternalProgram("vagrant", "ssh"); err != nil {
				fmt.Println(textformat.Fatal(err.Error()))

				os.Exit(1)
			}

			os.Exit(0)
		}

		if strings.EqualFold(os.Args[1], "destroy") {
			if err := execute.ExternalProgram("vagrant", "destroy", "--force"); err != nil {
				fmt.Println(textformat.Fatal(err.Error()))

				os.Exit(1)
			}

			os.Exit(0)
		}

		if err := execute.ExternalProgram("vagrant", "destroy", "--force"); err != nil {
			fmt.Println(textformat.Fatal(err.Error()))

			os.Exit(1)
		}

		if err := execute.ExternalProgram("vagrant", "up"); err != nil {
			fmt.Println(textformat.Fatal(err.Error()))

			os.Exit(1)
		}

	case 3:
		if strings.EqualFold(os.Args[2], "norecreate") {
			if err := execute.ExternalProgram("vagrant", "up"); err != nil {
				fmt.Println(textformat.Fatal(err.Error()))

				os.Exit(1)
			}
		}

		fmt.Println(textformat.Fatal("use 'norecreate' to skip recreating VM"))

		os.Exit(1)
	}

	if err := provision(os.Args[1]); err != nil {
		fmt.Println(textformat.Fatal(err.Error()))

		os.Exit(1)
	}

	os.Exit(0)
}
