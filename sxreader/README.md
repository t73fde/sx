# sxreader - a reader for symbolic expressions

This packages provides a reader to transform the textual representation of an symbolic expression into one or more internal `sx.Object`s.

`sxreader.Reader` uses a value that implements the `io.Reader` interface to read the textual representation.

Currently, the reader supports these representations:

* `( ... )` is transformed into a `sx.Pair`.
  Before the last element, a dot `.` is allowed to signal an improper list.
  A dot outside a list is not allowed.
* `" ... "` is transformed into a `sx.String`.
  Escape sequences, like `\"`, `\n`, `\x2a`, `\u2a2a`, or `\U2a2a2a`, are allowed to specify special Unicode code points.
* A sequence of digits `0` ... `9`, optional starting with a plus `+` or a minus `-` character is transformed into a `sx.Number`.
  Currently, `sx.Int64` is the only type supported.
  This will change.
* `'OBJ` is transformed into `(quote OBJ)`, ```OBJ`` into `(quasiquote OBJ)`, `,OBJ` into `(unquote OBJ)`, and `,@OBJ` into `(unquote-splicing OBJ)`.
  Other read macros are not supported.
* `; ...` is ignored until the end of the current line.
  Therefore, this works as a comment to the human reader.
* A printable sequence of other Unicode code points is transformed into a `sx.Symbol`.

Textual representations starting with `#` must not be used as a `sx.Symbol`, because such representation may encode other `sx.Object`s in the future.

When creating a `sxreader.Reader`, some options can be specified to customize the behaviour of the reader:

* `WithSymbolFactory(sx.SymbolFactory)` specifies a symbol factory to be used to create symbols.
  If no symbol factory is specified, or `WithDefaultSymbolFactory` is used, an internal created, reader specific symbol factory is used.
  Please note, that two symbols always differ, if they are created by different symbol factories.
* `WithListLimit(uint)` specifies a maximum list length.
  It should be used, if the reader reads from external sources to mitigate DOS attacks.
  If not specified, the value of `sxreader.DefaultListLimit` is used.
* `WithNestingLimit(uint)` specifies the maximum nesting of lists, for the same reasons.
  Since many functions works on nested lists recursively, you should use a value that does not exceed the available stack space.
  If not specified, the value of `sxreader.DefaultNestingLimit` is used.
