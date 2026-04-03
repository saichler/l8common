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
	"google.golang.org/protobuf/proto"
)

// ProtoMessage constrains T to be a protobuf message type.
type ProtoMessage[T any] interface {
	*T
	proto.Message
}

// ServiceConfig holds the configuration for activating a service.
type ServiceConfig = l8c.ServiceConfig

// ActivateService sets up and activates a service with the standard boilerplate.
// Generic wrapper over l8common's non-generic ActivateService.
func ActivateService[T any, TList any, PT ProtoMessage[T], PTL ProtoMessage[TList]](cfg ServiceConfig, creds, dbname string, vnic ifs.IVNic) {
	l8c.ActivateService(cfg, PT(new(T)), PTL(new(TList)), creds, dbname, vnic)
}

// ServiceHandler returns the service handler for the given service.
var ServiceHandler = l8c.ServiceHandler

// GetEntity retrieves a single entity by its filter, trying local first then remote.
// Generic wrapper — returns typed *T instead of interface{}.
func GetEntity[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) (*T, error) {
	result, err := l8c.GetEntity(serviceName, serviceArea, filter, vnic)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*T), nil
}

// GetEntities retrieves all entities matching a filter.
// Generic wrapper — returns typed []*T instead of []interface{}.
func GetEntities[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) ([]*T, error) {
	results, err := l8c.GetEntities(serviceName, serviceArea, filter, vnic)
	if err != nil {
		return nil, err
	}
	typed := make([]*T, 0, len(results))
	for _, r := range results {
		if t, ok := r.(*T); ok {
			typed = append(typed, t)
		}
	}
	return typed, nil
}

// PutEntity updates an entity via its service handler.
func PutEntity[T any](serviceName string, serviceArea byte, entity *T, vnic ifs.IVNic) error {
	return l8c.PutEntity(serviceName, serviceArea, entity, vnic)
}

// PostEntity creates a new entity via its service handler.
// Generic wrapper — returns typed *T instead of interface{}.
func PostEntity[T any](serviceName string, serviceArea byte, entity *T, vnic ifs.IVNic) (*T, error) {
	result, err := l8c.PostEntity(serviceName, serviceArea, entity, vnic)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*T), nil
}

// EntityExists checks if any entity matching the filter already exists.
func EntityExists[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) (bool, error) {
	return l8c.EntityExists(serviceName, serviceArea, filter, vnic)
}
