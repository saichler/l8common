package bs

import (
	l8c "github.com/saichler/l8common/go/common"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8business"
)

const (
	ServiceName = "BLabels"
	ServiceArea = byte(97)
)

func Activate(creds, dbname string, vnic ifs.IVNic) {
	l8c.ActivateService(l8c.ServiceConfig{
		ServiceName: ServiceName, ServiceArea: ServiceArea,
		PrimaryKey: "LabelId", Callback: newBusinessLabelsCallback(),
	}, &l8business.L8BusinessLabel{}, &l8business.L8BusinessLabelList{}, creds, dbname, vnic)
}

func Labels(vnic ifs.IVNic) (ifs.IServiceHandler, bool) {
	return l8c.ServiceHandler(ServiceName, ServiceArea, vnic)
}

func Label(labelId string, vnic ifs.IVNic) (*l8business.L8BusinessLabel, error) {
	result, err := l8c.GetEntity(ServiceName, ServiceArea, &l8business.L8BusinessLabel{LabelId: labelId}, vnic)
	if err != nil {
		return nil, err
	}
	return result.(*l8business.L8BusinessLabel), nil
}
