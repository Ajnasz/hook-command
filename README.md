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
curl -H 'X-hook-token: a screct TOKEN 1112$$&&% what Noone will know!' -H 'x-hook-job: example' localhost:10292
```
