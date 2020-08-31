package main

import (
	"fmt"
	"reflect"
	"time"
)

func decorateCallbackWithAttempt(decoPtr, fn interface{}, retryInterval int64, attempt int64) (err error) {
	defer func() {
		if err1 := recover(); err1 != nil {
			err = fmt.Errorf("decorater err: %v", err1)
		}
	}()

	var decoratedFunc, targetFunc reflect.Value

	decoratedFunc = reflect.ValueOf(decoPtr).Elem()
	targetFunc = reflect.ValueOf(fn)

	v := reflect.MakeFunc(targetFunc.Type(),
		func(in []reflect.Value) (out []reflect.Value) {
			for retry := attempt; retry > 0; retry-- {
				hasErr := false
				//Call callback func
				out = targetFunc.Call(in)

				//Has return val
				if valuesNum := len(out); valuesNum > 0 {
					resultItems := make([]interface{}, valuesNum)

					//Check value
					for k, val := range out {
						resultItems[k] = val.Interface()
						//Has error
						if _, ok := resultItems[k].(error); ok {
							hasErr = true
							break
						}
					}

					//Has err, retry
					if hasErr {
						time.Sleep(time.Duration(retryInterval) * time.Second)
						fmt.Println("retry &d\n", retry)
						continue
					}
					return
				}
			}
			return out
		})
	decoratedFunc.Set(v)
	return
}
