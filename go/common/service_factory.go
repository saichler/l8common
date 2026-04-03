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
package common

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/saichler/l8orm/go/orm/persist"
	"github.com/saichler/l8orm/go/orm/plugins/postgres"
	"github.com/saichler/l8srlz/go/serialize/object"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8web"
	"github.com/saichler/l8utils/go/utils/web"
	"google.golang.org/protobuf/proto"
	"reflect"
)

// ProtoMessage constrains T to be a protobuf message type.
type ProtoMessage[T any] interface {
	*T
	proto.Message
}

// ServiceConfig holds the configuration for activating a service.
type ServiceConfig struct {
	ServiceName  string
	ServiceArea  byte
	PrimaryKey   string
	Callback     ifs.IServiceCallback
	ServiceGroup string
}

// ActivateService sets up and activates a service with the standard boilerplate.
func ActivateService[T any, TList any, PT ProtoMessage[T], PTL ProtoMessage[TList]](cfg ServiceConfig, creds, dbname string, vnic ifs.IVNic) {
	if cfg.PrimaryKey == "" {
		panic(fmt.Sprintf("service %s (area %d): PrimaryKey is required", cfg.ServiceName, cfg.ServiceArea))
	}
	realdb, user, pass, _, err := vnic.Resources().Security().Credential(creds, dbname, vnic.Resources())
	if err != nil {
		panic(err)
	}
	db := OpenDBConection(realdb, user, pass)
	p := postgres.NewPostgres(db, vnic.Resources())

	sla := ifs.NewServiceLevelAgreement(&persist.OrmService{}, cfg.ServiceName, cfg.ServiceArea, true, cfg.Callback)
	sla.SetServiceItem(PT(new(T)))
	sla.SetServiceItemList(PTL(new(TList)))
	sla.SetPrimaryKeys(cfg.PrimaryKey)
	sla.SetArgs(p, true)
	sla.SetTransactional(true)
	sla.SetReplication(true)
	sla.SetReplicationCount(3)

	ws := web.New(cfg.ServiceName, cfg.ServiceArea, 0)
	ws.AddEndpoint(PT(new(T)), ifs.POST, &l8web.L8Empty{})
	ws.AddEndpoint(PTL(new(TList)), ifs.POST, &l8web.L8Empty{})
	ws.AddEndpoint(PT(new(T)), ifs.PUT, &l8web.L8Empty{})
	ws.AddEndpoint(PT(new(T)), ifs.PATCH, &l8web.L8Empty{})
	ws.AddEndpoint(&l8api.L8Query{}, ifs.DELETE, &l8web.L8Empty{})
	ws.AddEndpoint(&l8api.L8Query{}, ifs.GET, PTL(new(TList)))
	sla.SetWebService(ws)

	serviceGroup := cfg.ServiceGroup
	if serviceGroup == "" {
		serviceGroup = "L8SG"
	}
	sla.SetServiceGroup(serviceGroup)
	vnic.Resources().Services().Activate(sla, vnic)
}

// ServiceHandler returns the service handler for the given service.
func ServiceHandler(serviceName string, serviceArea byte, vnic ifs.IVNic) (ifs.IServiceHandler, bool) {
	return vnic.Resources().Services().ServiceHandler(serviceName, serviceArea)
}

// GetEntity retrieves a single entity by its filter, trying local first then remote.
func GetEntity[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) (*T, error) {
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Get(object.New(nil, filter), vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		if resp.Element() != nil {
			return resp.Element().(*T), nil
		}
		return nil, nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, filter, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	if resp.Element() != nil {
		return resp.Element().(*T), nil
	}
	return nil, nil
}

// isFilterEmpty returns true if the filter struct has all zero-value fields.
func isFilterEmpty(filter interface{}) bool {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.IsZero()
}

// filterTypeName returns the protobuf type name of the filter struct.
func filterTypeName(filter interface{}) string {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Type().Name()
}

// GetEntities retrieves all entities matching a filter.
// When the filter is empty (all zero values), it uses an L8Query to fetch all entities.
func GetEntities[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) ([]*T, error) {
	if isFilterEmpty(filter) {
		return getAllEntities[T](serviceName, serviceArea, filter, vnic)
	}
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Get(object.New(nil, filter), vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		return extractElements[T](resp.Elements()), nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, filter, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return extractElements[T](resp.Elements()), nil
}

// getAllEntities fetches all entities using an L8Query when the filter is empty.
func getAllEntities[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) ([]*T, error) {
	typeName := filterTypeName(filter)
	query := fmt.Sprintf("select * from %s", typeName)
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		elems, err := object.NewQuery(query, vnic.Resources())
		if err != nil {
			return nil, err
		}
		resp := handler.Get(elems, vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		return extractElements[T](resp.Elements()), nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, query, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return extractElements[T](resp.Elements()), nil
}

func extractElements[T any](elems []interface{}) []*T {
	result := make([]*T, 0, len(elems))
	for _, e := range elems {
		if t, ok := e.(*T); ok {
			result = append(result, t)
		}
	}
	return result
}

// PutEntity updates an entity via its service handler.
func PutEntity[T any](serviceName string, serviceArea byte, entity *T, vnic ifs.IVNic) error {
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Put(object.New(nil, entity), vnic)
		if resp.Error() != nil {
			return resp.Error()
		}
		return nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.PUT, entity, 30)
	if resp.Error() != nil {
		return resp.Error()
	}
	return nil
}

// PostEntity creates a new entity via its service handler.
func PostEntity[T any](serviceName string, serviceArea byte, entity *T, vnic ifs.IVNic) (*T, error) {
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Post(object.New(nil, entity), vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		if resp.Element() != nil {
			return resp.Element().(*T), nil
		}
		return entity, nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.POST, entity, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	if resp.Element() != nil {
		return resp.Element().(*T), nil
	}
	return entity, nil
}

// EntityExists checks if any entity matching the filter already exists.
func EntityExists[T any](serviceName string, serviceArea byte, filter *T, vnic ifs.IVNic) (bool, error) {
	existing, err := GetEntities[T](serviceName, serviceArea, filter, vnic)
	if err != nil {
		return false, err
	}
	return len(existing) > 0, nil
}
