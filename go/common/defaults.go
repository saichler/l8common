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
	"database/sql"
	"fmt"
	"github.com/saichler/l8reflect/go/reflect/introspecting"
	"github.com/saichler/l8services/go/services/manager"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/sec"
	"github.com/saichler/l8types/go/types/l8sysconfig"
	"github.com/saichler/l8utils/go/utils/logger"
	"github.com/saichler/l8utils/go/utils/registry"
	"github.com/saichler/l8utils/go/utils/resources"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var dbInstance *sql.DB
var dbMtx = &sync.Mutex{}

// CreateResources creates the standard Layer 8 resources with configurable parameters.
func CreateResources(alias string, logDirectory string, vnetPort uint32) ifs.IResources {
	if logDirectory != "" {
		logger.SetLogToFile(logDirectory, alias)
	}
	log := logger.NewLoggerImpl(&logger.FmtLogMethod{})
	log.SetLogLevel(ifs.Info_Level)
	res := resources.NewResources(log)

	res.Set(registry.NewRegistry())

	sec, _ := sec.LoadSecurityProvider(res)
	res.Set(sec)

	conf := &l8sysconfig.L8SysConfig{MaxDataSize: resources.DEFAULT_MAX_DATA_SIZE,
		RxQueueSize:              resources.DEFAULT_QUEUE_SIZE,
		TxQueueSize:              resources.DEFAULT_QUEUE_SIZE,
		LocalAlias:               alias,
		VnetPort:                 vnetPort,
		LogsDirectory:            logDirectory,
		KeepAliveIntervalSeconds: 30}
	res.Set(conf)

	res.Set(introspecting.NewIntrospect(res.Registry()))
	res.Set(manager.NewServices(res))

	return res
}

// WaitForSignal blocks until the process receives SIGINT or SIGTERM.
func WaitForSignal(resources ifs.IResources) {
	resources.Logger().Info("Waiting for os signal...")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	resources.Logger().Info("End signal received! ", sig)
}

// OpenDBConection establishes a connection to the PostgreSQL database.
// It uses localhost (127.0.0.1) on port 5432 with SSL disabled.
// Panics if the connection cannot be established or ping fails.
func OpenDBConection(dbname, user, pass string) *sql.DB {
	dbMtx.Lock()
	defer dbMtx.Unlock()
	if dbInstance != nil {
		return dbInstance
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"127.0.0.1", 5432, user, pass, dbname)
	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	dbInstance = db
	return dbInstance
}
