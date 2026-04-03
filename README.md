# l8common

Shared utilities, protobuf types, and service infrastructure for the [Layer 8 Ecosystem](https://github.com/saichler).

## Overview

`l8common` provides the common building blocks used across all Layer 8 projects. It eliminates boilerplate by offering reusable service activation, validation, CRUD operations, mock data generation utilities, and shared protobuf types.

## Project Structure

```
l8common/
├── proto/
│   ├── l8common.proto          # Shared protobuf types (Money, AuditInfo, Address, etc.)
│   └── make-bindings.sh        # Protobuf code generation script
└── go/
    ├── common/                  # Core utilities
    │   ├── defaults.go          # Resource creation, DB connection, signal handling
    │   ├── service_factory.go   # Service activation and CRUD operations
    │   ├── service_callback.go  # Generic service callback implementation
    │   ├── validation.go        # Reference validation helpers
    │   ├── validation_static.go # Static field validators (required, enum, date, money)
    │   ├── validation_builder.go# Fluent validation builder (VB)
    │   ├── status_machine.go    # Status transition state machine
    │   ├── type_registry.go     # Type and primary key registration
    │   ├── compute.go           # Money arithmetic and line-item aggregation
    │   └── currency.go          # Currency conversion helper
    ├── mocks/                   # Mock data generation utilities
    │   ├── client.go            # HTTP client for mock data upload
    │   ├── data_common.go       # Shared name/address data arrays
    │   └── utils.go             # ID generation, random data, audit/address/contact creators
    └── types/l8common/
        └── l8common.pb.go       # Generated protobuf Go bindings
```

## Shared Protobuf Types

Defined in `proto/l8common.proto`, these types are used across all Layer 8 projects:

| Type | Description |
|------|-------------|
| `Money` | Monetary amount with currency ID (amount in smallest unit, e.g., cents) |
| `AuditInfo` | Creation/modification timestamps and user tracking |
| `Address` | Physical/mailing address with type (Home, Work, Billing, etc.) |
| `ContactInfo` | Contact method with type (Phone, Email, Fax) |
| `DateRange` | Start and end date pair (Unix timestamps) |

## Core Utilities (`go/common`)

### Resource & Infrastructure Setup

- **`CreateResources(alias, logDir, vnetPort)`** — Creates standard Layer 8 resources (logger, registry, security provider, introspector, service manager)
- **`WaitForSignal(resources)`** — Blocks until SIGINT/SIGTERM
- **`OpenDBConection(dbname, user, pass)`** — Singleton PostgreSQL connection with connection pooling

### Service Activation

- **`ActivateService(cfg, item, itemList, creds, dbname, vnic)`** — One-call service setup: database connection, ORM, web endpoints (GET/POST/PUT/PATCH/DELETE), replication, and transactions

### CRUD Operations

- **`GetEntity(serviceName, area, filter, vnic)`** — Retrieve a single entity (local-first, remote fallback)
- **`GetEntities(serviceName, area, filter, vnic)`** — Retrieve multiple entities; auto-uses L8Query for empty filters
- **`PostEntity(serviceName, area, entity, vnic)`** — Create a new entity
- **`PutEntity(serviceName, area, entity, vnic)`** — Update an entity
- **`EntityExists(serviceName, area, filter, vnic)`** — Check if an entity matching a filter exists

### Validation

**Static validators:**
- `ValidateRequired`, `ValidateRequiredInt64` — Non-empty/non-zero checks
- `ValidateEnum` — Enum value against protobuf name map (rejects 0/UNSPECIFIED)
- `ValidateMoney`, `ValidateMoneyPositive` — Money nil/currency/amount checks
- `ValidateDateNotZero`, `ValidateDateInPast`, `ValidateDateAfter`, `ValidateDateRange` — Date checks
- `ValidateMinimumAge`, `ValidateConditionalRequired` — Domain-specific checks
- `GenerateID` — Auto-generate UUID for empty primary key fields

**Validation Builder (`VB`):**

Fluent API for building service callbacks with chained validators:

```go
callback := common.NewValidation(&mypackage.MyEntity{}, vnic).
    Require(func(v interface{}) string { return v.(*mypackage.MyEntity).Name }, "Name").
    Enum(func(v interface{}) int32 { return int32(v.(*mypackage.MyEntity).Status) }, StatusNameMap, "Status").
    Money(func(v interface{}) *l8common.Money { return v.(*mypackage.MyEntity).Amount }, "Amount").
    DateRange(func(v interface{}) *l8common.DateRange { return v.(*mypackage.MyEntity).Period }, "Period").
    StatusTransition(statusConfig).
    Build()
```

### Status State Machine

`StatusTransitionConfig` enforces allowed status transitions on PUT/PATCH, with automatic initial status assignment on POST.

### Money Utilities

- `MoneyAdd`, `MoneySubtract` — Currency-aware arithmetic
- `SumLineMoney`, `SumLineFloat64`, `SumLineInt64` — Line-item aggregation
- `MoneyAmount`, `MoneyIsZero` — Safe nil-checking accessors
- `ConvertMoney` — Exchange rate conversion

### Type Registration

- **`RegisterType(resources, typeInstance, listInstance, pkField)`** — Registers a protobuf type with its primary key decorator and list wrapper

## Mock Data Utilities (`go/mocks`)

Shared helpers for generating realistic test data across Layer 8 projects:

- **`MockClient`** — HTTP client with authentication for uploading mock data to running services
- **`PickRef`, `GenID`, `GenCode`** — Safe reference picking and ID/code generation
- **`RandomMoney`, `ExactMoney`** — Money value generators
- **`RandomPastDate`, `RandomFutureDate`, `RandomBirthDate`, `RandomHireDate`** — Date generators
- **`CreateAuditInfo`, `CreateAddress`, `CreateContact`** — Standard entity field generators
- **`RandomName`, `RandomPhone`, `RandomSSN`** — Personal data generators
- **`GenLines`** — Bulk child record generator for parent-child relationships
- Curated data arrays: `FirstNames`, `LastNames`, `StreetNames`, `Cities`, `States`

## Dependencies

Built on the Layer 8 framework libraries:
- `l8types` — Interfaces and shared API types
- `l8services` — Service management
- `l8orm` — ORM and PostgreSQL persistence
- `l8reflect` — Introspection and type metadata
- `l8srlz` — Serialization
- `l8utils` — Logger, registry, maps, queues
- `l8bus` — Overlay network health and protocol
- `l8ql` — Query language (L8Query)

## License

Apache License, Version 2.0 — see [LICENSE](LICENSE) for details.

Copyright 2025 Sharon Aicler (saichler@gmail.com)
