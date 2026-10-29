# Contributing

Only Golang and a container runtime are required that kind can work with.
Most tools required during development are version controlled and placed in `hack/tools/`.

Commits must be signed off using a `Signed-off-by` trailer, see the
documentation from the Linux Foundation for details:
<https://bestpractices.linuxfoundation.org/ip/contribution-mechanisms-dco.html>

Every file that isn't generated must carry the license headers.
The `check-fix` recipe documented below automatically adds these.

## Setup

1. Run `make tools` to download the tools
2. Start developing

## Testing

The central gate is the `check` recipe, which runs all relevant checks in sequence and is also used as the CI gate.
The `check-fix` recipe also lets the tools fix what they can (e.g. missing headers, linter errors).

In general `check-fix` is what should be used during development.
