/*
The logging package uses Hashicorp logutils to implement filtered logging: https://github.com/hashicorp/logutils.
The logging is directed to a file name 'log' in a directory named 'log' under the directory in which the app
is running (bin/server if the app is built using the supplied Make file.) The file and directory are created
if they don't exist - but only if logging is enabled.

Logging is enabled using an environment variable 'NBODYLOG' with a level value from any one of: 'DEBUG',
'INFO', 'WARN', or 'ERROR'. E.g.:

$ NBODYLOG=INFO bin/server

By default, logging is off. If an invalid log level is supplied, or no log level is supplied, then logging
is off. There is no logging in the client. Note: logging statements that don't contain the a filter specifier
like '[...] are always logged to stderr. (They are passed through the Hashicorp filter untouched.)
*/
package logging
