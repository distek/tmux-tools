package lib

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func runCmd(cmd string) (string, string, error) {
	c := exec.Command("sh", "-c", cmd)

	outBuf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	c.Stdout = outBuf
	c.Stderr = errBuf

	err := c.Run()

	return strings.TrimSuffix(outBuf.String(), "\n"), strings.TrimSuffix(errBuf.String(), "\n"), err
}

func formatArgs(args map[string]string) string {
	var bld strings.Builder
	bld.WriteString(" ")
	for k, v := range args {
		bld.WriteString(k)
		bld.WriteString(" ")
		if v != "" {
			bld.WriteString(v)
			bld.WriteString(" ")
		}
	}

	bld.WriteString(" ")

	return bld.String()
}

func Tmux(args map[string]string, cmd string, cmdArgs map[string]string, trailingCmd string) (string, string, error) {
	var argsStr string
	var cmdArgsStr string

	if len(args) > 0 {
		argsStr = formatArgs(args)
	}

	if len(cmdArgs) > 0 {
		cmdArgsStr = formatArgs(cmdArgs)
	}

	return runCmd("tmux " + argsStr + cmd + " " + cmdArgsStr + " " + trailingCmd)
}

// KillServer kills the server at sock or the current server if sock is
// an empty string
func KillServer(sock string) {
	tempSock, ok := GlobalArgs["-S"]
	if ok {
		if sock != "" {
			GlobalArgs["-S"] = sock
		}
	}

	_, _, err := Tmux(GlobalArgs, "kill-server", map[string]string{}, "")

	if err != nil {
		log.Println(err)
	}

	if ok {
		GlobalArgs["-S"] = tempSock
	}
}

// SockHasAttached returns true if a client is attached to the sock or the
// current socket if sock is an empty string
func SockHasAttached(sock string) bool {
	tempSock, ok := GlobalArgs["-S"]
	if ok {
		if sock != "" {
			GlobalArgs["-S"] = sock
		}
	}

	o, _, err := Tmux(GlobalArgs, "ls", map[string]string{
		"-F": "\"#{session_attached}\"",
	}, "")

	if ok {
		GlobalArgs["-S"] = tempSock
	}

	if err != nil {
		log.Println(err)
		return false
	}

	if strings.Contains(string(o), "1") {
		return true
	}

	return false
}

// CloseOnLastDetatch checks the socket `sock` every `n` milliseconds and kills
// the server if there are no more attached clients.
//
// `sock` can be a tmux socket or empty string if the socket provided at
// startup is the expected target server
func CloseOnLastDetatch(sock string, n int) {
	for {
		time.Sleep(time.Millisecond * time.Duration(n))

		log.Println(SockHasAttached(sock))
		if !SockHasAttached(sock) {
			KillServer(sock)
			return
		}
	}
}

// Get command line + args (hopefully (stares at darwin)) from pid
func GetProcCmd(pid int) (string, error) {
	var out []byte
	var err error

	switch runtime.GOOS {
	case "linux":
		out, err = exec.Command("ps", "--no-headers", "-o", "command", "--ppid", fmt.Sprint(pid)).CombinedOutput()
		if err != nil {
			return "", err
		}
	case "darwin":
		// TODO: Darwin needs testing
		// It wouldn't if they could just use modern tools
		// Like, you'll innovate on silcon design, making the fastest consumer-avaialble ARM chip
		// But the bash binary you ship is from 2007.
		out, err = exec.Command("pgrep", "-P", fmt.Sprint(pid)).CombinedOutput()
		if err != nil {
			if err.Error() == "exit status 1" {
				// try ps
				out, err = exec.Command("ps", "-o", "command", "-p", fmt.Sprint(pid)).CombinedOutput()
				if err != nil {
					return "", err
				}
			} else {
				return "", err
			}
		}
	}

	return strings.TrimSuffix(string(out), "\n"), nil
}
