package main

import "github.com/T4ko0522/spotify-cli/cmd"

var version = "dev"

func main() {
	cmd.Execute(version)
}
