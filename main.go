package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

func main() {
	switch len(os.Args) {
	case 1:
		fmt.Println(color.RedString("Please provide host name to test or action (ssh or destroy)"))
		fmt.Println(color.YellowString("To skip recreating, add 'norecreate' after hostname"))
		os.Exit(1)
	case 2:
		if strings.EqualFold(os.Args[1], "ssh") {
			run("vagrant", "ssh")
			os.Exit(0)
		}

		if strings.EqualFold(os.Args[1], "destroy") {
			run("vagrant", "destroy", "--force")
			os.Exit(0)
		}

		run("vagrant", "destroy", "--force")
		run("vagrant", "up")

	case 3:
		if strings.EqualFold(os.Args[2], "norecreate") {
			run("vagrant", "up")
		} else {
			fmt.Println(color.RedString("use 'norecreate' to skip recreating VM"))
			os.Exit(2)
		}
	}

	provision(os.Args[1])
	os.Exit(0)
}

func provision(hostname string) {
	count := "c"

	if runtime.GOOS == "windows" {
		count = "n"
	}

	err := run("ping", fmt.Sprintf("-%s", count), "1", "192.168.57.42")

	if err == nil {
		input, err := os.ReadFile("hosts.ini")
		if err != nil {
			fmt.Println(err)
			return
		}

		err = os.WriteFile("hosts.test.ini", input, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

		override := []byte(`
[all:vars]
ansible_host=192.168.57.42
ansible_port=22
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o CheckHostIP=no'
ansible_ssh_private_key_file=~/.ssh/insecure_private_key
ansible_user=vagrant`)

		file, err := os.OpenFile("hosts.test.ini", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer file.Close()

		if _, err = file.Write(override); err != nil {
			fmt.Println(err)
			return
		}

		run("ansible-host", "provision", "--verbose", "--inventory", "./hosts.test.ini", hostname)

		os.Remove("hosts.test.ini")
	}
}

func run(binary string, params ...string) error {
	cmd := exec.Command(binary, params...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
