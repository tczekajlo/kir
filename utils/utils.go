package utils

var (
	// VERSION contains a version of build.
	// The value is autocompleted during the build process
	VERSION string

	// Banner contains banner
	Banner = ` __  ___  __  .______      
|  |/  / |  | |   _  \     
|  '  /  |  | |  |_)  |    
|    <   |  | |      /     
|  .  \  |  | |  |\  \---.
|__|\__\ |__| | _| ._____|` + VERSION + ` `
)
