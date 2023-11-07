package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aksiksi/compose2nix"
)

var paths = flag.String("paths", "", "paths to Compose files")
var envFiles = flag.String("env_files", "", "paths to .env files")
var envFilesOnly = flag.Bool("env_files_only", false, "only use env files in the NixOS container definitions")
var output = flag.String("output", "", "path to output Nix file")
var project = flag.String("project", "", "project name used as a prefix for generated resources")
var projectSeparator = flag.String("project_separator", compose2nix.DefaultProjectSeparator, "seperator for project prefix")
var serviceInclude = flag.String("service_include", "", "regex pattern for services to include")
var autoStart = flag.Bool("auto_start", true, "control auto-start setting for containers")
var runtime = flag.String("runtime", "podman", `"podman" or "docker"`)
var useComposeLogDriver = flag.Bool("use_compose_log_driver", false, "if set, always use the Compose log driver.")

func main() {
	flag.Parse()

	ctx := context.Background()

	if *paths == "" {
		log.Fatalf("One or more paths must be specified")
	}

	paths := strings.Split(*paths, ",")
	envFiles := strings.Split(*envFiles, ",")

	var containerRuntime compose2nix.ContainerRuntime
	if *runtime == "podman" {
		containerRuntime = compose2nix.ContainerRuntimePodman
	} else if *runtime == "docker" {
		containerRuntime = compose2nix.ContainerRuntimeDocker
	} else {
		log.Fatalf("Invalid --runtime: %q", *runtime)
	}

	var serviceIncludeRegexp *regexp.Regexp
	if *serviceInclude != "" {
		pat, err := regexp.Compile(*serviceInclude)
		if err != nil {
			log.Fatalf("Failed to parse -service_include pattern %q: %v", *serviceInclude, err)
		}
		serviceIncludeRegexp = pat
	}

	start := time.Now()
	g := compose2nix.Generator{
		Project:             compose2nix.NewProjectWithSeparator(*project, *projectSeparator),
		Runtime:             containerRuntime,
		Paths:               paths,
		EnvFiles:            envFiles,
		EnvFilesOnly:        *envFilesOnly,
		ServiceInclude:      serviceIncludeRegexp,
		AutoStart:           *autoStart,
		UseComposeLogDriver: *useComposeLogDriver,
	}
	containerConfig, err := g.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated NixOS config in %v\n", time.Since(start))

	if *output != "" {
		dir := path.Dir(*output)
		if _, err := os.Stat(dir); err != nil {
			log.Fatalf("Directory %q does not exist", dir)
		}
		if err := os.WriteFile(*output, []byte(containerConfig.String()), 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Wrote NixOS config to %s\n", *output)
	}
}