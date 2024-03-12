# sxbuiltins - a collections of predefined functions to work with symbolic expressions

This package provides a collection of functions that can be used when symbolic
expressions are evaluated (as defined in package `sxeval`).

## Callable

Callable functions allow to compute with objects. All arguments are evaluated
before the function is called with these evaluated values as arguments.

For example, `(+ 2 5 7)` calculates the sum of these three numbers, resulting
in the number value `14`. Similar, `*` is the symbol that evaluates to the
multiplication function, while the symbol `/` evaluates to the division
function. `(= a b)` returns a boolean "true" value, if the symbols `a` and `b`
evaluate to equal values.

## Syntax

Syntax functions provide a way to interpret some symbolic expression with a
special evaluation scheme. In most cases, it is not desired to evaluate all
arguments of a function.

For example, the expression `(if COND TRUE FALSE)` should evaluate the value
of `COND`. Depending on its value, either `TRUE` or `FALSE` must be evaluated,
but not true. Otherwise, the expression `(if (= a 0) 'error (/ 10 a))` will not
work if `a` is equal to the number zero, when all arguments to a hypothetical
`if` function are evaluated before calling the function. An error would be
raised.

An interesting syntax function is `(lambda PARAMS OBJ1 OBJ2 ...)`. When
evaluated, it creates a user-defined callable function. `ARGS` is typically
a list that specifies the parameters of the function, while `OBJ1`, `OBJ2`,
... are objects that are evaluated sequentially in the freshly created new
environment of that function. The result of the last object is the result of
the function call. The environment is first the mapping of the parameter names
to the argument values, and second the pre-existing environment where `lambda`
was executed.

A simple user-defined function is a function that adds the number seven to
its single argument: `(lambda (x) (+ c 7))`. Since the result of the `lambda`
function is a callable, you are allowed to put it as a first object into a
list: `((lambda (x) (+ x 7)) 10)`. This binds the number `10` to the symbol `x`
and calls the function, resulting in a value of `17`.

Other syntax functions provide a way to update the current evaluation
environment. `(defvar SYMBOL OBJ)` binds the value of `OBJ` to the symbol
`SYMBOL`. Later, the binding may be changed.

Since `(defvar add7 (lambda (x) (+ x 7)))` is a little verbose, there is a
simpler form: `(defun add7 (x) (+x 7))`. Then you can evaluate `(add7 10)`.

While a `lambda` creates an user-defined callable function, the is a form to
create an user-defined syntax function: `(defmacro NAME PARAMS OBJ1 ...)`. Such
a syntax function typically creates a list that will be evaluated separately.
For example, there is an user-defined syntax function that sequentially
evaluates its arguments and stops if a boolean "false" value is found: `(and
OBJ1 ...)`. If all arguments evaluate to a boolean "true" value, `and` returns
a "true" value, of course. Here is its definition (`T` is bound to the symbol
`T`):

    (defmacro and args
        (cond ((null? args)       T)
              ((null? (cdr args)) (car args))
              (T                  `(if ,(car args) (and ,@(cdr args))))))

`cond` and `if` are predefined syntax functions, `null?`, `cdr`, and `car` are
predefined callable functions.
