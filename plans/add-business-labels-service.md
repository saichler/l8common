# Plan: Add BusinessLabels Service (L8BusinessLabel)

## Summary

Add a new `BusinessLabels` service in `go/bs/` with:
- **PrimeObject**: `L8BusinessLabel` (already defined in `l8types/proto/business.proto`)
- **ServiceName**: `BLabels` (7 chars, within 10-char limit)
- **ServiceArea**: `97`
- **ORM-based** via `common.ActivateService`

## Existing Proto Type

The `L8BusinessLabel` type already exists in `l8types/proto/business.proto` (package `l8business`, Go package `l8types/go/types/l8business`):

```protobuf
message L8BusinessLabel {
  string label_id = 1;
  string label_description = 2;
  AuditInfo audit_info = 3;
}

message L8BusinessLabelList {
  repeated L8BusinessLabel list = 1;
  l8api.L8MetaData metadata = 2;
}
```

The type follows the Layer 8 pattern:
- List type uses `repeated ... list = 1` + `l8api.L8MetaData metadata = 2`
- Primary key: `LabelId`
- Includes `AuditInfo` for creation/modification tracking
- `L8BusinessLabels` (map wrapper) matches the `IBusinessLabels` interface in `l8types/go/ifs/`

## Prerequisites

The `l8business` package is not currently vendored into l8common. The user must re-vendor before the service code can compile.

## Deliverables

### 1. Service File

**File**: `go/bs/BusinessLabelsService.go`

```go
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
```

### 2. Service Callback File

**File**: `go/bs/BusinessLabelsServiceCallback.go`

Uses `NewServiceCallback` with manual `setID` (the VB auto-derived `setID` would be a no-op here because the type isn't registered with the introspector at callback creation time):

```go
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
```

This manually defines `typeCheck`, `setID` (auto-generates `LabelId` on POST), and validates `LabelDescription` is required.

## Traceability Matrix

| # | Item | Phase |
|---|------|-------|
| 1 | Create `go/bs/BusinessLabelsService.go` | Phase 1 |
| 2 | Create `go/bs/BusinessLabelsServiceCallback.go` | Phase 1 |
| 3 | Verify: `go build ./...` passes | Phase 2 |

## Phase 1: Service Implementation

1. Create `go/bs/` directory
2. Create `go/bs/BusinessLabelsService.go` — constants, `Activate`, convenience getters (`Labels`, `Label`)
3. Create `go/bs/BusinessLabelsServiceCallback.go` — manual callback with `setID` and `LabelDescription` validation

## Phase 2: Verification

1. `go build ./...` — verify compilation
2. Verify `ServiceName` is 7 chars (within 10-char limit)
3. Verify `ServiceArea` is `97`
