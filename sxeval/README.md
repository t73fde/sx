# sxeval - an evaluator for symbolic expressions

Symbolic expressions may be interpreted, i.e. evaluated.

Most `sx.Object`s evaluate to themselves.

`sx.Symbol`s are resolved in an `sxeval.Environment` to a bound value. If no
bound value is found, some actions are taken. See below for details.

A non-empty list is treated differently: its first object must evaluate to a
"callable" object, i.e. a function. The other objects of the list are evaluated
recursively and are treated as arguments for that function.

The first object of a list may alternatively evaluate to a "syntax" object,
also a function. The function is called with the other objects of the list
as its arguments. The result of the function call, typically a list, is then
evaluated too. This allows some kind of meta-evaluation.

"Callable" and "syntax" objects are defined in the package `sxbuiltins`.

Evaluation works in three steps:

1. The object is parsed according to the evaluation rules, resulting in an
   "expression" object (`sxeval.Expr`).
2. Expression objects may be "improved", into possibly simpler expression
   objects. For example, if a symbols's value cannot be changed, the symbol
   lookup can be replaced with its value.
3. The expression is computed with respect to a given environment, resulting in
   an object.

This separation allowed to pre-compute the structure of an object, resulting in
possibly faster execution time or less memory to store. Parsing an expression
can be done in advance, while computing can be done much later.

To make the steps of evaluation easier to handle, `sxeval` defines an
"environment" type (`sxeval.Environment`) that provides appropriate functions.
Its central attribute is the current "binding".

`sxeval.Binding`s are effectively just a mapping of `sx.Symbol`s to an
`sx.Object`. A `sx.Symbol` is *bound* to a `sx.Object`.

The are two types of bindings: a *constant binding* does not allow to update
the `sx.Object` that is bound to the `sx.Symbol`. A *variable binding* allows
this update.

`sxeval.Binding`s form a hierarchy: all but one have a *parent
binding*. This allows to overwrite constant bindings somehow: create a
child parent and bind the `sx.Symbol` to another `sx.Object`, and evaluate a
`sx.Object` in the new child binding.

Resolving a `sx.Symbol` works as follows: when a `sx.Symbol` is looked up in a
given environment, and it is not bound in that environments binding, the
`sx.Symbol` is resolved in the parent binding.

Of course, there is a binding that does not have a parent binding: the
*root binding*. If a `sx.Symbol` is not bound in the root binding, the
lookup operation fails.
