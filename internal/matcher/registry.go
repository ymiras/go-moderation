package matcher

// MatcherCtor is a constructor function that creates a Matcher from a config.
type MatcherCtor func(cfg any) (Matcher, error)

// Registry holds matcher constructors and provides instantiation by name.
type Registry struct {
	ctors map[string]MatcherCtor
}

// RegistryInstance is the global matcher registry.
var RegistryInstance = &Registry{
	ctors: make(map[string]MatcherCtor),
}

// Register adds a matcher constructor to the registry.
func (r *Registry) Register(name string, fn MatcherCtor) {
	r.ctors[name] = fn
}

// Create instantiates a matcher by name using the provided config.
func (r *Registry) Create(name string, cfg any) (Matcher, error) {
	ctor, ok := r.ctors[name]
	if !ok {
		return nil, ErrMatcherNotFound
	}
	return ctor(cfg)
}

// ErrMatcherNotFound is returned when a matcher name is not registered.
var ErrMatcherNotFound = &Error{Message: "matcher not found"}

// Error MatcherError represents a matcher-related error.
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
