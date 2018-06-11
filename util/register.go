package util

type Register struct {
	data   map[string]interface{}
	parent *Register
}

func NewRegister() *Register {
	return &Register{data: make(map[string]interface{})}
}

// Walk all values in the register
func (r Register) Walk(visitor func(key string, val interface{}) bool) {
	for k, v := range r.data {
		if !visitor(k, v) {
			return
		}
	}
}

func (r *Register) SetParent(parent *Register) {
	r.parent = parent
}

func (r *Register) Set(name string, val interface{}) {
	r.data[name] = val
}

func (r Register) Get(name string) (interface{}, bool) {
	i, f := r.data[name]
	if !f && r.parent != nil {
		i, f = r.parent.Get(name)
	}
	return i, f
}

func (r Register) GetNumber(name string, def uint64) (uint64, bool) {
	i, f := r.Get(name)
	if f {
		if n, o := i.(uint64); o {
			return n, true
		}
	}
	return def, false
}

func (r Register) GetBoolean(name string, def bool) (bool, bool) {
	i, f := r.Get(name)
	if f {
		if n, o := i.(bool); o {
			return n, true
		}
	}
	return def, false
}

func (r Register) GetString(name string, def string) (string, bool) {
	i, f := r.Get(name)
	if f {
		if n, o := i.(string); o {
			return n, true
		}
	}
	return def, false
}

func (r Register) SetNumber(name string, val uint64) {
	r.Set(name, val)
}

func (r Register) SetBoolean(name string, val bool) {
	r.Set(name, val)
}

func (r Register) SetString(name string, val string) {
	r.Set(name, val)
}
