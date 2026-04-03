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
)

// SumLineMoney sums a Money field across a slice of line items.
// Generic wrapper — converts typed slice to []interface{} and delegates to l8common.
func SumLineMoney[L any](lines []*L, getter func(*L) *l8common.Money) *l8common.Money {
	ifaces := make([]interface{}, len(lines))
	for i, l := range lines {
		ifaces[i] = l
	}
	return l8c.SumLineMoney(ifaces, func(v interface{}) *l8common.Money { return getter(v.(*L)) })
}

// SumLineFloat64 sums a float64 field across a slice of line items.
func SumLineFloat64[L any](lines []*L, getter func(*L) float64) float64 {
	ifaces := make([]interface{}, len(lines))
	for i, l := range lines {
		ifaces[i] = l
	}
	return l8c.SumLineFloat64(ifaces, func(v interface{}) float64 { return getter(v.(*L)) })
}

// SumLineInt64 sums an int64 field across a slice of line items.
func SumLineInt64[L any](lines []*L, getter func(*L) int64) int64 {
	ifaces := make([]interface{}, len(lines))
	for i, l := range lines {
		ifaces[i] = l
	}
	return l8c.SumLineInt64(ifaces, func(v interface{}) int64 { return getter(v.(*L)) })
}

// Re-export non-generic money helpers from l8common.
var (
	MoneyAdd      = l8c.MoneyAdd
	MoneySubtract = l8c.MoneySubtract
	MoneyAmount   = l8c.MoneyAmount
	MoneyIsZero   = l8c.MoneyIsZero
)
