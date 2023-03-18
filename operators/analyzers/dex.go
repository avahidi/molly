package analyzers

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/avahidi/molly/util"
)

// helperAddMapMap is a helper function to insert items in a map-map
func helperAddMapMap(mm map[string]map[string]bool, key string, values ...string) {
	m1, found := mm[key]
	if !found {
		m1 = make(map[string]bool)
	}
	for _, value := range values {
		m1[value] = true
	}
	mm[key] = m1
}

// helperFilterMapMap remove items based on filter functions
func helperFilterMapMap(mm map[string]map[string]bool, f1 func(key string) bool,
	f2 func(key, value string) bool) map[string]map[string]bool {
	ret := make(map[string]map[string]bool)
	for k, vs1 := range mm {
		if !f1(k) {
			vs2 := make(map[string]bool)
			for v := range vs1 {
				if !f2(k, v) {
					vs2[v] = true
				}
			}
			if len(vs2) != 0 {
				ret[k] = vs2
			}
		}
	}
	return ret
}

// helperConvertMapMap is a helper function convert map-map to map-array
func helperConvertMapMap(mm map[string]map[string]bool) map[string][]string {
	ret := make(map[string][]string)
	for k, v1 := range mm {
		var v2 []string
		for v, _ := range v1 {
			v2 = append(v2, v)
		}
		sort.Strings(v2)
		ret[k] = v2
	}
	return ret
}

// DEX class analyzer based on
// https://source.android.com/devices/tech/dalvik/dex-format
//
// this code is pretty fragile since DEX is a bit hard to parse
// and we are ignording all corner cases and version differences

const dexMagic uint32 = 0x0A786564 // dex\n little-endian

type dexContext struct {
	util.Structured
	version    int
	clss       []*dexClass
	clssIdMap  map[uint32]*dexClass
	strings    []string
	type_ids   []uint32
	proto_ids  []dexProtoId
	method_ids []dexFieldMethod
	field_ids  []dexFieldMethod
}

type dexClassDef struct {
	ClassIdx        uint32 // -> type_ids
	AccessFlags     uint32
	SupperClassIdx  uint32 // -> type_ids or NO_INDEX
	InterfaceOffset uint32
	SourceIdx       uint32 // -> string_ids or 0
	AnnotationOff   uint32
	ClassDataOff    uint32
	StaticValuesOff uint32
}
type dexProtoId struct {
	ShortFormIdx     uint32 // -> string_ids
	ReturnTypeIdx    uint32 // -> type_ids
	ParametersOffset uint32
}
type dexFieldMethod struct {
	ClassIdx uint16 // -> type_ids
	ProtoIdx uint16 // -> proto_idx
	NameIdx  uint32 // -> string_ids
}
type dexFieldData struct {
	fieldIdx    uint32 // -> field_ids
	accessFlags uint32
}
type dexMethodData struct {
	methodIdx   uint32 // -> method_idx
	accessFlags uint32
	codeOffset  uint32
	insts       []uint16
}
type dexClass struct {
	classDef       *dexClassDef
	staticFields   []*dexFieldData
	instanceFields []*dexFieldData
	directMethods  []*dexMethodData
	virtualMethods []*dexMethodData
	// extracted by us
	name        string
	packageName string
}
type dexMapItem struct {
	Type     uint16
	Reserved uint16
	Count    uint32
	Offset   uint32
}

// GetString is a safe method to get a string
func (d dexContext) getString(n uint32) string {
	if n >= uint32(len(d.strings)) {
		return ""
	}
	return d.strings[n]
}

// GetTypeName is a safe method to get a type name from type_ids
func (d dexContext) getTypeName(n uint32) string {
	if n >= uint32(len(d.type_ids)) {
		return ""
	}
	return d.getString(d.type_ids[n])
}

// uleb128Skip reads past a uleb128
func uleb128Skip(r *bufio.Reader) error {
	_, err := util.Process(r, func(b uint8, n int) bool { return (b & 0x80) != 0 })
	return err
}

