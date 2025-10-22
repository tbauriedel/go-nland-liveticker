package model

import (
	"fmt"
	"strings"
	"time"
)

type Operation struct {
	Time     time.Time
	Units    string
	Report   string
	District string
	Location string
}

// getIdentifier generates a unique identifier for the operation based on its key attributes.
func (o *Operation) GetIdentifier() string {
	t := o.Time.Format("2006-01-02 15:04")

	identifier := fmt.Sprintf("%s;%s;%s;%s",
		strings.ToLower(strings.Replace(o.Units, " ", "", -1)),
		strings.Replace(t, " ", ";", -1),
		strings.ToLower(strings.Replace(o.Report, " ", "", -1)),
		strings.ToLower(o.Location),
	)

	return strings.Replace(identifier, "\n", "-", -1)
}
