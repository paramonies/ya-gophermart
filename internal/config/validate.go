package config

func (cfg *AppConfig) validate() error {
	if cfg.RunAddress == "" {
		return ErrMissingOption{
			Option: "run_address",
		}
	}

	return nil
}
