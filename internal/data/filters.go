package data

import (
	"strings"

	"github.com/pharsha1995/greenlight/internal/data/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f *Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f *Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(validator.WithinRange(f.Page, 1, 10_000_000), "page", "must be between 1 and 10 million")
	v.Check(validator.WithinRange(f.PageSize, 1, 100), "page_size", "must be between 1 and 100")
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
