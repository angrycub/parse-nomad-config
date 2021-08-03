# parse-nomad-config

`parse-nomad-config` parses one or more Nomad configuration files and prints the
effective configuration as a JSON object or using a Go template.

## Installation

Download a prebuilt release from the [releases]() page.

or

Use the [`go install`](https://golang.org/ref/mod#go-install) command to build and install
the command from source.

```bash
go install github.com/angrycub/parse-nomad-config
```

## Usage

```plaintext
usage: parse-nomad-config [options] [hcl-file ...]
  -j, --json              output in JSON format
  -o, --out string        write to the given file, instead of stdout
  -t, --template string   output using a go template
  -v, --version           show the version number and immediately exit
```

## Examples

Given the following Nomad configuration file--server.hcl:
```hcl
datacenter = "dc1"
data_dir   = "/opt/nomad/data"
log_level  = "DEBUG"

server {
  encrypt = "rn5HmQP/y1o2hvMxMfMYag=="
  enabled = true
}
```
### Print the effective configuration as JSON

```shell
$ parse-nomad-config --json server.hcl
```

Some of the fields are suppressed when printed in JSON format. For example,
the `server.encrypt` field is not included in JSON output. Public fields that
are not printed in JSON can be retrieved by using the Go Template output option.

These fields will have different capitalization and occasionally different names.
You can use the `"%+v"` format string to explore the Config structure or consult
the [nomad.command.agent.Config type][Config] for the full type specification.

### Retrieve the `encrypt` value using a Go Template

```shell
$ parse-nomad-config --template '{{ printf "%+v" .Server.EncryptKey}}' server.hcl
rn5HmQP/y1o2hvMxMfMYag==
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to
discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT][]

[Config]: https://github.com/hashicorp/nomad/blob/332dc88101fdaf6a819a37d63c3688dc2fac4bff/command/agent/config.go#L38
[MIT]: https://choosealicense.com/licenses/mit/
