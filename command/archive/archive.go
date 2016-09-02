package archive

import (
	"bitbucket.org/bosgood/dep-get/command"
	"bitbucket.org/bosgood/dep-get/lib/fs"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/cli"
	"io"
	"net/url"
	"os"
	"path"
)

type archiveCommand struct {
	os     fs.FileSystem
	config archiveCommandFlags
	s3     *s3.S3
}

type archiveCommandFlags struct {
	command.BaseFlags
	platform      string
	source        string
	region        string
	profile       string
	s3URL         string
	bucket        string
	s3Key         string
	maxNumWorkers uint
}

var (
	realOS fs.FileSystem = &fs.OSFS{}
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
	return "Archives application dependencies to S3"
}

func (c *archiveCommand) Help() string {
	_, flagSet, _ := getConfig([]string{})
	flagSet.PrintDefaults()
	return ""
}

func getConfig(args []string) (archiveCommandFlags, *flag.FlagSet, error) {
	var cmdConfig archiveCommandFlags

	cmdFlags := flag.NewFlagSet("archive", flag.ExitOnError)
	cmdFlags.BoolVar(&cmdConfig.Help, "help", false, "show command help")
	cmdFlags.StringVar(&cmdConfig.platform, "platform", "", "platform type (allowed: nodejs|python)")
	cmdFlags.StringVar(&cmdConfig.source, "source", "", "project directory (default: .)")
	cmdFlags.StringVar(&cmdConfig.profile, "profile", "", "AWS credentials profile (default: default)")
	cmdFlags.StringVar(&cmdConfig.region, "region", "", "AWS region")
	cmdFlags.StringVar(&cmdConfig.s3URL, "path", "", "S3 upload path")
	cmdFlags.UintVar(&cmdConfig.maxNumWorkers, "concurrency", 5, "Maximum number of workers (default: 5)")

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

func (c *archiveCommand) InitS3() error {
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

func (c *archiveCommand) uploadToS3(archiveFileInfo os.FileInfo, archiveFile io.ReadSeeker) error {
	s3Path := path.Join(c.config.s3Key, archiveFileInfo.Name())
	fmt.Printf(
		"%sUploading to path: s3://%s%s\n",
		command.LogInfoPrefix,
		c.config.bucket,
		s3Path,
	)
	_, err := c.s3.PutObject(&s3.PutObjectInput{
		Body:          archiveFile,
		Bucket:        aws.String(c.config.bucket),
		Key:           aws.String(s3Path),
		ContentLength: aws.Int64(archiveFileInfo.Size()),
		ContentType:   aws.String("application/gzip"),
	})

	return err
}

func (c *archiveCommand) Upload(fi os.FileInfo) error {
	archiveFilePath := path.Join(
		c.config.source,
		fi.Name(),
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
		return err
	}

	defer func() {
		if ferr := archiveFile.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	err = c.uploadToS3(fi, archiveFile)
	if err != nil {
		fmt.Printf(
			"%sFailed to archive object to s3://%s%s, %s\n",
			command.LogErrorPrefix,
			c.config.bucket,
			c.config.s3Key,
			err,
		)
		return err
	}

	fmt.Printf(
		"%sUploaded object to s3://%s%s%s\n",
		command.LogSuccessPrefix,
		c.config.bucket,
		c.config.s3Key,
		fi.Name(),
	)

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
	c.config = cmdConfig

	archives, err := c.os.ReadDir(c.config.source)
	if err != nil {
		fmt.Printf(
			"%sError reading archive path: %s\n",
			command.LogErrorPrefix,
			err,
		)
	}

	err = c.InitS3()
	if err != nil {
		fmt.Printf(
			"%sFailed to initialize AWS/S3 session: %s\n",
			command.LogErrorPrefix,
			err,
		)
		return 1
	}

	fmt.Printf(
		"%sUsing path s3://%s%s\n",
		command.LogInfoPrefix,
		c.config.bucket,
		c.config.s3Key,
	)

	remainingCount := len(archives)
	errors := make(chan error)
	successes := make(chan string)
	todo := make(chan os.FileInfo, remainingCount)

	// Enqueue all the files to be uploaded
	for _, archiveFileInfo := range archives {
		todo <- archiveFileInfo
	}

	fmt.Printf(
		"%sStarting %d workers\n",
		command.LogInfoPrefix,
		c.config.maxNumWorkers,
	)

	var i uint
	for i = 0; i < c.config.maxNumWorkers; i++ {
		go func() {
			for {
				archiveFileInfo := <-todo
				err := c.Upload(archiveFileInfo)
				if err != nil {
					errors <- err
				} else {
					successes <- archiveFileInfo.Name()
				}
			}
		}();
	}

	for {
		select {
		case <-errors:
			return 1
		case <-successes:
			remainingCount--
		}

		if remainingCount == 0 {
			break
		}
	}

	fmt.Printf(
		"%sUploaded %d object(s)\n",
		command.LogSuccessPrefix,
		len(archives),
	)

	return 0
}
