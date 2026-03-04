package example

const beforeConst = 1 // OK - const before any func

var beforeVar = 2 // OK - var before any func

type T struct{} // OK - type declarations are fine anywhere

func exported() {} // OK - first func

func (t T) method() {} // OK - method also sets seenFunc

const afterConst = 3 // want `const declaration after function declaration`

var afterVar = 4 // want `var declaration after function declaration`

const ( // want `const declaration after function declaration`
	groupedA = 5
	groupedB = 6
)

var ( // want `var declaration after function declaration`
	groupedC = 7
	groupedD = 8
)

func another() {} // OK - func after func is fine
