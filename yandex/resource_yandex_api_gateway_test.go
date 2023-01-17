package yandex

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/getenv"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
)

const apiGatewayResource = "yandex_api_gateway.test-api-gateway"
const specFile = "test-fixtures/serverless/main.yaml"
const specFileUpdated = "test-fixtures/serverless/main_updated.yaml"

var spec string
var specUpdated string

func init() {
	resource.AddTestSweepers("yandex_api_gateway", &resource.Sweeper{
		Name:         "yandex_api_gateway",
		F:            testSweepAPIGateway,
		Dependencies: []string{},
	})
	fileBytes, _ := ioutil.ReadFile(specFile)
	spec = string(fileBytes)
	fileBytes, _ = ioutil.ReadFile(specFileUpdated)
	specUpdated = string(fileBytes)
}

func testSweepAPIGateway(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &apigateway.ListApiGatewayRequest{FolderId: conf.FolderID}
	it := conf.sdk.Serverless().APIGateway().ApiGateway().ApiGatewayIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepAPIGateway(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep API Gateway %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepAPIGateway(conf *Config, id string) bool {
	return sweepWithRetry(sweepAPIGatewayOnce, conf, "API Gateway", id)
}

func sweepAPIGatewayOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexApiGatewayDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().APIGateway().ApiGateway().Delete(ctx, &apigateway.DeleteApiGatewayRequest{
		ApiGatewayId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexAPIGateway_basic(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")
	labelKey := acctest.RandomWithPrefix("tf-api-gateway-label")
	labelValue := acctest.RandomWithPrefix("tf-api-gateway-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			basicYandexAPIGatewayTestStep(apiGatewayName, apiGatewayDesc, labelKey, labelValue, spec, &apiGateway),
		},
	})
}

func TestAccYandexAPIGateway_update(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	apiGatewayName := acctest.RandomWithPrefix("tf-api-gateway")
	apiGatewayDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")
	labelKey := acctest.RandomWithPrefix("tf-api-gateway-label")
	labelValue := acctest.RandomWithPrefix("tf-api-gateway-label-value")

	apiGatewayNameUpdated := acctest.RandomWithPrefix("tf-api-gateway-updated")
	apiGatewayDescUpdated := acctest.RandomWithPrefix("tf-api-gateway-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-api-gateway-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-api-gateway-label-value-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			basicYandexAPIGatewayTestStep(apiGatewayName, apiGatewayDesc, labelKey, labelValue, spec, &apiGateway),
			basicYandexAPIGatewayTestStep(apiGatewayNameUpdated, apiGatewayDescUpdated, labelKeyUpdated, labelValueUpdated, specUpdated, &apiGateway),
		},
	})
}

func TestAccYandexAPIGateway_full(t *testing.T) {
	t.Parallel()

	var apiGateway apigateway.ApiGateway
	params := testYandexAPIGatewayParameters{}
	params.name = acctest.RandomWithPrefix("tf-api-gateway")
	params.desc = acctest.RandomWithPrefix("tf-api-gateway-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-api-gateway-label")
	params.labelValue = acctest.RandomWithPrefix("tf-api-gateway-label-value")

	paramsUpdated := testYandexAPIGatewayParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-api-gateway-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-api-gateway-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-api-gateway-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-api-gateway-label-value-updated")

	testConfigFunc := func(params testYandexAPIGatewayParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexAPIGatewayFull(params),
			Check: resource.ComposeTestCheckFunc(
				testYandexAPIGatewayExists(apiGatewayResource, &apiGateway),
				resource.TestCheckResourceAttr(apiGatewayResource, "name", params.name),
				resource.TestCheckResourceAttr(apiGatewayResource, "description", params.desc),
				resource.TestCheckResourceAttr(apiGatewayResource, "spec", spec),
				resource.TestCheckResourceAttrSet(apiGatewayResource, "folder_id"),
				testYandexAPIGatewayContainsLabel(&apiGateway, params.labelKey, params.labelValue),
				testYandexAPIGatewayContainsUserDomains(&apiGateway, make(map[string]struct{})),
				testAccCheckCreatedAtAttr(apiGatewayResource),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(params),
			testConfigFunc(paramsUpdated),
		},
	})
}

