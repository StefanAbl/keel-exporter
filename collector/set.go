package collector

// The Set type is a type alias of `map[string]struct{}`
type Set map[string]struct{}

// Adds a value to the set
func (s Set) add(value string) {
	s[value] = struct{}{}
}

// Removes a value from the set
func (s Set) remove(value string) {
	delete(s, value)
}

// Returns a boolean value describing if the value exists in the set
func (s Set) has(value string) bool {
	_, ok := s[value]
	return ok
}
