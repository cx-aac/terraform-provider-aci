package provider

import (
	"math"
	"math/rand"
	"time"

	"github.com/ciscoecosystem/aci-go-client/container"
)

const Retries = 3
const Factor = 2
const MinDelay = 4 * time.Second

func toStrMap(inputMap map[string]interface{}) map[string]string {
	rt := make(map[string]string)
	for key, value := range inputMap {
		rt[key] = value.(string)
	}

	return rt
}

func preparePayload(className string, inputMap map[string]string, children []interface{}) (*container.Container, error) {
	cont := container.New()
	cont.Object(className)
	cont.Object(className, "attributes")

	for attr, value := range inputMap {
		cont.Set(value, className, "attributes", attr)
	}
	cont.Array(className, "children")
	for _, child := range children {
		childMap := child.(map[string]interface{})
		childClassName := childMap["class_name"].(string)
		childContent := childMap["content"].(map[string]string)

		childCont := container.New()
		childCont.Object(childClassName)
		childCont.Object(childClassName, "attributes")

		for attr, value := range childContent {
			childCont.Set(value, childClassName, "attributes", attr)
		}
		cont.ArrayAppend(childCont.Data(), className, "children")
	}
	return cont, nil
}

func backoff(attempts int) bool {
	if attempts >= Retries {
		return false
	}
	min := float64(MinDelay)
	backoff := min * math.Pow(Factor, float64(attempts))
	backoff = rand.Float64()*(backoff-min) + min
	time.Sleep(time.Duration(backoff))
	return true
}
