// Package intern implements fast, immutable string interning.
//
// The package is a cgo binding for libintern:
//
//	https://github.com/chriso/intern
//
// Interning is a way of storing distinct strings only once in memory:
//
//	https://en.wikipedia.org/wiki/String_interning
//
// Each string is assigned an ID of type uint32. IDs start at 1 and
// increment towards 2^32-1:
//
//	repository := intern.NewRepository()
//
// 	id := repository.intern("foo")
// 	fmt.Println(id) // => 1
//
// 	id := repository.intern("bar")
// 	fmt.Println(id) // => 2
//
// 	id := repository.intern("foo")
// 	fmt.Println(id) // => 1
//
// 	id := repository.intern("qux")
// 	fmt.Println(id) // => 3
//
// Two-way lookup is provided:
//
//  if id, ok := repository.Lookup("foo"); ok {
//    fmt.Printf("string 'foo' has ID: %v", id)
//  }
//
//  if str, ok := repository.LookupID(1); ok {
//    fmt.Printf("string with ID 1: %v", str)
//  }
//
// This package is *NOT* safe to use from multiple goroutines without
// locking, e.g. https://golang.org/pkg/sync/#Mutex
package intern

// #include <intern/strings.h>
// #include <intern/optimize.h>
// #cgo LDFLAGS: -lintern
import "C"

import (
	"fmt"
	"runtime"
)

// ErrInvalidSnapshot is returned by Repository.Restore when the
// repository and snapshot are incompatible
var ErrInvalidSnapshot = fmt.Errorf("invalid snapshot")

// Repository stores a collection of unique strings
type Repository struct {
	ptr *C.struct_strings
}

// NewRepository creates a new string repository
func NewRepository() *Repository {
	ptr := C.strings_new()
	return newRepositoryFromPtr(ptr)
}

func newRepositoryFromPtr(ptr *C.struct_strings) *Repository {
	if ptr == nil {
		outOfMemory()
	}
	repo := &Repository{ptr}
	runtime.SetFinalizer(repo, (*Repository).free)
	return repo
}

func outOfMemory() {
	panic("out of memory")
}

func (repo *Repository) free() {
	C.strings_free(repo.ptr)
}

// Count returns the total number of unique strings in the repository
func (repo *Repository) Count() uint32 {
	return uint32(C.strings_count(repo.ptr))
}

// Intern interns a string and returns its unique ID. Note that IDs increment
// from 1. This function will panic if the string does not fit in one page:
// len(string) < repo.PageSize()
func (repo *Repository) Intern(str string) uint32 {
	id := uint32(C.strings_intern(repo.ptr, C.CString(str)))
	if id == 0 {
		outOfMemory()
	}
	return id
}

// Lookup returns the ID associated with a string, or false if the ID
// does not exist in the repository
func (repo *Repository) Lookup(str string) (uint32, bool) {
	id := uint32(C.strings_lookup(repo.ptr, C.CString(str)))
	return id, id != 0
}

// LookupID returns the string associated with an ID, or false if the string
// does not exist in the repository
func (repo *Repository) LookupID(id uint32) (string, bool) {
	str := C.strings_lookup_id(repo.ptr, C.uint32_t(id))
	if str == nil {
		return "", false
	}
	return C.GoString(str), true
}

// AllocatedBytes returns the total number of bytes allocated by the string
// repository
func (repo *Repository) AllocatedBytes() uint64 {
	return uint64(C.strings_allocated_bytes(repo.ptr))
}

// Cursor creates a new cursor for iterating strings
func (repo *Repository) Cursor() *Cursor {
	cursor := _Ctype_struct_strings_cursor{}
	C.strings_cursor_init(&cursor, repo.ptr)
	return &Cursor{repo, &cursor}
}

// Optimize creates a new, optimized string repository which stores the most
// frequently seen strings together. The string with the lowest ID (1) is the
// most frequently seen string
func (repo *Repository) Optimize(freq *Frequency) *Repository {
	ptr := C.strings_optimize(repo.ptr, freq.ptr)
	return newRepositoryFromPtr(ptr)
}

// Snapshot creates a new snapshot of the repository. It can later be
// restored to this position
func (repo *Repository) Snapshot() *Snapshot {
	snapshot := _Ctype_struct_strings_snapshot{}
	C.strings_snapshot(repo.ptr, &snapshot)
	return &Snapshot{repo, &snapshot}
}

// Restore restores the string repository to a previous snapshot
func (repo *Repository) Restore(snapshot *Snapshot) error {
	if ok := C.strings_restore(repo.ptr, snapshot.ptr); !ok {
		return ErrInvalidSnapshot
	}
	return nil
}

// PageSize returns the compile-time page size setting
func (repo *Repository) PageSize() uint64 {
	return uint64(C.strings_page_size())
}

// Snapshot is a snapshot of a string repository
type Snapshot struct {
	repo *Repository
	ptr  *C.struct_strings_snapshot
}

// Cursor is used to iterate strings in a repository
type Cursor struct {
	repo *Repository
	ptr  *C.struct_strings_cursor
}

// ID returns the ID that the cursor currently points to
func (cursor *Cursor) ID() uint32 {
	return uint32(C.strings_cursor_id(cursor.ptr))
}

// String returns the string that the cursor currently points to
func (cursor *Cursor) String() string {
	str := C.strings_cursor_string(cursor.ptr)
	if str == nil {
		return ""
	}
	return C.GoString(str)
}

// Next advances the cursor. It returns true if there is another
// string, and false otherwise
func (cursor *Cursor) Next() bool {
	return bool(C.strings_cursor_next(cursor.ptr))
}

// Frequency is used to track string frequencies
type Frequency struct {
	ptr *C.struct_strings_frequency
}

// NewFrequency creates a new string frequency tracker
func NewFrequency() *Frequency {
	ptr := C.strings_frequency_new()
	if ptr == nil {
		outOfMemory()
	}
	freq := &Frequency{ptr}
	runtime.SetFinalizer(freq, (*Frequency).free)
	return freq
}

func (freq *Frequency) free() {
	C.strings_frequency_free(freq.ptr)
}

// Add adds a string ID. This should be called after interning a string and
// getting back the ID
func (freq *Frequency) Add(id uint32) {
	if ok := C.strings_frequency_add(freq.ptr, C.uint32_t(id)); !ok {
		outOfMemory()
	}
}

// AddAll adds all string IDs, to ensure that each string is present in the
// optimized repository
func (freq *Frequency) AddAll(repo *Repository) {
	if ok := C.strings_frequency_add_all(freq.ptr, repo.ptr); !ok {
		outOfMemory()
	}
}
