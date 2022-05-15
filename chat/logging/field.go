package logging

import "go.uber.org/zap"

// mapFields maps ZapFieldAdapter to zap.Field
func mapFields(args []any) {
	for i, v := range args {
		// println(reflect.TypeOf(v).String())
		if t, ok := v.(*ZapFieldAdapter); ok {
			args[i] = zap.Field(t.f)
		}
	}
}

// ZapFieldAdapter implements Field for zap.
type ZapFieldAdapter struct{ f zap.Field }

// isField is a dummy method so we don't have an empty interface
func (a ZapFieldAdapter) isField() bool {
	return true
}

// Error wraps zap.Error
func Error(err error) Field {
	return &ZapFieldAdapter{f: zap.Error(err)}
}
