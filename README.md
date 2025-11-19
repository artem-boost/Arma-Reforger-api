## About

This is a backend for Arma Reforger written in GO and using a SQLite database.

## Get started
1. [Download Last Release](https://github.com/artem-boost/Arma-Reforger-api/releases/latest)
2. Create a new folder
3. Extract the archive into a folder
4. Move client files from the [Tools](Tools/Client) folder to the game folder
5. Move server files from the [Tools](Tools/Server) folder to the Dedicated Server folder (optional)
6. Execute program
7. Enjoy the game

## Change API address

You can set an ip address API other than localhost in the [client configuration file](Tools/Client/ClientConfig.json) and in [server configuration file](Tools/Server/ServerConfig.json)
```
{
	"Master":{
		"URL": "http://127.0.0.1:6122" # Change this
	},
	"Server":{
		"EnableProxy": false,
		"ProxyAddr": "127.0.0.1:8888",
		"VisibleInLicense": false
	}
}
```