// uleb128Read reads one uleb128
func uleb128Read(r *bufio.Reader) (uint64, error) {
	var ret uint64
	_, err := util.Process(r, func(b uint8, n int) bool {
		ret = ret | (uint64(b&0x7f) << uint(n*7))
		return (b & 0x80) != 0
	})
	return ret, err
}

// uleb128ReadN reads multiple uleb128's, as this seems to be a common operation
func uleb128ReadN(r *bufio.Reader, n int) (ret []uint64, err error) {
	ret = make([]uint64, n)
	for i := 0; i < n; i++ {
		ret[i], err = uleb128Read(r)
		if err != nil {
			return
		}
	}
	return
}

// dexExtractRefs returns list if methods this class invokes and types it uses
// some trivial cases will be excluded (e.g. call own methods)
func dexExtractRefs(c *dexContext, cls *dexClass) ([]*dexFieldMethod, []string) {
	idx := cls.classDef.ClassIdx
	sset := make(map[uint32]bool)
	mset := make(map[uint16]bool)

	mss := [][]*dexMethodData{cls.directMethods, cls.virtualMethods}
	for _, ms := range mss {
		for _, m := range ms {
			for i := 0; i < len(m.insts); {
				op, size, _ := dalvikAnalyze(m.insts, i)
				inst := m.insts[i : i+size]
				i += size

				if isNew, typeIdx := dalvikOpNew(op, inst); isNew {
					sset[c.type_ids[typeIdx]] = true
				}

				if isInvoke, methodIdx := dalvikOpInvoke(op, inst); isInvoke {
					m := &c.method_ids[methodIdx]
					if uint32(m.ClassIdx) != idx { // no point recording out own class
						mset[methodIdx] = true

						// obviously, if we invoke a method we should know its type
						sset[c.type_ids[m.ClassIdx]] = true
					}
				}
			}
		}
	}

	// convert sets to arrays of correct type and return
	var mret []*dexFieldMethod
	for methodIdx := range mset {
		mret = append(mret, &c.method_ids[methodIdx])
	}

	var sret []string
	for stringIdx := range sset {
		name, err := javaTypeToClassName(c.strings[stringIdx])
		if err == nil {
			sret = append(sret, name)
		}
	}

	return mret, sret
}

func dexLoadClass(c *dexContext, offset int64) (*dexClass, error) {
	ret := &dexClass{}
	if offset == 0 {
		return ret, nil
	}

	br, err := util.BufreaderAt(c.Reader, offset)
	if err != nil {
		return nil, err
	}
	// read the size fields:
	sizes, err := uleb128ReadN(br, 4)
	if err != nil {
		return nil, err
	}
	// read the field arrays
	fields := make([][]*dexFieldData, 2)
	for i := range fields {
		fields[i] = make([]*dexFieldData, int(sizes[i]))
		last := uint32(0)
		for j := range fields[i] {
			vals, err := uleb128ReadN(br, 2)
			if err != nil {
				return nil, err
			}
			fields[i][j] = &dexFieldData{
				fieldIdx:    uint32(vals[0]),
				accessFlags: uint32(vals[1]),
			}
			fields[i][j].fieldIdx += last
			last = fields[i][j].fieldIdx
		}
	}
	ret.staticFields = fields[0]
	ret.instanceFields = fields[1]

	// read the method arrays
	methods := make([][]*dexMethodData, 2)
	for i := range methods {
		methods[i] = make([]*dexMethodData, int(sizes[i+2]))
		last := uint32(0)
		for j := range methods[i] {
			vals, err := uleb128ReadN(br, 3)
			if err != nil {
				return nil, err
			}
			methods[i][j] = &dexMethodData{
				methodIdx:   uint32(vals[0]),
				accessFlags: uint32(vals[1]),
				codeOffset:  uint32(vals[2]),
			}
			methods[i][j].methodIdx += last
			last = methods[i][j].methodIdx
		}
	}
	ret.directMethods = methods[0]
	ret.virtualMethods = methods[1]

	// with all that in place, lets try to load the code
	for _, mss := range methods {
		for _, method := range mss {
			if method.codeOffset != 0 {
				var count uint32
				coff := int64(method.codeOffset) + 6*2
				if err := c.ReadAt(coff, &count); err != nil {
					return nil, err
				}
				method.insts = make([]uint16, count)
				if err := c.Read(&method.insts); err != nil {
					return nil, err
				}
			}
		}
	}
	return ret, nil
}

