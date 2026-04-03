/*
© 2025 Sharon Aicler (saichler@gmail.com)

Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package generic

import (
	l8c "github.com/saichler/l8common/go/common"
	"github.com/saichler/l8types/go/ifs"
)

// RegisterType registers a type and its list wrapper with the introspector and registry.
// Generic wrapper — creates instances via new(T) and new(TList).
func RegisterType[T any, TList any](resources ifs.IResources, pkField string) {
	l8c.RegisterType(resources, new(T), new(TList), pkField)
}
