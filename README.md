<div align="center">

[![License](https://img.shields.io/badge/License-MIT-blue)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/daytonaio/daytona-provider-aws)](https://goreportcard.com/report/github.com/daytonaio/daytona-provider-aws)
[![Issues - daytona](https://img.shields.io/github/issues/daytonaio/daytona-aws-provider)](https://github.com/daytonaio/daytona-provider-aws/issues)
![GitHub Release](https://img.shields.io/github/v/release/daytonaio/daytona-aws-provider)

</div>

<h1 align="center">Daytona AWS Provider</h1>
<div align="center">
This repository is the home of the <a href="https://github.com/daytonaio/daytona">Daytona</a> AWS Provider.
</div>
</br>

<p align="center">
  <a href="https://github.com/daytonaio/daytona-provider-aws/issues/new?assignees=&labels=bug&projects=&template=bug_report.md&title=%F0%9F%90%9B+Bug+Report%3A+">Report Bug</a>
    ·
  <a href="https://github.com/daytonaio/daytona-provider-aws/issues/new?assignees=&labels=enhancement&projects=&template=feature_request.md&title=%F0%9F%9A%80+Feature%3A+">Request Feature</a>
    ·
  <a href="https://go.daytona.io/slack">Join Our Slack</a>
    ·
  <a href="https://x.com/Daytonaio">X</a>
</p>

The AWS Provider allows Daytona to create and manage workspace projects on Amazon EC2 instances.

To use this provider, ensure your AWS programmatic access user has the `AmazonEC2FullAccess` permissions.
This policy grants the necessary permissions to manage EC2 instances, which is crucial for Daytona's workspace project creation and management.

## Target Options

| Property          | Type   | Optional | DefaultValue          | InputMasked | DisabledPredicate |
| ----------------- | ------ | -------- | --------------------- | ----------- | ----------------- |
| Region            | String | true     | us-east-1             | false       |                   |
| Image Id          | String | true     | ami-04a81a99f5ec58529 | false       |                   |
| Instance Type     | String | true     | t2.micro              | false       |                   |
| Device Name       | String | true     | t2./dev/sda1          | false       |                   |
| Volume Size       | String | true     | 10                    | false       |                   |
| Volume Type       | String | true     | gp3                   | false       |                   |
| Access Key Id     | String | false    |                       | true        |                   |
| Secret Access Key | String | false    |                       | true        |                   |

### Preset Targets

The AWS Provider has no preset targets. Before using the provider you must set the target using the `daytona target set` command.

## Code of Conduct

This project has adapted the Code of Conduct from the [Contributor Covenant](https://www.contributor-covenant.org/). For more information see the [Code of Conduct](CODE_OF_CONDUCT.md) or contact [codeofconduct@daytona.io.](mailto:codeofconduct@daytona.io) with any additional questions or comments.

## Contributing

The Daytona Docker Provider is Open Source under the [MIT License](LICENSE). If you would like to contribute to the software, you must:

1. Read the Developer Certificate of Origin Version 1.1 (https://developercertificate.org/)
2. Sign all commits to the Daytona Docker Provider project.

This ensures that users, distributors, and other contributors can rely on all the software related to Daytona being contributed under the terms of the [License](LICENSE). No contributions will be accepted without following this process.

Afterwards, navigate to the [contributing guide](CONTRIBUTING.md) to get started.

## Questions

For more information on how to use and develop Daytona, talk to us on
[Slack](https://go.daytona.io/slack).
