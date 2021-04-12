package enqueuestomp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestLogField(t *testing.T) {
	type input struct {
		key string
		value string
	}
	type output struct {
		fields []zap.Field
	}
	type testCase struct {
		name string
		inputs []input
		output output
	}
	testCases := []testCase{
		{
			name: "Putting key and value should be OK",
			inputs: []input{
				input{
					key: "testKey",
					value: "testValue",
				},
			},
			output: output{
				fields: []zap.Field{
					zap.String("testKey", "testValue"),
				},
			},
		},
		{
			name: "Putting more than one key and value should be OK",
			inputs: []input{
				input{
					key: "testKey",
					value: "testValue",
				},
				input{
					key: "testKey2",
					value: "testValue2",
				},
				input{
					key: "testKey3",
					value: "testValue3",
				},
			},
			output: output{
				fields: []zap.Field{
					zap.String("testKey", "testValue"),
					zap.String("testKey2", "testValue2"),
					zap.String("testKey3", "testValue3"),
				},
			},
		},
		{
			name: "Putting empty key and value should be nil",
			inputs: []input{
				input{
					key: "",
					value: "testValue",
				},
			},
			output: output{
				fields: nil,
			},
		},
		{
			name: "Putting key and empty value should be OK",
			inputs: []input{
				input{
					key: "testKey",
					value: "",
				},
			},
			output: output{
				fields: []zap.Field{
					zap.String("testKey", ""),
				},
			},
		},
	}
	for index, testCase := range testCases {
		logField := NewLogField()
		for _, input := range testCase.inputs {
			logField.SetNewField(input.key, input.value)
		}
		expectedResponse := logField.getFields()
		assert.Subset(t, testCase.output.fields, expectedResponse, fmt.Sprintf("Test #%d [%s] failed. Wrong response.", index, testCase.name))
	}
}
