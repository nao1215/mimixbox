package ftp

// NewFtpput returns the ftpput applet (STOR / upload). Its transfer logic and
// the shared FTP transport/session layer live in client.go; this file is the
// dedicated ftpput CLI surface.
func NewFtpput() *Command { return &Command{name: "ftpput", dir: dirPut} }
