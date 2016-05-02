# Proj

Access projects in their appropriate development environments, quickly and easily

    dustin@laptop ~ $ proj moz/cloud-tools
    dustin@moz-devel p/cloud-tools $

Features:

  * Automatically start and connect to Amazon EC2 instances

# Overview

Proj operates on a "project path", similar to the way `cd` operates on a directory path.
It begins in the current directory, then looks for a "child project" by the name of the first path component.
So `proj moz/cloud-tools` begins by looking for the child project `moz`.

Based on the configuration for that child project, the tool may perform some setup or connect to another host.
Once that is done, it continues on to the next path component.
When all path components are processed (or when an error occurs), it starts a shell, with a few simple hooks allowing you to set up the project's development environment.

# Usage

Proj is best used in concert with other tools.

## Screen or tmux

I have created an alias, `pp`, which invokes `proj` in a new screen window (I know tmux is the new hotness, kids, but I have screen sessions older than you are).

## Ansible or other configuration management

With multiple hosts and lots of develpoment environments, you'll probably want to use some tools to keep everything organized.

## Shell scripts

You'll need to write some shell scripts to coordinate development environment setup.
Development environments are as unique as developers (well, more -- diversity is not software engineering's strong point just yet), so you're on your own here.

When starting a shell, Proj searches for a file named `.projrc` (configurable in `~/.proj.cfg`) in the current directory and each enclosing directory until it finds one.
It then invokes the shell with `--rcfile path/to/.projrc -i`.
I use bash, so this assumes bash -- sorry!

Note that this does *not* pass the `-l` option to indicate a login shell.
It is up to `.projrc` to emulate this behavior by sourcing the appropriate files, if desired.

# Configuration

## Host-level

On each host, proj looks for configuration in `~/.proj.cfg`.
The file format is like an INI file, similar to `gitconfig`.

It has the following sections:

### ssh

Each `ssh` section specifies a host to connect to (but not automatically start).
The hostname defaults to the section name.

    [ssh myhost]
    hostname = "myhost.foo.com"
    # connection information
    user = dustin
    proj-path = /usr/local/bin/proj  # optional path to proj binary on the instance
    forward-agent = yes  # defaults to no
    ignore-known-hosts = yes  # defaults to no; if set, known hosts won't be checked or updated

### ec2

Each `ec2` section specifies an EC2 instance which can be started on demand with the `ec2` child type.

    [ec2 "devel"]
    # AWS account credentials
    access-key = ABC123
    secret-key = DEF456
    # instance information
    region = us-east-1
    name = devel-instance  # instance name ("Name" tag)

in addition to the SSH section options (with the exception of hostname) given above.

The following is the minimal policy that the access/secret key pair must have:

    {
      "Statement": [
        {
          "Resource": [
            "*"
          ],
          "Action": [
            "ec2:DescribeInstances"
          ],
          "Effect": "Allow",
          "Sid": "Stmt1462218813000"
        },
        {
          "Resource": [
            "arn:aws:ec2:<region>:<accountNumber>:instance/<instanceId>"
          ],
          "Action": [
            "ec2:StartInstances",
          ],
          "Effect": "Allow",
          "Sid": "Stmt1462218813001"
        }
      ],
      "Version": "2012-10-17"
    }

Note that this allows DescribeInstances for all instances.  Amazon does not
provide more granular control of this method.

### shell

The `shell` section configures the search for shell initialization files.

    [shell]
    rcfile = .projrc  # default
    no-search = true  # true to disable searching up the directory hierarchy

## Children

Proj searches for a child project `childproj` as follows, starting in the current directory:

 * If `.proj/childproj.cfg` exists, it is read to determine the configuration of the child project.
 * If a subdirectory named `childproj` exists, proj treats that directory as the child project.

When there are no more path components, proj implicitly looks for a child project named `DEFAULT`, starting it if present.
This allows short paths for the most common projects; for example, `proj moz` might correspond to a Gecko development project, while `proj moz/mig` opens the Mozilla InvestiGator development environment.

Each child configuration file begins with the child type as a section header, chosen from the child types below, followed by optional keys, for example

    [cd]
    dir = devel/bar-project

The following keys are optional for all child types

    # prepend this path to the project path given to the child project
    prepend = extra/path

### cd

The `cd` child type requires a `dir` key giving the directory of the child project.

### ssh

The `ssh` child type requires a `host` key which refers to an `ssh` section in the host configuration file.

### ec2

The `ec2` child type requires an `instance` key which refers to an `ec2` section in the host configuration.
