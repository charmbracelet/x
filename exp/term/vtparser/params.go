package parser

// DefaultMaxParameters is the maximum number of parameters allowed.
const DefaultMaxParameters = 32

// params is a parameters for VTE.
type params struct {
	// Number of subparameters in each parameter
	//
	// For each entry in the `params` slice, this stores the length of the param as number of
	// subparams at the same index as the param in the `params` slice.
	//
	// At the subparam positions the length will always be `0`.
	subparams []uint8

	// All parameters and subparameters
	params []uint16

	// Number of subparams in the current parameter.
	current_subparams uint8

	len int
}

func newParams(size int) *params {
	p := new(params)
	if size <= 0 {
		size = DefaultMaxParameters
	}
	p.params = make([]uint16, size)
	p.subparams = make([]uint8, size)
	return p
}

// Len returns the number of parameters.
func (p *params) Len() int {
	return p.len
}

// IsEmpty returns true if there are no parameters.
func (p *params) IsEmpty() bool {
	return p.len == 0
}

// IsFull returns true if there are no more parameters can be added.
func (p *params) IsFull() bool {
	return p.len == DefaultMaxParameters
}

// Clear clears all parameters.
func (p *params) Clear() {
	p.current_subparams = 0
	p.len = 0
}

// Push pushes a parameter.
func (p *params) Push(param uint16) {
	p.subparams[p.len-int(p.current_subparams)] = p.current_subparams + 1
	p.params[p.len] = param
	p.current_subparams = 0
	p.len++
}

// Extend extends the last parameter.
func (p *params) Extend(param uint16) {
	p.subparams[p.len-int(p.current_subparams)] = p.current_subparams + 1
	p.params[p.len] = param
	p.current_subparams++
	p.len++
}

// Range iterates over all parameters.
func (p *params) Range(f func(param []uint16)) {
	for i := 0; i < p.len; {
		numSubparams := p.subparams[i]
		param := p.params[i : i+int(numSubparams)]
		i += int(numSubparams)
		f(param)
	}
}

// Params returns all parameters and their subparameters as a slice.
func (p *params) Params() [][]uint16 {
	var params [][]uint16
	p.Range(func(param []uint16) {
		params = append(params, append([]uint16{}, param...))
	})
	return params
}

func isParamPrivate(code byte) bool {
	return code >= 0x3c && code <= 0x3f
}