func dexExtractStrings(c *dexContext, offset int64, count int) error {
	// get offset pointers
	offsets := make([]uint32, count)
	if err := c.ReadAt(offset, &offsets); err != nil {
		return err
	}
	// extract the strings one by one
	strs := make([]string, count)
	for i, off := range offsets {
		br, err := util.BufreaderAt(c.Reader, int64(off))
		if err != nil {
			return err
		}
		if err := uleb128Skip(br); err != nil {
			return err
		}
		data := make([]uint8, 0)
		_, err = util.Process(br, func(b uint8, n int) bool {
			if b == 0 {
				return false
			}
			data = append(data, b)
			return true
		})
		if err != nil {
			return err
		}
		strs[i] = string(data)
	}
	c.strings = strs

	return nil
}

func dexExtractTypes(c *dexContext, offset int64, count int) error {
	c.type_ids = make([]uint32, count)
	return c.ReadAt(offset, &c.type_ids)
}

func dexExtractClasses(c *dexContext, offset int64, count int) error {
	defs := make([]dexClassDef, count)
	if err := c.ReadAt(offset, &defs); err != nil {
		return err
	}
	c.clss = make([]*dexClass, count)
	c.clssIdMap = make(map[uint32]*dexClass)
	for i, _ := range defs {
		def := &defs[i] // don't use "for i, def := range ... "!
		cls, err := dexLoadClass(c, int64(def.ClassDataOff))
		if err != nil {
			return err
		}
		cls.classDef = def
		classname, err := javaTypeToClassName(c.getTypeName(def.ClassIdx))
		if err != nil {
			return err
		}

		cls.name = classname
		cls.packageName = javaExtractPackageName(classname)

		c.clss[i] = cls
		c.clssIdMap[cls.classDef.ClassIdx] = cls
	}

	return nil
}

func dexExtractProtos(c *dexContext, offset int64, count int) error {
	c.proto_ids = make([]dexProtoId, count)
	return c.ReadAt(offset, &c.proto_ids)
}

func dexExtractMethods(c *dexContext, offset int64, count int) error {
	c.method_ids = make([]dexFieldMethod, count)
	return c.ReadAt(offset, &c.method_ids)
}

func dexExtractFields(c *dexContext, offset int64, count int) error {
	c.field_ids = make([]dexFieldMethod, count)
	return c.ReadAt(offset, &c.field_ids)
}

// dexExtractMap extracts file map searchable by type
func dexExtractMap(c *dexContext, offset uint32) (map[uint16]dexMapItem, error) {
	var size uint32
	if err := c.ReadAt(int64(offset), &size); err != nil {
		return nil, err
	}
	items := make([]dexMapItem, size)
	if err := c.Read(&items); err != nil {
		return nil, err
	}
	ret := make(map[uint16]dexMapItem)
	for _, m := range items {
		ret[m.Type] = m
	}
	return ret, nil
}

func extractKeys(m map[interface{}]interface{}, a []interface{}) []interface{} {
	for k := range m {
		a = append(a, k)
	}
	return a
}

