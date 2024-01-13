package prolog

import (
	"encoding/hex"

	"github.com/ichiban/prolog/engine"
)

// TermHexToBytes try to convert an hexadecimal encoded atom to native golang []byte.
func TermHexToBytes(term engine.Term, env *engine.Env) ([]byte, error) {
	v, err := AssertAtom(env, term)
	if err != nil {
		return nil, err
	}

	result, err := hex.DecodeString(v.String())
	if err != nil {
		err = WithError(engine.DomainError(ValidEncoding("hex"), term, env), err, env)
	}
	return result, err
}
