AWS IAM command wrapper

====

## Description

Create a AWS of IAM user from this tool.

## Demo

## Requirement

This tool use a AWS command.
[Installing the AWS Command Line Interface](http://docs.aws.amazon.com/cli/latest/userguide/installing.html)

And make a AWS account config file.
```
aws configure --profile aws-hoge
AWS Access Key ID [None]: [input AWS Access Key ID]
AWS Secret Access Key [None]: [input AWS Secret Access Key]
Default region name [None]: [input use region]
Default output format [None]: text
```
```
[aws-hoge]
output = text
region = [region]
aws_access_key_id = [AWS access key ID]
aws_secret_access_key = [AWS secret access key]
```

## Usage
`./create-iam --user-name hoge,fuga --account aws-hoge --policy-json template.json`

## Install
`go build create-iam.go`
