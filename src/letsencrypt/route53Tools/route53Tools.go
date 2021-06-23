package route53Tools

import (
	"encoding/json"
	"goutils/logUtils"
	"io/ioutil"
	"net/http"
)

const (
	URL               = "http://169.254.169.254/latest/meta-data/"
	CREDS_PATH_STRING = "identity-credentials/ec2/security-credentials/"
)

/*
* This package provides all the necessary functions to update Route53 with the appropriate TXT record
* and update both the R53 record.
 */

// We want to pull the document that contains
// curl 169.254.169.254/latest/dynamic/instance-identity/document/

type AwsCredentials struct {
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Token           string `json:"Token"`
	Expiration      string `json:"Expiration"`
}

// We want to pull the role credentials here by using this path in the container:
// curl 169.254.169.254/latest/meta-data/identity-credentials/ec2/security-credentials/{role_name}
func GetCredentials(iamRole string) (creds AwsCredentials, err error) {
	credsURL := URL + CREDS_PATH_STRING + iamRole
	response, err := http.Get(credsURL)
	if err != nil {
		logUtils.Error.Fatal("Error pulling instance creds: %v", err)
	}
	defer response.Body.Close()
	htmlBody, err := ioutil.ReadAll(response.Body)
	creds = AwsCredentials{}
	err = json.Unmarshal(htmlBody, &creds)

	return creds, nil
}

//R53Record is an externally reference-able struct to transmit data to the function calls coming from the main application
type R53Record struct {
}
