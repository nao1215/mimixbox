package ftp

// NewFtpget returns the ftpget applet (RETR / download). Its transfer logic and
// the shared FTP transport/session layer live in client.go; this file is the
// dedicated ftpget CLI surface.
func NewFtpget() *Command { return &Command{name: "ftpget", dir: dirGet} }
