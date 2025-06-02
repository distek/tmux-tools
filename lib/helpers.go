package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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
	if sock != "" {
		GlobalArgs["-S"] = sock
	}
	_, _, err := Tmux(GlobalArgs, "kill-server", map[string]string{}, "")

	if ok {
		GlobalArgs["-S"] = tempSock
	}

	if err != nil {
		log.Println(err)
	}
}

// Check if sock exists
func SockExists(sock string) bool {
	_, err := os.Stat(sock)
	return err != nil && errors.Is(err, os.ErrNotExist)
}

// SockHasAttached returns true if a client is attached to the sock or the
// current socket if sock is an empty string
func SockHasAttached(sock string) bool {
	o, _, err := Tmux(map[string]string{"-S": sock}, "ls", map[string]string{
		"-F": "\"#{session_attached}\"",
	}, "")
	if err != nil {
		log.Println(err)
		return false
	}

	if strings.Contains(string(o), "1") {
		return true
	}

	return false
}

func SockActive(sock string) bool {
	_, e, _ := Tmux(map[string]string{"-S": sock}, "ls", nil, "")

	return !strings.HasPrefix(e, "no server running on")
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
		out, err = exec.Command("ps", "--no-headers", "-o", "command", "--ppid", fmt.Sprintf("%d", pid)).CombinedOutput()
		if err != nil {
			return "", err
		}
	case "darwin":
		outPre, err := exec.Command("ps", "-o", "ppid=,command=").CombinedOutput()
		if err != nil {
			return "", err
		}

		for v := range strings.SplitSeq(string(outPre), "\n") {
			split := strings.SplitN(strings.TrimSpace(v), " ", 2)
			if len(split) != 2 {
				continue
			}

			if split[1] == "" {
				continue
			}

			ppid, err := strconv.Atoi(split[0])
			if err != nil {
				fmt.Println(err)
				continue
			}

			if ppid == pid {
				out = []byte(split[1])
				break
			}
		}

	}

	return strings.TrimSuffix(string(out), "\n"), nil
}

func Fzf(list []string) (string, error) {
	data := bytes.NewBuffer([]byte(strings.Join(list, "\n")))

	var result strings.Builder
	cmd := exec.Command("fzf")
	cmd.Stdout = &result
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(stdin, data)
	if err != nil {
		return "", err
	}
	err = stdin.Close()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		// No selection made - don't error
		if cmd.ProcessState.ExitCode() == 130 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(result.String()), nil
}

func DirUp(path string) (string, error) {
	oldDir := os.Getenv("PWD")

	f, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !f.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", path)
	}

	err = os.Chdir(fmt.Sprintf("%s/..", path))
	if err != nil {
		return "", err
	}

	os.Chdir(oldDir)

	return filepath.Dir(path), nil
}

func IsGitWorktree(path string) bool {
	worktreesPath := filepath.Join(path, "worktrees")

	info, err := os.Stat(worktreesPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}
