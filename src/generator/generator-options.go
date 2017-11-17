package generator

type Options struct {
	validFrom string
	validFor  string
	password  string
	uid		  string
	did       string
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

func (o *Options) SetUid(value string) {
	o.uid = value
}

func (o Options) Uid() string {
	return o.uid
}

func (o *Options) SetDid(value string) {
	o.did = value
}

func (o Options) Did() string {
	return o.did
}
