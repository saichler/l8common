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

// StatusTransitionConfig defines a state machine for an entity's status field.
// Generic wrapper — typed getters/setters delegate to l8common's interface{}-based implementation.
type StatusTransitionConfig[T any] struct {
	StatusGetter  func(*T) int32
	StatusSetter  func(*T, int32)
	FilterBuilder func(*T) *T
	ServiceName   string
	ServiceArea   byte
	InitialStatus int32
	Transitions   map[int32][]int32
	StatusNames   map[int32]string
}

// BuildValidator returns an ActionValidateFunc that enforces status transitions.
func (cfg *StatusTransitionConfig[T]) BuildValidator() ActionValidateFunc[T] {
	inner := &l8c.StatusTransitionConfig{
		StatusGetter:  func(v interface{}) int32 { return cfg.StatusGetter(v.(*T)) },
		StatusSetter:  func(v interface{}, s int32) { cfg.StatusSetter(v.(*T), s) },
		FilterBuilder: func(v interface{}) interface{} { return cfg.FilterBuilder(v.(*T)) },
		ServiceName:   cfg.ServiceName,
		ServiceArea:   cfg.ServiceArea,
		InitialStatus: cfg.InitialStatus,
		Transitions:   cfg.Transitions,
		StatusNames:   cfg.StatusNames,
	}
	innerFn := inner.BuildValidator()
	return func(entity *T, action ifs.Action, vnic ifs.IVNic) error {
		return innerFn(entity, action, vnic)
	}
}
