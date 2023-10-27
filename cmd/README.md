# cmd - an example interactive shell to experiment with s-expressions

This package contains some example code that shows how to work with the
packages defined in this repository. All predefined callable and syntax
functions are available.

The shell produces some kind of verbose logging when it is evaluating symbolic
expression. To control this logging, some additional functions are available:

* `(log-reader)`: toggles the logging of the result of the `sxreader.Read()`
  call. The log output is prefixed with the string ";r ".
* `(log-parser)`: toggles the logging of the result of the `sxeval.Parse()`
  call. The log output is prefixed with the string ";p ". In addition, the
  internal parsing steps are aslo logged. The symbolic expression to be parsed
  is logged with the prefix ";P ", the resulting parsed expression with the
  prefix ";Q ".
* `(log-expr)`: enables logging of the result of calls to `sxeval.Rework`.
  Further processing is stopped, the resulting expression is not computed.
  This is to inspect the output of a `(quasiquote ...)` expression.
* `(log-executor)`: toggles the logging of the computing execution. The
  expression to be computed is logged with the prefix ";X ", the resulting
  value with the prefix ";O ".
* `(log-off)`: All loggings, except the rework log, are enabled. This command
  disables all logging.
* `(panic EXPR)`: produces an internal panic (a hard error within the Go
  environment). The optional expression is the argument. If not given, the
  string "common panic" is used instead.

Some symbols are bound to specific objects:

* `root-env` is bound to the root environment, where most predefined functions
  are defined.
* `repl-env` is bound to the environment that is used in the shell. The root
  environment is a (possible indirect) parent environment.

When the shell starts, the content of the file "prelude.sxn" is evaluated
before any input is read from the user (via the reader).
