package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-ini/ini"
	"os"
	"path/filepath"
)

// Returns the file path to the credentials files
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

// Writes out the configs generated from STS to the ~/.aws/credentials file or where AWS_SHARED_CREDENTIALS_FILE is define
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

// Parses the INI file to populate the sts credentials api call.
func parseConfig(section string, config string) error {

	cfg, err := ini.Load(config)
	if err != nil {
		return err
	}

	sec, err := cfg.GetSection(section)
	if err != nil {
		return err
	}

	base, err := sec.GetKey("base")
	if err != nil {
		return err
	}

	role, err := sec.GetKey("role_arn")
	if err != nil {
		return err
	}

	mfa_serial, err := sec.GetKey("mfa_serial")
	if err != nil {
		return err
	}

	profile, err := sec.GetKey("profile")
	if err != nil {
		return err
	}

	sess_name, err := sec.GetKey("session_name")
	if err != nil {

		flag.Set("sess_name", fmt.Sprintf("sts-creds-%s", profile.String()))

	} else {
		flag.Set("sess_name", sess_name.String())

	}

	flag.Set("base", base.String())
	flag.Set("arn", role.String())
	flag.Set("serial", mfa_serial.String())
	flag.Set("profile", profile.String())

	return nil
}

func checkFlags() error {

	if role_arn == "" {
		return errors.New("Role ARN must be set\n")
	}

	if mfa_token != "" {
		mfa_bool = true
	}

	if duration > 3600 {
		return errors.New("Duration must be between 900 and 3600 seconds\n")
	}

	return nil

}

func assumeRoleInput() *sts.AssumeRoleInput {
	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(role_arn),  // Required
		RoleSessionName: aws.String(sess_name), // Required
		DurationSeconds: aws.Int64(duration),
	}

	if mfa_bool {
		if mfa_token == "" {

			fmt.Println("MFA is enabled and token must be set\n")
			os.Exit(1)
		}
		if mfa_serial == "" {
			fmt.Println("MFA is enabled and serial must be set\n")
			os.Exit(1)
		}

		params = &sts.AssumeRoleInput{
			RoleArn:         aws.String(role_arn),  // Required
			RoleSessionName: aws.String(sess_name), // Required
			DurationSeconds: aws.Int64(duration),
			SerialNumber:    aws.String(mfa_serial),
			TokenCode:       aws.String(mfa_token),
		}
	}

	return params
}

var (
	base       string
	profile    string
	role_arn   string
	sess_name  string
	duration   int64
	mfa_bool   bool
	mfa_token  string
	mfa_serial string
	debug      bool
	config     string
)

func init() {

	flag.StringVar(&base, "base", "default", "base profile assuming")
	flag.StringVar(&profile, "profile", "", "profile to write creds out too")
	flag.StringVar(&role_arn, "arn", "", "Required - Role ARN")
	flag.StringVar(&sess_name, "name", "sts-session", "name of the session")
	flag.Int64Var(&duration, "duration", 900, "Number of seconds credentials will last, 900 - 3600")
	flag.BoolVar(&mfa_bool, "mfa", false, "indicates if a mfa is need for this role")
	flag.StringVar(&mfa_token, "token", "", "MFA token value. Requireed if MFA set.")
	flag.StringVar(&mfa_serial, "serial", "", "MFA serial number, ie arn:aws:iam::123456789012:mfa/user - Required if MFA set. ")
	flag.BoolVar(&debug, "debug", false, "debug output")
	flag.StringVar(&config, "config", "", "Config file that contains assume role information")
}

func printFlags() {
	fmt.Printf("\nDebug Printout of flags\n")
	fmt.Printf("base: %v\n", base)
	fmt.Printf("profile: %v\n", profile)
	fmt.Printf("role_arn: %v\n", role_arn)
	fmt.Printf("sess_name: %v\n", sess_name)
	fmt.Printf("duration: %v\n", duration)
	fmt.Printf("mfa_bool: %v\n", mfa_bool)
	fmt.Printf("mfa_token: %v\n", mfa_token)
	fmt.Printf("mfa_serial: %v\n", mfa_serial)
	fmt.Printf("debug: %v\n", debug)
	fmt.Printf("config: %v\n", config)
}

func main() {

	flag.Parse()

	if config != "" {
		err := parseConfig(profile, config)
		if err != nil {
			fmt.Printf("Error %v\n", err)
			os.Exit(1)
		}
	}

	err := checkFlags()
	if err != nil {
		fmt.Printf("Error %v\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if debug {
		printFlags()
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           base,
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sts.New(sess)

	params := assumeRoleInput()

	if debug {
		fmt.Printf("\nDEBUG: Params %v\n", params)
	}

	resp, err := svc.AssumeRole(params)
	if err != nil {
		fmt.Printf("Error with Assume Role %v\n", err.(awserr.Error))

		os.Exit(2)
	}

	if profile == "" {
		os.Setenv("AWS_ACCESS_KEY_ID", aws.StringValue(resp.Credentials.AccessKeyId))
		os.Setenv("AWS_SECRET_ACCESS_KEY", aws.StringValue(resp.Credentials.SecretAccessKey))
		os.Setenv("AWS_SESSION_TOKEN", aws.StringValue(resp.Credentials.SessionToken))
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", os.Getenv("AWS_ACCESS_KEY_ID"))
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", os.Getenv("AWS_SECRET_ACCESS_KEY"))
		fmt.Printf("export AWS_SESSION_TOKEN=%s\n", os.Getenv("AWS_SESSION_TOKEN"))
		fmt.Printf("export AWS_SECURITY_TOKEN=\"$AWS_SESSION_TOKEN\"\n")

	} else {
		writeIni(profile, resp)
	}

	// Pretty-print the response data.
	if debug {
		fmt.Println("\nRepsonse:\n")
		fmt.Println(resp)
	}

}
