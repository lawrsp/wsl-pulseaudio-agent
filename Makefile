all:
	env GOOS=windows go build -ldflags "-H=windowsgui"
