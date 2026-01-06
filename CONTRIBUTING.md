# How to Contribute to ProtoValidate

## Did you find a security issue?

**Don't open an issue yet!** Reporting an issue in Github could put users at risk.

Instead, follow our responsible disclosure policy at [proto.studio/security](http://proto.studio/security).

## Did you find a bug?

### Do you want to create a patch?

If you don't have a new bug to fix you can find a list of good starter bugs on the [contributor page](https://github.com/rails/rails/contribute).

1. Create a new branch off of the current feature branch.
2. Write a unit test that verifies the bug and your fix.
3. Make sure all new functions/methods are commented in Go style comments.
4. [Create a pull request](https://github.com/rails/rails/pulls).

### Do you want another contributor to fix it?

You may [create a new issue](https://github.com/proto-studio/protovalidate/issues). Make sure to follow the provided template.

It is preferable that you provide code to reproduce the issue but a description would also be fine.

## Do you want a new feature?

### Do you want to build it yourself?

We're trilled you want to help out by developing a new feature!

If you don't have a feature in mind you can find open issues on the [contributor page](https://github.com/rails/rails/contribute).

First, ask yourself these two questions:

1. Will this feature be generally useful to almost everyone?
2. Is the data structure being tested part of the Go standard library?

If you answered "no" to either of the questions then you probably want to consider creating a new library that imports ProtoValidate instead.

If you do think it is generally useful and does not need to import any third-party libraries:

1. [Join Discord](https://proto.studio/social/discord) and ask in chat if other people want this feature. ProtoStudio staff tends to be in chat to help out.
2. Create a fork off of the current major development branch (we do not introduce new features in minor releases). If you are unsure what is the current development branch, please ask.
3. Write your new feature and make sure it is tested and documented.
4. [Create a pull request](https://github.com/rails/rails/pulls).

### Do you want someone else to build it?

Please [Join us on Discord](https://proto.studio/social/discord) and discuss your idea! We'd love to hear about it.

## Documentation

Documentation is included in the `/docs` folder. We use MDX to 

## Quality Standards

Before submitting a new Pull Request you should check to make sure the change meets our quality standards.

We require 100% test coverage of all code in the root, `/pkg` and `/internal`:

```bash
# This will make sure all the tests pass and we still have full test coverage.
make coverage
```

We also require us to maintain a A+ rating on [Go Report Card](https://goreportcard.com/). If you  have it installed locally you can run it with:

```bash
# If you do not have it installed locally that is OK. It will be run when you make a pull request.
make reportcard
```

For those unfamiliar, Go Report Card will test:

- Commonly misspelled English words.
- Formatting.
- Cyclomatic complexity.
- Common issues.
- More...

Additionally, the target version of Go for this package is **1.20**. If you have a newer version of Go and you have Docker installed we've provided a script to
let you test easily using v1.20.

```bash
# Run the tests in a Docker container using v1.20 of Docker
# User this if your local version is newer to ensure we don't break compatibility.
make test-docker
```

Builds that fail in Go v1.20 will not pass the tests.

At the time of this writing, Go v1.21 is out but is relatively new and does not provide any new features that are required to use ProtoValidate so we decided to not upgrade yet to give users time to upgrade at their own pace. Future versions we may increase this requirement.

## Thank You ❤️

Thank you to our contributors and users!

— The ProtoStudio Team