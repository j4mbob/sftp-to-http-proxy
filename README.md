SFTP to HTTP(s) go Proxy

translates sftp client file get requests to http get requests

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
