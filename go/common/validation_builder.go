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
type VB[T any] struct {
	typeName         string
	setID            SetIDFunc[T]
	validators       []func(*T, ifs.IVNic) error
	actionValidators []ActionValidateFunc[T]
	afterActions     []ActionValidateFunc[T]
}

// NewValidation creates a validation builder for a ServiceCallback.
func NewValidation[T any](typeName string, setID SetIDFunc[T]) *VB[T] {
	return &VB[T]{typeName: typeName, setID: setID}
}

// Require adds a required string field validation.
func (b *VB[T]) Require(getter func(*T) string, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateRequired(getter(e), name)
	})
	return b
}

// RequireInt64 adds a required int64 field validation.
func (b *VB[T]) RequireInt64(getter func(*T) int64, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateRequiredInt64(getter(e), name)
	})
	return b
}

// Enum adds an enum field validation using the protobuf _name map.
// Value 0 (UNSPECIFIED) is rejected; unknown values are rejected.
func (b *VB[T]) Enum(getter func(*T) int32, nameMap map[int32]string, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateEnum(getter(e), nameMap, name)
	})
	return b
}

// Money adds a required money field validation (nil + CurrencyId check).
func (b *VB[T]) Money(getter func(*T) *l8common.Money, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateMoney(getter(e), name)
	})
	return b
}

// MoneyPositive adds a required money field validation with positive amount.
func (b *VB[T]) MoneyPositive(getter func(*T) *l8common.Money, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateMoneyPositive(getter(e), name)
	})
	return b
}

// OptionalMoney validates a money field only when non-nil (skips nil).
func (b *VB[T]) OptionalMoney(getter func(*T) *l8common.Money, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		m := getter(e)
		if m == nil {
			return nil
		}
		return ValidateMoney(m, name)
	})
	return b
}

// DateNotZero adds a required date (non-zero timestamp) validation.
func (b *VB[T]) DateNotZero(getter func(*T) int64, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateDateNotZero(getter(e), name)
	})
	return b
}

// DateAfter validates that date1 > date2 (skips if either is zero).
func (b *VB[T]) DateAfter(getter1, getter2 func(*T) int64, name1, name2 string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		d1, d2 := getter1(e), getter2(e)
		if d1 == 0 || d2 == 0 {
			return nil
		}
		return ValidateDateAfter(d1, d2, name1, name2)
	})
	return b
}

// DateRange validates a required *l8common.DateRange field (nil + StartDate < EndDate).
func (b *VB[T]) DateRange(getter func(*T) *l8common.DateRange, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidateDateRange(getter(e), name)
	})
	return b
}

// Compute adds a function that derives/computes entity fields before validation.
// Chain Compute() before Require() so computed fields can be validated.
func (b *VB[T]) Compute(fn func(*T) error) *VB[T] {
	return b.Custom(func(e *T, _ ifs.IVNic) error {
		return fn(e)
	})
}

// Custom adds a custom validation function.
func (b *VB[T]) Custom(fn func(*T, ifs.IVNic) error) *VB[T] {
	b.validators = append(b.validators, fn)
	return b
}

// StatusTransition adds a status state-machine validator.
func (b *VB[T]) StatusTransition(cfg *StatusTransitionConfig[T]) *VB[T] {
	b.actionValidators = append(b.actionValidators, cfg.BuildValidator())
	return b
}

// After adds a function to run after successful persistence (PUT/PATCH only).
func (b *VB[T]) After(fn ActionValidateFunc[T]) *VB[T] {
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
func (b *VB[T]) Period(getter func(*T) *l8api.L8Period, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		return ValidatePeriod(getter(e), name)
	})
	return b
}

// OptionalPeriod validates an L8Period field only when non-nil.
func (b *VB[T]) OptionalPeriod(getter func(*T) *l8api.L8Period, name string) *VB[T] {
	b.validators = append(b.validators, func(e *T, _ ifs.IVNic) error {
		p := getter(e)
		if p == nil {
			return nil
		}
		return ValidatePeriod(p, name)
	})
	return b
}

// Build creates the IServiceCallback from the chained validators.
func (b *VB[T]) Build() ifs.IServiceCallback {
	validate := func(item *T, vnic ifs.IVNic) error {
		for _, v := range b.validators {
			if err := v(item, vnic); err != nil {
				return err
			}
		}
		return nil
	}
	if len(b.afterActions) > 0 {
		return NewServiceCallbackWithAfter(b.typeName, b.setID, validate,
			b.actionValidators, b.afterActions)
	}
	return NewServiceCallback(b.typeName, b.setID, validate, b.actionValidators...)
}
