package provider

import (
	"github.com/ciscoecosystem/aci-go-client/client"
	"github.com/ciscoecosystem/aci-go-client/container"
	"github.com/ciscoecosystem/aci-go-client/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func preparePayload(className string, inputMap map[string]string, children []interface{}, addAnnotation bool) (*container.Container, error) {
	cont := container.New()
	cont.Object(className)
	cont.Object(className, "attributes")

	if addAnnotation && !containsString(NoAnnotationClasses, className) {
		cont.Set("orchestrator:terraform", className, "attributes", "annotation")
	}
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

		if addAnnotation && !containsString(NoAnnotationClasses, childClassName) {
			childCont.Set("orchestrator:terraform", childClassName, "attributes", "annotation")
		}
		for attr, value := range childContent {
			childCont.Set(value, childClassName, "attributes", attr)
		}
		cont.ArrayAppend(childCont.Data(), className, "children")
	}
	return cont, nil
}

func ApicRest(d *schema.ResourceData, meta interface{}, method string, children bool) (*container.Container, diag.Diagnostics) {
	aciClient := meta.(apiClient).Client
	path := "/api/mo/" + d.Get("dn").(string) + ".json"
	className := d.Get("class_name").(string)
	if method == "GET" {
		if children {
			path += "?rsp-subtree=children"
		} else if !containsString(FullClasses, className) {
			path += "?rsp-prop-include=config-only"
		}
	}
	var cont *container.Container = nil
	var err error

	if method == "POST" {
		content := d.Get("content")
		contentStrMap := toStrMap(content.(map[string]interface{}))

		childrenSet := make([]interface{}, 0, 1)

		for _, child := range d.Get("child").(*schema.Set).List() {
			childMap := make(map[string]interface{})
			childClassName := child.(map[string]interface{})["class_name"]
			childContent := child.(map[string]interface{})["content"]
			childMap["class_name"] = childClassName.(string)
			childMap["content"] = toStrMap(childContent.(map[string]interface{}))
			childrenSet = append(childrenSet, childMap)
		}

		cont, err = preparePayload(className, contentStrMap, childrenSet, meta.(apiClient).IsAnnotation)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	req, err := aciClient.MakeRestRequest(method, path, cont, true)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	respCont, _, err := aciClient.Do(req)
	if err != nil {
		return respCont, diag.FromErr(err)
	}
	if respCont.S("imdata").Index(0).String() == "{}" {
		return nil, nil
	}
	err = client.CheckForErrors(respCont, method, false)
	if err != nil {
		if method == "DELETE" {
			errCode := models.StripQuotes(models.StripSquareBrackets(respCont.Search("imdata", "error", "attributes", "code").String()))
			// Ignore errors of type "Cannot delete object"
			if errCode == "1" || errCode == "107" {
				return respCont, nil
			}
		}
		return respCont, diag.FromErr(err)
	}
	if method == "POST" {
		return cont, nil
	} else {
		return respCont, nil
	}
}
