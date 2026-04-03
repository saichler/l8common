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

// ValidateFunc is a function that validates an entity.
type ValidateFunc[T any] func(*T, ifs.IVNic) error

// ActionValidateFunc is a function that validates an entity with access to the CRUD action.
type ActionValidateFunc[T any] func(*T, ifs.Action, ifs.IVNic) error

// SetIDFunc is a function that generates/sets the primary key on an entity.
type SetIDFunc[T any] func(*T)

// NewServiceCallback creates a standard IServiceCallback with generic type safety.
// Delegates to l8common's non-generic implementation.
func NewServiceCallback[T any](typeName string, setID SetIDFunc[T], validate ValidateFunc[T], actionValidators ...ActionValidateFunc[T]) ifs.IServiceCallback {
	avs := make([]l8c.ActionValidateFunc, len(actionValidators))
	for i, av := range actionValidators {
		fn := av
		avs[i] = func(v interface{}, action ifs.Action, vnic ifs.IVNic) error {
			return fn(v.(*T), action, vnic)
		}
	}
	var vf l8c.ValidateFunc
	if validate != nil {
		vf = func(v interface{}, vnic ifs.IVNic) error {
			return validate(v.(*T), vnic)
		}
	}
	return l8c.NewServiceCallback(
		typeName,
		func(v interface{}) bool { _, ok := v.(*T); return ok },
		func(v interface{}) { setID(v.(*T)) },
		vf,
		avs...,
	)
}

// NewServiceCallbackWithAfter creates a ServiceCallback with both action validators
// and after-actions that run after successful PUT/PATCH persistence.
func NewServiceCallbackWithAfter[T any](typeName string, setID SetIDFunc[T], validate ValidateFunc[T], actionValidators []ActionValidateFunc[T], afterActions []ActionValidateFunc[T]) ifs.IServiceCallback {
	avs := make([]l8c.ActionValidateFunc, len(actionValidators))
	for i, av := range actionValidators {
		fn := av
		avs[i] = func(v interface{}, action ifs.Action, vnic ifs.IVNic) error {
			return fn(v.(*T), action, vnic)
		}
	}
	aas := make([]l8c.ActionValidateFunc, len(afterActions))
	for i, aa := range afterActions {
		fn := aa
		aas[i] = func(v interface{}, action ifs.Action, vnic ifs.IVNic) error {
			return fn(v.(*T), action, vnic)
		}
	}
	var vf l8c.ValidateFunc
	if validate != nil {
		vf = func(v interface{}, vnic ifs.IVNic) error {
			return validate(v.(*T), vnic)
		}
	}
	return l8c.NewServiceCallbackWithAfter(
		typeName,
		func(v interface{}) bool { _, ok := v.(*T); return ok },
		func(v interface{}) { setID(v.(*T)) },
		vf,
		avs,
		aas,
	)
}
