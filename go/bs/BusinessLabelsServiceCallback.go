package bs

import (
	l8c "github.com/saichler/l8common/go/common"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8business"
)

func newBusinessLabelsCallback() ifs.IServiceCallback {
	return l8c.NewServiceCallback(
		"L8BusinessLabel",
		func(e interface{}) bool { _, ok := e.(*l8business.L8BusinessLabel); return ok },
		setBusinessLabelID,
		validateBusinessLabel,
	)
}

func setBusinessLabelID(e interface{}) {
	entity := e.(*l8business.L8BusinessLabel)
	l8c.GenerateID(&entity.LabelId)
}

func validateBusinessLabel(e interface{}, vnic ifs.IVNic) error {
	entity := e.(*l8business.L8BusinessLabel)
	return l8c.ValidateRequired(entity.LabelDescription, "LabelDescription")
}
