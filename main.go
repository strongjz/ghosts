package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"flag"
	"github.com/go-ini/ini"
	"os"
	"path/filepath"
)

func getIniLocation() string {
	if filename := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); filename != "" {
		return filename
	}

	homeDir := os.Getenv("HOME") // *nix
	if homeDir == "" {           // Windows
		homeDir = os.Getenv("USERPROFILE")
	}
	if homeDir == "" {
		fmt.Println("home folder not found")
		os.Exit(1)
		return ""
	}

	return filepath.Join(homeDir, ".aws", "credentials")
}

func writeIni(sectionName string, credentials *sts.AssumeRoleOutput) {
	iniLocation := getIniLocation()

	var cfg *ini.File
	if _, err := os.Stat(iniLocation); os.IsNotExist(err) {
		fmt.Printf("No config file found at: %s. Creating new one.\n", iniLocation)
		cfg = ini.Empty()
	} else {
		cfg, err = ini.Load(iniLocation)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	cfg.NewSection(sectionName)
	cfg.Section(sectionName).NewKey("aws_access_key_id", aws.StringValue(credentials.Credentials.AccessKeyId))
	cfg.Section(sectionName).NewKey("aws_secret_access_key", aws.StringValue(credentials.Credentials.SecretAccessKey))
	cfg.Section(sectionName).NewKey("aws_session_token", aws.StringValue(credentials.Credentials.SessionToken))
	cfg.SaveTo(iniLocation)

}

//Profile Name

func main() {

	base := flag.String("base", "default", "base profile assuming")

	profile := flag.String("profile", "", "profile to write creds out too")

	role_arn := flag.String("arn", "", "Role ARN")
	sess_name := flag.String("name", "sts-sesssion","name of the session")

	duration := flag.Int64("duration", 900, "number of seconds credentails will last")
	mfa_bool := flag.Bool("mfa", false, "indicates if a mfa is need for this role")
	mfa_token := flag.String("token", "","MFA token value")
	mfa_serial := flag.String("serial","","MFA serial number, arn:aws:iam::123456789012:mfa/user")

	flag.Parse()


	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: *base,
		SharedConfigState: session.SharedConfigEnable,

	}))


	svc := sts.New(sess)

	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(*role_arn),             // Required
		RoleSessionName: aws.String(*sess_name), // Required
		DurationSeconds: aws.Int64(*duration),

	}

	if *mfa_bool {
		if (*mfa_token == "" ){

			fmt.Println("MFA is enabled and token must be set")
			os.Exit(1)
		}
		if (*mfa_serial == "" ){
			fmt.Println("MFA is enabled and serial must be set")
			os.Exit(1)
		}

		params = &sts.AssumeRoleInput{
			RoleArn:         aws.String(*role_arn),             // Required
			RoleSessionName: aws.String(*sess_name), // Required
			DurationSeconds: aws.Int64(*duration),
			SerialNumber:    aws.String(*mfa_serial),
			TokenCode:       aws.String(*mfa_token),
		}
	}

	resp, err := svc.AssumeRole(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	if *profile == "" {
		os.Setenv("AWS_ACCESS_KEY_ID", aws.StringValue(resp.Credentials.AccessKeyId))
		os.Setenv("AWS_SECRET_ACCESS_KEY", aws.StringValue(resp.Credentials.SecretAccessKey))
		os.Setenv("AWS_SESSION_TOKEN", aws.StringValue(resp.Credentials.SessionToken))


	}else {
		writeIni(*profile, resp)
	}

	// Pretty-print the response data.
	//fmt.Println(resp)

}