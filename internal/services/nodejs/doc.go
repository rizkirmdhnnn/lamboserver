// Package node manages Node.js version installation and activation for
// LamboServer.
//
// The Manager type downloads Node.js releases from nodejs.org, maintains
// multiple installed versions under the LamboServer data directory, and
// activates a version by updating a symlink on the PATH. It uses FileSystem
// and CommandRunner interfaces for testability.
//
// This package does not manage npm packages or global Node.js modules.
package nodejs
