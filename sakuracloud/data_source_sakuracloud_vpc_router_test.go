// Copyright 2016-2019 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakuracloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccSakuraCloudDataSourceVPCRouter_Basic(t *testing.T) {
	randString1 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	randString2 := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	name := fmt.Sprintf("%s_%s", randString1, randString2)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		CheckDestroy:              testAccCheckSakuraCloudVPCRouterDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraCloudDataSourceVPCRouterBase(name),
				Check:  testAccCheckSakuraCloudDataSourceExists("sakuracloud_vpc_router.foobar"),
			},
			{
				Config: testAccCheckSakuraCloudDataSourceVPCRouterConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudDataSourceExists("data.sakuracloud_vpc_router.foobar"),
					resource.TestCheckResourceAttr("data.sakuracloud_vpc_router.foobar", "name", name),
					resource.TestCheckResourceAttr("data.sakuracloud_vpc_router.foobar", "description", "description_test"),
					resource.TestCheckResourceAttr("data.sakuracloud_vpc_router.foobar", "tags.#", "3"),
					resource.TestCheckResourceAttr("data.sakuracloud_vpc_router.foobar", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("data.sakuracloud_vpc_router.foobar", "tags.1", "tag2"),
					resource.TestCheckResourceAttr("data.sakuracloud_vpc_router.foobar", "tags.2", "tag3"),
				),
			},
			{
				Config: testAccCheckSakuraCloudDataSourceVPCRouterConfig_With_Tag(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudDataSourceExists("data.sakuracloud_vpc_router.foobar"),
				),
			},
			{
				Config: testAccCheckSakuraCloudDataSourceVPCRouterConfig_NotExists(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudDataSourceNotExists("data.sakuracloud_vpc_router.foobar"),
				),
				Destroy: true,
			},
			{
				Config: testAccCheckSakuraCloudDataSourceVPCRouterConfig_With_NotExists_Tag(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudDataSourceNotExists("data.sakuracloud_vpc_router.foobar"),
				),
				Destroy: true,
			},
		},
	})
}

func testAccCheckSakuraCloudDataSourceVPCRouterBase(name string) string {
	return fmt.Sprintf(`
resource sakuracloud_vpc_router "foobar" {
  plan = "standard"

  name = "%s"
  description = "description_test"
  tags = ["tag1","tag2","tag3"]
}
`, name)
}

func testAccCheckSakuraCloudDataSourceVPCRouterConfig(name string) string {
	return fmt.Sprintf(`
resource sakuracloud_vpc_router "foobar" {
  plan = "standard"

  name = "%s"
  description = "description_test"
  tags = ["tag1","tag2","tag3"]
}
data "sakuracloud_vpc_router" "foobar" {
  filters {
	names = ["%s"]
  }
}`, name, name)
}

func testAccCheckSakuraCloudDataSourceVPCRouterConfig_With_Tag(name string) string {
	return fmt.Sprintf(`
resource sakuracloud_vpc_router "foobar" {
  plan = "standard"

  name = "%s"
  description = "description_test"
  tags = ["tag1","tag2","tag3"]
}
data "sakuracloud_vpc_router" "foobar" {
  filters {
	tags = ["tag1","tag3"]
  }
}`, name)
}

func testAccCheckSakuraCloudDataSourceVPCRouterConfig_With_NotExists_Tag(name string) string {
	return fmt.Sprintf(`
resource sakuracloud_vpc_router "foobar" {
  plan = "standard"

  name = "%s"
  description = "description_test"
  tags = ["tag1","tag2","tag3"]
}
data "sakuracloud_vpc_router" "foobar" {
  filters {
	tags = ["tag1-xxxxxxx","tag3-xxxxxxxx"]
  }
}`, name)
}

func testAccCheckSakuraCloudDataSourceVPCRouterConfig_NotExists(name string) string {
	return fmt.Sprintf(`
resource sakuracloud_vpc_router "foobar" {
  plan = "standard"

  name = "%s"
  description = "description_test"
  tags = ["tag1","tag2","tag3"]
}
data "sakuracloud_vpc_router" "foobar" {
  filters {
	names = ["xxxxxxxxxxxxxxxxxx"]
  }
}`, name)
}
