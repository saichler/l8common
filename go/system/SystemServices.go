package system

import (
	evtservices "github.com/saichler/l8events/go/services"
	notifyservices "github.com/saichler/l8notify/go/services"
	"github.com/saichler/l8types/go/ifs"
)

// Activate activates the required system services shared by every project:
// l8events (Events) and l8notify (Notify, IntegCfg).
// Lives outside the common package because l8events and l8notify import common.
func Activate(creds, dbname string, vnic ifs.IVNic) {
	evtservices.ActivateEvents(creds, dbname, vnic)
	notifyservices.ActivateNotify(creds, dbname, vnic)
	notifyservices.ActivateIntegrationConfig(creds, dbname, vnic)
}
