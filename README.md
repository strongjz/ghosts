# ghosts
Ghosts - Go implementation that will help manage multiple STS credentials for shared AWS accounts

-profile will write the the temporary credentails to the defaul ~/.aws/credentials profile name in the option


Leaving -profile blank will make GHOSTS print the export varables so that they can be set as environment variables

### Ghost CLI options
ghosts --help

Usage of ghosts:

  -arn string Role ARN

  -base string base profile assuming (default "default")

  -config string  Config file that contains assume role informations

  -debug debug output

  -duration int number of seconds credentials will last (default 900)

  -mfa indicates if a mfa is need for this role

  -name string name of the session (default "sts-sesssion")

  -profile string profile to write credentials out too

  -serial string MFA serial number, arn:aws:iam::123456789012:mfa/user

  -token string MFA token value

    	
    	
 ### Using a Config file
 
 GHOSTS will parse the file provided in the --config option and set the flags for 
 
  -arn 
     	
  -base 
   
  -serial 
      	
  -profile
 
 Config File Syntax
 
 [profile1]
 base="base1"
 role="arn:aws:iam::[ACCOUNT_NUMBER]:role/[ROLENAME]"
 profile="[PROFILE_TO_UPDATE]"
 mfa_serial="arn:aws:iam::[ACCOUNT_NUMBER]:mfa/[IAM_USERNAME]"
 
 [profile2]
 base="base1"
 role="arn:aws:iam::[ACCOUNT_NUMBER]:role/[ROLENAME]"
 profile="[PROFILE_TO_UPDATE]"
 mfa_serial="arn:aws:iam::[ACCOUNT_NUMBER]:mfa/[IAM_USERNAME]"
 
 
 Go STS overview
 
 https://docs.aws.amazon.com/sdk-for-go/api/service/sts/#pkg-overview
 
 Inspired by 
 
 https://github.com/wernerb/aws-adfs/blob/master/aws-adfs.go
 
 