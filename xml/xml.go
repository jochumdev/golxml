// Copyright 2012 Rene Jochum.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"encoding/xml"
	"errors"
	// "fmt"
	//"github.com/kicool/kicool.go/dump"
	gokoxml "github.com/moovweb/gokogiri/xml"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// An UnmarshalError represents an error in the unmarshalling process.
type UnmarshalError string

func (e UnmarshalError) Error() string { return string(e) }

func Unmarshal(data []byte, v interface{}) error {
	return new(Decoder).Decode(data, v)
}

type Decoder struct {
	doc *gokoxml.XmlDocument
}

// TODO: Make the Parser options configureable.
func (d *Decoder) Decode(data []byte, v interface{}) error {
	doc, err := gokoxml.Parse(data, gokoxml.DefaultEncodingBytes, nil, gokoxml.DefaultParseOption, gokoxml.DefaultEncodingBytes)
	d.doc = doc
	if err != nil {
		return err
	}
	defer doc.Free()

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("non-pointer passed to Decode")
	}

	return d.unmarshal(val.Elem(), nil)
}

func (p *Decoder) unmarshal(val reflect.Value, start gokoxml.Node) error {
	// Find first xml node.
	if start == nil {
		start = p.doc.Root().XmlNode
	}

	// Unpacks a pointer
	if pv := val; pv.Kind() == reflect.Ptr {
		if pv.IsNil() {
			pv.Set(reflect.New(pv.Type().Elem()))
		}
		val = pv.Elem()
	}

	var (
		sv    reflect.Value
		tinfo *typeInfo
		err   error
	)

	switch v := val; v.Kind() {
	default:
		return errors.New("unknown type " + v.Type().String())

		// TODO: Implement this once i understand Skip()
		// case reflect.Interface:
		// 	return p.Skip()

	case reflect.Slice:
		typ := v.Type()
		if typ.Elem().Kind() == reflect.Uint8 {
			// []byte
			if err := copyValue(v, start.Content()); err != nil {
				return err
			}
			break
		}

		// Slice of element values.
		// Grow slice.
		n := v.Len()
		if n >= v.Cap() {
			ncap := 2 * n
			if ncap < 4 {
				ncap = 4
			}
			new := reflect.MakeSlice(typ, n, ncap)
			reflect.Copy(new, v)
			v.Set(new)
		}
		v.SetLen(n + 1)

		// Recur to read element into slice.
		if err := p.unmarshal(v.Index(n), start); err != nil {
			v.SetLen(n)
			return err
		}
		return nil

	case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.String:
		if err := copyValue(v, start.Content()); err != nil {
			return err
		}

	case reflect.Struct:
		typ := v.Type()
		if typ == nameType {
			v.Set(reflect.ValueOf(xml.Name{Local: start.Name()}))
			break
		}
		if typ == timeType {
			if err := copyValue(v, start.Content()); err != nil {
				return err
			}
			break
		}

		sv = v
		tinfo, err = getTypeInfo(typ)
		if err != nil {
			return err
		}

		// Validate and assign element name.
		if tinfo.xmlname != nil {
			// var space string
			finfo := tinfo.xmlname
			if finfo.name != "" && finfo.name != start.Name() {
				return UnmarshalError("expected element type <" + finfo.name + "> but have <" + start.Name() + ">")
			}

			fv := sv.FieldByIndex(finfo.idx)
			if _, ok := fv.Interface().(xml.Name); ok {
				fv.Set(reflect.ValueOf(xml.Name{Local: start.Name()}))
			}
		}

		for i := range tinfo.fields {
			finfo := &tinfo.fields[i]
			if finfo.flags&fMode == fAttr {
				strv := sv.FieldByIndex(finfo.idx)
				for name, a := range start.Attributes() {
					if name == finfo.name {
						copyValue(strv, a.Content())
					}
				}
			}
		}

		for cur_node := start.FirstChild(); cur_node != nil; cur_node = cur_node.NextSibling() {
			if sv.IsValid() {
				if cur_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
					continue
				}

				err = p.unmarshalPath(tinfo, sv, nil, cur_node)
				if err != nil {
					return err
				}
			}
		}

		// Assign attributes.
		// Also, determine whether we need to save character data or comments.
		// for i := range tinfo.fields {
		// 	finfo := &tinfo.fields[i]
		// 	switch finfo.flags & fMode {
		// 	case fAttr:
		// 		strv := sv.FieldByIndex(finfo.idx)
		// 		// Look for attribute.
		// 		for _, a := range start.Attr {
		// 			if a.Name.Local == finfo.name {
		// 				copyValue(strv, []byte(a.Value))
		// 				break
		// 			}
		// 		}

		// 	case fCharData:
		// 		if !saveData.IsValid() {
		// 			saveData = sv.FieldByIndex(finfo.idx)
		// 		}

		// 	case fComment:
		// 		if !saveComment.IsValid() {
		// 			saveComment = sv.FieldByIndex(finfo.idx)
		// 		}

		// 	case fAny:
		// 		if !saveAny.IsValid() {
		// 			saveAny = sv.FieldByIndex(finfo.idx)
		// 		}

		// 	case fInnerXml:
		// 		if !saveXML.IsValid() {
		// 			saveXML = sv.FieldByIndex(finfo.idx)
		// 			if p.saved == nil {
		// 				saveXMLIndex = 0
		// 				p.saved = new(bytes.Buffer)
		// 			} else {
		// 				saveXMLIndex = p.savedOffset()
		// 			}
		// 		}
		// 	}
		// }

	} // switch v := val; v.Kind() {

	return nil
}

