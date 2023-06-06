# hr

This is a cooler HTTP router for Go. The implementation of the route tree and the way it works is slightly different than you may have seen in other famous libraries.

## Features

- Named route variables
- Group routes
- Request Binding
- Easy error handling
- Easy plugins (middlewares)

## Example

Check out [example](./example/).

## For Contributors

Contributors should read the content below carefully.

### Commit Title Format

Any commit title **MUST** follow one of

- `:sparkles:` Introducing new features.
- `:construction:` Work in progress.
- `:memo:` Updated documentations, including README.
- `:bug:` Fixed a bug.

For example, if you fixed a bug in the route tree, you must commit like

```bash
git commit -m ":bug: fixed a bug in route tree"
```

where `:bug:` will be displayed as an emoji on GitHub to better present what a commit is about.

## License

This library is MIT licensed.
