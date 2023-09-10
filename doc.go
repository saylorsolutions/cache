/*
Package cache provides a simple and easy to use method for adding caching aspects to your Go code.
Everything revolves around a cache Value, which is a goroutine-safe, generic container for any value type.

# Creating a Value

A Value is created with a LoaderFunc (or a derivative) that determines how new values are loaded.
These functions may return an error, so a call to Value.Get also returns an error.

Different behaviors can be assigned to a Value depending on how they're created.

# New and NewEager

[New] creates a Value with the given LoaderFunc, and makes the Value lazily initialized.
If eager initialization is needed, then NewEager is provided for that purpose.

These functions cannot be initialized with a nil LoaderFunc, and will panic if one is passed.

# NewWithTTL

This function is similar to New, and is useful for cases where the time to live is determined by the process loading the cached value.
It accepts a LoaderTTLFunc that returns the time to live duration.

This function cannot be initialized with a nil LoaderTTLFunc, and will panic if one is passed.

# Internals

A Value stores a pointer to your cached value type, so it can easily determine if it's set or not.
A Value has other, optional attributes that may help it fit with your caching needs.

# Time to Live

By default, a cached value will not expire unless a call to Value.Invalidate is received.
However, to meet this need a Value may have a time to live duration set with [Value.SetTTL].
This allows a Value to automatically re-load itself when a call to Value.Get is received after the time to live duration has elapsed.

With a time to live duration set, a call to Value.Get will not refresh this time to live timer.
That behavior may be enabled with [Value.EnableGetTTLRefresh].

If you need to respond to a call to Value.Invalidate (but not timed expiration), then a handler function can be registered with Value.OnInvalidate.

If you no longer want a Value to have a time to live, then use Value.RemoveTTL.

# Caching multiple values

You may have a need to store many of the same values.
The MultiCache is provided for this use-case.

The MultiCache behaves similarly to the Value type, except that each underlying Value in a MultiCache is assigned a user-specified, comparable key.
A MultiCache can contain more [MultiCache]'s if a sort of hierarchy is desired.

To eagerly load values into a MultiCache, use MultiCache.Preheat with a set of keys.

Note that setting a TTL on a MultiCache sets that policy for all newly added Values.

Also, MultiCache doesn't provide a means of setting the underlying persistence where cached values are sourced. This is the role of the MultiLoaderFunc.
*/
package cache
