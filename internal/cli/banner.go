package cli

import (
	"fmt"
)

const banner = `
	    ___              __             
	   /   |  __________/ /_  ___  _____
	  / /| | / ___/ ___/ __ \/ _ \/ ___/
	 / ___ |/ /  / /__/ / / /  __/ /    
	/_/  |_/_/   \___/_/ /_/\___/_/ v2.0.0
`
const Version = "v2.0.0"

func ShowBanner() {
	fmt.Printf("%s\n", banner)
	fmt.Print("\t  Distributed scanner for the masses\n\n")
}
