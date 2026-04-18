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

// ServiceConfig holds the configuration for activating a service.
type ServiceConfig struct {
	ServiceName  string
	ServiceArea  byte
	PrimaryKey   string
	Callback     ifs.IServiceCallback
	ServiceGroup string
}

// ActivateService sets up and activates a service with the standard boilerplate.
// serviceItem and serviceItemList must be proto.Message instances (e.g., &MyType{}, &MyTypeList{}).
func ActivateService(cfg ServiceConfig, serviceItem proto.Message, serviceItemList proto.Message, creds, dbname string, vnic ifs.IVNic) {
	if cfg.PrimaryKey == "" {
		panic(fmt.Sprintf("service %s (area %d): PrimaryKey is required", cfg.ServiceName, cfg.ServiceArea))
	}
	_, user, pass, _, err := vnic.Resources().Security().Credential(creds, dbname, vnic.Resources())
	if err != nil {
		panic("Did not find credentials " + creds + " or db " + dbname + ":" + err.Error())
	}
	fmt.Println("Test ", creds, " ", dbname, " ", user, " ", pass)
	db := OpenDBConection(dbname, user, pass)
	p := postgres.NewPostgres(db, vnic.Resources())

	sla := ifs.NewServiceLevelAgreement(&persist.OrmService{}, cfg.ServiceName, cfg.ServiceArea, true, cfg.Callback)
	sla.SetServiceItem(serviceItem)
	sla.SetServiceItemList(serviceItemList)
	sla.SetPrimaryKeys(cfg.PrimaryKey)
	sla.SetArgs(p, true)
	sla.SetTransactional(true)
	sla.SetReplication(true)
	sla.SetReplicationCount(3)

	ws := web.New(cfg.ServiceName, cfg.ServiceArea, 0)
	ws.AddEndpoint(serviceItem, ifs.POST, &l8web.L8Empty{})
	ws.AddEndpoint(serviceItemList, ifs.POST, &l8web.L8Empty{})
	ws.AddEndpoint(serviceItem, ifs.PUT, &l8web.L8Empty{})
	ws.AddEndpoint(serviceItem, ifs.PATCH, &l8web.L8Empty{})
	ws.AddEndpoint(&l8api.L8Query{}, ifs.DELETE, &l8web.L8Empty{})
	ws.AddEndpoint(&l8api.L8Query{}, ifs.GET, serviceItemList)
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
