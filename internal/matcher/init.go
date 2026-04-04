package matcher

func init() {
	RegistryInstance.Register("ac", func(cfg any) (Matcher, error) {
		return NewAC(cfg.(*ACConfig))
	})
	RegistryInstance.Register("regex", func(cfg any) (Matcher, error) {
		return NewRegex(cfg.(*RegexConfig))
	})
	RegistryInstance.Register("external", func(cfg any) (Matcher, error) {
		return NewExternal(cfg.(*ExternalConfig))
	})
}
