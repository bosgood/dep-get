package retrieve

import (
	"bitbucket.org/bosgood/dep-get/command"
	"bitbucket.org/bosgood/dep-get/lib/fs"
	"bitbucket.org/bosgood/dep-get/nodejs"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/cli"
	"net/url"
	"path"
	"regexp"
)

type retrieveCommand struct {
	os     fs.FileSystem
	config retrieveCommandFlags
	s3     *s3.S3
}

type retrieveCommandFlags struct {
	command.BaseFlags
	platform     string
	source       string
	destination  string
	region       string
	profile      string
	s3URL        string
	bucket       string
	s3Key        string
	whitelistStr string
	whitelist    *regexp.Regexp
}

var (
	realOS fs.FileSystem = &fs.OSFS{}
)

func newRetrieveCommandWithFS(os fs.FileSystem) (cli.Command, error) {
	cmd := &retrieveCommand{
		os: os,
	}
	return cmd, nil
}

// NewRetrieveCommand is used to generate a command object
// which downloads dependencies from S3
func NewRetrieveCommand() (cli.Command, error) {
	return newRetrieveCommandWithFS(realOS)
}

func (c *retrieveCommand) Synopsis() string {
	return "Retrieves archived application dependencies from S3"
}

func (c *retrieveCommand) Help() string {
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (retrieveCommandFlags, *flag.FlagSet, error) {
	var cmdConfig retrieveCommandFlags

	cmdFlags := flag.NewFlagSet("retrieve", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false, "show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")
	cmdFlags.StringVar(&cmdConfig.destination, "destination", "", "dependency download directory (default: .)")
	cmdFlags.StringVar(&cmdConfig.profile, "profile", "", "AWS credentials profile (default: default)")
	cmdFlags.StringVar(&cmdConfig.region, "region", "", "AWS region")
	cmdFlags.StringVar(&cmdConfig.s3URL, "path", "", "S3 storage path")
	cmdFlags.StringVar(&cmdConfig.whitelistStr, "whitelist", "", "dependency name whitelist regexp")

	if err := cmdFlags.Parse(args); err != nil {
		errMsg := fmt.Sprintf(
			"%sError parsing args: %s\n",
			command.LogErrorPrefix,
			err,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	if cmdConfig.Help {
		return cmdConfig, cmdFlags, &command.ConfigError{}
	}

	// All missing required argument checks go here
	var missingArg string
	if cmdConfig.platform == "" {
		missingArg = "platform"
	}

	if cmdConfig.region == "" {
		missingArg = "region"
	}

	if cmdConfig.s3URL == "" {
		missingArg = "path"
	}

	if missingArg != "" {
		errMsg := fmt.Sprintf(
			"%sMissing required argument: %s\n",
			command.LogErrorPrefix,
			missingArg,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	if cmdConfig.platform != "nodejs" {
		errMsg := fmt.Sprintf(
			"%sOnly nodejs supported at the moment\n",
			command.LogErrorPrefix,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

	// Parameter validation goes here
	s3URL, err := url.Parse(cmdConfig.s3URL)
	if err != nil {
		errMsg := fmt.Sprintf(
			"%sInvalid S3 path: %s\n",
			command.LogErrorPrefix,
			cmdConfig.s3URL,
		)
		return cmdConfig, cmdFlags, &command.ConfigError{
			Explanation: errMsg,
		}
	}

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

	cmdConfig.bucket = s3URL.Host
	cmdConfig.s3Key = s3URL.Path

	return cmdConfig, cmdFlags, nil
}

func (c *retrieveCommand) InitS3() error {
	cfg := aws.NewConfig().
		WithRegion(c.config.region)

	if c.config.profile != "" {
		creds := credentials.NewSharedCredentials("", c.config.profile)
		cfg = cfg.WithCredentials(creds)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	c.s3 = s3.New(sess)
	return nil
}

func (c *retrieveCommand) Download(archiveFileName string) (*s3.GetObjectOutput, error) {
	s3Path := path.Join(c.config.s3Key, archiveFileName)
	fmt.Printf(
		"%sDownloading: s3://%s%s\n",
		command.LogInfoPrefix,
		c.config.bucket,
		s3Path,
	)
	out, err := c.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.config.bucket),
		Key:    aws.String(s3Path),
	})

	return out, err
}

func (c *retrieveCommand) Run(args []string) int {
	cmdConfig, _, err := getConfig(args)
	if err != nil {
		errMsg := err.Error()
		if errMsg != "" {
			fmt.Print(err.Error())
		}
		return cli.RunResultHelp
	}
	c.config = cmdConfig

	npmShrinkwrap, err := nodejs.ReadDependencies(c.os, c.config.source)
	if err != nil {
		return 1
	}

	allDeps := nodejs.CollectDependencies(npmShrinkwrap)
	var deps []nodejs.NodeDependency

	// Filter deps according to whitelist if present
	if c.config.whitelistStr == "" {
		deps = allDeps
	} else {
		for _, dep := range allDeps {
			if c.config.whitelist.MatchString(dep.Name) {
				deps = append(deps, dep)
			}
		}
	}

	fmt.Printf(
		"%sWill download %d matching dependencies.\n",
		command.LogInfoPrefix,
		len(deps),
	)

	err = c.InitS3()
	if err != nil {
		fmt.Printf(
			"%sFailed to initialize AWS/S3 session: %s\n",
			command.LogErrorPrefix,
			err,
		)
		return 1
	}

	for _, depInfo := range deps {
		_, err := c.Download(depInfo.GetCanonicalName()+".tgz")
		if err != nil {
			fmt.Printf(
				"%sError downloading dependency %s: %s\n",
				command.LogErrorPrefix,
				depInfo.GetCanonicalName(),
				err,
			)
			return 1
		}
		fmt.Printf(
			"%sDownloaded dependency: %s\n",
			command.LogSuccessPrefix,
			depInfo.GetCanonicalName(),
		)
	}

	return 0
}
