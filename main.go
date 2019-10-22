package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/itsbalamurali/wgu-s3/commands"

	"github.com/urfave/cli"
)

func main() {

	var account string
	var storageClass string
	var accessKeyID string
	var secretAccessKey string
	var bucket string
	var bucketPattern string
	var groupByRegion bool

	app := cli.NewApp()
	app.Name = "s3metrics"
	app.Usage = "AWS S3 Bucket metrics tool"
	app.Description = "AWS S3 Bucket metrics tool"
	app.Version = "0.0.1"
	app.Author = "Balamurali Pandranki <balamurali@live.com>"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "account",
			Usage:       "returns the information about the specified account in the aws cli credentials file",
			Destination: &account,
		},
		cli.StringFlag{
			Name:        "access_key_id",
			Usage:       "AWS API Access Key ID for using with alternate manual API credentials",
			Destination: &accessKeyID,
		},
		cli.StringFlag{
			Name:        "secret_access_key",
			Usage:       "AWS API Secret Access Key for using with alternate manual API credentials",
			Destination: &secretAccessKey,
		},
		cli.StringFlag{
			Name:        "bucket",
			Usage:       "returns the information about the specified bucket eg: s3://example-bucket",
			Destination: &bucket,
		},
		cli.StringFlag{
			Name:        "pattern",
			Usage:       "returns only the buckets/objects inside matching pattern eg: s3://example-bucket/a/folder/b/c/*",
			Destination: &bucketPattern,
		},
		cli.BoolFlag{
			Name:        "group-by-region",
			Usage:       "returns the buckets list grouped by their aws region eg: ap-south-1",
			Destination: &groupByRegion,
		},
		cli.StringFlag{
			Name:        "storageclass",
			Usage:       "returns buckets which contain objects of the specified storage class eg: STANDARD_IA",
			Destination: &storageClass,
		},
	}
	app.Action = func(c *cli.Context) error {
		// Create AWS SDK client session.
		sess := session.Must(session.NewSession())
		return commands.ListBuckets(c, sess)
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
