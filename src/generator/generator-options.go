package generator

type Options struct {
	host       string
	validFrom  string
	validFor   string
	rsaBits    int
	ecdsaCurve string
	defaultTTL int
}

func (o *Options) SetHost(value string) {
	o.host = value
}

func (o Options) Host() string {
	return o.host
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

func (o *Options) SetRsaBits(value int) {
	o.rsaBits = value
}

func (o Options) RsaBits() int {
	return o.rsaBits
}

func (o *Options) SetEcdsaCurve(value string) {
	o.ecdsaCurve = value
}

func (o Options) EcdsaCurve() string {
	return o.ecdsaCurve
}

func (o *Options) SetDefaultTTL(value int) {
	o.defaultTTL = value
}

func (o Options) DefaultTTL() int {
	return o.defaultTTL
}
