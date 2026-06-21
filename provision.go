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
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
)

// getSSHHost returns the DHCP-assigned IP address of the running Vagrant VM
// by parsing the output of "vagrant ssh-config".
func getSSHHost() (string, error) {
	out, err := execute.ExternalProgramCapture("vagrant", "ssh-config")
	if err != nil {
		return "", fmt.Errorf("vagrant ssh-config: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "HostName ") {
			return strings.TrimPrefix(line, "HostName "), nil
		}
	}

	return "", fmt.Errorf("HostName not found in vagrant ssh-config output")
}

// provision runs an Ansible playbook against the Vagrant test VM for the
// given hostname.
//
// It first discovers the VM's DHCP-assigned IP via "vagrant ssh-config",
// then pings it to verify network connectivity. If the ping succeeds, it:
//
//  1. Copies the local hosts.ini inventory file to hosts.test.ini.
//  2. Appends an [all:vars] block with Vagrant-specific SSH connection
//     parameters using the discovered IP, port, private key, user, and
//     options to bypass strict host-key checking.
//  3. Invokes [ansible-host provision] with the temporary inventory and the
//     specified hostname.
//  4. Removes the temporary hosts.test.ini file.
func provision(hostname string) error {
	addr, err := getSSHHost()
	if err != nil {
		return err
	}

	count := "c"

	if runtime.GOOS == "windows" {
		count = "n"
	}

	fmt.Printf("VM address: %s\n", addr)

	if err := execute.ExternalProgram("ping",
		fmt.Sprintf("-%s", count),
		"1",
		addr); err != nil {
		return err
	}

	if err := filesystem.CopyFile("hosts.ini", "hosts.test.ini"); err != nil {
		return err
	}

	connectionParameters := fmt.Sprintf(`
[all:vars]
ansible_host=%s
ansible_port=22
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o CheckHostIP=no'
ansible_ssh_private_key_file=~/.ssh/insecure_private_key
ansible_user=vagrant`, addr)

	if err := filesystem.AppendFile("hosts.test.ini", []byte(connectionParameters)); err != nil {
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
