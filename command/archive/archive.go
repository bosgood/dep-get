package archive

import (
	"bitbucket.org/bosgood/dep-get/command"
	"bitbucket.org/bosgood/dep-get/lib/fs"
	"flag"
	"fmt"
	"net/url"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/cli"
	"path"
)

type archiveCommand struct {
	os     fs.FileSystem
	config archiveCommandFlags
}

type archiveCommandFlags struct {
	command.BaseFlags
	platform string
	source   string
	region   string
	profile  string
	s3URL    string
	bucket   string
	s3Key    string
}

var (
	realOS       fs.FileSystem = &fs.OSFS{}
)

func newArchiveCommandWithFS(os fs.FileSystem) (cli.Command, error) {
	cmd := &archiveCommand{
		os: os,
	}
	return cmd, nil
}

// NewArchiveCommand is used to generate a command object
// which downloads dependencies from S3 and installs them
func NewArchiveCommand() (cli.Command, error) {
	return newArchiveCommandWithFS(realOS)
}

func (c *archiveCommand) Synopsis() string {
	return "Installs archived application dependencies"
}

func (c *archiveCommand) Help() string {
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (archiveCommandFlags, *flag.FlagSet, error) {
	var cmdConfig archiveCommandFlags

	cmdFlags := flag.NewFlagSet("install", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false, "show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")
	cmdFlags.StringVar(&cmdConfig.profile, "profile", "", "AWS credentials profile (default: default)")
	cmdFlags.StringVar(&cmdConfig.region, "region", "", "AWS region")
	cmdFlags.StringVar(&cmdConfig.s3URL, "path", "", "S3 upload path")

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

	cmdConfig.bucket = s3URL.Host
	cmdConfig.s3Key = s3URL.Path

	return cmdConfig, cmdFlags, nil
}

func (c *archiveCommand) Upload() error {
	return nil
}

func (c *archiveCommand) Run(args []string) int {
	cmdConfig, _, err := getConfig(args)
	if err != nil {
		errMsg := err.Error()
		if errMsg != "" {
			fmt.Print(err.Error())
		}
		return cli.RunResultHelp
	}

	archives, err := c.os.ReadDir(cmdConfig.source)
	if err != nil {
		fmt.Printf(
			"%sError reading archive path: %s\n",
			command.LogErrorPrefix,
			err,
		)
	}

	cfg := aws.NewConfig().
		WithRegion(cmdConfig.region)

	if cmdConfig.profile != "" {
		creds := credentials.NewSharedCredentials("", cmdConfig.profile)
		cfg = cfg.WithCredentials(creds)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		fmt.Printf(
			"%sFailed to create AWS session: %s\n",
			command.LogErrorPrefix,
			err,
		)
		return 1
	}
	svc := s3.New(sess)

	fmt.Printf(
		"%sUsing path s3://%s%s\n",
		command.LogInfoPrefix,
		cmdConfig.bucket,
		cmdConfig.s3Key,
	)

	archiveFileInfo := archives[0]
	archiveFilePath := path.Join(
		cmdConfig.source,
		archiveFileInfo.Name(),
	)
	fmt.Printf(
		"%sReading dependency file: %s\n",
		command.LogInfoPrefix,
		archiveFilePath,
	)

	archiveFile, err := c.os.Open(archiveFilePath)
	if err != nil {
		fmt.Printf(
			"%sFailed to open file %s: %s\n",
			command.LogErrorPrefix,
			archiveFilePath,
			err,
		)
		return 1
	}

	defer func() {
		if ferr := archiveFile.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	s3Path := path.Join(cmdConfig.s3Key, archiveFileInfo.Name())
	fmt.Printf(
		"%sUploading to path: s3://%s%s\n",
		command.LogInfoPrefix,
		cmdConfig.bucket,
		s3Path,
	)
	uploadResult, err := svc.PutObject(&s3.PutObjectInput{
		Body:          archiveFile,
		Bucket:        aws.String(cmdConfig.bucket),
		Key:           aws.String(s3Path),
		ContentLength: aws.Int64(archiveFileInfo.Size()),
		ContentType:   aws.String("application/gzip"),
	})

	if err != nil {
		fmt.Printf(
			"%sFailed to archive object to s3://%s%s, %s\n",
			command.LogErrorPrefix,
			cmdConfig.bucket,
			cmdConfig.s3Key,
			err,
		)
		return 1
	}

	fmt.Println(uploadResult)

	return 0
}
