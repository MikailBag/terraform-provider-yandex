package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"strconv"
	"testing"
)

const albBgDataSourceResource = "data.yandex_alb_backend_group.test-bg-ds"

func TestAccDataSourceALBBackendGroup_byID(t *testing.T) {
	t.Parallel()

	bgName := acctest.RandomWithPrefix("tf-bg")
	bgDesc := "tf-bg-description"
	folderID := getExampleFolderID()

	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBBackendGroupConfigByID(bgName, bgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckResourceIDField(albBgDataSourceResource, "backend_group_id"),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "name", bgName),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "description", bgDesc),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(albBgDataSourceResource),
					testAccCheckALBBackendGroupValues(&bg, false, false),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_byName(t *testing.T) {
	t.Parallel()

	bgName := acctest.RandomWithPrefix("tf-bg")
	bgDesc := "tf-bg-description"
	folderID := getExampleFolderID()

	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBBackendGroupConfigByName(bgName, bgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckResourceIDField(albBgDataSourceResource, "backend_group_id"),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "name", bgName),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "description", bgDesc),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(albBgDataSourceResource),
					testAccCheckALBBackendGroupValues(&bg, false, false),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_fullWithHttpBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsHttpBackend = true

	backendPath := ""
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false),
					testExistsFirstElementWithAttr(
						albBgDataSourceResource, "http_backend", "tls", &backendPath,
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "tls.0.sni", func(value string) error {
							tlsSni := bg.GetHttp().GetBackends()[0].Tls.Sni
							if value != tlsSni {
								return fmt.Errorf("BackendGroup's http backend's tls sni doesnt't match. %s != %s", value, tlsSni)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.locality_aware_routing_percent", func(value string) error {
							lbConfigPercent := bg.GetHttp().GetBackends()[0].LoadBalancingConfig.LocalityAwareRoutingPercent
							if value != strconv.FormatInt(lbConfigPercent, 10) {
								return fmt.Errorf("BackendGroup's http backend's load balancing config locality aware routing percent doesnt't match. %s != %d", value, lbConfigPercent)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.panic_threshold", func(value string) error {
							lbConfigPanicThreshold := bg.GetHttp().GetBackends()[0].LoadBalancingConfig.PanicThreshold
							if value != strconv.FormatInt(lbConfigPanicThreshold, 10) {
								return fmt.Errorf("BackendGroup's http backend's load balancing config panic threshold doesnt't match. %s != %d", value, lbConfigPanicThreshold)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.strict_locality", func(value string) error {
							lbConfigStrictLocality := bg.GetHttp().GetBackends()[0].LoadBalancingConfig.StrictLocality
							if value != strconv.FormatBool(lbConfigStrictLocality) {
								return fmt.Errorf("BackendGroup's http backend's load balancing config panic threshold doesnt't match. %s != %t", value, lbConfigStrictLocality)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_fullWithGrpcBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsGrpcBackend = true

	backendPath := ""
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, true),
					testExistsFirstElementWithAttr(
						albBgDataSourceResource, "grpc_backend", "tls", &backendPath,
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "tls.0.sni", func(value string) error {
							tlsSni := bg.GetGrpc().GetBackends()[0].Tls.Sni
							if value != tlsSni {
								return fmt.Errorf("BackendGroup's grpc backend's tls sni doesnt't match. %s != %s", value, tlsSni)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.locality_aware_routing_percent", func(value string) error {
							lbConfigPercent := bg.GetGrpc().GetBackends()[0].LoadBalancingConfig.LocalityAwareRoutingPercent
							if value != strconv.FormatInt(lbConfigPercent, 10) {
								return fmt.Errorf("BackendGroup's grpc backend's load balancing config locality aware routing percent doesnt't match. %s != %d", value, lbConfigPercent)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.panic_threshold", func(value string) error {
							lbConfigPanicThreshold := bg.GetGrpc().GetBackends()[0].LoadBalancingConfig.PanicThreshold
							if value != strconv.FormatInt(lbConfigPanicThreshold, 10) {
								return fmt.Errorf("BackendGroup's Grpc backend's load balancing config panic threshold doesnt't match. %s != %d", value, lbConfigPanicThreshold)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.strict_locality", func(value string) error {
							lbConfigStrictLocality := bg.GetGrpc().GetBackends()[0].LoadBalancingConfig.StrictLocality
							if value != strconv.FormatBool(lbConfigStrictLocality) {
								return fmt.Errorf("BackendGroup's grpc backend's load balancing config panic threshold doesnt't match. %s != %t", value, lbConfigStrictLocality)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func testAccDataSourceALBBackendGroupExists(bgName string, bg *apploadbalancer.BackendGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[bgName]
		if !ok {
			return fmt.Errorf("Not found: %s", bgName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().BackendGroup().Get(context.Background(), &apploadbalancer.GetBackendGroupRequest{
			BackendGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Backend Group not found")
		}

		*bg = *found

		return nil
	}
}

func testAccDataSourceALBBackendGroupConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_backend_group" "test-bg-ds" {
  backend_group_id = "${yandex_alb_backend_group.test-bg.id}"
}

resource "yandex_alb_backend_group" "test-bg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceALBBackendGroupConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_backend_group" "test-bg-ds" {
  name = "${yandex_alb_backend_group.test-bg.name}"
}

resource "yandex_alb_backend_group" "test-bg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}