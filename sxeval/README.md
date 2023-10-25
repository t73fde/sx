# sxeval - an evaluator for symbolic expressions

Symbolic expressions may be interpreted, i.e. evaluated.

Most `sx.Object`s evaluate to themselves.

`sx.Symbol`s are looked up in an `sx.Environment` to evaluate to a stored value.
It is an error, if a symbol's value is not stored in the environment.

A non-empty list is treated differently: its first object must evaluate to a "callable" object, i.e. a function.
The other objects of the list are evaluated recursively and are treated as arguments for that function.

The first object of a list may alternatively evaluate to a "syntax" object, also a function.
The function is called with the other objects of the list as its arguments.
The result of the function call, typically a list, is then evaluated too.
This allows some kind of meta-evaluation.

"Callable" and "syntax" objects are defined in the package `sxbuiltins`.

Evaluation works in three steps:

1. The object is parsed according to the evaluation rules, resulting in an "expression" object (`sxeval.Expr`).
2. Expression objects may be "reworked", into possibly simpler expression objects.
   For example, if a symbols's value cannot be changed, the symbol lookup can be replaced with its value.
3. The expression is computed with respect to a given environment, resulting in an object.

This separation allowed to pre-compute the structure of an object, resulting in possibly faster execution time or less memory to store.
Parsing an reworking can be done in advance, while computing can be done much later.

To make the steps of evaluation easier to handle, `sxeval` defines an "engine" type (`sxeval.Engine`) that provides appropriate functions.
In addition, it provides function to create an initial environment to evaluate symbols.