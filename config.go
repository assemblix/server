package main

const ( // server
	port string = ":8080"

	appName string = "AssemblixServer" // used for log directory
)

const ( // website
	minimumUsernameLength int    = 3
	maximumUsernameLength int    = 20
	usernameRegexp        string = "^[a-zA-Z0-9_]+$"

	minimumPasswordLength int = 8
	maximumPasswordLength int = 72

	recaptchaSecret string = "6LcNHVUnAAAAAICv4oKEzhh6UTHk3QraFDfdde01"

	joinCash int = 20
)

// const ( // debugging
// 	debugPrefix   string = "\033[0;33m[DEBUG]\033[0m"
// 	errorPrefix   string = "\033[0;91m[ERROR]\033[0m"
// 	warningPrefix string = "\033[0;93m[WARNING]\033[0m"
// )
