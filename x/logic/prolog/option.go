package prolog

import (
	"fmt"

	"github.com/ichiban/prolog/engine"
)

// GetOption returns the value of the first option with the given name in the given options.
// An option is a compound with the given name as functor and one argument which is
// a term, for instance `opt(v)`.
// The options are either a list of options or an option.
// If no option is found nil is returned.
func GetOption(name engine.Atom, options engine.Term, env *engine.Env) (engine.Term, error) {
	extractOption := func(term engine.Term) (engine.Term, error) {
		switch v := term.(type) {
		case engine.Compound:
			if v.Functor() == name {
				if v.Arity() != 1 {
					return nil, fmt.Errorf("invalid arity for compound '%s': %d but expected 1", name, v.Arity())
				}

				return v.Arg(0), nil
			}
			return nil, nil
		case nil:
			return nil, nil
		default:
			return nil, fmt.Errorf("invalid term '%s' - expected engine.Compound but got %T", term, v)
		}
	}

	resolvedTerm := env.Resolve(options)

	if IsEmptyList(resolvedTerm) {
		return nil, nil
	}

	if IsList(resolvedTerm) {
		iter := engine.ListIterator{List: resolvedTerm, Env: env}

		for iter.Next() {
			opt := env.Resolve(iter.Current())

			term, err := extractOption(opt)
			if err != nil {
				return nil, err
			}

			if term != nil {
				return term, nil
			}
		}
		return nil, nil
	}

	return extractOption(resolvedTerm)
}

// GetOptionWithDefault returns the value of the first option with the given name in the given options, or the given
// default value if no option is found.
func GetOptionWithDefault(
	name engine.Atom, options engine.Term, defaultValue engine.Term, env *engine.Env,
) (engine.Term, error) {
	if term, err := GetOption(name, options, env); err != nil {
		return nil, err
	} else if term != nil {
		return term, nil
	}
	return defaultValue, nil
}

// GetOptionAsAtomWithDefault is a helper function that returns the value of the first option with the given name in the
// given options.
func GetOptionAsAtomWithDefault(
	name engine.Atom, options engine.Term, defaultValue engine.Term, env *engine.Env,
) (engine.Atom, error) {
	term, err := GetOptionWithDefault(name, options, defaultValue, env)
	if err != nil {
		return AtomEmpty, err
	}
	atom, err := AssertAtom(env, term)
	if err != nil {
		return AtomEmpty, err
	}

	return atom, nil
}