func TestAccYandexAPIGateway_domainsUpdate(t *testing.T) {
	t.Parallel()

	testName := acctest.RandomWithPrefix("tf-api-gateway")
	testDesc := acctest.RandomWithPrefix("tf-api-gateway-desc")
	testCertificateId := getTestCertificateId()
	testDomain1 := fmt.Sprintf("%s.tf-acc-tests.prod.apigwtest.ru", acctest.RandomWithPrefix("test"))
	testDomain2 := fmt.Sprintf("%s.tf-acc-tests.prod.apigwtest.ru", acctest.RandomWithPrefix("test"))

	// initial API Gateway creation
	createConfig := fmt.Sprintf(`
resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  custom_domains {
    certificate_id = "%s"
    domain = "%s"
  }
  spec = <<EOF
%sEOF
}`, testName, testDesc, testCertificateId, testDomain1, spec)
	testCreateFunc := resource.TestStep{
		Config: createConfig,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.0.certificate_id", testCertificateId),
			resource.TestCheckResourceAttrSet(apiGatewayResource, "custom_domains.0.domain_id"),
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.0.domain", testDomain1),
			resource.TestCheckNoResourceAttr(apiGatewayResource, "custom_domains.1"),
		),
	}

	// add domain
	addDomainConfig := fmt.Sprintf(`
resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  custom_domains {
    certificate_id = "%s"
    domain = "%s"
  }
  custom_domains {
    certificate_id = "%s"
    domain = "%s"
  }
  spec = <<EOF
%sEOF
}`, testName, testDesc, testCertificateId, testDomain1, testCertificateId, testDomain2, spec)
	testAddDomainFunc := resource.TestStep{
		Config: addDomainConfig,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.0.certificate_id", testCertificateId),
			resource.TestCheckResourceAttrSet(apiGatewayResource, "custom_domains.0.domain_id"),
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.0.domain", testDomain2),
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.1.certificate_id", testCertificateId),
			resource.TestCheckResourceAttrSet(apiGatewayResource, "custom_domains.1.domain_id"),
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.1.domain", testDomain1),
			resource.TestCheckNoResourceAttr(apiGatewayResource, "custom_domains.2"),
		),
	}

	// remove domain
	removeDomainConfig := fmt.Sprintf(`
resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  custom_domains {
    certificate_id = "%s"
    domain = "%s"
  }
  spec = <<EOF
%sEOF
}`, testName, testDesc, testCertificateId, testDomain2, spec)
	testRemoveDomainFunc := resource.TestStep{
		Config: removeDomainConfig,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.0.certificate_id", testCertificateId),
			resource.TestCheckResourceAttrSet(apiGatewayResource, "custom_domains.0.domain_id"),
			resource.TestCheckResourceAttr(apiGatewayResource, "custom_domains.0.domain", testDomain2),
			resource.TestCheckNoResourceAttr(apiGatewayResource, "custom_domains.1"),
		),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexAPIGatewayDestroy,
		Steps: []resource.TestStep{
			testCreateFunc,
			testAddDomainFunc,
			testRemoveDomainFunc,
		},
	})
}

func basicYandexAPIGatewayTestStep(apiGatewayName, apiGatewayDesc, labelKey, labelValue string, spec string, apiGateway *apigateway.ApiGateway) resource.TestStep {
	return resource.TestStep{
		Config: testYandexAPIGatewayBasic(apiGatewayName, apiGatewayDesc, labelKey, labelValue, spec),
		Check: resource.ComposeTestCheckFunc(
			testYandexAPIGatewayExists(apiGatewayResource, apiGateway),
			resource.TestCheckResourceAttr(apiGatewayResource, "name", apiGatewayName),
			resource.TestCheckResourceAttr(apiGatewayResource, "description", apiGatewayDesc),
			resource.TestCheckResourceAttr(apiGatewayResource, "spec", spec),
			resource.TestCheckResourceAttrSet(apiGatewayResource, "folder_id"),
			testYandexAPIGatewayContainsLabel(apiGateway, labelKey, labelValue),
			testYandexAPIGatewayContainsUserDomains(apiGateway, make(map[string]struct{})),
			testAccCheckCreatedAtAttr(apiGatewayResource),
		),
	}
}

func testYandexAPIGatewayDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_api_gateway" {
			continue
		}

		_, err := testGetAPIGatewayByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("APIGateway still exists")
		}
	}

	return nil
}

func testYandexAPIGatewayExists(name string, apiGateway *apigateway.ApiGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetAPIGatewayByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway not found")
		}

		*apiGateway = *found
		return nil
	}
}

func testGetAPIGatewayByID(config *Config, ID string) (*apigateway.ApiGateway, error) {
	req := apigateway.GetApiGatewayRequest{
		ApiGatewayId: ID,
	}

	return config.sdk.Serverless().APIGateway().ApiGateway().Get(context.Background(), &req)
}

func testYandexAPIGatewayContainsLabel(apiGateway *apigateway.ApiGateway, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := apiGateway.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexAPIGatewayContainsUserDomains(apiGateway *apigateway.ApiGateway, domains map[string]struct{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attachedDomains := apiGateway.AttachedDomains
		expectedLen := len(domains)
		actualLen := len(attachedDomains)
		if actualLen != expectedLen {
			return fmt.Errorf("Incorrect number of attached domains: expected '%q' but found '%q'", expectedLen, actualLen)
		}

		for _, domain := range attachedDomains {
			domainId := domain.DomainId
			if _, ok := domains[domainId]; !ok {
				return fmt.Errorf("Domain '%s' was not expected to be attached", domainId)
			}

			delete(domains, domainId)
		}

		return nil
	}
}

func testYandexAPIGatewayBasic(name, desc, labelKey, labelValue string, spec string) string {
	return fmt.Sprintf(`
resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  spec = <<EOF
%sEOF
}
	`,
		name,
		desc,
		labelKey,
		labelValue,
		spec)
}

type testYandexAPIGatewayParameters struct {
	name          string
	desc          string
	labelKey      string
	labelValue    string
	certificateId string
	domain        string
}

func testYandexAPIGatewayFull(params testYandexAPIGatewayParameters) string {
	return fmt.Sprintf(`
resource "yandex_api_gateway" "test-api-gateway" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  spec = <<EOF
%sEOF
}
	`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		spec)
}

func getTestCertificateId() string {
	return getenv.Strict("APIGW_TEST_CERTIFICATE_ID")
}
