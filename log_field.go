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
	return &LogFieldImpl{}
}

func (log *LogFieldImpl) SetNewField(key, value string) {
	log.fields[key] = value
}

func (log *LogFieldImpl) getFields() []zap.Field {
	var fields []zap.Field
	for key, value := range log.fields {
		fields = append(fields, zap.String(key, value))
	}
	return fields
}