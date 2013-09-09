# AWS Docker Config

This tool is designed to deliver an app's configuration from a central config database into the process environment for an app. This tool enables [Twelve-Factor style configuration](http://12factor.net/config) for apps running on EC2.

## Install
Setup steps:

* Create DynamoDB table
* Creating IAM Role with access to the DynamoDB table
* Downloading ADF-CONFIG onto the EC2 instance

### Database Setup
Create a table named **adf-config** with a *hash key* of **App** and a *range key* of **Value**.

### IAM Role
Create an IAM role with the following policy:
```json
{
  "Version": "timestamp",
  "Statement": [
    {
      "Action": [
        "dynamodb:Scan"
        "dynamodb:PutItem",
        "dynamodb:DeleteItem"
      ],
      "Sid": "Stmt:timestamp:0",
      "Resource": [
        "*"
      ],
      "Effect": "Allow"
    }
  ]
}
```

Then be sure to boot your EC2 instances with a role that has this policy.

### Client Installation

```bash
$ curl -L drone.io/github.com/ryandotsmith/adf-config/files/adf-config.tar.gz \
  | tar xvz
```

## Usage

```bash
$ adf-config -l -a x
FOO=f
BAR=b

#TODO
$ adf-config -s BAZ=b -a x
BAZ=b

#TODO
$ adf-config -d BAZ -a x
BAZ=b
```
