package vcs

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func getRef(dir string) (string, error) {
	cmd := exec.Command(
		"git",
		"branch",
		"--show-current")
	cmd.Dir = dir
	r, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer r.Close()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), cmd.Wait()
}

func init() {
	Register(newGit, "git")
}

type GitDriver struct{}

func newGit(b []byte) (Driver, error) {
	return &GitDriver{}, nil
}

func (g *GitDriver) HeadRev(dir string) (string, error) {
	cmd := exec.Command(
		"git",
		"rev-parse",
		"HEAD")
	cmd.Dir = dir
	r, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer r.Close()

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), cmd.Wait()
}

func run(desc, dir, cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = dir

	//log.Printf("Executing %s", c)
	if out, err := c.CombinedOutput(); err != nil {
		log.Printf(
			"Failed to %s %s, see output below\n%sContinuing...",
			desc,
			dir,
			out)
		return err
	}
	return nil
}

func (g *GitDriver) Pull(dir string) (string, error) {
	//log.Printf("Pull %s", dir)

	branch, err := getRef(dir)
	if err != nil {
		log.Printf(
			"Failed to retrieve branch for %s",
			dir)
		return "", err
	}

	if err := run("git fetch", dir,
		"git",
		"fetch",
		"--prune",
		"--no-tags",
		"--depth", "1",
		"origin",
		fmt.Sprintf("+%s:remotes/origin/%s", branch, branch)); err != nil {
		return "", err
	}

	if err := run("git reset", dir,
		"git",
		"reset",
		"--hard",
		fmt.Sprintf("origin/%s", branch)); err != nil {
		return "", err
	}

	return g.HeadRev(dir)
}

func (g *GitDriver) Clone(dir, url string) (string, error) {
	par, rep := filepath.Split(dir)
	cmd := exec.Command(
		"git",
		"clone",
		"--depth", "1",
		url,
		rep)
	cmd.Dir = par
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to clone %s, see output below\n%sContinuing...", url, out)
		return "", err
	}

	return g.HeadRev(dir)
}

func (g *GitDriver) SpecialFiles() []string {
	return []string{
		".git",
	}
}
