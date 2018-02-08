# Understanding Golang Vendoring

We would like to understand how the golang vendor directory interacts with the
go compiler and runtime behaviour. Once we understand this, we can experiment
with tools that populate this directory, such as dep and godep.

If you want to follow along with these experiments, you can check out this git
repo, outside your normal `$GOPATH`.  This repo contains an `.envrc` file to
set your $GOPATH to this directory, whenever you `cd` in.

# Without Additional Tools

The branches in this git repo are experiments in what can happen when we vendor
things entirely by hand. We start with three packages:

- `a`
- `b`
- `c`

We have `a` depend on both `b` and `c`, and `b` depend on `c` like so:

![Dependency Diagram](https://g.gravizo.com/source/custom_mark10?https%3A%2F%2Fraw.githubusercontent.com%2Ftotherme%2Fgo-vendor-exploration%2Fmaster%2FREADME.md)
<details> 
<summary></summary>
custom_mark10
  digraph G {
    a -> b;
    b -> c;
    a -> c;
  }
custom_mark10
</details>

To begin with our `$GOPATH` contains only these three packages, and no
vendoring whatsoever.

## Static Types

In the `vendoring-static-types` branch, each dependency is an import of a
struct type. Package `c` defines a struct used in `a` and `b`, and package `b`
defines a struct used in `a`. In `a/main.go` we print out values of these
structs. If we can successfully `go run a/main.go`, then everything is
basically ok.

You can follow the experiment in the commit messages in that branch.

Here are some circumstances in which we discovered that things might break:
- If you vendor one of your dependencies, but not all of them
- If you vendor the vendor folder of one of your dependencies

## Runtime Behaviour

In the `vendoring-runtime-behaviour` branch, each dependency is the usage of a
function declared in the depended-upon package. Package `c` defines
`c.CFunc()`, which is used in `a` and `b`. Package `b` defines `b.BFunc()`,
which uses `c.CFunc()` and is used in `a`. Since `b.BFunc()` depends on the
behaviour of `c`, it has a unit test to check that behaviour.

Since we're not declaring different types in this experiment, we can do more
than we could in the static types experiment. In particular, note that once
we've made B vendor C, we can then change the behaviour of the global C,
without breaking the unit test in B. This is misleading however, since if we do
this: 
- part of the behaviour of A is changed by the change to C
- while part of it retains the behaviour of the old C, which persists vendored in B

This is definitely confusing behaviour, and is one more reason why if you
vendor one dependency, you should vendor all your dependencies. For full
details, read the code and commit messages in the `vendoring-runtime-behaviour`
branch.

In the `nested-vendoring-runtime-behaviour` branch, we cause the same confusing
behaviour, despite vendoring all the things. We manage this by vendoring the
vendor folder of `a`'s dependency `b`. In this way, we allow `a/vendor/c` to be
different from `a/vendor/b/vendor/c`, and are able to observe the mixed
behaviour from both copies of `c` in `a/main.go`.

This is also extremely confusing behaviour, and is one more reason why you
should never vendor a vendor folder. For full details, read the code and commit
messages in the `nested-vendoring-runtime-behaviour` branch.

# With One Tool
## With `dep`

## With `godep`

# With Multiple Tools
## When vendoring a `dep`-managed library into a `godep`-managed project
