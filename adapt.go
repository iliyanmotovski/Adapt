package adapt

import (
	"errors"
	"reflect"
)

type adapter struct {
	adapters map[string]func(reflect.Value) interface{}
}

func NewAdapter() *adapter {
	return &adapter{adapters: make(map[string]func(reflect.Value) interface{})}
}

func (a *adapter) RegisterAdaptFunc(key string, f func(reflect.Value) interface{}) {
	a.adapters[key] = f
}

func (a *adapter) SrcToDst(src interface{}, dst interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case *reflect.ValueError:
				if rflErr := r.(*reflect.ValueError).Error(); rflErr == "reflect: call of reflect.flag.mustBeAssignable on zero Value" {
					err = errors.New(`src and dst field name not matching and no "dstName" tag is provided`)
					break
				}
				err = r.(*reflect.ValueError)
			case string:
				err = errors.New(r.(string))
			}
		}
	}()

	srcType := reflect.TypeOf(src)

	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst).Elem()

	srcFieldsLen := srcType.NumField()
	for i := 0; i < srcFieldsLen; i++ {
		fieldType := srcType.Field(i)

		adapter := fieldType.Tag.Get("adapter")
		dstName := fieldType.Tag.Get("dstName")

		// handles the case when the src field is to be skipped and not adapted
		if adapter == "skip" {
			continue
		}

		fieldName := ""
		fieldValue := reflect.Value{}

		// handles the case when the src and dst field have the same names and
		// are of the same type (no adapter func is needed)
		if adapter == "" && dstName == "" {
			fieldName = fieldType.Name
			fieldValue = srcValue.Field(i)
		}

		// handles the case when the src and dst field have different names, but
		// are of the same type (no adapter func is needed)
		if adapter == "" && dstName != "" {
			fieldName = dstName
			fieldValue = srcValue.Field(i)
		}

		// handles the case when the src and dst field have the same names, but
		// are of different types (adapter func is needed to be registered)
		if adapter != "" && dstName == "" {
			value, ok := a.adapt(adapter, srcValue.Field(i))
			if !ok {
				return errors.New("no adapter func registered for the provided tag key")
			}

			fieldName = fieldType.Name
			fieldValue = value
		}

		// handles the case when the src and dst field have different names and
		// are of different types (adapter func is needed to be registered)
		if adapter != "" && dstName != "" {
			value, ok := a.adapt(adapter, srcValue.Field(i))
			if !ok {
				return errors.New("no adapter func registered for the provided tag key")
			}

			fieldName = dstName
			fieldValue = value
		}

		dstValue.FieldByName(fieldName).Set(fieldValue)
	}

	return nil
}

func (a *adapter) adapt(adapterKey string, value reflect.Value) (reflect.Value, bool) {
	adapter, ok := a.adapters[adapterKey]
	if !ok {
		return reflect.Value{}, false
	}

	return reflect.ValueOf(adapter(value)), true
}
