// Package boc is a Go implementation of [paper] "When Concurrency Matters: Behaviour-Oriented Concurrency" in OOPSLA2023.
//
//   - The core data structure is [CownPtr], which is a pointer to a cown.
//   - The core primitive is [When]. Use [When1], [When2], [When3], [When4], [WhenVec] if possible to leverage generic and type checking.
//   - Use [EnableTypeCheck] or set "TYPE_CHECK=1" as environment variable can also enable type checking.
//   - [TypeCheckHelper] are used to skip some operations except [When] statements.
//   - [TypeCheckWait] is used to wait for all [When] statements(goroutines spawned) to finish.
//
// This package is rewritten from [boc.rs].
// The benchmark is provided by [kaist-cp/cs431] and rewritten in Go.
//
// [boc.rs]: https://github.com/Lslightly-courses/cs431/blob/main/homework/src/boc.rs
// [paper]: https://dl.acm.org/doi/10.1145/3622852
// [kaist-cp/cs431]: https://github.com/kaist-cp/cs431
package boc
