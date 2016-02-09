package intern_test

import (
	"fmt"

	"github.com/chriso/go-intern"
)

func ExampleRepository_Intern() {
	repo := intern.NewRepository()

	fmt.Println(repo.Intern("foo"))
	fmt.Println(repo.Intern("bar"))
	fmt.Println(repo.Intern("baz"))
	fmt.Println(repo.Intern("foo"))

	// Output:
	// 1
	// 2
	// 3
	// 1
}

func ExampleRepository_Count() {
	repo := intern.NewRepository()

	fmt.Printf("Initial count is %d\n", repo.Count())

	strings := []string{"foo", "bar", "qux", "qux", "qux", "foo"}
	for _, str := range strings {
		repo.Intern(str)
	}

	fmt.Printf("There are now %d unique strings\n", repo.Count())

	// Output:
	// Initial count is 0
	// There are now 3 unique strings
}

func ExampleRepository_Lookup() {
	repo := intern.NewRepository()
	repo.Intern("foo")

	for _, str := range []string{"foo", "bar"} {
		if id, ok := repo.Lookup(str); ok {
			fmt.Printf("Found string %#v with id %d\n", str, id)
		} else {
			fmt.Printf("Did not find string %#v\n", str)
		}
	}

	// Output:
	// Found string "foo" with id 1
	// Did not find string "bar"
}

func ExampleRepository_LookupID() {
	repo := intern.NewRepository()
	repo.Intern("foo")

	for _, id := range []uint32{1, 2} {
		if str, ok := repo.LookupID(id); ok {
			fmt.Printf("Found string %#v with id %d\n", str, id)
		} else {
			fmt.Printf("Did not find id %d\n", id)
		}
	}

	// Output:
	// Found string "foo" with id 1
	// Did not find id 2
}

func ExampleRepository_Cursor() {
	repo := intern.NewRepository()

	strings := []string{"foo", "bar", "baz"}
	for _, str := range strings {
		repo.Intern(str)
	}

	cursor := repo.Cursor()
	for cursor.Next() {
		fmt.Printf("String %#v has id %d\n", cursor.String(), cursor.ID())
	}

	// Output:
	// String "foo" has id 1
	// String "bar" has id 2
	// String "baz" has id 3
}

func ExampleRepository_Optimize() {
	repo := intern.NewRepository()
	frequencies := intern.NewFrequency()

	strings := []string{"foo", "bar", "qux", "qux", "qux", "foo"}
	for _, str := range strings {
		id := repo.Intern(str)
		frequencies.Add(id)
	}

	optimized := repo.Optimize(frequencies)
	cursor := optimized.Cursor()
	for cursor.Next() {
		fmt.Printf("String %#v has id %d\n", cursor.String(), cursor.ID())
	}

	// Output:
	// String "qux" has id 1
	// String "foo" has id 2
	// String "bar" has id 3
}

func ExampleRepository_Restore() {
	repo := intern.NewRepository()

	repo.Intern("foo")
	snapshot := repo.Snapshot()
	repo.Intern("bar")
	repo.Intern("qux")

	repo.Restore(snapshot)
	repo.Intern("xyz")

	cursor := repo.Cursor()
	for cursor.Next() {
		fmt.Printf("String %#v has id %d\n", cursor.String(), cursor.ID())
	}

	// Output:
	// String "foo" has id 1
	// String "xyz" has id 2
}
