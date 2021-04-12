package enqueuestomp

import "go.uber.org/zap"

type LogField interface {
	SetNewField(key, value string)
	getFields() []zap.Field
}

type LogFieldImpl struct {
	fields map[string]string
}

func NewLogField() LogField {
	return &LogFieldImpl{
		fields: map[string]string{},
	}
}

func (log *LogFieldImpl) SetNewField(key, value string) {
	if len(key) > 0 {
		log.fields[key] = value
	}
}

func (log *LogFieldImpl) getFields() []zap.Field {
	var fields []zap.Field
	if len(log.fields) > 0 {
		for key, value := range log.fields {
			fields = append(fields, zap.String(key, value))
		}
	}
	return fields
}