# acbuild

acbuild is a command line utility to build and modify App Container images (ACIs).

## Rationale

Dockerfiles are powerful and feature useful concepts such as build layers,
controlled build environment. At the same time, they lack flexibility
(impossible to extend, re-use environment variables) and don't play nicely with
the app spec and Linux toolchain (e.g. shell, makefiles)

This proposal introduces the concept of a command-line build tool, `acbuild`,
that natively supports ACI builds and integrates well with shell, makefiles and
other Unix tools.

## Dependencies

acbuild requires a handful of commands be available:

- `systemd-nspawn`
- `cp`
- `mount`
- `modprobe`
- `tar`

## Usage

A build with `acbuild` is explicitly started with `init` and finished with
`finish`. While a build is in progress the current ACI is stored expanded in the
current working directory at `.acbuild.tmp`. A build can be started with an
empty ACI, or an initial ACI can be provided.

The following commands are supported:

* `acbuild abort`

  Abort the current build, throwing away any changes since `init` was called.

* `acbuild add-anno NAME VALUE`

  Updates the ACI to contain an annotation with the given name and value. If the
  annotation already exists, its value will be changed.

* `acbuild add-dep IMAGE_NAME --image-id sha512-... --label env=canary`

  Updates the ACI to contain a dependency with the given name. If the dependency
  already exists, its values will be changed.

* `acbuild add-env NAME VALUE`

  Updates the ACI to contain an environment variable with the given name and
  value. If the variable already exists, its value will be changed.

* `acbuild add-label NAME VALUE`

  Updates the ACI to contain a label with the given name and value. If the label
  already exists, its value will be changed.

* `acbuild add-mount NAME PATH`

  Updates the ACI to contain a mount point with the given name and path. If the
  mount point already exists, its path will be changed.

* `acbuild add-port NAME PROTOCOL PORT`

  Updates the ACI to contain a port with the given name, protocol, and port. If
  the port already exists, its values will be changed.

* `acbuild copy PATH_ON_HOST PATH_IN_ACI`

  Copy a file or directory into an ACI.

* `acbuild exec CMD [ARGS]`

  Run a given command in an ACI, and save the resulting container as a new ACI.

* `acbuild name ACI_NAME`

  Changes the name of an ACI in its manifest.

* `acbuild set-group GROUP`

  Set the group the app will run as inside the container.

* `acbuild set-user USER`
  
  Set the user the app will run as inside the container

* `acbuild set-run CMD [ARGS]`

  Sets the run command in the ACI's manifest.

Every `add` command has an accompanying `rm` command

### acbuild exec

`acbuild exec` builds the root filesystem with any dependencies the ACI has
using overlayfs, and then executes the given command using systemd-nspawn. The
current ACI being built is the upper level in the overlayfs, and thus modified
files that came from the ACI's dependencies will be copied into the ACI. More
information on this behavior is available
[here](https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt).

`acbuild exec` requires overlayfs if the ACI being operated upon has
dependencies.

`acbuild exec` also requires root.

## Planned features

### Context-free mode

There are scenarios in which it is not convenient to need to call `init` and
`finish`, the most obvious being when a single change is made to an existing
ACI. A flag will be added to allow every subcommand to be performed on a given
ACI, instead of looking for a current build in progress.

### Image signing

It would be convenient if the appropriate gpg keys could be passed into `acbuild
finish`, and a ASC file would then be produced in addition to the ACI file.

### Squash

`acbuild squash`: fetch all the dependencies for the given image and squash them
together into an ACI without dependencies.

### Image fetching with init

Pass in an image name, instead of a path to an ACI for `acbuild init`, and the
image will be fetched and used as the starting point for the build.

## Examples

Use apt-get to install nginx.

```
acbuild init
acbuild add-dep quay.io/fermayo/ubuntu
acbuild exec -- apt-get update
acbuild exec -- apt-get -y install nginx
acbuild finish ubuntu-nginx.aci
```

## Related work

- https://github.com/sgotti/baci
- https://github.com/appc/spec/tree/master/actool - particularly the `build` and
  `patch-manifest` subcommands. `acbuild` may subsume such functionality,
  leaving `actool` as a validator only.


