/*
Package file provides some easy to use methods for caching file information.

The contents of a file may be cached, with automatic invalidation provided with the [github.com/fsnotify/fsnotify] library.
The actual file bytes may be used with [NewFileCache] or [NewEagerFileCache].

Alternatively, reading the file's contents and decoding it can be combined into a single function with [NewReaderCache].
The Value returned will store the decoded form for easy retrieval.

If you need to perform some action in response to the file being changed, then use OnInvalidate on the Value returned from any of these functions.
*/
package file
