// Package nginx manages Nginx installation and LaunchDaemon lifecycle for local
// reverse-proxy and static file serving.
//
// Config generation (nginx.conf, fastcgi_params, mime.types, default site) is
// implemented as Manager methods in manager.go so they can use the injected
// FileSystem interface rather than direct os.* calls.
package nginx