func DexAnalyzer(filename string, r io.ReadSeeker) (interface{}, error) {
	var header struct {
		DexMagic    uint32
		DexVersion  [8]uint8
		Ignored     [7]uint32
		EndianTag   uint32
		MoreIgnored [2]uint32
		MapOffset   uint32
	}
	ctx := &dexContext{Structured: util.Structured{Reader: r, Order: binary.LittleEndian}}
	if err := ctx.ReadAt(0, &header); err != nil {
		return nil, err

	}
	// extract version and check sanity before we do anything
	for _, c := range header.DexVersion[:3] {
		ctx.version = ctx.version*10 + int(c-'0')
	}
	if header.DexMagic != dexMagic ||
		header.DexVersion[0] != '0' || header.DexVersion[3] != 0 {
		return nil, fmt.Errorf("Not a dex file or unknown dex version\n")
	}
	if header.EndianTag == 0x78563412 {
		ctx.Order = binary.BigEndian
	}
	mp, err := dexExtractMap(ctx, header.MapOffset)
	if err != nil {
		return nil, err
	}
	// process map entries in the order we want
	handlers := []struct {
		typ      uint16
		optional bool
		f        func(*dexContext, int64, int) error
	}{
		{0x0001, false, dexExtractStrings},
		{0x0002, false, dexExtractTypes},
		{0x0003, false, dexExtractProtos},
		{0x0004, false, dexExtractFields},
		{0x0005, false, dexExtractMethods},
		{0x0006, false, dexExtractClasses},
	}
	for _, h := range handlers {
		if p, found := mp[h.typ]; !found {
			if !h.optional {
				return nil, fmt.Errorf("dex internal error: missing type %d", h.typ)
			}
		} else if err := h.f(ctx, int64(p.Offset), int(p.Count)); err != nil {
			return nil, err
		}
	}
	report := map[string]interface{}{
		"dex-version":   header.DexVersion,
		"00-disclaimer": "dex analyzer is still under construction",
		"01-note":       "innerclasses and subpackage classes are excluded",
	}
	dexCreateReport(ctx, report)
	return report, err
}

// dexCreateReport generates the report from the extract dexContext
func dexCreateReport(ctx *dexContext, report map[string]interface{}) {
	// some helper functions used for filtering
	uninterestingcall := func(caller, callee string) bool {
		return javaIsInnerName(callee) || strings.HasPrefix(callee, caller) ||
			strings.HasSuffix(callee, "init>") // .<init> and .<cinit>
	}

	// extract typenames as set
	typenames := make(map[string]bool)
	for _, idx := range ctx.type_ids {
		typenames[ctx.strings[idx]] = true
	}

	// extract STRINGS, exclude empty strings and typenames
	var strs []string
	for _, str := range ctx.strings {
		if _, found := typenames[str]; !found && len(str) > 0 {
			strs = append(strs, str)
		}
	}
	sort.Strings(strs)
	report["strings"] = strs

	// extract CLASSNAMES
	var classNames []string
	for _, cls := range ctx.clss {
		if !javaIsInnerName(cls.name) {
			classNames = append(classNames, cls.name)
		}
	}
	sort.Strings(classNames)
	report["classes"] = classNames

	// extract method and types seen in each class/package
	// we use sets of strings since we can have doubles due to polymorphism etc
	typemap_class := make(map[string]map[string]bool)
	callmap_class := make(map[string]map[string]bool)

	for _, cls := range ctx.clss {
		methods, objects := dexExtractRefs(ctx, cls)

		// record types
		helperAddMapMap(typemap_class, cls.name, objects...)

		// record calls
		for _, method := range methods {
			// extract target, but don't use ctx.clss since target class  might be a foreign type
			calleeName := ctx.getString(uint32(method.NameIdx))
			calleeClass, _ := javaTypeToClassName(ctx.getTypeName(uint32(method.ClassIdx)))
			call := fmt.Sprintf("%s.%s", calleeClass, calleeName)
			helperAddMapMap(callmap_class, cls.name, call)
		}
	}

	// simplify it, because it is mostly garbage
	callmap_class = helperFilterMapMap(callmap_class, javaIsInnerName, uninterestingcall)
	typemap_class = helperFilterMapMap(typemap_class, javaIsInnerName, uninterestingcall)

	report["callmap-class"] = helperConvertMapMap(callmap_class)
	report["typemap-class"] = helperConvertMapMap(typemap_class)

}
