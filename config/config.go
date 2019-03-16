package config

import (
	"flag"
	"regexp"
)

var (
	WebsocketPort    int
	WebhookPort      int
	AppServerAddress string
	BoundAddress     string
	Debug            bool
)

func init() {
	flag.IntVar(&WebsocketPort, "websocket-port", 5568, "port number for websocket (default: 5568)")
	flag.IntVar(&WebhookPort, "webhook-port", 5580, "port for webpush (default: 5580)")
	flag.StringVar(&AppServerAddress, "app", "", "app server address")
	flag.StringVar(&BoundAddress, "bind", "localhost", "bound addresss (default: localhost)")
	flag.BoolVar(&Debug, "debug", false, "debug?")
	flag.Parse()
	rep := regexp.MustCompile(`(/)$`)
	AppServerAddress = rep.ReplaceAllString(AppServerAddress, "")
}
