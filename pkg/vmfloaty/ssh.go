package vmfloaty

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"log"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/communicator/ssh"
	"github.com/hashicorp/terraform/terraform"
	"github.com/puppetlabs/go-floaty/pkg/rand"
)

func Exec(host Host, key string, file *os.File) {

	script := bufio.NewReader(file)

	r := &terraform.InstanceState{
		Ephemeral: terraform.EphemeralState{
			ConnInfo: map[string]string{
				"type":        "ssh",
				"user":        "root",
				"private_key": key,
				"host":        host.Hostname,
				"port":        "22",
				"timeout":     "1s",
			},
		},
	}

	c, err := ssh.New(r)
	if err != nil {
		log.Fatalf(err.Error())
	}

	execScript(c, fmt.Sprintf("/tmp/%s", rand.String(16)), script, nil)
}

func ExecScriptFromBuffer(host Host, username string, key string, data []byte, parameters []string) error {
	script := bytes.NewReader(data)

	// TODO we should probably capture this output, so that it
	// can be printed in the event of a failure.
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stdout)

	r := &terraform.InstanceState{
		Ephemeral: terraform.EphemeralState{
			ConnInfo: map[string]string{
				"type":        "ssh",
				"user":        username,
				"private_key": key,
				"host":        fmt.Sprintf("%s.%s", host.Hostname, host.Domain),
				"port":        "22",
				"timeout":     "30s",
				"ssh_pty":     "false",
			},
		},
	}

	c, err := ssh.New(r)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return execScript(c, fmt.Sprintf("/tmp/%s", rand.String(16)), script, parameters)

}

func execScript(client communicator.Communicator, path string, input io.Reader, parameters []string) error {
	reader := bufio.NewReader(input)
	var script bytes.Buffer

	script.ReadFrom(reader)
	if err := client.Upload(path, &script); err != nil {
		return fmt.Errorf("Error occured while uploading setup script : %s", err)
	}

	var stdout, stderr bytes.Buffer
	var cmdString string
	if parameters != nil && len(parameters) > 0 {
		cmdString = fmt.Sprintf("chmod 0777 %s && %s %s; rm -f %s",
			path,
			path,
			strings.Trim(fmt.Sprint(parameters), "[]"),
			path,
		)
	} else {
		cmdString = fmt.Sprintf("chmod 0777 %s && %s; rm -f %s", path, path, path)
	}

	cmd := &remote.Cmd{
		Command: cmdString,
		Stdout:  &stdout,
		Stderr:  &stderr,
	}

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stdout)
	if err := client.Start(cmd); err != nil {
		return fmt.Errorf(
			"Error chmodding script file to 0777 in remote "+
				"machine: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf(
			"Error chmodding script file to 0777 in remote "+
				"machine %v: %s %s", err, stdout.String(), stderr.String())
	}

	//log.Print(string(stdout.String()))

	return nil
}
