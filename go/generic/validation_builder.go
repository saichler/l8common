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
package generic

import (
	l8c "github.com/saichler/l8common/go/common"
	l8common "github.com/saichler/l8common/go/types/l8common"
	"github.com/saichler/l8types/go/ifs"
	l8api "github.com/saichler/l8types/go/types/l8api"
)

// VB (Validation Builder) chains validators for a ServiceCallback.
// Generic wrapper — delegates to l8common's non-generic VB with type-safe adapters.
type VB[T any] struct {
	inner *l8c.VB
}

// NewValidation creates a validation builder for a ServiceCallback.
func NewValidation[T any](typeName string, setID SetIDFunc[T]) *VB[T] {
	return &VB[T]{inner: l8c.NewValidation(typeName,
		func(v interface{}) bool { _, ok := v.(*T); return ok },
		func(v interface{}) { setID(v.(*T)) })}
}

// Require adds a required string field validation.
func (b *VB[T]) Require(getter func(*T) string, name string) *VB[T] {
	b.inner.Require(func(v interface{}) string { return getter(v.(*T)) }, name)
	return b
}

// RequireInt64 adds a required int64 field validation.
func (b *VB[T]) RequireInt64(getter func(*T) int64, name string) *VB[T] {
	b.inner.RequireInt64(func(v interface{}) int64 { return getter(v.(*T)) }, name)
	return b
}

// Enum adds an enum field validation using the protobuf _name map.
func (b *VB[T]) Enum(getter func(*T) int32, nameMap map[int32]string, name string) *VB[T] {
	b.inner.Enum(func(v interface{}) int32 { return getter(v.(*T)) }, nameMap, name)
	return b
}

// Money adds a required money field validation (nil + CurrencyId check).
func (b *VB[T]) Money(getter func(*T) *l8common.Money, name string) *VB[T] {
	b.inner.Money(func(v interface{}) *l8common.Money { return getter(v.(*T)) }, name)
	return b
}

// MoneyPositive adds a required money field validation with positive amount.
func (b *VB[T]) MoneyPositive(getter func(*T) *l8common.Money, name string) *VB[T] {
	b.inner.MoneyPositive(func(v interface{}) *l8common.Money { return getter(v.(*T)) }, name)
	return b
}

// OptionalMoney validates a money field only when non-nil (skips nil).
func (b *VB[T]) OptionalMoney(getter func(*T) *l8common.Money, name string) *VB[T] {
	b.inner.OptionalMoney(func(v interface{}) *l8common.Money { return getter(v.(*T)) }, name)
	return b
}

// DateNotZero adds a required date (non-zero timestamp) validation.
func (b *VB[T]) DateNotZero(getter func(*T) int64, name string) *VB[T] {
	b.inner.DateNotZero(func(v interface{}) int64 { return getter(v.(*T)) }, name)
	return b
}

// DateAfter validates that date1 > date2 (skips if either is zero).
func (b *VB[T]) DateAfter(getter1, getter2 func(*T) int64, name1, name2 string) *VB[T] {
	b.inner.DateAfter(
		func(v interface{}) int64 { return getter1(v.(*T)) },
		func(v interface{}) int64 { return getter2(v.(*T)) },
		name1, name2)
	return b
}

// DateRange validates a required *l8common.DateRange field (nil + StartDate < EndDate).
func (b *VB[T]) DateRange(getter func(*T) *l8common.DateRange, name string) *VB[T] {
	b.inner.DateRange(func(v interface{}) *l8common.DateRange { return getter(v.(*T)) }, name)
	return b
}

// Compute adds a function that derives/computes entity fields before validation.
func (b *VB[T]) Compute(fn func(*T) error) *VB[T] {
	b.inner.Compute(func(v interface{}) error { return fn(v.(*T)) })
	return b
}

// Custom adds a custom validation function.
func (b *VB[T]) Custom(fn func(*T, ifs.IVNic) error) *VB[T] {
	b.inner.Custom(func(v interface{}, vnic ifs.IVNic) error { return fn(v.(*T), vnic) })
	return b
}

// StatusTransition adds a status state-machine validator.
func (b *VB[T]) StatusTransition(cfg *StatusTransitionConfig[T]) *VB[T] {
	innerCfg := &l8c.StatusTransitionConfig{
		StatusGetter:  func(v interface{}) int32 { return cfg.StatusGetter(v.(*T)) },
		StatusSetter:  func(v interface{}, s int32) { cfg.StatusSetter(v.(*T), s) },
		FilterBuilder: func(v interface{}) interface{} { return cfg.FilterBuilder(v.(*T)) },
		ServiceName:   cfg.ServiceName,
		ServiceArea:   cfg.ServiceArea,
		InitialStatus: cfg.InitialStatus,
		Transitions:   cfg.Transitions,
		StatusNames:   cfg.StatusNames,
	}
	b.inner.StatusTransition(innerCfg)
	return b
}

// After adds a function to run after successful persistence (PUT/PATCH only).
func (b *VB[T]) After(fn ActionValidateFunc[T]) *VB[T] {
	b.inner.After(func(v interface{}, action ifs.Action, vnic ifs.IVNic) error {
		return fn(v.(*T), action, vnic)
	})
	return b
}

// Re-export ValidatePeriod from l8common.
var ValidatePeriod = l8c.ValidatePeriod

// Period adds a required L8Period field validation.
func (b *VB[T]) Period(getter func(*T) *l8api.L8Period, name string) *VB[T] {
	b.inner.Period(func(v interface{}) *l8api.L8Period { return getter(v.(*T)) }, name)
	return b
}

// OptionalPeriod validates an L8Period field only when non-nil.
func (b *VB[T]) OptionalPeriod(getter func(*T) *l8api.L8Period, name string) *VB[T] {
	b.inner.OptionalPeriod(func(v interface{}) *l8api.L8Period { return getter(v.(*T)) }, name)
	return b
}

// Build creates the IServiceCallback from the chained validators.
func (b *VB[T]) Build() ifs.IServiceCallback {
	return b.inner.Build()
}
