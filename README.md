# A Web server which accepts requests to run any predifined command

Copy configuration.example.json to configuration.json

```
cp configuration.example.json configuration.json
```

Run the server

```
SCM_TOKEN='a screct TOKEN 1112$$&&% what Noone will know!' go run main.go
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
