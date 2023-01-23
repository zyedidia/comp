package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fatal(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

// Returns true if 'path' exists and is not a directory.
func existsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func execCmd(prog string, args ...string) []byte {
	cmd := exec.Command(prog, args...)
	log.Println(cmd)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fatal(err)
	}
	return buf.Bytes()
}

type CompCommand struct {
	Directory string `json:"directory"`
	File      string `json:"file"`
	Command   string `json:"command"`
}

func FindCompDb() (string, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	path := wd
	name := "compile_commands.json"
	for filepath.Dir(path) != path {
		if existsFile(filepath.Join(path, name)) {
			p, e := filepath.Rel(wd, path)
			return name, p, e
		}
		path = filepath.Dir(path)
	}
	return "", "", errors.New("could not find compile_commands.json")
}

func main() {
	debug := flag.Bool("debug", false, "show debug information")
	knit := flag.Bool("knit", false, "run knit to generate the compdb")

	flag.Parse()
	args := flag.Args()

	if *debug {
		log.SetOutput(os.Stdout)
		log.SetFlags(0)
		log.SetPrefix("[debug] ")
	} else {
		log.SetOutput(io.Discard)
	}

	if len(args) <= 0 {
		fatal("no file to build")
	}

	var data []byte
	var path string
	if *knit {
		path = filepath.Dir(string(execCmd("knit", "-t", "path")))
		data = execCmd("knit", "-t", "compdb", "all")
	} else {
		var name string
		var err error
		name, path, err = FindCompDb()
		if err != nil {
			fatal(err)
		}
		compdb := filepath.Join(path, name)

		log.Println("compdb:", compdb)

		data, err = os.ReadFile(compdb)
		if err != nil {
			fatal(err)
		}
	}

	var compcmds []CompCommand
	err := json.Unmarshal(data, &compcmds)
	if err != nil {
		fatal(err)
	}

	wd, _ := os.Getwd()
	target, _ := filepath.Rel(filepath.Join(wd, path), filepath.Join(wd, args[0]))
	curdir, _ := filepath.Rel(filepath.Join(wd, path), wd)

	for _, c := range compcmds {
		file := filepath.Join(wd, path, c.Directory, c.File)
		frel, _ := filepath.Rel(filepath.Join(wd, path), file)
		if target == frel {
			log.Println(c.Command)
			buf := &bytes.Buffer{}
			cmd := exec.Command("sh", "-c", c.Command)
			cmd.Dir = filepath.Join(wd, path, c.Directory)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = buf
			cmd.Run()
			fmt.Print(strings.ReplaceAll(buf.String(), curdir+"/", ""))
		}
	}
}
