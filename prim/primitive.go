package prim

type Primitive interface {
	Get() interface{}
	Binary(o Primitive, op Operation) (Primitive, error)
	Unary(op Operation) (Primitive, error)
}

func ValueToPrimitive(i interface{}) Primitive {
	switch v := i.(type) {
	case bool:
		return NewBoolean(v)
	case uint8:
		return NewNumber(uint64(v), 1, false)
	case uint16:
		return NewNumber(uint64(v), 2, false)
	case uint32:
		return NewNumber(uint64(v), 4, false)
	case uint64:
		return NewNumber(uint64(v), 8, false)
	case int8:
		return NewNumber(uint64(v), 1, true)
	case int16:
		return NewNumber(uint64(v), 2, true)
	case int32:
		return NewNumber(uint64(v), 4, true)
	case int64:
		return NewNumber(uint64(v), 8, true)
	case []byte:
		return NewStringRaw(v)
	case string:
		return NewString(v)
	default:
		panic("Unknown value -> primitive conversion")
		return nil
	}
}
