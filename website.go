package main

import "github.com/overflow3d/website/route"

func main() {
	//Run the server and fire up routes
	route.Run(route.DoRoutes())
}
