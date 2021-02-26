package evaluate

import (
	"fmt"
	"math"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
)

func TestEvaluateResultWithSuccess(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "true",
		FailureCondition: "false",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseSuccessful, status)
	assert.NoError(t, err)
}

func TestEvaluateResultWithFailure(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "true",
		FailureCondition: "true",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseFailed, status)
	assert.NoError(t, err)
}

func TestEvaluateResultInconclusive(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "false",
		FailureCondition: "false",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseInconclusive, status)
	assert.NoError(t, err)
}

func TestEvaluateResultNoSuccessConditionAndNotFailing(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "",
		FailureCondition: "false",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseSuccessful, status)
	assert.NoError(t, err)
}

func TestEvaluateResultNoFailureConditionAndNotSuccessful(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "false",
		FailureCondition: "",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseFailed, status)
	assert.NoError(t, err)
}

func TestEvaluateResultNoFailureConditionAndNoSuccessCondition(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "",
		FailureCondition: "",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseSuccessful, status)
	assert.NoError(t, err)
}

func TestEvaluateResultWithErrorOnSuccessCondition(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "a == true",
		FailureCondition: "true",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, status)
	assert.EqualError(t, err, "unknown name a")
}

func TestEvaluateResultWithErrorOnFailureCondition(t *testing.T) {
	metric := v1alpha1.Metric{
		SuccessCondition: "true",
		FailureCondition: "a == true",
	}
	logCtx := logrus.WithField("test", "test")
	status, err := EvaluateResult(true, metric, *logCtx)
	assert.Equal(t, v1alpha1.AnalysisPhaseError, status)
	assert.EqualError(t, err, "unknown name a")
}

