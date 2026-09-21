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

// NewOrmSLA builds the SLA for an ORM-backed service with the house defaults:
// transactional, replicated 3 ways, service group "L8SG", not a voter.
//
// The SLA *is* the configuration. Anything the service needs beyond the
// defaults is set on it directly by the caller — SetVoter, SetUniqueKeys,
// SetNonUniqueKeys, SetReplication, ... — so a setter added to the SLA later
// is immediately usable everywhere, instead of silently unavailable until
// this package grows a matching field.
func NewOrmSLA(serviceName string, serviceArea byte, primaryKey string, callback ifs.IServiceCallback,
	serviceItem proto.Message, serviceItemList proto.Message) *ifs.ServiceLevelAgreement {

	sla := ifs.NewServiceLevelAgreement(&persist.OrmService{}, serviceName, serviceArea, true, callback)
	sla.SetServiceItem(serviceItem)
	sla.SetServiceItemList(serviceItemList)
	sla.SetPrimaryKeys(primaryKey)
	sla.SetTransactional(true)
	sla.SetReplication(true)
	sla.SetReplicationCount(3)
	sla.SetServiceGroup("L8SG")
	return sla
}

// ActivateService does what the SLA cannot describe on its own: it opens the
// service's database, attaches the Postgres plugin, registers the standard
// REST surface and activates the service.
func ActivateService(sla *ifs.ServiceLevelAgreement, creds, dbname string, vnic ifs.IVNic) {
	if len(sla.PrimaryKeys()) == 0 {
		panic(fmt.Sprintf("service %s (area %d): PrimaryKey is required", sla.ServiceName(), sla.ServiceArea()))
	}

	// Register the primary key with the INTROSPECTOR as well as the SLA.
	//
	// The ORM reads the primary key off the SLA, so persistence worked without
	// this. But everything that resolves a type's identity through the
	// introspector needs a Primary decorator, and on a backend node nothing
	// else registers one -- RegisterType/AddPrimaryKeyDecorator is only called
	// on the UI node. The visible consequence was that NewValidation could not
	// derive its auto-id setter, so POST never generated an id and every
	// service whose callback does Require("<Pk>Id") rejected every create with
	// "<Pk>Id is required". Only services that also called GenerateID
	// explicitly could be created at all; the mock generator never noticed
	// because it supplies its own ids.
	if intro := vnic.Resources().Introspector(); intro != nil {
		if item, ok := sla.ServiceItem().(proto.Message); ok {
			if err := intro.Decorators().AddPrimaryKeyDecorator(item, sla.PrimaryKeys()...); err != nil {
				panic(fmt.Sprintf("service %s (area %d): could not register primary key %v: %s",
					sla.ServiceName(), sla.ServiceArea(), sla.PrimaryKeys(), err.Error()))
			}
		}
	}

	_, user, pass, port, err := vnic.Resources().Security().Credential(creds, dbname, vnic.Resources())
	if err != nil {
		panic("Did not find credentials " + creds + " or db " + dbname + ":" + err.Error())
	}
	db := OpenDBConection(dbname, user, pass, port)
	sla.SetArgs(postgres.NewPostgres(db, vnic.Resources()), true)

	// The SLA holds these as interface{}; the REST surface needs the protos.
	serviceItem, ok := sla.ServiceItem().(proto.Message)
	if !ok {
		panic(fmt.Sprintf("service %s (area %d): service item is not a proto.Message", sla.ServiceName(), sla.ServiceArea()))
	}
	serviceItemList, ok := sla.ServiceItemList().(proto.Message)
	if !ok {
		panic(fmt.Sprintf("service %s (area %d): service item list is not a proto.Message", sla.ServiceName(), sla.ServiceArea()))
	}

	ws := web.New(sla.ServiceName(), sla.ServiceArea(), 0)
	ws.AddEndpoint(serviceItem, ifs.POST, &l8web.L8Empty{})
	ws.AddEndpoint(serviceItemList, ifs.POST, &l8web.L8Empty{})
	ws.AddEndpoint(serviceItem, ifs.PUT, &l8web.L8Empty{})
	ws.AddEndpoint(serviceItem, ifs.PATCH, &l8web.L8Empty{})
	ws.AddEndpoint(&l8api.L8Query{}, ifs.DELETE, &l8web.L8Empty{})
	ws.AddEndpoint(&l8api.L8Query{}, ifs.GET, serviceItemList)
	sla.SetWebService(ws)

	vnic.Resources().Services().Activate(sla, vnic)
}

