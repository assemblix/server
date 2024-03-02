package main

const ( // server
	port string = ":8080"

	appName   string = "AssemblixServer" // used for log directory
	logFormat string = "[dateTtime] [type] error"
)

const ( // website
	minimumUsernameLength int    = 3
	maximumUsernameLength int    = 20
	usernameRegexp        string = "^[a-zA-Z0-9_]+$"

	minimumPasswordLength int = 8
	maximumPasswordLength int = 72

	joinCash int = 20
)
