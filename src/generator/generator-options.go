package generator

type Options struct {
	validFrom string
	validFor  string
	password  string
}

func (o *Options) SetValidFrom(value string) {
	o.validFrom = value
}

func (o Options) ValidFrom() string {
	return o.validFrom
}

func (o *Options) SetValidFor(value string) {
	o.validFor = value
}

func (o Options) ValidFor() string {
	return o.validFor
}

func (o *Options) SetPassword(value string) {
	o.password = value
}

func (o Options) Password() string {
	return o.password
}