func copyValue(dst reflect.Value, src string) (err error) {
	// Helper functions for integer and unsigned integer conversions
	var itmp int64
	getInt64 := func() bool {
		itmp, err = strconv.ParseInt(src, 10, 64)
		// TODO: should check sizes
		return err == nil
	}
	var utmp uint64
	getUint64 := func() bool {
		utmp, err = strconv.ParseUint(src, 10, 64)
		// TODO: check for overflow?
		return err == nil
	}
	var ftmp float64
	getFloat64 := func() bool {
		ftmp, err = strconv.ParseFloat(src, 64)
		// TODO: check for overflow?
		return err == nil
	}

	// Save accumulated data.
	switch t := dst; t.Kind() {
	case reflect.Invalid:
		// Probably a comment.
	default:
		return errors.New("cannot happen: unknown type " + t.Type().String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if !getInt64() {
			return err
		}
		t.SetInt(itmp)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if !getUint64() {
			return err
		}
		t.SetUint(utmp)
	case reflect.Float32, reflect.Float64:
		if !getFloat64() {
			return err
		}
		t.SetFloat(ftmp)
	case reflect.Bool:
		value, err := strconv.ParseBool(strings.TrimSpace(src))
		if err != nil {
			return err
		}
		t.SetBool(value)
	case reflect.String:
		t.SetString(src)
	case reflect.Slice:
		t.SetBytes([]byte(src))
	case reflect.Struct:
		if t.Type() == timeType {
			tv, err := time.Parse(time.RFC3339, src)
			if err != nil {
				return err
			}
			t.Set(reflect.ValueOf(tv))
		}
	}
	return nil
}

// unmarshalPath walks down an XML structure looking for wanted
// paths, and calls unmarshal on them.
// The consumed result tells whether XML elements have been consumed
// from the Decoder until start's matching end element, or if it's
// still untouched because start is uninteresting for sv's fields.
func (p *Decoder) unmarshalPath(tinfo *typeInfo, sv reflect.Value, parents []string, start gokoxml.Node) (err error) {
	recurse := false
	name := start.Name() // For speed

Loop:
	for i := range tinfo.fields {
		finfo := &tinfo.fields[i]
		if finfo.flags&fElement == 0 || len(finfo.parents) < len(parents) {
			continue
		}
		for j := range parents {
			if parents[j] != finfo.parents[j] {
				continue Loop
			}
		}
		if len(finfo.parents) == len(parents) && finfo.name == name {
			// It's a perfect match, unmarshal the field.
			return p.unmarshal(sv.FieldByIndex(finfo.idx), start)
		}
		if len(finfo.parents) > len(parents) && finfo.parents[len(parents)] == name {
			// It's a prefix for the field. Break and recurse
			// since it's not ok for one field path to be itself
			// the prefix for another field path.
			recurse = true

			// We can reuse the same slice as long as we
			// don't try to append to it.
			parents = finfo.parents[:len(parents)+1]
			break
		}
	}

	if !recurse {
		// We have no business with this element.
		return nil
	}

	// The element is not a perfect match for any field, but one
	// or more fields have the path to this element as a parent
	// prefix. Recurse and attempt to match these.
	for cur_node := start.FirstChild(); cur_node != nil; cur_node = cur_node.NextSibling() {
		if cur_node.NodeType() != gokoxml.XML_ELEMENT_NODE {
			continue
		}

		if err := p.unmarshalPath(tinfo, sv, parents, cur_node); err != nil {
			return err
		}
	}

	// No more XML Nodes.
	return nil
}
