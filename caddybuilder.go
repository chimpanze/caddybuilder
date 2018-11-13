// Copyright (c) 2017 Fabio Cagliero
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var pwd, buildDir, binDir, caddySourceDir string
var goos, goarch string
var plugins pluginsArray
var dev bool
var disableTelemetry bool

// var avaiblePlugins bool = false

func init() {
	flag.StringVar(&goos, "goos", "", "OS for which to build")
	flag.StringVar(&goarch, "goarch", "", "ARCH for which to build")
	flag.Var(&plugins, "plugin", "Plugin to integrate in the build")
	flag.BoolVar(&dev, "dev", false, "Build the current master branch")
	flag.BoolVar(&disableTelemetry, "disable-telemetry", false, "Disable built-in telemetry")
	// TODO
	//flag.BoolVar(&avaiblePlugins, "listplugins", false, "Display all the available plugins")
}

func main() {
	flag.Parse()

	var err error
	// Check if git is installed
	cmd := exec.Command("git", "--version")
	err = cmd.Run()
	check(err)

	// Getting current work directory
	pwd, err = os.Getwd()
	check(err)

	buildDir = pwd + "/build"
	binDir = pwd + "/bin"
	caddySourceDir = buildDir + "/src/github.com/mholt/caddy"

	os.Mkdir(buildDir, 0755)
	os.Mkdir(binDir, 0755)

	os.Setenv("GOPATH", buildDir)

	fmt.Println("Downloading caddy source code...")
	cmd = exec.Command("go", "get", "github.com/mholt/caddy/caddy")
	err = cmd.Run()
	check(err)

	cmd = exec.Command("go", "get", "github.com/caddyserver/builds")
	err = cmd.Run()
	check(err)

	// Git checkout to last tagged version
	// Skip for building the current master branch
	if !dev {
		cmd = exec.Command("git", "describe", "--abbrev=0", "--tags")
		cmd.Dir = caddySourceDir
		tag, err := cmd.Output()
		check(err)

		caddyVersion := strings.TrimSpace(string(tag))

		cmd = exec.Command("git", "checkout", caddyVersion)
		cmd.Dir = caddySourceDir
		err = cmd.Run()
		check(err)

		fmt.Println("Tag to build: ", caddyVersion)
	} else {
		fmt.Println("Branch to build: master")
	}

	pluginRepos := caddyAvailablePlugins()

	var selectedPlugins []string

	for _, plugin := range plugins {
		if _, found := pluginRepos[plugin]; !found {
			// TODO
			// fmt.Printf("Plugin %s not found. Run with option -listplugins to see available plugins.\n", plugin)
			fmt.Printf("Plugin %s not found.\n", plugin)
			os.Exit(0)
		}

		selectedPlugins = append(selectedPlugins, pluginRepos[plugin])
	}

	if len(selectedPlugins) > 0 {
		addPlugins(selectedPlugins)
	}

	if disableTelemetry {
		caddyDisableTelemetry()
		fmt.Println("Telemetry disabled")
	}

	fmt.Println("Building...")

	cmd = exec.Command("go", "run", "build.go", "-goos", goos, "-goarch", goarch)
	cmd.Dir = caddySourceDir + "/caddy"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err = cmd.Run()
	check(err)

	err = os.Rename(caddySourceDir+"/caddy/caddy", binDir+"/caddy")

	fmt.Println("Removing build dir...")
	os.RemoveAll(buildDir)

	fmt.Println("Done! Your caddy executable is in ", binDir)
}

// Define plugin type
type pluginsArray []string

func (p *pluginsArray) String() string {
	return fmt.Sprintf("%d", *p)
}

func (p *pluginsArray) Set(plugin string) error {
	*p = append(*p, plugin)
	return nil
}

// Other functions
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func caddyAvailablePlugins() map[string]string {
	file, err := os.Open(caddySourceDir + "/caddyhttp/httpserver/plugin.go")
	check(err)

	var varDirectives bool = false
	scanner := bufio.NewScanner(file)

	pluginRepos := make(map[string]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "// The ordering of this list is important." {
			varDirectives = true
		}

		if varDirectives && line == "}" {
			varDirectives = false
			break
		}

		if varDirectives {
			pluginRegExp := regexp.MustCompile(`^\"([a-zA-Z_-]+)\",\s+\/\/\s([a-zA-Z_\-.\/]+)$`)
			check(err)

			if pluginRegExp.MatchString(line) {
				subMatches := (pluginRegExp.FindStringSubmatch(line))[1:]
				pluginRepos[subMatches[0]] = subMatches[1]
			}
		}
	}

	return pluginRepos
}

func addPlugins(selectedPlugins []string) {
	for k, plugin := range selectedPlugins {
		fmt.Printf("Downloading %s plugin source code...\n", plugin)
		cmd := exec.Command("go", "get", plugin)
		err := cmd.Run()
		check(err)

		selectedPlugins[k] = fmt.Sprintf("\t_ \"%s\"", plugin)
	}

	fileRunGo, err := ioutil.ReadFile(caddySourceDir + "/caddy/caddymain/run.go")
	check(err)

	lines := strings.Split(string(fileRunGo), "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "// This is where other plugins get plugged in (imported)" {
			lines = append(lines, selectedPlugins...)
			copy(lines[i+1+len(selectedPlugins):], lines[i+1:])
			copy(lines[i+1:], selectedPlugins)
			break
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(caddySourceDir+"/caddy/caddymain/run.go", []byte(output), 0644)
	check(err)
}

func caddyDisableTelemetry() {
	fileRunGo, err := ioutil.ReadFile(caddySourceDir + "/caddy/caddymain/run.go")
	check(err)

	// Until Caddy v0.11.0
	output := strings.Replace(string(fileRunGo), "const enableTelemetry = true", "const enableTelemetry = false", 1)

	// From Caddy v0.11.1
	output = strings.Replace(string(fileRunGo), "var EnableTelemetry = true", "var EnableTelemetry = false", 1)

	err = ioutil.WriteFile(caddySourceDir+"/caddy/caddymain/run.go", []byte(output), 0644)
	check(err)
}
