package actions

import (
	"fmt"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/processors"
)

func configChecked(
	constr processors.Constructor,
	checks ...func(common.Config) error,
) processors.Constructor {
	validator := checkAll(checks...)
	return func(c common.Config) (processors.Processor, error) {
		err := validator(c)
		if err != nil {
			return nil, fmt.Errorf("%v in %v", err.Error(), c.Path())
		}

		return constr(c)
	}
}

func checkAll(checks ...func(common.Config) error) func(common.Config) error {
	return func(c common.Config) error {
		for _, check := range checks {
			if err := check(c); err != nil {
				return err
			}
		}
		return nil
	}
}

func requireFields(fields ...string) func(common.Config) error {
	return func(c common.Config) error {
		for _, field := range fields {
			if !c.HasField(field) {
				return fmt.Errorf("missing %v option", field)
			}
		}
		return nil
	}
}

func allowedFields(fields ...string) func(common.Config) error {
	return func(c common.Config) error {
		for _, field := range c.GetFields() {
			found := false
			for _, allowed := range fields {
				if field == allowed {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("unexpected %v option", field)
			}
		}
		return nil
	}
}
