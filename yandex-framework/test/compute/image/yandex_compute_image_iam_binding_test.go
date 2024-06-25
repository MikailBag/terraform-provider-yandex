package image

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test/compute/iam"
)

const (
	standardImagesFolderID = "standard-images"
	timeout                = time.Minute * 15
)

func TestAccComputeImage_basicIamMember(t *testing.T) {
	var (
		image       compute.Image
		userID      = "allUsers"
		role        = "editor"
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeImageWithIAM_basic("image-test-"+acctest.RandString(8), role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeImageExists("yandex_compute_image.foobar", &image),
					iam.TestAccCheckIamBindingExists(ctx, func() iam.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().Image()
					}, &image, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccCheckComputeImageExists(n string, image *compute.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().Image().Get(context.Background(), &compute.GetImageRequest{
			ImageId: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Image not found")
		}

		*image = *found

		return nil
	}
}

func testAccComputeImageWithIAM_basic(name, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_compute_image" "foobar" {
  name          = "%s"
  description   = "description-test"
  family        = "ubuntu-1804-lts"
  source_family = "ubuntu-1804-lts"
  min_disk_size = 10
  os_type       = "linux"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}

resource "yandex_compute_image_iam_binding" "test-image-bind" {
  role = "%s"
  members = ["system:%s"]
  image_id = yandex_compute_image.foobar.id
}
`, name, role, userID)
}

func testAccCheckComputeImageDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_image" {
			continue
		}

		r, err := config.SDK.Compute().Image().Get(context.Background(), &compute.GetImageRequest{
			ImageId: rs.Primary.ID,
		})

		// Do not trigger error on images from "standard-images" folder
		if err == nil && r.FolderId != standardImagesFolderID {
			return fmt.Errorf("Image still exists: %q", r)
		}
	}

	return nil
}
