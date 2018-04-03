# A Web server which accepts requests to run any predifined command

Copy configuration.example.json to configuration.json

```
cp configuration.example.json configuration.json
```

Start the server:

```
$ HCMD_SCRIPTS_DIR=scripts HCMD_CONFIG_FILE=configuration.json HCMD_TOKEN='a screct TOKEN 1112$$&&% what Noone will know!' go run main.go
```

Start a job *example* on the server

```
$ curl -H 'X-Hook-Token: a screct TOKEN 1112$$&&% what Noone will know!' -H 'X-Hook-Job: example' localhost:10292
randomjobidreturned
```

When a job started, a 16 char long job ID returned (`[a-z]{16}`), which can be used to retreive logs from the job:

Poll for job status:

```
$ curl -H 'X-Hook-Token: a screct TOKEN 1112$$&&% what Noone will know!'  localhost:10292/job/asdffdsajklhhjkl
```

Last message will contain `EOL` string, so you can grep for it to see if job run finished.


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

To generate a token you can try this:
`head -c 32 /dev/random | sha256sum`

Example:
`HCMD_TOKEN=a06c43b409c72be4cd8421da451f40e4a94cb53a0ff48fa233c10312437a5d41`


## Job configuration

The job configurations are stored in a json file, see configuration.example.json.

The following configuration options are available for a command:

- *job*: The name of the job. When you call the service, the value of the `X-Hook-Job` header must match with this value.
- *command*: The name of the executable
- *dir*: The name of the working directory for the command
- *env*: List of environment variables
