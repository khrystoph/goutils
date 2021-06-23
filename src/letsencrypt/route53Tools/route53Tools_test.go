package route53Tools

import (
	"fmt"
	"os"
)

func init() {

}

func main() {
	ec2Creds, err := route53Tools.getCredentials()
	if err != nil {
		fmt.Printf("Failed to retrieve credentials.")
		os.Exit(1)
	}
	fmt.Printf("Token Expires on: %v", ec2Creds.Expiration)
	os.Exit(0)
}
