# Cache

A simple and easy to use method for adding caching aspects to your Go code.
Everything revolves around a cache `Value`, which is a goroutine-safe, generic container for any value type.

## Features

* Cache whatever value you want in a type safe container, made possible by generics.
* Set time to live policies that make sense to you.
* Group multiple cached values together in a `MultiCache`.
* Cache the contents of a file and be notified of changes.
* Automate translations of file contents.

Check out [the package docs](https://pkg.go.dev/github.com/saylorsolutions/cache) for more information about how these functions work.
