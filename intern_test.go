package intern

import (
	"fmt"
	"testing"
)

func TestIntern(t *testing.T) {
	repo := NewRepository()
	if repo.Intern("foo") != 1 ||
		repo.Intern("bar") != 2 {
		t.Error("invalid Intern() result")
	}
	if repo.Intern("qux") != 3 ||
		repo.Intern("foo") != 1 ||
		repo.Intern("bar") != 2 {
		t.Error("invalid Intern() result")
	}
}

func TestLookup(t *testing.T) {
	repo := NewRepository()
	repo.Intern("foo")
	if id, ok := repo.Lookup("foo"); !ok || id != 1 {
		t.Error("invalid Lookup() result")
	}
	if _, ok := repo.Lookup("bar"); ok {
		t.Error("invalid Lookup() result")
	}
}

func TestAllocatedBytes(t *testing.T) {
	repo := NewRepository()
	if repo.AllocatedBytes() == 0 {
		t.Error("invalid AllocatedBytes() result")
	}
}

func TestLookupID(t *testing.T) {
	repo := NewRepository()
	repo.Intern("foo")
	if str, ok := repo.LookupID(1); !ok || str != "foo" {
		t.Error("invalid Lookup() result")
	}
	if _, ok := repo.LookupID(2); ok {
		t.Error("invalid Lookup() result")
	}
}

func TestCount(t *testing.T) {
	repo := NewRepository()
	if repo.Count() != 0 {
		t.Error("invalid Count() result")
	}
	repo.Intern("foo")
	if repo.Count() != 1 {
		t.Error("invalid Count() result")
	}
	repo.Intern("foo")
	if repo.Count() != 1 {
		t.Error("invalid Count() result")
	}
	repo.Intern("bar")
	if repo.Count() != 2 {
		t.Error("invalid Count() result")
	}
}

func TestLargeRepository(t *testing.T) {
	count := 100000

	repo := NewRepository()
	startSize := repo.AllocatedBytes()
	for i := 1; i <= count; i++ {
		str := fmt.Sprintf("%d", i)
		id := repo.Intern(str)
		if int(id) != i {
			t.Error("invalid Intern() result")
		}
		if id != repo.Intern(str) {
			t.Error("Intern() is not idempotent")
		}
		if lookupID, ok := repo.Lookup(str); !ok || lookupID != id {
			t.Error("invalid Lookup() result")
		}
		if lookupStr, ok := repo.LookupID(id); !ok || lookupStr != str {
			t.Error("invalid LookupID() result")
		}
	}
	endSize := repo.AllocatedBytes()

	if startSize >= endSize {
		t.Error("invalid AllocatedBytes() result")
	}

	if int(repo.Count()) != count {
		t.Error("invalid Count() result")
	}

	cursor := repo.Cursor()
	i := 0
	// the cursor is invalid until Next() is called
	if cursor.String() != "" || cursor.ID() != 0 {
		t.Error("invalid cursor position")
	}
	for cursor.Next() {
		i++
		str := fmt.Sprintf("%d", i)
		if int(cursor.ID()) != i || cursor.String() != str {
			t.Error("invalid cursor position")
		}
	}
	if i != count {
		t.Error("invalid cursor operation(s)")
	}
	// the cursor is now invalid
	if cursor.String() != "" || cursor.ID() != 0 {
		t.Error("invalid cursor position")
	}
}

func assertStrings(t *testing.T, repo *Repository, expected []string) {
	if len(expected) != int(repo.Count()) {
		t.Error("unexpected count")
	}
	for i, str := range expected {
		if id, ok := repo.Lookup(str); !ok || int(id) != i+1 {
			t.Error("invalid repository")
		}
	}
	if _, ok := repo.LookupID(uint32(len(expected) + 1)); ok {
		t.Error("unexpected extra strings in repository")
	}
}

func TestSnapshotRestore(t *testing.T) {
	repo := NewRepository()
	start := repo.Snapshot()
	repo.Intern("foo")
	repo.Intern("bar")
	mid := repo.Snapshot()
	repo.Intern("qux")
	repo.Intern("xyz")
	end := repo.Snapshot()

	if err := repo.Restore(end); err != nil {
		t.Error(err)
	}
	assertStrings(t, repo, []string{"foo", "bar", "qux", "xyz"})

	if err := repo.Restore(mid); err != nil {
		t.Error(err)
	}
	assertStrings(t, repo, []string{"foo", "bar"})

	if err := repo.Restore(start); err != nil {
		t.Error(err)
	}
	assertStrings(t, repo, []string{})

	// restoring to start invalidates the mid snapshot
	if err := repo.Restore(mid); err == nil {
		t.Error("expected an error when restoring to an invalid snapshot")
	} else if err != ErrInvalidSnapshot {
		t.Error("unexpected error")
	}
}

func TestOptimize(t *testing.T) {
	repo := NewRepository()
	for _, str := range []string{"foo", "bar", "baz"} {
		repo.Intern(str)
	}

	freq := NewFrequency()
	freq.Add(2)
	optimized := repo.Optimize(freq)
	assertStrings(t, optimized, []string{"bar"})
	freq.Add(3)
	freq.Add(3)
	optimized = repo.Optimize(freq)
	assertStrings(t, optimized, []string{"baz", "bar"})
	freq.AddAll(repo)
	freq.Add(3)
	freq.Add(2)
	optimized = repo.Optimize(freq)
	assertStrings(t, optimized, []string{"baz", "bar", "foo"})
}
