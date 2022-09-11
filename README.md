# Chrome bookmarks to markdown

[![Latest version](https://img.shields.io/github/v/tag/daishe/chrome-bookmarks-to-markdown?label=latest%20version&sort=semver)](https://github.com/daishe/chrome-bookmarks-to-markdown/releases)
[![Latest release status](https://img.shields.io/github/workflow/status/daishe/chrome-bookmarks-to-markdown/Release?label=release%20build&logo=github&logoColor=fff)](https://github.com/daishe/chrome-bookmarks-to-markdown/actions/workflows/release.yaml)

[![Go version](https://img.shields.io/github/go-mod/go-version/daishe/chrome-bookmarks-to-markdown?label=version&logo=go&logoColor=fff)](https://golang.org/dl/)
[![License](https://img.shields.io/github/license/daishe/chrome-bookmarks-to-markdown)](https://github.com/daishe/chrome-bookmarks-to-markdown/blob/master/LICENSE)

A simple CLI utility to convert Chrome bookmarks to markdown format.

## Usage

Just run the CLI ad it will produce chrome bookmarks for all profiles in markdown format:

```sh
chrome-bookmarks-to-markdown
```

If you wish to save returned document to file instead of printing to stdout use `--output` flag:

```sh
chrome-bookmarks-to-markdown --output 'path/to/store/generated/document.md'
```

You can also limit profiles with `--profiles` flag:

```sh
chrome-bookmarks-to-markdown --profiles 'Default,Profile 1'
```

and override default path with Chrome configuration with `--input` flag.

That's it!

## Help

To get the complete list of all flags, use

```sh
chrome-bookmarks-to-markdown --help
```

## License

Chrome bookmarks to markdown is open-sourced software licensed under the [Apache License 2.0](http://www.apache.org/licenses/).
