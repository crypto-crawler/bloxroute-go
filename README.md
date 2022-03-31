# bloxroute-go

A bloXroute websocket client written in Go.

This library provides typesafe APIs, not just a raw `chan string` channel. For example, `SubscribeNewTxs()` outputs messages to a `chan *types.Transaction` channel.
