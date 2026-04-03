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
	l8common "github.com/saichler/l8common/go/types/l8common"
	"github.com/saichler/l8types/go/ifs"
	l8api "github.com/saichler/l8types/go/types/l8api"
)

// VB (Validation Builder) chains validators for a ServiceCallback.
// Use NewValidation to start building, chain .Require() calls, then .Build().
// All getter functions accept interface{} — the caller must type-assert inside.
type VB struct {
	typeName         string
	typeCheck        func(interface{}) bool
	setID            SetIDFunc
	validators       []func(interface{}, ifs.IVNic) error
	actionValidators []ActionValidateFunc
	afterActions     []ActionValidateFunc
}

// NewValidation creates a validation builder for a ServiceCallback.
// typeName is a human-readable name for error messages.
// typeCheck verifies the entity type (e.g., func(v interface{}) bool { _, ok := v.(*MyType); return ok }).
// setID generates/sets the primary key on a new entity.
func NewValidation(typeName string, typeCheck func(interface{}) bool, setID SetIDFunc) *VB {
	return &VB{typeName: typeName, typeCheck: typeCheck, setID: setID}
}

// Require adds a required string field validation.
func (b *VB) Require(getter func(interface{}) string, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateRequired(getter(e), name)
	})
	return b
}

// RequireInt64 adds a required int64 field validation.
func (b *VB) RequireInt64(getter func(interface{}) int64, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateRequiredInt64(getter(e), name)
	})
	return b
}

// Enum adds an enum field validation using the protobuf _name map.
// Value 0 (UNSPECIFIED) is rejected; unknown values are rejected.
func (b *VB) Enum(getter func(interface{}) int32, nameMap map[int32]string, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateEnum(getter(e), nameMap, name)
	})
	return b
}

// Money adds a required money field validation (nil + CurrencyId check).
func (b *VB) Money(getter func(interface{}) *l8common.Money, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateMoney(getter(e), name)
	})
	return b
}

// MoneyPositive adds a required money field validation with positive amount.
func (b *VB) MoneyPositive(getter func(interface{}) *l8common.Money, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateMoneyPositive(getter(e), name)
	})
	return b
}

// OptionalMoney validates a money field only when non-nil (skips nil).
func (b *VB) OptionalMoney(getter func(interface{}) *l8common.Money, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		m := getter(e)
		if m == nil {
			return nil
		}
		return ValidateMoney(m, name)
	})
	return b
}

// DateNotZero adds a required date (non-zero timestamp) validation.
func (b *VB) DateNotZero(getter func(interface{}) int64, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateDateNotZero(getter(e), name)
	})
	return b
}

// DateAfter validates that date1 > date2 (skips if either is zero).
func (b *VB) DateAfter(getter1, getter2 func(interface{}) int64, name1, name2 string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		d1, d2 := getter1(e), getter2(e)
		if d1 == 0 || d2 == 0 {
			return nil
		}
		return ValidateDateAfter(d1, d2, name1, name2)
	})
	return b
}

// DateRange validates a required *l8common.DateRange field (nil + StartDate < EndDate).
func (b *VB) DateRange(getter func(interface{}) *l8common.DateRange, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidateDateRange(getter(e), name)
	})
	return b
}

// Compute adds a function that derives/computes entity fields before validation.
// Chain Compute() before Require() so computed fields can be validated.
func (b *VB) Compute(fn func(interface{}) error) *VB {
	return b.Custom(func(e interface{}, _ ifs.IVNic) error {
		return fn(e)
	})
}

// Custom adds a custom validation function.
func (b *VB) Custom(fn func(interface{}, ifs.IVNic) error) *VB {
	b.validators = append(b.validators, fn)
	return b
}

// StatusTransition adds a status state-machine validator.
func (b *VB) StatusTransition(cfg *StatusTransitionConfig) *VB {
	b.actionValidators = append(b.actionValidators, cfg.BuildValidator())
	return b
}

// After adds a function to run after successful persistence (PUT/PATCH only).
func (b *VB) After(fn ActionValidateFunc) *VB {
	b.afterActions = append(b.afterActions, fn)
	return b
}

// ValidatePeriod validates an L8Period field.
func ValidatePeriod(p *l8api.L8Period, name string) error {
	if p == nil {
		return fmt.Errorf("%s is required", name)
	}
	if p.PeriodType == l8api.L8PeriodType_invalid_period_type {
		return fmt.Errorf("%s type is required", name)
	}
	if p.PeriodYear < 1970 || p.PeriodYear > 2100 {
		return fmt.Errorf("%s year must be between 1970 and 2100", name)
	}
	switch p.PeriodType {
	case l8api.L8PeriodType_Quarterly:
		if p.PeriodValue < l8api.L8PeriodValue_Q1 || p.PeriodValue > l8api.L8PeriodValue_Q4 {
			return fmt.Errorf("%s quarterly value must be Q1-Q4", name)
		}
	case l8api.L8PeriodType_Monthly:
		if p.PeriodValue < l8api.L8PeriodValue_January || p.PeriodValue > l8api.L8PeriodValue_December {
			return fmt.Errorf("%s monthly value must be January-December", name)
		}
	}
	return nil
}

// Period adds a required L8Period field validation.
func (b *VB) Period(getter func(interface{}) *l8api.L8Period, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		return ValidatePeriod(getter(e), name)
	})
	return b
}

// OptionalPeriod validates an L8Period field only when non-nil.
func (b *VB) OptionalPeriod(getter func(interface{}) *l8api.L8Period, name string) *VB {
	b.validators = append(b.validators, func(e interface{}, _ ifs.IVNic) error {
		p := getter(e)
		if p == nil {
			return nil
		}
		return ValidatePeriod(p, name)
	})
	return b
}

// Build creates the IServiceCallback from the chained validators.
func (b *VB) Build() ifs.IServiceCallback {
	validate := func(item interface{}, vnic ifs.IVNic) error {
		for _, v := range b.validators {
			if err := v(item, vnic); err != nil {
				return err
			}
		}
		return nil
	}
	if len(b.afterActions) > 0 {
		return NewServiceCallbackWithAfter(b.typeName, b.typeCheck, b.setID, validate,
			b.actionValidators, b.afterActions)
	}
	return NewServiceCallback(b.typeName, b.typeCheck, b.setID, validate, b.actionValidators...)
}
