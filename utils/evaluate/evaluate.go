package evaluate

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/checker"
	"github.com/antonmedv/expr/compiler"
	"github.com/antonmedv/expr/conf"
	"github.com/antonmedv/expr/file"
	"github.com/antonmedv/expr/parser"
	"github.com/sirupsen/logrus"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
)

var (
	// A regexp to detect when nil is being compared to non-nil
	mismatchedTypesWithNil = regexp.MustCompile("mismatched types.*<nil>")
)

// EvaluateResult returns the AnalysisPhase when evaluating the measurement result
func EvaluateResult(result interface{}, metric v1alpha1.Metric, logCtx logrus.Entry) (v1alpha1.AnalysisPhase, error) {
	successCondition := false
	failCondition := false
	var err error

	if metric.SuccessCondition != "" {
		successCondition, err = EvalCondition(result, metric.SuccessCondition)
		if err != nil {
			return v1alpha1.AnalysisPhaseError, err
		}
	}
	if metric.FailureCondition != "" {
		failCondition, err = EvalCondition(result, metric.FailureCondition)
		if err != nil {
			return v1alpha1.AnalysisPhaseError, err
		}
	}

	switch {
	case metric.SuccessCondition == "" && metric.FailureCondition == "":
		// Always return success unless there is an error
		return v1alpha1.AnalysisPhaseSuccessful, nil
	case metric.SuccessCondition != "" && metric.FailureCondition == "":
		// Without a failure condition, a measurement is considered a failure if the measurement's success condition is not true
		failCondition = !successCondition
	case metric.SuccessCondition == "" && metric.FailureCondition != "":
		// Without a success condition, a measurement is considered a successful if the measurement's failure condition is not true
		successCondition = !failCondition
	}

	if failCondition {
		return v1alpha1.AnalysisPhaseFailed, nil
	}

	if !failCondition && !successCondition {
		return v1alpha1.AnalysisPhaseInconclusive, nil
	}

	// If we reach this code path, failCondition is false and successCondition is true
	return v1alpha1.AnalysisPhaseSuccessful, nil
}

// EvalCondition evaluates the condition with the resultValue as an input. This function supports
// lazy evaluation (e.g. result != nil && result > 0)
func EvalCondition(resultValue interface{}, condition string) (bool, error) {
	env := map[string]interface{}{
		"result":  resultValue,
		"asInt":   asInt,
		"asFloat": asFloat,
		"isNaN":   math.IsNaN,
		"isInf":   isInf,
	}
	// first try with strict mode, if it passes, then great
	output, err := eval(resultValue, condition, env, true)
	if err == nil {
		return output, nil
	}
	// if the error is anything but mismatched type check against nil, return the error
	if !mismatchedTypesWithNil.MatchString(err.Error()) {
		return false, err
	}
	// otherwise re-run evaluation but disabling strict mode
	return eval(resultValue, condition, env, false)
}

// unwrapFileErr is a helper to remove multi-line formatting from expr errors
func unwrapFileErr(e error) error {
	if fileErr, ok := e.(*file.Error); ok {
		e = errors.New(fileErr.Message)
	}
	return e
}

// eval is a wrapper on expr Compile and Run, but with an option to disable strict mode.
// Disabling strict allows us to relax checks against strict type checking, unknown variables,
// in order for us to do things like lazy evaluation.
func eval(resultValue interface{}, condition string, env map[string]interface{}, strict bool) (bool, error) {
	//expr.Compile()
	config := conf.New(env)
	config.Strict = strict
	config.Optimize = true
	err := config.Check()
	if err != nil {
		return false, unwrapFileErr(err)
	}
	tree, err := parser.Parse(condition)
	if err != nil {
		return false, unwrapFileErr(err)
	}
	if strict {
		_, err = checker.Check(tree, config)
		if err != nil {
			return false, unwrapFileErr(err)
		}
	}
	program, err := compiler.Compile(tree, config)
	if err != nil {
		return false, unwrapFileErr(err)
	}
	output, err := expr.Run(program, env)
	if err != nil {
		return false, unwrapFileErr(err)
	}
	val, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("expected bool, but got %T", output)
	}
	return val, nil
}

func isInf(f float64) bool {
	return math.IsInf(f, 0)
}

func asInt(in interface{}) int64 {
	switch i := in.(type) {
	case float64:
		return int64(i)
	case float32:
		return int64(i)
	case int64:
		return i
	case int32:
		return int64(i)
	case int16:
		return int64(i)
	case int8:
		return int64(i)
	case int:
		return int64(i)
	case uint64:
		return int64(i)
	case uint32:
		return int64(i)
	case uint16:
		return int64(i)
	case uint8:
		return int64(i)
	case uint:
		return int64(i)
	case string:
		inAsInt, err := strconv.ParseInt(i, 10, 64)
		if err == nil {
			return inAsInt
		}
		panic(err)
	}
	panic(fmt.Sprintf("asInt() not supported on %v %v", reflect.TypeOf(in), in))
}

func asFloat(in interface{}) float64 {
	switch i := in.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int32:
		return float64(i)
	case int16:
		return float64(i)
	case int8:
		return float64(i)
	case int:
		return float64(i)
	case uint64:
		return float64(i)
	case uint32:
		return float64(i)
	case uint16:
		return float64(i)
	case uint8:
		return float64(i)
	case uint:
		return float64(i)
	case string:
		inAsFloat, err := strconv.ParseFloat(i, 64)
		if err == nil {
			return inAsFloat
		}
		panic(err)
	}
	panic(fmt.Sprintf("asFloat() not supported on %v %v", reflect.TypeOf(in), in))
}
