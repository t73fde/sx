# sx - symbolic expressions

Sx is a collection of libraries and frameworks to work with
[s-expressions](https://en.wikipedia.org/wiki/S-expression).

See for more information on its [project page](https://zettelstore.de/sx).

## Types

Sx support the following atomic, immutable types:

* **Numbers** contain numeric values.
  Currently, only integer values are supported. There is no maximum or minimum
  integer value. They optionally start with a `+` or `-` sign and contain only
  digits `0`, ..., `9`.
* **Strings** are UTF-8 encoded Unicode character sequences.
  They are delimited by `"` characters. Special characters inside the string,
  like the `"` character itself, are escaped by the `\` character.
* **Symbols** are sequences of printable / visible Unicode characters.
  They are typically used to bind them to values within an environment. Another
  use case is symbolic computation.

Sx supports nested lists. A list is delimited by parentheses: `( ... )`. Within
a list, all values are separated by space characters, including new line. Lists
can be nested. Internally, lists are created by pairs. The first part of a
pair, called "car", contains the actual value stored at the beginning of a
list. The second part, called "cdr", typically links to the next pair. This
allows multiple lists to share elements. A proper list is a list, where the
last elements second part is the empty list. The last element of a list may be
a pair, where the second part references a values except a list. Such lists are
improper lists. Since the second part may reference any value, even earlier
elements of a list, lists may be circular. Single pairs are denoted as `(X .
Y)`, where the car references S and the cdr references Y (Y is not a list).

All other types supported by Sx cannot be specified via the reader.

* **Vector** is a mutable sequence of values, to be used if direct access to
  values of a longer sequence is needed. A list has only O(n) access. Vectors
  are typically more memory efficient, compared to pair lists. However, they
  cannot be process recursively and they are not able to share elements.
* **Undefined** contains just the _undefined_ value.
  It is signalled by some functions that should not abort with an error.

There is no special **boolean** value. The empty list `()`, the empty string
`""`, and the undefined value are considered as a "false" value. All other
values are treated as a "true" value. Most functions that return a boolean
value currently return the empty list to signal a "false" value or return
either the number `1` or the symbol `T` as a "true" value.

Vectors and lists are both **sequence**s and share some common functions /
methods. You can calculate the length of a sequence, check for a length less
than a value, fetch the n-th element of a sequence, convert a sequence to a
pair list, and iterate over the elements in a ordered way.
