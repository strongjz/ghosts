package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"flag"
)

//Profile Name

func main() {

	profile := flag.String("profile", "default", "profile to update")
	role_arn := flag.String("arn", "notset", "Role ARN")
	sess_name := flag.String("name", "sts-sesssion","name of the session")

	duration := flag.Int64("duration", 900, "number of seconds credentails will last")
	mfa_bool := flag.Bool("mfa", false, "indicates if a mfa is need for this role")
	mfa_token := flag.String("token", "","MFA token value")
	mfa_serial := flag.String("serial","","MFA serial number, arn:aws:iam::123456789012:mfa/user")

	flag.Parse()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: *profile,
		SharedConfigState: session.SharedConfigEnable,

	}))

	svc := sts.New(sess)

	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(*role_arn),             // Required
		RoleSessionName: aws.String(*sess_name), // Required
		DurationSeconds: aws.Int64(*duration),

	}

	if *mfa_bool {
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

	// Pretty-print the response data.
	fmt.Println(resp)

}