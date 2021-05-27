package provider

import (
	"github.com/ciscoecosystem/aci-go-client/client"
	"github.com/ciscoecosystem/aci-go-client/container"
	"github.com/ciscoecosystem/aci-go-client/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ApicRest(d *schema.ResourceData, meta interface{}, method string, children bool) (*container.Container, diag.Diagnostics) {
	aciClient := meta.(*client.Client)
	path := "/api/mo/" + d.Get("dn").(string) + ".json"
	if method == "GET" && children {
		path += "?rsp-subtree=children"
	}
	var cont *container.Container = nil
	var err error

	if method == "POST" {
		content := d.Get("content")
		contentStrMap := toStrMap(content.(map[string]interface{}))

		className := d.Get("class_name").(string)

		childrenSet := make([]interface{}, 0, 1)

		for _, child := range d.Get("child").([]interface{}) {
			childMap := make(map[string]interface{})
			childClassName := child.(map[string]interface{})["class_name"]
			childContent := child.(map[string]interface{})["content"]
			childMap["class_name"] = childClassName.(string)
			childMap["content"] = toStrMap(childContent.(map[string]interface{}))
			childrenSet = append(childrenSet, childMap)
		}

		cont, err = preparePayload(className, contentStrMap, childrenSet)
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
