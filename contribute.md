# Contributing

When contributing to this repository, please first discuss the change you wish to make via issue,
email, or any other method with the owners of this repository before making a change.

Please note we have a code of conduct, please follow it in all your interactions with the project.

## Reporting Bug

1. **Use a clear and descriptive title for the issue to identify the problem**.
2. **Describe the exact steps which reproduce the problem in as many details as possible**. For example, start by explaining how you started node, e.g. which command exactly you used in the terminal, or how you the node. When listing steps,don't just say what you did, but explain how you did it. For example, if you add an extra arguments, explain which arguments you've added.
3. **Explain which behavior you expected to see instead and why**.
4. **Include screenshots and animated GIFs** which show you following the described steps and clearly demonstrate the problem.
5. **Follow the issue template**.

## Tackling Issue

Before taking these steps, please fork the repository, and create branch with the following format: `{issue_number}-{short-issue-title}`

1. **Take the one of the ticket in the issue list** and discuss with the other developer to make sure no one have taken the same issue, and self assign the ticket, so there'll be no duplicate work.
2. **Discuss in the ticket comment section**, any discussion should happen in the comment section.
3. **Thou Shalt Write Tests**, write tests on your the code you added, it save thousands of life.
4. **Document your update** always document your code when necessary, some changes will require us to update the [wiki-page](https://github.com/zoobc/zoobc-core/wiki)
5. **Knock knock, is it safe to PR?**, run `golangci-lint` and `go test ./...` to make sure every test suite working fine before submitting PR. Read more on [readme.md](readme.md)

## Pull Request Process

All Pull Requests should be submitted to `develop` branch, or any other feature branch.

1. **Ensure any install or build dependencies are removed before the end of the layer when doing a build**.
2. **Update the [readme.md](readme.md) with details of changes to the interface**, this includes new environment variables, exposed ports, useful file locations and container parameters.
3. **Increase the version numbers** in any examples files and the [readme.md](readme.md) to the new version that this Pull Request would represent. The versioning scheme we use is [SemVer](http://semver.org/).
4. Your Pull Request will be merged in once you have the **approval of two other developers**, or if you do not have permission to do that, you may request the second reviewer to merge it for you.
5. **Pull request are `squashed` into a single commit** to keep the history clean.
6. If your PR containing a packages update, don't mind to ***tidying go modules***, it is will sliming of go modules, removing unused package and updates your current go.mod to include the dependencies needed for tests in your module.