// © 2025 Sharon Aicler (saichler@gmail.com)
//
// Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generic

import (
	"errors"
	l8c "github.com/saichler/l8common/go/common"
)

// SafeCast safely casts an interface{} to a pointer of type T.
// Returns an error instead of panicking if the element is nil.
func SafeCast[T any](element interface{}) (*T, error) {
	if element == nil {
		return nil, errors.New("entity not found")
	}
	result, ok := element.(*T)
	if !ok {
		return nil, errors.New("unexpected response type")
	}
	return result, nil
}

// ValidateEnum validates an enum value against its name map.
// Generic wrapper — accepts any ~int32 enum type and casts to int32.
func ValidateEnum[T ~int32](value T, nameMap map[int32]string, enumName string) error {
	return l8c.ValidateEnum(int32(value), nameMap, enumName)
}

// Re-export validators from l8common.
var (
	ValidateRequired                 = l8c.ValidateRequired
	ValidateRequiredInt64            = l8c.ValidateRequiredInt64
	ValidateDateInPast               = l8c.ValidateDateInPast
	ValidateDateNotZero              = l8c.ValidateDateNotZero
	ValidateDateAfter                = l8c.ValidateDateAfter
	ValidateMinimumAge               = l8c.ValidateMinimumAge
	ValidateConditionalRequired      = l8c.ValidateConditionalRequired
	ValidateConditionalRequiredInt64 = l8c.ValidateConditionalRequiredInt64
	ValidateMoney                    = l8c.ValidateMoney
	ValidateMoneyPositive            = l8c.ValidateMoneyPositive
	ValidateDateRange                = l8c.ValidateDateRange
	GenerateID                       = l8c.GenerateID
)
