package example

// clean has no underscore — should pass.
func clean() {}

// _ is the blank identifier — should pass.
func _() {}

// TestSomethingCase has no underscore — should pass.
func TestSomethingCase() {}

// with_underscore contains an underscore — want diagnostic.
func with_underscore() {} // want `Rename function "with_underscore" to match the regular expression \^`

// TestSomething_Case contains an underscore — want diagnostic.
func TestSomething_Case() {} // want `Rename function "TestSomething_Case" to match the regular expression \^`

// T is a dummy type for method tests.
type T struct{}

// bad_method is a method with an underscore — want diagnostic.
func (t T) bad_method() {} // want `Rename function "bad_method" to match the regular expression \^`
