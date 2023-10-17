/*
Package buffered introduces double-buffered cache support.
This trades memory for throughput by way of reduced chance of lock contention, and allows directly setting a value in a [MultiCache].

A buffered MultiCache will maintain an internal read cache and write buffer.

If there's a cache miss on a [MultiCache.Get], then the read cache will pass through to the write buffer.
The write buffer is where a new value is set with [MultiCache.Set].
This implicitly invalidates the matching key in the read cache.

With this setup and good cache hit rates, there can be near zero contention between reading and writing values.
*/
package buffered
