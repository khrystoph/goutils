package main

import (
	"flag"
	"os"

	//import our local package for route53Tools
	"goutils/letsencrypt/route53Tools"
	"goutils/logUtils"
)

var (
	instanceRole string
)

func init() {

	//set up flags to be parsed from command-line
	flag.StringVar(&instanceRole, "r", "foo", "Enter the role that you want to use to pull credentials")
	flag.StringVar(&instanceRole, "role", "foo", "Enter the role that you want to use to pull credentials")
}

func main() {
	//call parse on the flags that you've provided
	flag.Parse()

	ec2Creds := route53Tools.AwsCredentials{}
	ec2Creds, err := route53Tools.GetCredentials(instanceRole)
	if err != nil {
		logUtils.Error.Printf("Failed to retrieve credentials.")
	}
	logUtils.Info.Printf("Token Expires on: %v", ec2Creds.Expiration)
	logUtils.Info.Printf("Access Key Id: %v", ec2Creds.AccessKeyId)
	os.Exit(0)
}
