# A Web server which accepts requests to run any predifined command

Copy configuration.example.json to configuration.json

```
cp configuration.example.json configuration.json
```

Run the server

```
HCMD_SCRIPTS_DIR=scripts HCMD_CONFIG_FILE=configuration.json HCMD_TOKEN='a screct TOKEN 1112$$&&% what Noone will know!' go run main.go
```

And call the server:

```
curl -H 'X-Hook-Token: a screct TOKEN 1112$$&&% what Noone will know!' -H 'X-Hook-Job: example' localhost:10292
```


## Configuration

You can configure the service by setting environment variables:


### HCMD_PORT

The port where the server should listen. The default is `10292`

Example:
`HCMD_PORT=9322`


### HCMD_CONFIG_FILE

Path to the command configurations.
See _configuration.exmple.json_ file

Example:
`HCMD_CONFIG_FILE=/etc/hcmd.json`


### HCMD_SCRIPTS_DIR

The directory where the commands are stored.

Example:
`HCMD_SCRIPTS_DIR=/usr/local/share/hcmd`


### HCMD_TOKEN

A random string to authorize access to the commands. Use very long token to prevent brute force attacks

Example:
`HCMD_TOKEN=2292kjdfkjasdf923rkjlakjfd0239jalkdsjf201laanb56jjxxwq`


## Job configuration

The job configurations are stored in a json file, see configuration.example.json.

The following configuration options are available for a command:

- *job*: The name of the job. When you call the service, the value of the `X-Hook-Job` header must match with this value.
- *command*: The name of the executable
- *dir*: The name of the working directory for the command
- *env*: List of environment variables
