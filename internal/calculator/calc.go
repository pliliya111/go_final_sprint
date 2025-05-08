package calculator

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pliliya111/go_final_sprint/internal/model"
)

var (
	timeAdditionMS       = getEnvInt("TIME_ADDITION_MS", 1000)
	timeSubtractionMS    = getEnvInt("TIME_SUBTRACTION_MS", 1000)
	timeMultiplicationMS = getEnvInt("TIME_MULTIPLICATIONS_MS", 1000)
	timeDivisionMS       = getEnvInt("TIME_DIVISIONS_MS", 1000)
	Results              = make(map[string]interface{})
)

func getEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func PerformOperation(task *model.Task) interface{} {

	if arg1ID, ok := task.Arg1.(string); ok {
		if result, exists := Results[arg1ID]; exists {
			task.Arg1 = result
		}
	}
	if arg2ID, ok := task.Arg2.(string); ok {
		if result, exists := Results[arg2ID]; exists {
			task.Arg2 = result
		}
	}

	if arg1Str, ok := task.Arg1.(string); ok {
		arg1, err := strconv.ParseFloat(arg1Str, 64)
		if err != nil {
			return fmt.Errorf("invalid argument: arg1=%v (cannot convert to float64)", task.Arg1)
		}
		task.Arg1 = arg1
	}
	if arg2Str, ok := task.Arg2.(string); ok {
		arg2, err := strconv.ParseFloat(arg2Str, 64)
		if err != nil {
			return fmt.Errorf("invalid argument: arg2=%v (cannot convert to float64)", task.Arg2)
		}
		task.Arg2 = arg2
	}

	arg1, ok1 := task.Arg1.(float64)
	arg2, ok2 := task.Arg2.(float64)

	if !ok1 || !ok2 {
		return fmt.Errorf("invalid arguments for operation: arg1=%v, arg2=%v", task.Arg1, task.Arg2)
	}

	switch task.Operation {
	case "+":
		time.Sleep(time.Duration(timeAdditionMS) * time.Millisecond)
		return arg1 + arg2
	case "-":
		time.Sleep(time.Duration(timeSubtractionMS) * time.Millisecond)
		return arg1 - arg2
	case "*":
		time.Sleep(time.Duration(timeMultiplicationMS) * time.Millisecond)
		return arg1 * arg2
	case "/":
		time.Sleep(time.Duration(timeDivisionMS) * time.Millisecond)
		if arg2 == 0 {
			return fmt.Errorf("division by zero")
		}
		return arg1 / arg2
	default:
		return fmt.Errorf("unknown operation: %s", task.Operation)
	}
}