func TestEvaluateConditionWithSuccess(t *testing.T) {
	b, err := EvalCondition(true, "result == true")
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestEvaluateConditionWithFailure(t *testing.T) {
	b, err := EvalCondition(true, "result == false")
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestErrorWithNonBoolReturn(t *testing.T) {
	b, err := EvalCondition(true, "1")
	assert.Equal(t, fmt.Errorf("expected bool, but got int"), err)
	assert.False(t, b)
}

func TestErrorWithInvalidReference(t *testing.T) {
	b, err := EvalCondition(true, "invalidVariable")
	assert.Equal(t, fmt.Errorf("unknown name invalidVariable"), err)
	assert.False(t, b)
}

func TestErrorWithInvalidReference2(t *testing.T) {
	b, err := EvalCondition(true, "invalidVariable == true")
	assert.Equal(t, fmt.Errorf("unknown name invalidVariable"), err)
	assert.False(t, b)
}

func TestEvaluateArray(t *testing.T) {
	floats := []float64{float64(2), float64(2)}
	b, err := EvalCondition(floats, "all(result, {# > 1})")
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestEvaluateInOperator(t *testing.T) {
	floats := []float64{float64(2), float64(2)}
	b, err := EvalCondition(floats, "2 in result")
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestEvaluateFloat64(t *testing.T) {
	b, err := EvalCondition(float64(5), "result > 1")
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestEvaluateInvalidStruct(t *testing.T) {
	b, err := EvalCondition(true, "result.Name() == 'hi'")
	assert.EqualError(t, err, "type bool has no method Name")
	assert.False(t, b)
}

func TestEvaluateAsIntPanic(t *testing.T) {
	b, err := EvalCondition("1.1", "asInt(result) == 1.1")
	assert.EqualError(t, err, "strconv.ParseInt: parsing \"1.1\": invalid syntax")
	assert.False(t, b)
}

func TestEvaluateNil(t *testing.T) {
	tests := []struct {
		input       interface{}
		expression  string
		expectation bool
		expectedErr string
	}{
		{nil, "result == nil", true, ""},
		{nil, "result != nil", false, ""},
		{nil, "result == 0", false, ""},
		{nil, "result != 0", true, ""},
		{nil, "result >= 0", false, "invalid operation: <nil> >= int"},
		{nil, "result == \"nil\"", false, ""},
		{nil, "result != \"nil\"", true, ""},
		{nil, "result == false", false, ""},
		{nil, "result != false", true, ""},
		{nil, "result == true", false, ""},
		{nil, "result != true", true, ""},
		{nil, "result == nil || result > 0", true, ""},
		{nil, "result == nil && result != nil", false, ""},
		{nil, "result == nil || result == false", true, ""},
		{nil, "result == nil ? true : result > 0", true, ""},
		{nil, "result == nil || result.subfield == 0", true, ""},
		{nil, "result == nil || result[0] > 0.9", true, ""},
		{nil, "result == nil ? true : shouldnotevaluate", false, "unknown name shouldnotevaluate"},
		{1, "result == nil", false, ""},
		{1, "result != nil && result == 1 ", true, ""},
		{1, "result != nil && result > 0", true, ""},
		{1.1, "result != nil && result <= 0 ", false, ""},
		{1, "result != nil || result > 0", true, ""},
		{1, "result == nil ? false : result > 0", true, ""},
		{map[string]int{"foo": 1}, "result == nil", false, ""},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			b, err := EvalCondition(test.input, test.expression)
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectation, b)
		})
	}
}

func TestEvaluateAsInt(t *testing.T) {
	tests := []struct {
		input       interface{}
		expression  string
		expectation bool
	}{
		{"1", "asInt(result) == 1", true},
		{1, "asInt(result) == 1", true},
		{1.123, "asInt(result) == 1", true},
	}
	for _, test := range tests {
		b, err := EvalCondition(test.input, test.expression)
		assert.NoError(t, err)
		assert.Equal(t, test.expectation, b)
	}
}

func TestEvaluateAsFloatError(t *testing.T) {
	tests := []struct {
		input      interface{}
		expression string
		errRegexp  string
	}{
		{"NotANum", "asFloat(result) == 1.1", `strconv.ParseFloat: parsing "NotANum": invalid syntax`},
		{"1.1", "asFloat(result) == \"1.1\"", `invalid operation: == \(mismatched types float64 and string\)`},
	}
	for _, test := range tests {
		b, err := EvalCondition(test.input, test.expression)
		if assert.Error(t, err) {
			assert.Regexp(t, test.errRegexp, err.Error())
		}
		assert.False(t, b)
	}
}

func TestEvaluateAsFloat(t *testing.T) {
	tests := []struct {
		input       interface{}
		expression  string
		expectation bool
	}{
		{"1.1", "asFloat(result) == 1.1", true},
		{"1.1", "asFloat(result) >= 1.1", true},
		{"1.1", "asFloat(result) <= 1.1", true},
		{1.1, "asFloat(result) == 1.1", true},
		{1, "asFloat(result) == 1", true},
		{1, "asFloat(result) >= 1", true},
		{1, "asFloat(result) >= 1", true},
	}
	for _, test := range tests {
		b, err := EvalCondition(test.input, test.expression)
		assert.NoError(t, err)
		assert.Equal(t, test.expectation, b)
	}
}

func TestAsInt(t *testing.T) {
	tests := []struct {
		input       string
		output      int64
		shouldPanic bool
	}{
		{"1", 1, false},
		{"notint", 1, true},
		{"1.1", 1, true},
	}

	for _, test := range tests {
		if test.shouldPanic {
			assert.Panics(t, func() { asInt(test.input) })
		} else {
			assert.Equal(t, test.output, asInt(test.input))
		}
	}
}

func TestAsFloat(t *testing.T) {
	tests := []struct {
		input       string
		output      float64
		shouldPanic bool
	}{
		{"1", 1, false},
		{"notfloat", 1, true},
		{"1.1", 1.1, false},
	}

	for _, test := range tests {
		if test.shouldPanic {
			assert.Panics(t, func() { asFloat(test.input) })
		} else {
			assert.Equal(t, test.output, asFloat(test.input))
		}
	}
}

func TestIsInf(t *testing.T) {
	inf, notInf := math.Inf(0), 0.0
	assert.True(t, isInf(inf))
	assert.False(t, isInf(notInf))
}
