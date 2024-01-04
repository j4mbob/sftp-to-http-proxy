# SFTP to HTTP(S) go Proxy

can be used to faciliate ZTP for switches that support SFTP as a transport mechanism but not HTTP(S)

sets up a real sftp server with custom handlers for operations which allows us to back off requests to a web server

this works by tricking the sftp client into believing the file they requested is on the sftp servers local file system. the proxy handles the request and pulls the file 
from a remote http(s) server and serves up the file over the established SFTP session back to the client

intended to run as a systemd service on a server so supports PID file handling and concurrent sftp clients to handle multiple switches being simultanously deployed

```
  -listenip string
    	IP for SFTP server to bind to (default "0.0.0.0")
  -listenport string
    	port for SFTP server to listen on (default "2122")
  -loadconfig string
    	load json config file (default "none")
  -password string
    	password to use for authentication (default "sftp")
  -pyroscope
    	enable sending application metrics to pyroscope host
  -pyroscopehost string
    	remote pyroscope to send application metrics to (default "http://pyroscope-host:4040")
  -remoteurl string
    	remote web server to send requests to (default "http://remote-webserver:8080")
  -sslkey string
    	ssl private key to use (default "id_rsa")
  -username string
    	username to use for authentication (default "sftp")

```

logs to stdout and includes status of file requests and duration transfers took so we get some visability over what the switch is doing:

```
2024/01/02 10:32:28 loading JSON config: config.json
2024/01/02 10:32:28 sending application metrics to remote pyroscope host: http://pyroscope:4040
2024/01/02 10:32:28 sftp proxy listening on 10.20.58.1:2122
2024/01/02 10:32:39 successful login from: 10.20.58.2:65427
2024/01/02 10:32:43 client 10.20.58.2:65427 attempting to get: http://remote-url:8080/S5735-L-V2_V600R022C10SPC500.cc
2024/01/02 10:33:28 proxy downloaded: http://remote-url:8080/S5735-L-V2_V600R022C10SPC500.cc for client 10.20.58.2:65427 duration: 44.496820739s
```
