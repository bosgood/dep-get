package fetch

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bosgood/dep-get/command"
	"github.com/bosgood/dep-get/lib/fs"
	"github.com/bosgood/dep-get/nodejs"
	"github.com/mitchellh/cli"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var realOS fs.FileSystem = &fs.OSFS{}

type fetchCommand struct {
	npmShrinkwrap nodejs.NPMShrinkwrap
	os            fs.FileSystem
	config        fetchCommandFlags
}

type fetchCommandFlags struct {
	command.BaseFlags
	platform     string
	source       string
	destination  string
	whitelistStr string
	whitelist    *regexp.Regexp
}

func newFetchCommandWithFS(os fs.FileSystem) (cli.Command, error) {
	cmd := &fetchCommand{
		os: os,
	}
	return cmd, nil
}

// NewFetchCommand is used to generate a command object
// which orchestrates package dependency discovery
// and archiving to S3
func NewFetchCommand() (cli.Command, error) {
	return newFetchCommandWithFS(realOS)
}

func (c *fetchCommand) Synopsis() string {
	return "Archives application dependencies"
}

func (c *fetchCommand) Help() string {
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (fetchCommandFlags, *flag.FlagSet, error) {
	var cmdConfig fetchCommandFlags

	cmdFlags := flag.NewFlagSet("fetch", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false,
		"show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")
	cmdFlags.StringVar(&cmdConfig.destination, "destination", "", "dependencies download destination")
	cmdFlags.StringVar(&cmdConfig.whitelistStr, "whitelist", "", "dependency name whitelist regexp")

	if err := cmdFlags.Parse(args); err != nil {
		errMsg := fmt.Sprintf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Error parsing args",
			err,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	if cmdConfig.Help {
		return cmdConfig, cmdFlags, &command.ConfigError{}
	}

	// All missing required arguments checks go here
	var missingArg string
	if cmdConfig.platform == "" {
		missingArg = "platform"
	} else if cmdConfig.destination == "" {
		missingArg = "destination"
	}

	if missingArg != "" {
		errMsg := fmt.Sprintf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Missing required argument",
			missingArg,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	if cmdConfig.platform != "nodejs" {
		errMsg := fmt.Sprintf(
			"%s%s\n",
			command.LogErrorPrefix,
			"Only nodejs supported at the moment",
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	// Additional command parsing goes here
	if cmdConfig.whitelistStr != "" {
		rgx, err := regexp.Compile(cmdConfig.whitelistStr)
		if err != nil {
			errMsg := fmt.Sprintf(
				"%sMalformed dependency whitelist regexp: %s\n",
				command.LogErrorPrefix,
				cmdConfig.whitelistStr,
			)
			return cmdConfig, cmdFlags, &command.ConfigError{
				Explanation: errMsg,
			}
		}
		cmdConfig.whitelist = rgx
	}

	return cmdConfig, cmdFlags, nil
}

func (c *fetchCommand) readDependencies(dirPath string) (nodejs.NPMShrinkwrap, error) {
	var npmShrinkwrap nodejs.NPMShrinkwrap

	packageFilePath := path.Join(dirPath, nodejs.DependenciesFileName)
	packageFileContents, err := c.os.ReadFile(packageFilePath)
	if err != nil {
		fmt.Printf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Can't open the dependencies file",
			err,
		)
		return npmShrinkwrap, err
	}

	err = json.Unmarshal(packageFileContents, &npmShrinkwrap)
	if err != nil {
		fmt.Printf(
			"%s%s: %s\n",
			command.LogErrorPrefix,
			"Failed to decode the dependencies file file",
			err,
		)
		return npmShrinkwrap, err
	}

	return npmShrinkwrap, nil
}

var repoPattern = regexp.MustCompile(`^/.+/.+\.git$`)

func (c *fetchCommand) resolveDependencyURL(depURL string) (string, error) {
	urlObj, err := url.Parse(depURL)
	if err != nil {
		return depURL, nil
	}

	// Plain old HTTP(s) URLs
	scheme := urlObj.Scheme
	if scheme == "http" || scheme == "https" {
		return depURL, nil
	}

	// git URLs whitelisted by site
	if scheme == "git" {
		if !repoPattern.MatchString(urlObj.Path) {
			return depURL, fmt.Errorf("Unknown git URL path format: %s", urlObj.Path)
		}
		repoParts := strings.Split(strings.Split(urlObj.Path, ".")[0], "/")[1:]
		owner := repoParts[0]
		repo := repoParts[1]

		commit := urlObj.Fragment
		if commit == "" {
			commit = "master"
		}

		if urlObj.Host == "github.com" {
			httpURL := fmt.Sprintf(
				"https://github.com/%s/%s/archive/%s.tgz",
				owner,
				repo,
				commit,
			)
			return httpURL, nil
		} else if urlObj.Host == "bitbucket.com" {
			httpURL := fmt.Sprintf(
				"https://bitbucket.org/%s/%s/get/%s.tgz",
				owner,
				repo,
				commit,
			)
			return httpURL, nil
		}
	}

	return depURL, fmt.Errorf("Unknown URL scheme: %s", urlObj.Scheme)
}

func (c *fetchCommand) fetchDependency(dep nodejs.NodeDependency) (string, error) {
	depURL, err := c.resolveDependencyURL(dep.PackageURL)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(depURL)
	if err != nil {
		return "", err
	}
	defer func() {
		if rerr := resp.Body.Close(); rerr != nil && err == nil {
			err = rerr
		}
	}()

	outFilePath := path.Join(c.config.destination, dep.GetCanonicalName()+".tgz")
	outFile, err := c.os.Create(outFilePath)
	defer func() {
		if ferr := outFile.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	if err != nil {
		return "", err
	}
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	return outFilePath, err
}

func (c *fetchCommand) fetchDependencies(deps []nodejs.NodeDependency) ([]string, error) {
	numDeps := len(deps)

	var outFilePaths []string
	for i, dep := range deps {
		fmt.Printf(
			"%s(%d/%d) Downloading %s (%s)\n",
			command.LogInfoPrefix,
			i, numDeps,
			dep.GetCanonicalName(),
			dep.PackageURL,
		)
		outFilePath, err := c.fetchDependency(dep)
		if err != nil {
			return outFilePaths, err
		}
		outFilePaths = append(outFilePaths, outFilePath)
	}

	return outFilePaths, nil
}

func (c *fetchCommand) Run(args []string) int {
	cmdConfig, _, err := getConfig(args)
	if err != nil {
		errMsg := err.Error()
		if errMsg != "" {
			fmt.Print(err.Error())
		}
		return cli.RunResultHelp
	}

	c.config = cmdConfig

	var dirPath string
	if cmdConfig.source == "" {
		cwd, err := c.os.Getwd()
		if err != nil {
			fmt.Printf(
				"%s%s: %s\n",
				command.LogErrorPrefix,
				"Can't read current directory",
				err,
			)
			return 1
		}
		dirPath = cwd
	} else {
		dirPath = cmdConfig.source
	}

	npmShrinkwrap, err := c.readDependencies(dirPath)
	if err != nil {
		return 1
	}

	c.npmShrinkwrap = npmShrinkwrap

	allDeps := nodejs.CollectDependencies(npmShrinkwrap)
	var deps []nodejs.NodeDependency

	// Filter deps according to whitelist if present
	if cmdConfig.whitelistStr == "" {
		deps = allDeps
	} else {
		for _, dep := range allDeps {
			if cmdConfig.whitelist.MatchString(dep.Name) {
				deps = append(deps, dep)
			}
		}
	}

	fmt.Printf(
		"%sFound %d matching dependencies.\n",
		command.LogSuccessPrefix,
		len(deps),
	)

	fetchedDeps, err := c.fetchDependencies(deps)
	if err != nil {
		fmt.Printf(
			"%s%s: %s",
			command.LogErrorPrefix,
			"Error fetching dependencies",
			err,
		)
		return 1
	}

	for _, dep := range fetchedDeps {
		fmt.Printf(
			"%sFetched: %s\n",
			command.LogInfoPrefix,
			dep,
		)
	}

	fmt.Printf(
		"%sFetched %d dependencies.\n",
		command.LogSuccessPrefix,
		len(deps),
	)

	return 0
}