// ServiceHandler returns the service handler for the given service.
func ServiceHandler(serviceName string, serviceArea byte, vnic ifs.IVNic) (ifs.IServiceHandler, bool) {
	return vnic.Resources().Services().ServiceHandler(serviceName, serviceArea)
}

// GetEntity retrieves a single entity by its filter, trying local first then remote.
// Returns the entity as interface{} — caller must type-assert.
func GetEntity(serviceName string, serviceArea byte, filter interface{}, vnic ifs.IVNic) (interface{}, error) {
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Get(object.New(nil, filter), vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		return resp.Element(), nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, filter, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return resp.Element(), nil
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
// Returns []interface{} — caller must type-assert each element.
func GetEntities(serviceName string, serviceArea byte, filter interface{}, vnic ifs.IVNic) ([]interface{}, error) {
	if isFilterEmpty(filter) {
		return getAllEntities(serviceName, serviceArea, filter, vnic)
	}
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Get(object.New(nil, filter), vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		return filterNilElements(resp.Elements()), nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, filter, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return filterNilElements(resp.Elements()), nil
}

// getAllEntities fetches all entities using an L8Query when the filter is empty.
func getAllEntities(serviceName string, serviceArea byte, filter interface{}, vnic ifs.IVNic) ([]interface{}, error) {
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
		return filterNilElements(resp.Elements()), nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, query, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return filterNilElements(resp.Elements()), nil
}

// filterNilElements removes nil entries from a slice returned by resp.Elements().
func filterNilElements(elems []interface{}) []interface{} {
	if elems == nil {
		return nil
	}
	result := make([]interface{}, 0, len(elems))
	for _, e := range elems {
		if e != nil {
			result = append(result, e)
		}
	}
	return result
}

// PutEntity updates an entity via its service handler.
func PutEntity(serviceName string, serviceArea byte, entity interface{}, vnic ifs.IVNic) error {
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
// Returns the created entity as interface{}.
func PostEntity(serviceName string, serviceArea byte, entity interface{}, vnic ifs.IVNic) (interface{}, error) {
	handler, ok := ServiceHandler(serviceName, serviceArea, vnic)
	if ok {
		resp := handler.Post(object.New(nil, entity), vnic)
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		if resp.Element() != nil {
			return resp.Element(), nil
		}
		return entity, nil
	}
	resp := vnic.Request("", serviceName, serviceArea, ifs.POST, entity, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	if resp.Element() != nil {
		return resp.Element(), nil
	}
	return entity, nil
}

// GetEntitiesByQuery retrieves multiple entities using an L8Query string.
// Use this when you need a WHERE clause (e.g., "select * from Alarm where State=1").
// For simple all-or-filter retrieval, use GetEntities instead.
// Returns []interface{} — caller must type-assert each element.
func GetEntitiesByQuery(serviceName string, serviceArea byte, query string, vnic ifs.IVNic) ([]interface{}, error) {
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
		return resp.Elements(), nil
	}
	q := &l8api.L8Query{Text: query}
	resp := vnic.Request("", serviceName, serviceArea, ifs.GET, q, 30)
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	return resp.Elements(), nil
}

// EntityExists checks if any entity matching the filter already exists.
func EntityExists(serviceName string, serviceArea byte, filter interface{}, vnic ifs.IVNic) (bool, error) {
	existing, err := GetEntities(serviceName, serviceArea, filter, vnic)
	if err != nil {
		return false, err
	}
	return len(existing) > 0, nil
}
