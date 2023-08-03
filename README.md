# sysdoc

`sysdoc` allows to model system dependencies in a text based manner and process these information in an automated way.

## Installation

Head to https://github.com/unprofession-al/sysdoc/releases/latest and grab the release fitting your need. Unpack the archive
and put the `sysdoc` binary somewhere in your `$PATH`.

## Usage

### Document the System

`sysdoc` reads a directory structure and finds files in this structure which are expected to describe your system architecture.
It assumes a hirachical structure (as for example the [C4 Model](https://c4model.com/) suggests) where every folder represents
layer. Each layer (and therefore folder) can contain a Frontmatter file (Markdown with a YAML header, usually a `README.md` file)
do describe the entity of the given layer.

_Details about the documentation format of system need to be documented here_ 

### Use the Documentation with `sysdoc`

```
# sysdoc -h
sysdoc allows to document dependencies between systems

Usage:
  sysdoc [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  render      renders system documentation in a given template to standard output
  serve       renders system documentation in a given d2lang template to a svg file and serves it over http
  svg         renders system documentation in a given d2lang template to a svg file
  version     Print version info

Flags:
      --base string     base directory of the sysdoc definitions (default ".")
      --config string   configuration file path (default "./sysdoc.yaml")
      --focus strings   elements to be focussed
      --glob string     glob to find sysdoc definitions (default "README.md")
  -h, --help            help for sysdoc

Use "sysdoc [command] --help" for more information about a command.
```

### Writing Renderers


_Details about the templating of renderers need to be documented here_ 

## Why `sysdoc`?

Commonly there are a few ways to document a system architecture:

__No documentation whatsoever:__ This is bad for apparent reasons.

__Some wiki pages:__ This is a good and easy approach but requires a decent amount of sustained editorial work to keep things
up to date. Particularly dependencies between systems owned by differend parties seem to be hard to get right and even harder
to keep maintained. Also, no added benefit is generated with an up to date documetation (other than for the documtations sake)
as the data is unstructured and usually hard to access from third party applications.

__Self-generated documentation:__ In an ideal setting, documentation is generated either by the software itself or by its run
time (kubernetes and some service mesh can do a pretty neat job with this). This should be the goal of a green field project
but seems to be an unrealistic effort for most project that are build ontop of some legacy or around some software that is not
build in-house/consumend as SaaS/of-the-shelve and therefore not exactly built to spec.

__Specialist tool:__ There are a bunch of [specialist tools](https://content.ardoq.com/en/gartner-magic-quadrant-for-enterprise-architecture-tools)
available to create a proper enterprise architecture documentation. These tools often cost a good amount of money to use and are a bit
overkill in many situations.

`sysdoc` attempts to provide an low profile, relatively easy and cost efficient way to document your system environment. It is
based on Markdown (with structured data provided as YAML or JSON), which allows you to keep everything neatly under control with
a SCM of your choice. As your system descriptions are provided in simple Markdown, the documentation is already usable as is, as
for example GitHub can render the information in a user consumable way. However, as the meta data of your systems cosist of stuctured
data (in particular the interfaces of a system and the dependencies from systems to interfaces of other systems), more information
can be generated. For example, the following questions can be answered quite easily:

- A system experiences a downtime. Which other systems are affected?
- A system will be subject of change. What dependencies do we have to take care of?
- ...
