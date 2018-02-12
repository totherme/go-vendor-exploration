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

In these experiments, we make use of two additional github repos: one for
[module B](https://github.com/totherme/govendorexplorationb), and one for
[module C](https://github.com/totherme/govendorexplorationc). We submodule
those into this `$GOPATH` repo as if we had performed a `go get`, and we
experiment with using `dep` to vendor them in to our program in `a`.

In our simplest experiment in branch `vendoring-with-dep`, we follow the same
path we followed in the `vendoring-runtime-behaviour` experiment, only using
`dep` to perform our vendoring, instead of doing it by hand. In this case `dep`
actually saves us from some unexpected behaviour, by ignoring the `$GOPATH`,
and always preferring the default branch of whatever it fetches unless we
explicitly ask for something else. Since `dep` flattens all imports, we are
unable to perform an equivalent of the `nested-vendoring-runtime-behaviour`
experiment.

In `vendoring-specific-versions-with-dep` we learn how to cause `dep ensure` to
fail. We ask it for constraints which cannot be solved. We then relax the
constraints of the importing module (module `a`), to allow `dep` to find a
solution. In `vendoring-deliberately-vague-versions-with-dep` we solve the same
problem, but this time by relaxing the constraints of the imported module (in
this case `b`).

Relaxing either constraint seems to be good enough for `dep`. If you have
control over either importing or imported modules, and you are able to ensure
compatibility with a wide range of common dependencies, it is probably worth
explicitly relaxing your constraints in your `Gopkg.toml`

Finally, in the last commit of
`vendoring-deliberately-vague-versions-with-dep`, we check that having 0
constraints on a dependency does not mean pretending we don't have that
dependency at all. It just means that `dep` can pick whatever version of that
dependency it wants.

## With `godep`

`godep` will, unlike `dep` always use the packages as they are in the
`$GOPATH`, in the exactly version as they are checked out there.

In `vendoring-with-godep` we show that, if you you have a dependency which uses
`godep`, it is very important to do a `godep restore` to make sure you have the
same versions of transient dependencies in the `$GOPATH` as they exist in a
dependency's vendor tree. If you don't do that and you start vendoring a
dependency with a vendor tree, `godep` will just go and use whatever version it
finds in `$GOPATH` and generates a flat vendor tree (no nested vendor
directories).

If it cannot find a dependency in your `$GOPATH` because you only have it in a
vendor tree of some dependency, `godep save` will just bails out.

Also `godep` does not warn you if you have different versions in a dependency's
vendor tree and in your `$GOPATH`. It does not do checks, again, it just uses
what's available in the `$GOPATH`.

# With Multiple Tools

## When vendoring a `dep`-managed library into a `godep`-managed project

Vendoring the `dep`-managed package into the `godep` managed package is shown
in `vendoring-dep-project-into-godep-project`.

When initially vendoring the `dep` package into a `godep` package, `godep`
completely removes the vendor directory and from the dependency and generates a
flat vendor tree. However, `gedep` of course has no idea about versioning of
dependencies in the `dep`-managed package. Thus, when flattening the vendor
tree `godep` picks again whatever version of dependencies are checked out in
`$GOPATH`, regardless of which version would have actually been configured by
`dep`.
This means: the `dep`-managed package has to be super flexible on which
versions of its dependencies it eventually ends up with. It's not in the
control of `dep` anymore -- and `godep` just uses the version from the
`$GOPATH`. It might be necessary to test various versions of your dependencies
in your CI.

Updating dependencies with `godep update ...` seems [to be
broken](https://github.com/tools/godep/issues/498). In our experiment in
`vendoring-dep-project-into-godep-project-fail` we tried to do so, and see that
`godep` generates a nested vendor tree and vendors two versions of the
dependency which is also in the vendor tree of another package. It vendors once
from the `$GOPATH` and once from that other vendor directory. As described, it
does not flatten the vendor tree anymore. That means that it probably is
easiest, to completely delete the `vendor` directory and vendor from scratch
with a `godep save ...`.

# `godep` & `git` Submodules

When setting up these test repositories, at some point in time we did a `go get
...` of our dependencies, module A & module B. This essentially placed a git
repository somewhere into our `$GOPATH`, which itself is a git repository.
Therefore we used `git submodule add ...` to "link" the dependencies with the
main repository. Interestingly enough we now ended up with a full blown git
repository for the submodules with a `.git` directory instead of the expected
`.git` file with just a pointer to the main repository's git directory.
However, everything still seemed to have worked as expected so we didn't even notice.

Then, after some experiments and changing, deleting and re-init of submodules
we eventually ended up with a mixed case where one submodule had a real
directory in `.git`, the other one had a file in `.git`.
Now things started to become strange, `godep` seemed to have been broken where
it just worked before, without no notable difference -- git would have told us
if things would have changed; or so we thought.

Eventually it turned out, that `godep` checks if a dependency is actually of
some sort of VCS. It actually checks [if a certain directory is
present](https://github.com/tools/godep/blob/ce0bfadeb516ab23c845a85c4b0eae421ec9614b/vendor/golang.org/x/tools/go/vcs/vcs.go#L350).
Now that git submodules normally won't have those `.git` directories, `godep`
won't work with dependencies that are submoduled in. However, as git seems to
be happy when a submodule has a `.git` directory, one can make `godep` and
submodules work together.
We do not recommend that, because a user needs to do extra steps after cloning,
however for the sake of this exploration it should be OK.
