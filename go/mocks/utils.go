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
package mocks

import (
	"fmt"
	"math/rand"
	"time"

	l8common "github.com/saichler/l8common/go/types/l8common"
)

// PickRef safely picks a reference ID by modulo index, returns "" if slice empty.
func PickRef(ids []string, index int) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[index%len(ids)]
}

// RandomMoney generates a Money with random amount in [min, min+rangeSize) cents.
func RandomMoney(currencyIDs []string, min, rangeSize int) *l8common.Money {
	return &l8common.Money{Amount: int64(rand.Intn(rangeSize) + min), CurrencyId: PickRef(currencyIDs, rand.Intn(100))}
}

// ExactMoney creates a Money with exact amount in cents.
func ExactMoney(currencyIDs []string, amount int64) *l8common.Money {
	return &l8common.Money{Amount: amount, CurrencyId: PickRef(currencyIDs, 0)}
}

// RandomPastDate returns Unix timestamp randomly in the past.
func RandomPastDate(maxMonths, maxDays int) int64 {
	return time.Now().AddDate(0, -rand.Intn(maxMonths), -rand.Intn(maxDays)).Unix()
}

// RandomFutureDate returns Unix timestamp randomly in the future.
func RandomFutureDate(maxMonths, maxDays int) int64 {
	return time.Now().AddDate(0, rand.Intn(maxMonths), rand.Intn(maxDays)).Unix()
}

// GenID creates an ID like "prefix-001".
func GenID(prefix string, index int) string {
	return fmt.Sprintf("%s-%03d", prefix, index+1)
}

// GenCode creates a code like "PREFIX001".
func GenCode(prefix string, index int) string {
	return fmt.Sprintf("%s%03d", prefix, index+1)
}

// CreateAuditInfo creates a standard AuditInfo with current time.
func CreateAuditInfo() *l8common.AuditInfo {
	now := time.Now().Unix()
	return &l8common.AuditInfo{
		CreatedAt:  now,
		CreatedBy:  "mock-generator",
		ModifiedAt: now,
		ModifiedBy: "mock-generator",
	}
}

// CreateAddress creates a random work address.
func CreateAddress() *l8common.Address {
	return &l8common.Address{
		AddressType:   l8common.AddressType_ADDRESS_TYPE_WORK,
		Line1:         fmt.Sprintf("%d %s", rand.Intn(9999)+1, StreetNames[rand.Intn(len(StreetNames))]),
		City:          Cities[rand.Intn(len(Cities))],
		StateProvince: States[rand.Intn(len(States))],
		PostalCode:    fmt.Sprintf("%05d", rand.Intn(90000)+10000),
		CountryCode:   "US",
		IsPrimary:     true,
	}
}

// CreateContact creates a random work phone contact.
func CreateContact() *l8common.ContactInfo {
	return &l8common.ContactInfo{
		ContactType: l8common.ContactType_CONTACT_TYPE_PHONE_WORK,
		Value:       RandomPhone(),
		IsPrimary:   true,
	}
}

// RandomName generates a random full name.
func RandomName() string {
	return fmt.Sprintf("%s %s", FirstNames[rand.Intn(len(FirstNames))], LastNames[rand.Intn(len(LastNames))])
}

// RandomPhone generates a random US phone number.
func RandomPhone() string {
	return fmt.Sprintf("(%03d) %03d-%04d", rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(9000)+1000)
}

// RandomSSN generates a random SSN string.
func RandomSSN() string {
	return fmt.Sprintf("%03d-%02d-%04d", rand.Intn(900)+100, rand.Intn(90)+10, rand.Intn(9000)+1000)
}

// RandomBirthDate generates a random birth date between 25-60 years ago.
func RandomBirthDate() int64 {
	yearsAgo := rand.Intn(35) + 25
	return time.Now().AddDate(-yearsAgo, -rand.Intn(12), -rand.Intn(28)).Unix()
}

// RandomHireDate generates a random hire date in the last 10 years.
func RandomHireDate() int64 {
	yearsAgo := rand.Intn(10)
	return time.Now().AddDate(-yearsAgo, -rand.Intn(12), -rand.Intn(28)).Unix()
}

// SanitizeEmail strips non-alphabetic characters from a string.
func SanitizeEmail(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			result += string(c)
		}
	}
	return result
}

// GetIssuingOrg returns the issuing organization for a certification name.
func GetIssuingOrg(certName string) string {
	orgs := map[string]string{
		"PMP":       "Project Management Institute",
		"AWS":       "Amazon Web Services",
		"Google":    "Google Cloud",
		"Scrum":     "Scrum Alliance",
		"SHRM":      "Society for Human Resource Management",
		"CPA":       "Certified Public Accountant",
		"Six":       "ASQ",
		"CISSP":     "ISC2",
		"CompTIA":   "CompTIA",
		"Microsoft": "Microsoft",
	}
	for key, org := range orgs {
		if len(certName) >= len(key) && certName[:len(key)] == key {
			return org
		}
	}
	return "Professional Certification Body"
}

// GenLines generates N child items per parent, calling create(idx, parentIdx, childIdx, parentID).
func GenLines[L any](parentIDs []string, n int, create func(idx, pIdx, j int, parentID string) *L) []*L {
	lines := make([]*L, 0, len(parentIDs)*n)
	idx := 1
	for pIdx, parentID := range parentIDs {
		for j := 0; j < n; j++ {
			lines = append(lines, create(idx, pIdx, j, parentID))
			idx++
		}
	}
	return lines
}

// MinInt returns the smaller of two ints.
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
