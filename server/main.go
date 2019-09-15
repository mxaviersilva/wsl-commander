package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NixUser describes a *nix user most basic properties
type NixUser struct {
	Username string `json:"userName"`
	HomeDir  string `json:"homeDir"`
	Shell    string `json:"loginShell"`
}

// WSLDistro describes a WSL distro and its state and version
type WSLDistro struct {
	Default bool   `json:"isDefault"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Version int    `json:"version"`
}

func main() {
	// test if wsl is present
	exists, path := CheckCmd("wsl")
	if !exists {
		println("wsl is not installed")
		os.Exit(1)
	}
	path = filepath.ToSlash(path)

	// compiles a list of the distros registered with wsl
	_, out, err := RunCmd(filepath.Base(path), []string{"-l", "-v"})
	if err != nil {
		panic(err)
	}
	outStr := string(cleanReturn(out[0]))
	ParseWSLDistroTable(outStr)

	/*
		// print a table with the found distros
		distroTable := tablewriter.NewWriter(os.Stdout)
		distroTable.SetHeader([]string{"name"})
		for _, distro := range distros {
			distroTable.Append([]string{distro})
		}
		distroTable.Render()

		// prints a table with the user for each found distro
		for _, distro := range distros {
			// gets a list opf the users in the distro for details see the function
			users := GetUsers(string(distro), filepath.Base(path))
			// creates a new table writer from the tablewriter library
			tablew := tablewriter.NewWriter(os.Stdout)
			tablew.SetHeader([]string{"username", "homedir", "login shell"})
			tablew.SetFooter([]string{distro, "", fmt.Sprintf("%d", len(users))})
			for _, user := range users {
				tablew.Append([]string{user.Username, user.HomeDir, user.Shell})
			}
			tablew.Render()
		}*/
}

// CheckCmd checks if a given command is in the system's path
func CheckCmd(cmd string) (bool, string) {
	path, err := exec.LookPath("wsl")
	return err == nil, path
}

// RunCmd Execustes a given command and returns the stdou and stderr along with the return code and any erros
func RunCmd(inCmd string, args []string) (int, [2][]byte, error) {
	cmd := exec.Command(inCmd, args...)
	// creates two buffers to hold the commands outputs
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// in case of error pass everything ahead with a return code of -1 and pass the error
		return -1, [2][]byte{stdout.Bytes(), stderr.Bytes()}, err
	}
	return cmd.ProcessState.ExitCode(), [2][]byte{stdout.Bytes(), stderr.Bytes()}, nil
}

// cleanReturn fixes the return of windows commands by removing the needless whitespaces and making the output LF line feeds
func cleanReturn(origin []byte) []byte {
	var output []byte
	for _, b := range origin {
		if b != 0x00 && b != 0x0D {
			output = append(output, b)
		}
	}
	return output
}

// GetUsers this function returns a slice of NixUsers from running cat /etc/passwd and treating the output
func GetUsers(distro string, path string) []NixUser {
	// calls RunCmd to get the list of users as the root user
	_, out, _ := RunCmd(path, []string{"-d", distro, "--user", "root", "-e", "cat", "/etc/passwd"})
	treated := cleanReturn(out[0])
	usersLines := strings.Fields(string(treated))
	var users []NixUser
	for _, userLine := range usersLines {
		split := strings.Split(userLine, ":")
		if len(split) > 5 {
			nixu := NixUser{
				Username: split[0],
				HomeDir:  split[len(split)-2],
				Shell:    split[len(split)-1],
			}
			users = append(users, nixu)
		}
	}
	return users
}

func ParseWSLDistroTable(table string) {
	lines := strings.Split(table, "\n")
	for i, l := range lines {
		if i == 0 {
			continue
		} else if i != 0 {
			lineContents := strings.Split(l, " ")
			for _, lc := range lineContents {
				fmt.Printf(lc)
			}
		}
	}
}
