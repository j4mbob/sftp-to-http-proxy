SFTP to HTTP(s) go Proxy

translates sftp client file get requests to http get requests
```
  -listenip string
    	IP for SFTP server to bind to (default "0.0.0.0")
  -listenport string
    	port for SFTP server to listen on (default "2022")
  -loadconfig string
    	load json config file (default "none")
  -password string
    	password to use for authentication (default "sftp")
  -pyroscope
    	enable sending application metrics to pyroscope host
  -pyroscopehost string
    	remote pyroscope to send application metrics to (default "http://grafana.networks-util.ask4.net")
  -remoteurl string
    	remote web server to send requests to (default "http://grafana.networks-util.ask4.net:8080")
  -sslkey string
    	ssl private key to use (default "id_rsa")
  -username string
    	username to use for authentication (default "sftp")

```

logs to stdout and includes status of file requests and duration transfers took so we get some visability over what the switch is doing:

```
2023/12/22 13:17:51 sftp proxy listening on 0.0.0.0:2022
2023/12/22 13:18:04 successful login from: [::1]:55032
2023/12/22 13:18:07 client [::1]:55032 attempting to get: http://grafana.networks-util.ask4.net:8080/test.file
2023/12/22 13:18:07 error getting file: 404 Not Found
2023/12/22 13:18:25 client [::1]:55032 attempting to get: http://grafana.networks-util.ask4.net:8080/S5730HI-V200R019C00SPC500.cc
2023/12/22 13:22:57 client [::1]:55032 downloaded: http://grafana.networks-util.ask4.net:8080/S5730HI-V200R019C00SPC500.cc duration: 4m31.636666625s
```
