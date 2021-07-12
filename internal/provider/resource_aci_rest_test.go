package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ciscoecosystem/aci-go-client/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAciRest_tenant(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckAciRestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAciRestConfig_tenant(name, "Create description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAciRestObject("aci_rest.fvTenant"),
				),
			},
			{
				ResourceName:      "aci_rest.fvTenant",
				ImportState:       true,
				ImportStateId:     "fvTenant:uni/tn-" + name,
				ImportStateVerify: true,
			},
			{
				Config: testAccAciRestConfig_tenant(name, "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAciRestObject("aci_rest.fvTenant"),
				),
			},
		},
	})
}

func TestAccAciRest_connPref(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckAciRestStillExists,
		Steps: []resource.TestStep{
			{
				Config: testAccAciRestConfig_connPref("ooband"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAciRestObject("aci_rest.mgmtConnectivityPrefs"),
				),
			},
			{
				ResourceName:      "aci_rest.mgmtConnectivityPrefs",
				ImportState:       true,
				ImportStateId:     "mgmtConnectivityPrefs:uni/fabric/connectivityPrefs",
				ImportStateVerify: true,
			},
			{
				Config: testAccAciRestConfig_connPref("inband"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAciRestObject("aci_rest.mgmtConnectivityPrefs"),
				),
			},
		},
	})
}

func TestAccAciRest_noContent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckAciRestStillExists,
		Steps: []resource.TestStep{
			{
				Config: testAccAciRestConfig_connPref(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAciRestObject("aci_rest.mgmtConnectivityPrefs"),
				),
			},
			{
				ResourceName:      "aci_rest.mgmtConnectivityPrefs",
				ImportState:       true,
				ImportStateId:     "mgmtConnectivityPrefs:uni/fabric/connectivityPrefs",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAciRest_tenantVrf(t *testing.T) {
	name := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCheckAciRestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAciRestConfig_tenantVrf(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aci_rest.fvTenant", "child.0.class_name", "fvCtx"),
					resource.TestCheckResourceAttr("aci_rest.fvTenant", "child.0.rn", "ctx-"+name),
					resource.TestCheckResourceAttr("aci_rest.fvTenant", "child.0.content.name", name),
				),
			},
		},
	})
}

func testAccAciRestConfig_tenant(name string, description string) string {
	return fmt.Sprintf(`
	resource "aci_rest" "fvTenant" {
		dn = "uni/tn-%[1]s"
		class_name = "fvTenant"
		content = {
			name = "%[1]s"
			descr = "%[2]s"
			nameAlias = "Testacc_Tenant"
		}
	}
	`, name, description)
}

func testAccAciRestConfig_connPref(status string) string {
	if status != "" {
		return fmt.Sprintf(`
		resource "aci_rest" "mgmtConnectivityPrefs" {
			dn = "uni/fabric/connectivityPrefs"
			class_name = "mgmtConnectivityPrefs"
			content = {
				interfacePref = "%[1]s"
			}
		}
		`, status)
	} else {
		return `
		resource "aci_rest" "mgmtConnectivityPrefs" {
			dn = "uni/fabric/connectivityPrefs"
			class_name = "mgmtConnectivityPrefs"
		}
		`
	}
}

func testAccAciRestConfig_tenantVrf(name string) string {
	return fmt.Sprintf(`
	resource "aci_rest" "fvTenant" {
		dn = "uni/tn-%[1]s"
		class_name = "fvTenant"
		content = {
			name = "%[1]s"
		}

		child {
			rn         = "ctx-%[1]s"
			class_name = "fvCtx"
			content = {
			  name = "%[1]s"
			}
		}
	}
	`, name)
}

func testAccCheckAciRestObject(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Resource aci_rest %s not found", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No aci_rest dn attribute was set")
		}

		client := testAccProvider.Meta().(apiClient).Client

		cont, err := client.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		className := rs.Primary.Attributes["class_name"]
		dn := models.StripQuotes(models.StripSquareBrackets(cont.Search("imdata", className, "attributes", "dn").String()))

		if dn != rs.Primary.ID {
			return fmt.Errorf("APIC object %s not found", rs.Primary.ID)
		}

		for key, value := range rs.Primary.Attributes {
			if strings.Contains(key, "content.") && key != "content.%" {
				attr := key[8:]
				v := models.StripQuotes(models.StripSquareBrackets(cont.Search("imdata", className, "attributes", attr).String()))
				if v != value {
					return fmt.Errorf("APIC object %s, expected: %s, got: %s", rs.Primary.ID, value, v)
				}
			}
		}

		return nil
	}
}

func testAccCheckAciRestDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(apiClient).Client

	for _, rs := range s.RootModule().Resources {

		if rs.Type == "aci_rest" {
			_, err := client.Get(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Resource aci_rest %s still exists", rs.Primary.ID)
			}

		} else {
			continue
		}
	}

	return nil
}

func testAccCheckAciRestStillExists(s *terraform.State) error {
	client := testAccProvider.Meta().(apiClient).Client

	for _, rs := range s.RootModule().Resources {

		if rs.Type == "aci_rest" {
			_, err := client.Get(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("Error retrieving resource aci_rest %s", rs.Primary.ID)
			}

		} else {
			continue
		}
	}

	return nil
}
