# Faktory [![Travis Build Status](https://travis-ci.org/contribsys/faktory.svg?branch=master)](https://travis-ci.org/contribsys/faktory?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/hunter-io/faktory)](https://goreportcard.com/report/github.com/hunter-io/faktory) [![Gitter](https://badges.gitter.im/contribsys/facktory.svg)](https://gitter.im/contribsys/faktory)

At a high level, Faktory is a work server.  It is the repository for
background jobs within your application. Jobs have a type and a set of
arguments and are placed into queues for workers to fetch and execute.

You can use this server to distribute jobs to one or hundreds of
machines. Jobs can be executed with any language by clients using
the Faktory API to fetch a job from a queue.

![webui](https://raw.githubusercontent.com/contribsys/faktory/master/docs/webui.png)

## Basic Features

- Jobs are represented as JSON hashes.
- Jobs are pushed to and fetched from queues.
- Jobs are reserved with a timeout, 30 min by default.
- Jobs `FAIL`'d or not `ACK`'d within the reservation timeout are requeued.
- FAIL'd jobs trigger a retry workflow with exponential backoff.
- Contains a comprehensive Web UI for management and monitoring.

## Installation

See the [Installation wiki page](https://github.com/hunter-io/faktory/wiki/Installation) for current installation methods.

## Deployment

To deploy a new version of Faktory on the Hunter infrastructure:
- update the dependencies with `make prepare`
- execute the test suite with `make test`
- compile the binary for Linux with `GOOS=linux GOARCH=amd64 make build`
- upload the binary on the Faktory server using scp `scp -i ... faktory ubuntu@faktory:DEST`
- use `chmod +x faktory` to ensure the binary can be executed
- archive the previous version of faktory in `/usr/bin/faktory` and replace it with the new one
- restart the server with `systemctl faktory restart`

## Documentation

Please [see the Faktory wiki](https://github.com/hunter-io/faktory/wiki) for full documentation.

## Support

You can find help in the [contribsys/faktory](https://gitter.im/contribsys/faktory) chat channel. Stop by and say hi!

## Author

Mike Perham, @mperham, mike @ contribsys.com
