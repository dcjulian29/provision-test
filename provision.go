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

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
)

// provision runs an Ansible playbook against the Vagrant test VM for the given
// hostname.
//
// It first pings the VM at 192.168.57.42 to verify network connectivity. If
// the ping succeeds, it:
//
//  1. Copies the local hosts.ini inventory file to hosts.test.ini.
//  2. Appends an [all:vars] block with Vagrant-specific SSH connection
//     parameters (host, port, private key, user, and options to bypass strict
//     host-key checking).
//  3. Invokes [ansible-host provision] with the temporary inventory and the
//     specified hostname.
//  4. Removes the temporary hosts.test.ini file.
//
// The hardcoded Vagrant VM address is 192.168.57.42, and the SSH private key
// is expected at ~/.ssh/insecure_private_key.
func provision(hostname string) error {
	count := "c"

	if runtime.GOOS == "windows" {
		count = "n"
	}

	if err := execute.ExternalProgram("ping",
		fmt.Sprintf("-%s", count),
		"1",
		"192.168.57.42"); err != nil {
		return err
	}

	if err := filesystem.CopyFile("hosts.ini", "hosts.test.ini"); err != nil {
		return err
	}

	if err := filesystem.AppendFile("hosts.test.ini", []byte(`
[all:vars]
ansible_host=192.168.57.42
ansible_port=22
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o CheckHostIP=no'
ansible_ssh_private_key_file=~/.ssh/insecure_private_key
ansible_user=vagrant`)); err != nil {
		return err
	}

	if err := execute.ExternalProgram("ansible-host",
		"provision",
		"--verbose",
		"--inventory",
		"./hosts.test.ini",
		hostname); err != nil {
		return err
	}

	return os.Remove("hosts.test.ini")
}
