package wcontext

// Key is our
type Key int

// ContextKeyRequestIP is the context key for our middleware that adds the request ip
var ContextKeyRequestIP = Key(1)

// ContextKeyReferer is the context key for our middleware that adds the referer
var ContextKeyReferer = Key(2)

// ContextKeyUserAgent is the context key for our middleware that adds the user agent
var ContextKeyUserAgent = Key(3)
